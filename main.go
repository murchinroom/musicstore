package main

import (
	"musicstore/model"

	"github.com/cdfmlr/crud/log"
	"github.com/cdfmlr/crud/orm"
	"github.com/glebarez/sqlite" // pure go sqlite driver: supports math functions
	"gorm.io/gorm"
)

func main() {
	// orm.ConnectDB(orm.DBDriverSqlite, "musicstore.db")
	connectDB("musicstore.db")

	orm.RegisterModel(&model.Track{})

	r := MakeRouter()
	r.Run(":8086")
}

// TODO: crud should support custom driver
func connectDB(dsn string) error {
	var err error
	orm.DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: log.Logger4Gorm,
	})
	return err
}
