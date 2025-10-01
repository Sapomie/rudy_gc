package movie_repo

import (
	"context"

	"rudy_gc/data/modelx/moviex"
)

type MovieRepo interface {
	UpsertByJavId(ctx context.Context, mv *moviex.AMovie) (*moviex.AMovie, error)
}
