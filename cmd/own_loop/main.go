package main

import (
	"context"
	"flag"
	"fmt"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/sc"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.Disable()

	var err error
	ctx := context.Background()
	deps, err := svc.NewDeps(c)
	if err != nil {
		fmt.Println(ctx)
		panic(err)
	}

	err = sc.NewScService(deps).PickProcession()
	if err != nil {
		panic(err)
	}

}
