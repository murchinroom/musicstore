package main

import (
	"gopkg.in/yaml.v3"
	"io"
)

type MusicstoreConfig struct {
	HttpListenAddr  string
	Metadata        MetadataConfig
	AudioFileStores []AudioFileStoreConfig
	Emomusic        EmomusicConfig
}

func (c *MusicstoreConfig) Write(dst io.Writer) error {
	return yaml.NewEncoder(dst).Encode(&c)
}

type MetadataConfig struct {
	DB string
}

type AudioFileStoreConfig struct {
	Name           string
	FileDir        string
	BaseUrl        string
	EnableEmomusic bool
	LoadFromDir    bool
}

type EmomusicConfig struct {
	Server string
}
