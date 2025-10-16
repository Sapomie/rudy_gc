package main

import (
	"context"
	"flag"
	"rudy_gc/internal/config"
	"rudy_gc/internal/domain/spider/logic"
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

	err = logic.NewCrawlLogic(ctx, deps).DailyBestProcession()
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

	//err = film.NewFilmService(deps).ProcessFilm(ctx)
	//if err != nil {
	//	panic(err)
	//}

	//movieType, err := movie.New(deps).GetMovieType(ctx, "javmezriqa")
	//if err != nil {
	//	panic(err)
	//}
	//deps.Log.Info("movieType", movieType.Name)

	//l :=
	//err = l.CrawlActiveSeeds()
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
