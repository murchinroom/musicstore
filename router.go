package main

import (
	"github.com/cdfmlr/crud/router"
	"github.com/gin-gonic/gin"
)

func MakeRouter() *gin.Engine {
	r := router.NewRouter()

	// basic CRUDs
	router.Crud[Track](r, "/tracks")

	// upload track: new metadata & file
	r.POST("/new", UploadTrack)

	// static audio file
	r.Static(AudioStaticServePath, AudioFileDir())

	return r
}
