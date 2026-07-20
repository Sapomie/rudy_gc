package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dailybestmigration "rudy_gc/remote/api/internal/migration/dailybest"
	dailybestremotehttp "rudy_gc/remote/api/internal/modules/dailybest/adapter/inbound/http"
	dailybestremotex "rudy_gc/remote/api/internal/modules/dailybest/adapter/outbound/mysql"
	dailybestsource "rudy_gc/remote/api/internal/modules/dailybest/adapter/outbound/source"
	dailybestcollector "rudy_gc/remote/api/internal/modules/dailybest/application/collector"
	"rudy_gc/remote/api/internal/platform/config"
	"rudy_gc/remote/api/internal/platform/database"
	"rudy_gc/remote/api/internal/platform/sourceclient"

	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("f", "config/server.yaml", "DailyBest 远端服务配置文件")
	dataSource = flag.String("datasource", "", "仅本次运行使用的远端数据库连接；留空时读取配置文件")
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
		log.Fatalf("DailyBest 远端服务连接数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("DailyBest 远端服务获取数据库连接失败: %v", err)
	}
	defer sqlDB.Close()
	if err := dailybestmigration.Run(db); err != nil {
		log.Fatalf("DailyBest 远端服务迁移数据库失败: %v", err)
	}

	loc, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		log.Fatalf("DailyBest 远端服务加载时区失败: %v", err)
	}
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	repo := dailybestremotex.New(db)
	sourceHTTPClient := sourceclient.New(sourceclient.Config{
		UserAgent: cfg.Source.UserAgent,
		Cookie:    cfg.Source.Cookie,
		Timeout:   time.Duration(cfg.Source.Timeout) * time.Second,
	})
	collector := dailybestcollector.NewService(
		rootCtx,
		repo,
		dailybestsource.New(sourceHTTPClient, cfg.Source.JavAddress),
		loc,
	)
	if err := collector.StartScheduler(rootCtx, cfg.DailyFetchTime, cfg.FetchOnStart); err != nil {
		log.Fatalf("DailyBest 远端服务启动定时任务失败: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	engine := dailybestremotehttp.New(repo, collector, cfg.ReadToken, cfg.AdminToken)
	server := &http.Server{Addr: cfg.ListenAddr, Handler: engine}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("DailyBest 远端服务监听 %s", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-rootCtx.Done():
	case err := <-errCh:
		log.Printf("DailyBest 远端服务退出: %v", err)
		stop()
	}
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("DailyBest 远端服务关闭 HTTP 失败: %v", err)
	}
	collector.Wait()
}
