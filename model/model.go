package model

import "github.com/cdfmlr/crud/orm"

// copy from murecom-chorus-1/unistructs/models.go
// with TrackText & TrackEmotion removed
// and AudioFileURL added

type Track struct {
	orm.BasicModel

	Name          string
	Artist        string
	Album         string
	CoverImageURL string
	AudioFileURL  string

	Emotion Emotion `gorm:"embedded"`

	// emmm, 就当作文档型数据库吧
}

type Emotion struct {
	Valence float64 `json:"valence"`
	Arousal float64 `json:"arousal"`
}
