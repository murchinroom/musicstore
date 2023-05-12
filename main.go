package main

import (
	"github.com/cdfmlr/crud/orm"
)

func main() {
	orm.ConnectDB(orm.DBDriverSqlite, "musicstore.db")
	orm.RegisterModel(&Track{})

	r := MakeRouter()
	r.Run(":8086")
}
