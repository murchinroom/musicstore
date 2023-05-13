package main

import (
	"context"
	"errors"
	"mime/multipart"
	"musicstore/emomusic"
	"musicstore/model"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
)

const EnvEmomusicOnPostNew = "EMOMUSIC_ON_POST_NEW"

// EmomusicOnPostNew == true: do emotion analyze for new track from POST /new.
//
// set by env EMOMUSIC_ON_POST_NEW
var EmomusicOnPostNew bool = true

func init() {
	if e, ok := os.LookupEnv("EMOMUSIC_ON_POST_NEW"); ok {
		if b, err := strconv.ParseBool(e); err == nil {
			EmomusicOnPostNew = b
		}
	}
}

// this file implement a controller for uploading & storing music files.

type PostNewTrackRequest struct {
	model.Track
	File *multipart.FileHeader
}

// PostNewTrackResponse when uploading successful:
type PostNewTrackResponse struct {
	Track model.Track
}

// PostNewTrack handles: POST /new
//
// Body: multipart/form-data
//
//   - Name: the name of the track
//   - Artist: the artists of the track
//   - Album: the albums of the track
//   - CoverImageURL: track'scover image
//
// and one of:
//
//   - File: curl -F 'File=@audio.mp3'
//   - AudioFileURL: curl -F 'AudioFileURL=https://example.com/audio.mp3'
//
// The metadata of the track will be saved to the database,
// and the music file will be saved to the disk.
func PostNewTrack(c *gin.Context) {
	// bind file: https://github.com/gin-gonic/examples/blob/master/file-binding/main.go
	req := new(PostNewTrackRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := checkUploadRequest(c, req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// save the metadata
	if err := saveMetadata(c, req); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	// assert the track ID is not 0, which is required by following codes
	if req.Track.ID == 0 {
		panic("track ID is 0. this should not happen.")
	}

	// metadata saved, now save the file (if any)
	if req.File != nil {
		if err := saveFile(c, req); err != nil {
			_, _ = service.Delete(c, &req.Track) // rollback
			c.JSON(422, gin.H{"error": err.Error()})
			return
		}

		// update the AudioFileURL field
		if err := updateAudioFileURL(c, &req.Track); err != nil {
			_, _ = service.Delete(c, &req.Track) // rollback
			c.JSON(422, gin.H{"error": err.Error()})
			return
		}
	}

	// analyze the emotion
	if EmomusicOnPostNew {
		if err := analyzeAndUpdateEmotion(c, &req.Track); err != nil {
			_ = os.Remove(req.Track.AudioFilePath())
			_, _ = service.Delete(c, &req.Track)

			c.JSON(422, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(200, PostNewTrackResponse{Track: req.Track})
}

// success returns true
func checkUploadRequest(c *gin.Context, req *PostNewTrackRequest) error {
	if req.File == nil && req.AudioFileURL == "" {
		return errors.New("neither File nor AudioFileURL is provided")
	}

	if req.Name == "" && req.File != nil {
		req.Name = strings.TrimSuffix(filepath.Base(req.File.Filename), filepath.Ext(req.Name))
	}
	if req.Name == "" {
		return errors.New("track name is empty")
	}

	// check if the track already exists
	if trackExists(c, &req.Track) {
		return errors.New("track already exists")
	}

	return nil
}

func trackExists(ctx context.Context, track *model.Track) bool {
	cnt, err := service.Count[model.Track](ctx,
		service.FilterBy("name", track.Name),
		service.FilterBy("artist", track.Artist))

	if err != nil {
		log.Logger.WithContext(ctx).
			WithField("name", track.Name).
			WithField("artist", track.Artist).
			WithError(err).
			Error("trackExists: failed to select tracks")
		return false
	}

	return cnt > 0
}

func saveMetadata(c *gin.Context, req *PostNewTrackRequest) error {
	err := service.Create(c, &req.Track, service.IfNotExist())
	return err
}

func saveFile(c *gin.Context, req *PostNewTrackRequest) error {
	file := req.File
	dst := req.Track.AudioFilePath()
	err := c.SaveUploadedFile(file, dst)
	return err
}

// requires: track.ID != 0
func updateAudioFileURL(ctx context.Context, track *model.Track) error {
	track.AudioFileURL = track.AudioFileUrlRelevant()
	_, err := service.Update(ctx, track)
	return err
}

// requires: track.AudioFileURL != ""
func analyzeAndUpdateEmotion(ctx context.Context, track *model.Track) error {
	mp3URL := track.AudioFileURL
	if mp3URL == "" {
		return errors.New("empty track.AudioFileURL")
	}
	if strings.HasPrefix(mp3URL, "/") {
		mp3URL = track.AudioFileUrlAbsolute()
	}

	// call emomusic

	emotion, err := emomusic.AnalyzeURI(mp3URL)
	if err != nil {
		return err
	}

	// update db

	track.Emotion = emotion
	_, err = service.Update(ctx, track)

	return err
}
