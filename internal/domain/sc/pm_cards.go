package sc

import (
	"context"
	"rudy_gc/internal/types"
)

func (l *ScService) PickMovieCards(ctx context.Context, req *types.ListMovieFullRequest, n int) ([]*types.MovieType, error) {

	movieTypes, err := l.PickMovieOnce(ctx, req, n)
	if err != nil {
		return nil, err
	}

	return movieTypes, nil
}
