package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// This file implements an API client for the emomusic API.

// emomusicServerURL returns the emomusic server address.
func emomusicServerURL() string {
	s := os.Getenv("EMOMUSIC_SERVER")
	if s == "" {
		s = "http://localhost:8000/"
	}

	return s
}

// emomusicPredictmp3URL returns the emomusic API address.
func emomusicPredictmp3URL() string {
	r, err := url.JoinPath(emomusicServerURL(), "predictmp3")
	if err != nil {
		panic(err)
	}
	return r
}

func AnalyzeEmotion(mp3Filepath string) (Emotion, error) {
	// build body
	form, err := predictmp3RequestForm(mp3Filepath)
	if err != nil {
		return Emotion{}, err
	}

	// build request
	client := &http.Client{}
	req, err := predictmp3Request(form)
	if err != nil {
		return Emotion{}, err
	}

	// send request
	resp, err := client.Do(req)
	if err != nil {
		return Emotion{}, err
	}
	defer resp.Body.Close()

	// check response
	if resp.StatusCode != http.StatusOK {
		return Emotion{}, errors.New("emomusic API error")
	}

	// parse response
	var emotion Emotion
	err = json.NewDecoder(resp.Body).Decode(&emotion)
	if err != nil {
		return Emotion{}, err
	}

	return emotion, nil
}

// -F "file=@{mp3Filepath}"
func predictmp3RequestForm(mp3Filepath string) (*bytes.Buffer, error) {
	form := new(bytes.Buffer)

	writer := multipart.NewWriter(form)

	fw, err := writer.CreateFormFile("file.name", mp3Filepath)
	if err != nil {
		return nil, err
	}

	fd, err := os.Open(mp3Filepath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	_, err = io.Copy(fw, fd)
	if err != nil {
		return nil, err
	}

	writer.Close()

	return form, nil
}

// POST {emomusicPredictmp3URL()}/predictmp3 with {form}
func predictmp3Request(form *bytes.Buffer) (*http.Request, error) {
	req, err := http.NewRequest("POST", emomusicPredictmp3URL(), form)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "multipart/form-data")

	return req, nil
}
