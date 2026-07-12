package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"rudy_gc/internal/config"
	"rudy_gc/internal/model/modelg"
	"rudy_gc/pkg/orm"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	flag.Parse()

	c := config.MustLoad(*configFile)
	dsn := strings.TrimSpace(c.DataSource)
	if dsn == "" {
		log.Fatal("数据库迁移失败：配置中的 DataSource 为空")
	}

	db, err := orm.NewGormDBEngine(&orm.Config{
		DSN:      dsn,
		LogLevel: strings.TrimSpace(c.LogursLevel),
	})
	if err != nil {
		log.Fatalf("数据库迁移失败：连接数据库失败：%v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("数据库迁移失败：获取数据库连接失败：%v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败：%v", err)
		}
	}()

	log.Printf("开始执行数据库迁移，配置文件：%s", *configFile)
	if err := modelg.AutoMigrate(db); err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：%w", err))
	}
	log.Println("数据库迁移完成")
}
