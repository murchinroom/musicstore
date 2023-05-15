package model

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

// this file implements a Track contributor that
// read track metadata from a audio file.
//
// This function only fills the Name, Artist and Album fields of the Track.
// The CoverImageURL and AudioFileURL fields are left blank.
func TrackFromAudioFile(path string) (*Track, error) {
	// open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// read metadata
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	// construct track
	track := &Track{
		Name:   m.Title(),
		Artist: m.Artist(),
		Album:  m.Album(),
		// CoverImageURL: "",
		// AudioFileURL: "",
	}

	if track.Name == "" {
		track.Name = strings.TrimSuffix(
			filepath.Base(path), filepath.Ext(path))
	}

	return track, nil
}
