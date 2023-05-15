package metadata

import (
	"musicstore/model"
	"musicstore/murecom"

	"github.com/cdfmlr/crud/router"

	"github.com/gin-gonic/gin"
)

func registerRoutes(r gin.IRouter) {
	// basic CRUDs
	router.Crud[model.Track](r, "/tracks")

	// murecom
	r.GET("/murecom", murecom.GetMurecom)
}
