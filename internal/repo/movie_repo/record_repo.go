package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type RecordRepo interface {
	// TryInsert: 以 Name 作为幂等键，已存在返回 (false, nil)，否则插入返回 (true, nil)
	TryInsert(ctx context.Context, rec *types.Record) (bool, error)

	// Find: 按起始时间与类型查询，limit 为最大返回条数（<=0 表示不限制）
	Find(ctx context.Context, startTimeFrom int64, typ string, limit int) ([]*types.Record, error)
}
