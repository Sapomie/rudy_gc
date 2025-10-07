package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type FilmRepo interface {
	UpsertFilm(ctx context.Context, in *types.Film) (*types.Film, types.UpsertStatus, error)
	FindAll(ctx context.Context) ([]*types.Film, error)

	FindOne(ctx context.Context, id int64) (*types.Film, error)
	FindOneByMovieJavId(ctx context.Context, javId string) (*types.Film, error)
	FindOneByMovieName(ctx context.Context, name string) (*types.Film, error)
}
