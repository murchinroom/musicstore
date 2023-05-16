// Package audiofilestore stores audio files in a local directory.
// Exposure an AudioFileStore with the following methods:
//   - AddTrack: add a track (from audio file path) to the audiofilestore
//   - AddTracksFromDir: read self.FileDir and add all the tracks in it
//
// Exposure Routes:
//   - /audio: static audio file
//   - /new: add track (upload file or download from url)
package audiofilestore

import (
	"context"
	"errors"
	"fmt"
	"musicstore/emomusic"
	"musicstore/metadata"
	"musicstore/model"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cdfmlr/crud/log"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger = log.ZoneLogger("musicstore/audiofilestore")

func init() {
	// logger.Logger.SetLevel(logrus.DebugLevel)
	logger.Logger.SetLevel(logrus.InfoLevel)
}

// AudioFileStore stores audio files in a local directory.
type AudioFileStore struct {
	Name           string
	FileDir        string
	BaseUrl        string
	EnableEmomusic bool
}

func NewAudioFileStore(name, fileDir, baseUrl string, enableEmomusic bool, router gin.IRouter) *AudioFileStore {
	a := &AudioFileStore{
		Name:           name,
		FileDir:        fileDir,
		BaseUrl:        baseUrl,
		EnableEmomusic: enableEmomusic,
	}

	a.registerRoutes(router)

	return a
}

// this file implement an app that turns a local directory into a music store.
// that is: build a database from the music files in the directory,
// and generate config files for the music store.
// After that, the musicstore can be run with the generated config files & database.

// AddTrack adds a track (from audio file path) to the database.
//
// File will be hard linked to the FileDir. And named as:
//
//	{name_of_the_track}-{name_of_the_track}-{name_of_the_track}.mp3
func (a *AudioFileStore) AddTrack(path string, options ...AddTrackOption) (*model.Track, error) {
	// get track metadata
	track, err := model.TrackFromAudioFile(path)
	if err != nil {
		return nil, fmt.Errorf("AudioFileToTrack: TrackFromAudioFile failed: %w", err)
	}

	// apply options
	for _, opt := range options {
		opt(a, track)
	}

	// check if track exists
	if metadata.TrackExists(context.Background(), track) {
		return nil, fmt.Errorf("AudioFileToTrack: track already exists: %s", track.Name)
	}

	// Save audio file to FileDir: hard link it
	oldpath := path
	path, err = a.hardLinkAudioFile(track, path)
	if err != nil {
		return nil, fmt.Errorf("AudioFileToTrack: hardLinkAudioFile failed: %w", err)
	}

	// fill url
	track.AudioFileURL, err = a.audioUrl(path)
	if err != nil {
		os.Remove(path) // rollback

		return nil, fmt.Errorf("AudioFileToTrack: AudioFileURL failed: %w", err)
	}
	// TODO: Image??

	// emotion analyze
	if a.EnableEmomusic {
		emotion, err := emomusic.AnalyzeURI(track.AudioFileURL)
		// emotion, err := emomusic.AnalyzeFile(path)
		if err != nil {
			os.Remove(path) // rollback

			return nil, fmt.Errorf("AudioFileToTrack: AnalyzeURI failed: %w", err)
		}
		track.Emotion = emotion
	}

	// save to db
	err = metadata.CreateTrack(context.Background(), track)
	if err != nil {
		os.Remove(path) // rollback

		return nil, fmt.Errorf("AudioFileToTrack: Create failed: %w", err)
	}

	logger.WithField("ID", track.ID).
		WithField("Name", track.Name).
		WithField("AudioFileURL", track.AudioFileURL).
		Info("AddTrack: success")

	oldpathAbs, err1 := filepath.Abs(oldpath)
	aFileDirAbs, err2 := filepath.Abs(a.FileDir)
	if err1 == nil && err2 == nil &&
		strings.HasPrefix(oldpathAbs, aFileDirAbs) {
		// 原来就在 FileDir 下，rm 原文件，相当于只是重命名
		os.Remove(oldpath)
	}

	return track, nil
}

// AddTrackOption is the option type for AddTrack.
// Options are applied in order after the track is constructed from the audio file
// and before the track is saved (both metadata & audio file).
type AddTrackOption func(*AudioFileStore, *model.Track)

// OverrideTrackMetadata overrides the track metadata with the given track.
// Fields that are not empty will be used to override the track metadata.
//
// The AudioFileURL field of the given track will be ignored.
func OverrideTrackMetadata(track *model.Track) AddTrackOption {
	return func(a *AudioFileStore, t *model.Track) {
		if track == nil {
			return
		}
		if track.Name != "" {
			t.Name = track.Name
		}
		if track.Artist != "" {
			t.Artist = track.Artist
		}
		if track.Album != "" {
			t.Album = track.Album
		}
		if track.CoverImageURL != "" {
			t.CoverImageURL = track.CoverImageURL
		}
	}
}

// hardLinkAudioFile hard link the audio file to the FileDir.
// It returns the new path.
//
// The new path is constructed as:
//
//	{FileDir}/{name_of_the_track}-{name_of_the_track}-{name_of_the_track}.mp3
//
// If the file already exists, it returns an error.
func (a *AudioFileStore) hardLinkAudioFile(track *model.Track, path string) (newpath string, err error) {
	filename := fmt.Sprintf("%s-%s-%s%s",
		stringToSnake(track.Name), stringToSnake(track.Artist), stringToSnake(track.Album),
		filepath.Ext(path)) // Ext includes the dot

	newpath = filepath.Join(a.FileDir, filename)

	// check if the file exists
	if _, err := os.Stat(newpath); !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("hardLinkAudioFile: file already exists: %s. err=%w", newpath, err)
	}

	// hard link
	err = os.Link(path, newpath)
	if err != nil {
		return "", fmt.Errorf("hardLinkAudioFile: Link failed: %w", err)
	}

	return newpath, nil
}

