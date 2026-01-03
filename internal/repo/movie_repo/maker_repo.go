package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type MakerRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
	FindOne(ctx context.Context, id int64) (*moviex.AmMaker, error)
	// 更新统计字段（仅单行）
	UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error

	// 全量 ID 列表
	ListAllIDs(ctx context.Context) ([]int64, error)
}
