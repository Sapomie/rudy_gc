// internal/spider/svc/deps.go
package svc

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/config"
	"rudy_gc/internal/infra"
	"rudy_gc/internal/repo"
	"rudy_gc/internal/spider/fetcher"
)

// Deps 聚合 Spider 模块运行所需的依赖
type Deps struct {
	Config        config.Config
	SeedRepo      repo.SeedRepo
	InventoryRepo repo.InventoryRepo
	Fetcher       *fetcher.Fetcher
}

// NewDeps 构造依赖：数据库模型 → repo → fetcher
func NewDeps(cfg config.Config) *Deps {
	// DB 连接
	conn := sqlx.NewMysql(cfg.DataSource)

	// Model & Repo
	seedModel := spiderx.NewDSeedModel(conn)
	invModel := spiderx.NewDInventoryModel(conn)

	seedRepo := infra.NewSeedRepoSqlx(seedModel)
	inventoryRepo := infra.NewInventoryRepoSqlx(invModel)

	// Fetcher
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
		Fetcher:       f,
	}
}
