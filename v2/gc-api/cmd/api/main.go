package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/dep"
	"rudy-gc-api/internal/router"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	configFile = flag.String("f", "config/config.yaml", "the config file")
	addrFlag   = flag.String("addr", ":2041", "the api listen address")
)

func main() {
	flag.Parse()

	var cfg config.Config
	conf.MustLoad(*configFile, &cfg)

	// Match the v2 API process model: let this server own SIGINT/SIGTERM shutdown.
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	switch cfg.LogxMode {
	case "off":
	case "disable":
		logx.Disable()
	case "", "disable_stat":
		logx.DisableStat()
	default:
		logx.DisableStat()
	}

	dp := dep.Build(cfg)
	logger := dp.Log
	if logger == nil {
		logger = logrus.New()
	}

	engine := router.New(dp)
	server := &http.Server{
		Addr:              normalizeAddr(*addrFlag),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	logger.Infof("gc-api listening on %s", server.Addr)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		logger.Info("gc-api shutting down")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("gc-api start failed")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("gc-api graceful shutdown failed")
	}
}

func normalizeAddr(addr string) string {
	if addr == "" {
		return ":2041"
	}
	if addr[0] == ':' {
		return addr
	}
	return fmt.Sprintf(":%s", addr)
}
