// Package metadata provides CRUDs for tracks metadata.
// And provides murecom api.
package metadata

import (
	"musicstore/model"

	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/gin-gonic/gin"

	"github.com/glebarez/sqlite" // pure go sqlite driver: supports math functions
	"gorm.io/gorm"
)

var logger = log.ZoneLogger("musicstore/metadata")

// Start the metadata module.
//
// There should be only one metadata module in a program.
// The metadata module should be run before audiofilestore modules.
func Start(dbDSN string, router gin.IRouter) {
	// orm.ConnectDB(orm.DBDriverSqlite, "musicstore.db")
	connectDB(dbDSN)

	orm.RegisterModel(&model.Track{})

	registerRoutes(router)
}

// TODO: crud should support custom driver
func connectDB(dsn string) error {
	var err error
	orm.DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: log.Logger4Gorm,
	})
	return err
}
