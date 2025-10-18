package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/loop"
	"rudy_gc/internal/svc"
	http2 "rudy_gc/internal/transport/http"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Disable()

	deps, err := svc.NewDeps(c)
	if err != nil {
		log.Fatal(err)
	}

	//实例化并启动 loop
	engine := http2.NewEngine(deps)

	go loop.NewFetchLoopService(deps).DetailFetchLoopSingle(context.Background(), deps.DetailJobs, time.Second*10, 100)

	// 4) 启动 HTTP 服务器（可换成 graceful/shutdown）
	srv := &http.Server{
		Addr:         c.Port,
		Handler:      engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("HTTP listening on %s\n", c.Port)
	log.Fatal(srv.ListenAndServe())

	select {} // 阻塞主协程，或按你的方式优雅退出
}
