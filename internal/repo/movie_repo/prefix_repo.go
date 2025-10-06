package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type PrefixRepo interface {
	GetOrCreateByName(ctx context.Context, name string) (int64, error)
	FindOne(ctx context.Context, id int64) (*moviex.AmPrefix, error)
}
