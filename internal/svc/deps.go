package svc

import (
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/config"
	"rudy_gc/internal/infra"
	"rudy_gc/internal/repo"
	"rudy_gc/internal/spider/fetcher"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Deps struct {
	Config        config.Config
	SeedRepo      repo.SeedRepo
	InventoryRepo repo.InventoryRepo
	ItemRepo      repo.ItemRepo
	DetailRepo    repo.DetailRepo
	Fetcher       *fetcher.Fetcher
}

func NewDeps(cfg config.Config) *Deps {
	conn := sqlx.NewMysql(cfg.DataSource)

	seedModel := spiderx.NewDSeedModel(conn)
	invModel := spiderx.NewDInventoryModel(conn)
	itemModel := spiderx.NewEItemModel(conn)
	detailModel := spiderx.NewDDetailModel(conn)

	seedRepo := infra.NewSeedRepoSqlx(seedModel)
	inventoryRepo := infra.NewInventoryRepoSqlx(invModel)
	itemRepo := infra.NewItemRepoSqlx(itemModel)
	detailRepo := infra.NewDetailRepoSqlx(detailModel)

	f := fetcher.NewFetcher(fetcher.Config{
		UserAgent: cfg.Fetcher.UserAgent,
		Cookie:    cfg.Fetcher.Cookie,
		Proxy:     cfg.Fetcher.Proxy,
		Timeout:   15 * time.Second,
	})

	return &Deps{
		Config:        cfg,
		SeedRepo:      seedRepo,
		InventoryRepo: inventoryRepo,
		ItemRepo:      itemRepo,
		DetailRepo:    detailRepo,
		Fetcher:       f,
	}
}
