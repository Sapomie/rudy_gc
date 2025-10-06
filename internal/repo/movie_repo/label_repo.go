package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type LabelRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
	FindOne(ctx context.Context, id int64) (*moviex.AmLabel, error)
}
