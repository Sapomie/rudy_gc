// cmd/db_migrate
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/model/modelg"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/pkg/orm"
)

func main() {
	cfg := orm.Config{
		DSN:      "root:4521822123@tcp(127.0.0.1:3306)/rudy_gc?charset=utf8mb4",
		LogLevel: "info",
	}

	dbGorm := orm.MustNewGormDBEngine(&cfg)
	modelg.MustAutoMigrate(dbGorm)
	mustBackfillCPersonSc(cfg.DSN)
}

func mustBackfillCPersonSc(dsn string) {
	const batchSize = 200

	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	personScModel := moviex.NewCPersonScModel(conn, nil)

	var personIDs []int64
	err := conn.QueryRowsCtx(ctx, &personIDs, "select `id` from `c_person` where `id` > 0 order by `id` asc")
	if err != nil {
		panic(fmt.Sprintf("list c_person ids failed: %v", err))
	}
	if len(personIDs) == 0 {
		return
	}

	now := time.Now().Unix()
	for start := 0; start < len(personIDs); start += batchSize {
		end := start + batchSize
		if end > len(personIDs) {
			end = len(personIDs)
		}
		if err := personScModel.RebuildByPersonIDs(ctx, personIDs[start:end], now); err != nil {
			panic(fmt.Sprintf("backfill c_person_sc failed, batch=%d-%d, err=%v", start, end, err))
		}
	}
}
