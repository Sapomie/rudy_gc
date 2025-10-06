package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type DirectorRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
	FindOne(ctx context.Context, id int64) (*moviex.AmDirector, error)
}
