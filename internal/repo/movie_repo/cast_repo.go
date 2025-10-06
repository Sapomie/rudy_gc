package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type CastRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)

	// 新增：按主键查
	FindOne(ctx context.Context, id int64) (*moviex.AmCast, error)
}
