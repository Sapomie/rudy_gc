package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"rudy_gc/remote/api/internal/config"
	"rudy_gc/remote/api/internal/model/modelg"
	dailybestremotex "rudy_gc/remote/api/internal/model/modelx/dailybest"
	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
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
	db, err := modelg.NewGormDBEngine(&modelg.Config{
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
	if err := modelg.AutoMigrate(db); err != nil {
		log.Fatalf("DailyBest 远端数据库迁移失败: %v", err)
	}
	loc, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		log.Fatalf("DailyBest 远端时区加载失败: %v", err)
	}
	service := dailybestservice.New(context.Background(), dailybestremotex.New(db), nil, loc)
	if err := service.EnsureRankProjection(context.Background()); err != nil {
		log.Fatalf("DailyBest 远端排名派生数据重建失败: %v", err)
	}
	log.Println("DailyBest 远端数据库迁移完成")
}
