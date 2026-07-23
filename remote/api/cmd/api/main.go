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

	"github.com/gin-gonic/gin"

	"rudy_gc/remote/api/internal/config"
	"rudy_gc/remote/api/internal/dep"
	"rudy_gc/remote/api/internal/model/modelg"
	"rudy_gc/remote/api/internal/router"
	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
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

	deps, err := dep.BuildDep(cfg)
	if err != nil {
		log.Fatalf("DailyBest 远端服务构建依赖失败: %v", err)
	}
	defer func() {
		if err := deps.Close(); err != nil {
			log.Printf("DailyBest 远端服务关闭依赖失败: %v", err)
		}
	}()
	if err := modelg.AutoMigrate(deps.DB); err != nil {
		log.Fatalf("DailyBest 远端服务迁移数据库失败: %v", err)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	dailyBestSvc := dailybestservice.New(rootCtx, deps.DailyBestRepository, deps.DailyBestSource, deps.Location, deps.DailyBestBrowser)
	if err := dailyBestSvc.PrepareRuntime(rootCtx); err != nil {
		log.Fatalf("DailyBest 远端服务准备运行状态失败: %v", err)
	}
	if err := dailyBestSvc.EnsureRankDayStats(rootCtx); err != nil {
		log.Fatalf("DailyBest 远端服务重建排名派生数据失败: %v", err)
	}
	if err := dailyBestSvc.StartScheduler(rootCtx, cfg.DailyFetchTime); err != nil {
		log.Fatalf("DailyBest 远端服务启动定时任务失败: %v", err)
	}
	if err := dailyBestSvc.StartFailureRecovery(rootCtx, cfg.FailureRetryDuration()); err != nil {
		log.Fatalf("DailyBest 远端服务启动失败恢复任务失败: %v", err)
	}
	if err := dailyBestSvc.StartTargetProbeScheduler(rootCtx, cfg.TargetProbeDuration(), cfg.TargetProbe.URL); err != nil {
		log.Fatalf("DailyBest 远端服务启动目标站探测失败: %v", err)
	}
	if err := dailyBestSvc.StartBrowserMonitor(rootCtx, parseDuration(cfg.BrowserMonitorInterval)); err != nil {
		log.Fatalf("DailyBest 远端服务启动浏览器监控失败: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	server := &http.Server{Addr: cfg.ListenAddr, Handler: router.New(dailyBestSvc)}
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
	dailyBestSvc.Wait()
}

func parseDuration(value string) time.Duration {
	duration, _ := time.ParseDuration(strings.TrimSpace(value))
	return duration
}
