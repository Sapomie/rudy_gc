package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/vfilm"
	"rudy_gc/internal/svc"
	"rudy_gc/pkg/redis"

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

	//flushDb()

	//err = logic.NewCrawlLogic(deps).CrawlDailyBestProcession(ctx)
	//if err != nil {
	//	panic(err)
	//}

	err = vfilm.NewFilmService(deps).ProcessFilm(ctx)
	if err != nil {
		panic(err)
	}

	//err = logic.NewCrawlLogic(ctx, deps).ParseDetails()
	//if err != nil {
	//	panic(err)
	//}

	//err = logic.NewCrawlLogic(ctx, deps).DownLoadAllPicture()
	//err = logic.NewCrawlLogic(ctx, deps).TranslateTitle()
	//if err != nil {
	//	panic(err)
	//}
	//

	//err = sc.NewScService(deps).AddSc(ctx, "/Users/gaojinwei/Desktop/temp/sc/2025-10-20-10-15")
	//if err != nil {
	//	panic(err)
	//}

	//movieType, err := movie.New(deps).GetMovieType(ctx, "javmezriqa")
	//if err != nil {
	//	panic(err)
	//}
	//deps.Log.Info("movieType", movieType.Name)

	//l :=
	//err = l.FetchAndParseInventoryBySeed()
	//if err != nil {
	//	panic(err)
	//}
	//_, err = l.FetchAndParseDetails()
	//if err != nil {
	//	panic(err)
	//}

	//err = l.FetchAndParseDailyBestinv()
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

func flushDb() {
	redis.FlushDB("6378")
	redis.FlushDB("63789")
}
