package fetchsukebei

import (
	"context"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) loadMovieReleaseDate(ctx context.Context, movieJavID string) (int64, error) {
	row, err := s.deps.MovieModel.FindOneByJavId(ctx, movieJavID)
	if err == nil {
		return row.ReleasingDate, nil
	}
	if err == moviex.ErrNotFound {
		return 0, nil
	}
	return 0, err
}
