package main

import (
	"flag"
	"rudy_gc/internal/config"
	migrate "rudy_gc/internal/domain/migrate_old"
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
	//ctx := context.Background()
	deps, err := svc.NewDeps(c)
	if err != nil {
		panic(err)
	}

	//err = migrate.New(deps).MigrateLocalCover()

	//err = migrate.New(deps).MigrateDetail()
	//err = migrate.New(deps).MigrateSc()
	//if err != nil {
	//	panic(err)
	//}
	//err = migrate.New(deps).MigrateGlist()
	//if err != nil {
	//	panic(err)
	//}
	//err = migrate.New(deps).AddScInfoToMinfo()
	//if err != nil {
	//	panic(err)
	//}

	//err = migrate.New(deps).MigrateFilm()
	//err = migrate.New(deps).MigrateRank()
	//err = migrate.New(deps).MigrateTranslationAndNeedDown()
	//err = migrate.New(deps).UpDateAllRankInfo()
	if err != nil {
		panic(err)
	}

}
