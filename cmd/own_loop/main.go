package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	logic2 "rudy_gc/internal/domain/spider/logic"
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
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logic := logic2.NewCrawlLogic(ctx, deps)

	// 1) 创建通道（带缓冲更平滑）
	detailJobs := make(chan []string, 16)

	// 2) 启动 loop
	go logic.d(detailJobs)

	// 3) 在主流程的合适位置投递任务
	detailJobs <- []string{"IPZ-001", "SSIS-123"}
	detailJobs <- []string{"ABP-888"}

	// 4) 退出时：先 cancel ctx 或关闭通道
	// cancel()          // 方案A：通过 ctx 让 loop 退出
	// close(detailJobs) // 方案B：显式关闭通道（两者择一即可）

}
