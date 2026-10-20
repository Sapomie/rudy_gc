package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/movie"
)

var configFile = flag.String("f", "config/config.yaml", "the config file")

func main() {
	flag.Parse()

	cfg := config.MustLoad(*configFile)
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		log.Fatalf("构建排行统计回填依赖失败：%v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("开始回填排行日统计和周期统计")
	if err := movie.NewService(deps).BackfillMissingRankStats(ctx); err != nil {
		log.Fatalf("排行统计回填失败：%v", err)
	}
	log.Printf("排行统计回填完成")
}
