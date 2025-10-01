// cmd/db_migrate
package main

import (
	"rudy_gc/data/modelg"
	"rudy_gc/pkg/orm"
)

func main() {

	cfg := orm.Config{
		DSN:      "root:4521822123@tcp(127.0.0.1:3306)/rudy_gc?charset=utf8mb4",
		LogLevel: "info",
	}

	dbGorm := orm.MustNewGormDBEngine(&cfg)
	modelg.MustAutoMigrate(dbGorm)

}
