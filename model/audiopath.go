package model

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// this file defines local (fs) & remote (http) path for tracks' audio files

const (
	EnvMusicstoreFilePath = "MUSICSTORE_FILEPATH"
	EnvMusicstoreBaseUrl  = "MUSICSTORE_BASEURL"
)

const (
	AudioDirname         = "audio"
	AudioStaticServePath = "/" + AudioDirname
)

// -------- base dir & base url --------

// AudioFileDir returns the directory path of the music files:
//
//	{MUSICSTORE_FILEPATH}/audio
//
// where {MUSICSTORE_FILEPATH} is an environment variable,
// and defaults to the current directory (./).
func AudioFileDir() string {
	base, ok := os.LookupEnv(EnvMusicstoreFilePath)
	if !ok {
		base = "."
	}
	return filepath.Join(base, AudioDirname)
}

// AudioBaseURL returns the base url of the music resources.
//
//	http://{MUSICSTORE_BASEURL}/audio
//
// if env {MUSICSTORE_FILEPATH} not found or eq to ""
// return relevant path: /audio
func AudioBaseURL() string {
	base := os.Getenv(EnvMusicstoreBaseUrl)
	u, err := url.JoinPath(base, AudioDirname)
	if err != nil {
		return AudioStaticServePath
	}
	return u
}

// -------- file name --------

// AudioFileName returns "{trackID}.mp3".
//
// It's used to construct the AudioFileURL field of the Track model.
// And it's also used to construct the filepath of the music file of the track.
func AudioFileName(trackID uint) string {
	return fmt.Sprintf("%d.mp3", trackID)
}

func (t *Track) AudioFileName() string {
	return AudioFileName(t.ID)
}

// -------- base dir|url + file name --------

// AudioFilePath returns the filepath of the music file of the track:
//
//	{MUSICSTORE_FILEPATH}/audio/{trackID}.mp3
//
// where {MUSICSTORE_FILEPATH} is an environment variable,
// and defaults to the current directory (./).
func AudioFilePath(trackID uint) string {
	return filepath.Join(AudioFileDir(), AudioFileName(trackID))
}

// AudioFilePath returns the filepath of the music file of the track:
//
//	{MUSICSTORE_FILEPATH}/audio/{trackID}.mp3
//
// where {MUSICSTORE_FILEPATH} is an environment variable,
// and defaults to the current directory (./).
func (t *Track) AudioFilePath() string {
	return AudioFilePath(t.ID)
}

// AudioFileUrlAbsolute returns the url of the music file of the track:
//
//	{MUSICSTORE_BASEURL}/audio/{trackID}.mp3
//
// where {MUSICSTORE_BASEURL} is an environment variable,
// and defaults to "".
func AudioFileUrlAbsolute(trackID uint) string {
	u, err := url.JoinPath(AudioBaseURL(), AudioFileName(trackID))
	if err != nil {
		// never happen
		u = filepath.Join(AudioBaseURL(), AudioFileName(trackID))
	}
	return u
}

// AudioFileUrlAbsolute returns the url of the music file of the track:
//
//	{MUSICSTORE_BASEURL}/audio/{trackID}.mp3
//
// where {MUSICSTORE_BASEURL} is an environment variable,
// and defaults to "".
func (t *Track) AudioFileUrlAbsolute() string {
	return AudioFileUrlAbsolute(t.ID)
}

// AudioFileUrlRelevant returns the url of the music file of tthe track:
//
//	/audio/{trackID}.mp3
func AudioFileUrlRelevant(trackID uint) string {
	return AudioStaticServePath + "/" + AudioFileName(trackID)
}

// AudioFileUrlRelevant returns the url of the music file of tthe track:
//
//	/audio/{trackID}.mp3
func (t *Track) AudioFileUrlRelevant() string {
	return AudioFileUrlRelevant(t.ID)
}
