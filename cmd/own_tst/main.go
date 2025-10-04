package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/spider/logic"
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
	var err error

	l := logic.NewCrawlLogic(ctx, deps)

	//err = l.CrawlActiveSeeds()
	//if err != nil {
	//	panic(err)
	//}
	_, err = l.FetchAndParseDetails()
	if err != nil {
		panic(err)
	}
	//
	//err = l.CrawlDailyBestinv()
	//if err != nil {
	//	panic(err)
	//}
	//_, err = l.FetchAndParseDetails()
	//if err != nil {
	//	panic(err)
	//}
	//err = l.ProcessBestinvRank()
	//if err != nil {
	//	panic(err)
	//}

}
