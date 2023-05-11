package main

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/service"
	"github.com/gin-gonic/gin"
)

// this file implement a controller for uploading & storing music files.

type UploadTrackRequest struct {
	Track
	File *multipart.FileHeader
}

// UploadTrackResponse when uploading successful:
type UploadTrackResponse struct {
	Track Track
}

// UploadTrack handles: POST /tracks/upload
//
// Body: multipart/form-data
//   - File: the music file
//   - Name: the name of the track
//   - Artist: the artists of the track
//   - Album: the albums of the track
//
// The metadata of the track will be saved to the database,
// and the music file will be saved to the disk.
func UploadTrack(c *gin.Context) {
	// bind file: https://github.com/gin-gonic/examples/blob/master/file-binding/main.go
	req := new(UploadTrackRequest)
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

	// if no file provided: just save the metadata
	if req.File == nil {
		c.JSON(200, UploadTrackResponse{Track: req.Track})
		return
	}
	// else: metadata saved, now save the file
	if err := saveFile(c, req); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	// update the AudioFileURL field
	if err := updateAudioFileURL(c, &req.Track); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	// TODO: uncomment this
	// analyze the emotion
	// if err := analyzeAndUpdateEmotion(c, &req.Track); err != nil {
	// 	c.JSON(422, gin.H{"error": err.Error()})
	// 	return
	// }

	c.JSON(200, UploadTrackResponse{Track: req.Track})
}

// success returns true
func checkUploadRequest(c *gin.Context, req *UploadTrackRequest) error {
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

func trackExists(ctx context.Context, track *Track) bool {
	cnt, err := service.Count[Track](ctx,
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

func saveMetadata(c *gin.Context, req *UploadTrackRequest) error {
	err := service.Create(c, &req.Track, service.IfNotExist())
	return err
}

func saveFile(c *gin.Context, req *UploadTrackRequest) error {
	file := req.File
	dst := TrackFilepath(req.Track.ID)
	err := c.SaveUploadedFile(file, dst)
	return err
}

// requires: track.ID != 0
func updateAudioFileURL(ctx context.Context, track *Track) error {
	track.AudioFileURL = AudioFileURL(track.ID)
	_, err := service.Update(ctx, track)
	return err
}

// TODO: implement me
func analyzeAndUpdateEmotion(ctx context.Context, track *Track) error {
	return errors.New("not implemented")
}

// TrackFilepath returns the filepath of the music file of the track:
//
//	{MUSICSTORE_FILEPATH}/audio/{trackID}.mp3
//
// where {MUSICSTORE_FILEPATH} is an environment variable,
// and defaults to the current directory (./).
func TrackFilepath(trackID uint) string {
	return filepath.Join(AudioFileDir(), AudioFileName(trackID))
}

const AudioDirname = "audio"
const AudioStaticServePath = "/" + AudioDirname

// AudioFileDir returns the directory path of the music files:
//
//	{MUSICSTORE_FILEPATH}/audio
//
// where {MUSICSTORE_FILEPATH} is an environment variable,
// and defaults to the current directory (./).
func AudioFileDir() string {
	base, ok := os.LookupEnv("MUSICSTORE_FILEPATH")
	if !ok {
		base = "."
	}
	return filepath.Join(base, AudioDirname)
}

// AudioFileName returns "{trackID}.mp3".
//
// It's used to construct the AudioFileURL field of the Track model.
// And it's also used to construct the filepath of the music file of the track.
func AudioFileName(trackID uint) string {
	return fmt.Sprintf("%d.mp3", trackID)
}

func AudioFileURL(trackID uint) string {
	return AudioStaticServePath + "/" + AudioFileName(trackID)
}
