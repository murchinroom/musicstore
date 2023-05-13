package main

import (
	"github.com/cdfmlr/crud/orm"
	"musicstore/model"
)

func main() {
	orm.ConnectDB(orm.DBDriverSqlite, "musicstore.db")
	orm.RegisterModel(&model.Track{})

	r := MakeRouter()
	r.Run(":8086")
}
