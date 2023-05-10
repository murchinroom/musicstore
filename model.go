package main

import "github.com/cdfmlr/crud/orm"

// copy from murecom-chorus-1/unistructs/models.go
// with TrackText & TrackEmotion removed
// and AudioFileURL added

type Track struct {
	orm.BasicModel
	Name          string
	Artist       string
	Album         string
	CoverImageURL string
	AudioFileURL  string

	// emmm, 就当作文档型数据库吧
}
