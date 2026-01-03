package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type GenreRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
	// 新增：按主键查
	FindOne(ctx context.Context, id int64) (*moviex.AmGenre, error)
	// 更新统计字段（仅单行）
	UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error
}
