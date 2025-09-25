package svc

import (
	"rudy_gc/data/modelx/spider"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/infra"
)

func NewDeps(conn sqlx.SqlConn) *Deps {
	seedModel := spider.NewDSeedModel(conn)
	seedRepo := infra.NewSeedRepoSqlx(seedModel)

	return &Deps{
		SeedRepo: seedRepo,
	}
}
