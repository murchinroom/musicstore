package audiofilestore

import "github.com/gin-gonic/gin"

func (a *AudioFileStore) registerRoutes(r gin.IRouter) {
	group := r.Group(a.Name)

	// static audio file
	group.Static("/audio", a.FileDir)  // a.audioStaticBasePath

	// add track
	group.POST("/new", a.PostNewTrack)
}
