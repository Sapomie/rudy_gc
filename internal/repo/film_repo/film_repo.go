// internal/repo/film_repo/film_repo.go
package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type FilmRepo interface {
	UpsertFilm(ctx context.Context, in *types.Film) (*types.Film, error)
}