// stringToSnake converts "a string with spaces" to "a_string_with_spaces".
func stringToSnake(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

// audioRelevantPath = Abs(path) - Abs(FileDir)
func (a *AudioFileStore) audioRelevantPath(path string) (string, error) {
	fileAbsPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	dirAbsPath, err := filepath.Abs(a.FileDir)
	if err != nil {
		return "", err
	}

	relevant := strings.TrimPrefix(fileAbsPath, dirAbsPath)

	return relevant, nil
}

func (a *AudioFileStore) audioStaticBasePath() string {
	return "/" + a.Name + "/audio"
}

// audioUrl = BaseUrl + audioStaticBasePath + audioRelevantPath
func (a *AudioFileStore) audioUrl(path string) (string, error) {
	relevant, err := a.audioRelevantPath(path)
	if err != nil {
		return "", err
	}

	u, err := url.JoinPath(a.BaseUrl, a.audioStaticBasePath(), relevant)
	return u, err
}

// AddTracksFromDir adds all the tracks in the directory to the database.
func (a *AudioFileStore) AddTracksFromDir() error {
	logger.WithField("FileDir", a.FileDir).Info("AddTracksFromDir: start")

	// enumerate music files
	ch, err := enumMusicFiles(a.FileDir)
	if err != nil {
		return fmt.Errorf("AddTracksFromDir: enumMusicFiles failed: %w", err)
	}

	// add tracks
	for path := range ch {
		logger.WithField("path", path).Debug("AddTracksFromDir: AddTrack")
		_, err := a.AddTrack(path)
		if err != nil {
			logger.Errorf("AddTracksFromDir: AddTrack failed: %v", err)
		}
	}

	return nil
}

// isMusicFile returns true if the file is a music file.
// It checks the file extension.
// supported extensions: .mp3, .wav, .m4a
func isMusicFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3", ".wav", ".m4a":
		return true
	default:
		return false
	}
}

// enumMusicFiles enumerates all the music files in the directory.
// It returns a channel of the file paths.
func enumMusicFiles(dir string) (chan string, error) {
	if dir == "" {
		return nil, errors.New("empty dir")
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, errors.New("not a dir")
	}

	ch := make(chan string, 3)

	go func() {
		defer close(ch)

		// walk the directory
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// skip non-music files
			if d.IsDir() || !isMusicFile(path) {
				return nil
			}

			// send the path to the channel
			ch <- path

			return nil
		})

		if err != nil {
			fmt.Println("[ERR] filepath.Walk:", err)
		}
	}()

	return ch, nil
}
