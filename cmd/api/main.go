package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/spider/loop"
	"rudy_gc/internal/svc"
	"rudy_gc/pkg/mylog"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "cmd/api/config.yaml", "the config file")

func main() {
	var c config.Config
	conf.MustLoad(*configFile, &c)

	logx.DisableStat()
	logx.SetWriter(mylog.NewLogrusWriter(mylog.Options{
		JSON:            false,
		TimestampFormat: "2006-01-02 15:04:05",
		Level:           logrus.InfoLevel,
	}))

	ctx := context.Background()

	deps := svc.NewDeps(c)

	// 实例化并启动 loop
	ls := loop.NewLoopServer(ctx, deps, 1)
	ls.Start()

	// 示例：手动触发（实际应由管理接口写入 invCh）
	// invCh <- &spiderx.Notification{Info: spiderx.NotifyCrawActiveQueries}

	select {} // 阻塞主协程，或按你的方式优雅退出
}
