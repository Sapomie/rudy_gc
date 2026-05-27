package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	"rudy_gc/internal/svc"
	"rudy_gc/pkg/redis"

	"github.com/zeromicro/go-zero/core/conf"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	var err error
	ctx := context.Background()
	deps, err := svc.NewDeps(c)
	if err != nil {
		panic(err)
		deps.Log.Error(ctx, "NewDeps err:", err)
	}

	flushDb()

}

func flushDb() {
	redis.FlushDB("6378")
	redis.FlushDB("6379")
}
