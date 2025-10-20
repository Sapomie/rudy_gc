package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/loop"
	"rudy_gc/internal/svc"
	http2 "rudy_gc/internal/transport/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	flag.Parse()

	// 1) 读取配置
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 2) 关闭 go-zero 自带日志（你已有）
	logx.Disable()

	// 3) 构建依赖
	deps, err := svc.NewDeps(c)
	if err != nil {
		log.Fatal(err)
	}

	// 4) 根上下文：可被信号取消
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := loop.NewFetchLoopService(deps)

	// 先启动详情抓取 loop（可随时被 DailyBest 暂停/恢复）
	srv.StartDetailLoop(ctx, 10*time.Second, 100)

	// 启动“每日榜触发”loop
	go srv.ProcessionTriggerLoop(ctx, deps.BestTrigger)

	// 6) 启动 HTTP Server（建议放协程里）
	engine := http2.NewEngine(deps)
	s := &http.Server{
		Addr:         c.Port,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP listening on %s\n", c.Port)
		// http.ErrServerClosed 视为正常关闭
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	//select {
	//case deps.BestTrigger <- contracts.TriggerMsg{Kind: contracts.ProcSeeds}:
	//default:
	//}

	// 7) 等待：信号 or 服务器异常退出
	select {
	case <-ctx.Done():
		// 收到 SIGINT/SIGTERM
	case err := <-errCh:
		log.Printf("HTTP server exited with error: %v", err)
	}

	// 8) 优雅关闭 HTTP，再停止后台 loop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// 9) 通知 loop 退出：依赖于 ctx 取消；此处无需 close(deps.DetailJobs)
	//    如果你确实要 close，确保所有生产者已停止后再 close：
	// close(deps.DetailJobs)

	log.Println("bye")
}
