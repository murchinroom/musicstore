package audiofilestore

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"musicstore/model"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// this file implements a api controller handling the upload of new tracks.

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
func (a *AudioFileStore) PostNewTrack(c *gin.Context) {
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

	// save file
	savedpath, err := a.saveFile(c, req)
	if err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	// add track to lib
	track, err := a.AddTrack(savedpath, OverrideTrackMetadata(&req.Track))
	if err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"track": track})
}

// success returns true
func checkUploadRequest(c *gin.Context, req *PostNewTrackRequest) error {
	if req.File == nil && req.AudioFileURL == "" {
		return errors.New("neither File nor AudioFileURL is provided")
	}

	if req.File != nil && req.AudioFileURL != "" {
		return errors.New("both File and AudioFileURL are provided, but only one is allowed")
	}

	return nil
}

// saveFile saves the file to the disk.
func (a *AudioFileStore) saveFile(c *gin.Context, req *PostNewTrackRequest) (savedpath string, err error) {
	switch {
	case req.File != nil:
		return a.saveFileFromMultipart(c, req)
	case req.AudioFileURL != "":
		return a.saveFileFromURL(c, req)
	default:
		return "", errors.New("neither File nor AudioFileURL is provided")
	}
}

func (a *AudioFileStore) tmpDir() string {
	tmp := filepath.Join(a.FileDir, ".tmp")

	// create tmp dir if not exists
	if _, err := os.Stat(tmp); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(tmp, 0755)
		if err != nil {
			panic(fmt.Errorf("AudioFileStore.tmpDir(): Mkdir failed: %w", err))
		}
	}

	return tmp
}

// saveFileFromMultipart saves the file from the multipart request.
func (a *AudioFileStore) saveFileFromMultipart(c *gin.Context, req *PostNewTrackRequest) (savedpath string, err error) {
	file := req.File

	filename := filepath.Base(file.Filename)
	filename = guardFilename(filename)
	dst := filepath.Join(a.tmpDir(), filename)

	err = c.SaveUploadedFile(file, dst)
	return dst, err
}

// saveFileFromURL saves the file from the given URL.
func (a *AudioFileStore) saveFileFromURL(c *gin.Context, req *PostNewTrackRequest) (savedpath string, err error) {
	// download file
	resp, err := http.Get(req.AudioFileURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// get filename from URL
	tokens := strings.Split(req.AudioFileURL, "/")
	filename := tokens[len(tokens)-1]
	filename = guardFilename(filename)

	// save file
	dst := filepath.Join(a.tmpDir(), filename)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return dst, err
}

// guardFilename guards the filename:
//   - If it is empty, generate a random filename.
//   - If it has no extension, add ".mp3" to the end.
func guardFilename(filename string) string {
	if filename == "" {
		// random filename
		filename = fmt.Sprintf("%d.mp3", time.Now().UnixNano())
	}
	if filepath.Ext(filename) == "" {
		filename += ".mp3"
	}
	return filename
}
