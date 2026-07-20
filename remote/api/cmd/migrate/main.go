package main

import (
	"flag"
	"log"
	"strings"

	dailybestmigration "rudy_gc/remote/api/internal/migration/dailybest"
	"rudy_gc/remote/api/internal/platform/config"
	"rudy_gc/remote/api/internal/platform/database"
)

var (
	configFile = flag.String("f", "config/server.yaml", "DailyBest 远端服务配置文件")
	dataSource = flag.String("datasource", "", "仅本次迁移使用的远端数据库连接；留空时读取配置文件")
)

func main() {
	flag.Parse()
	cfg := config.MustLoad(*configFile)
	if value := strings.TrimSpace(*dataSource); value != "" {
		cfg.DataSource = value
	}
	db, err := database.NewGormDBEngine(&database.Config{
		DSN:      strings.TrimSpace(cfg.DataSource),
		LogLevel: strings.TrimSpace(cfg.LogursLevel),
	})
	if err != nil {
		log.Fatalf("DailyBest 远端数据库连接失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("DailyBest 远端数据库连接获取失败: %v", err)
	}
	defer sqlDB.Close()
	if err := dailybestmigration.Run(db); err != nil {
		log.Fatalf("DailyBest 远端数据库迁移失败: %v", err)
	}
	log.Println("DailyBest 远端数据库迁移完成")
}
