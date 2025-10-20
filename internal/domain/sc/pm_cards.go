package sc

import (
	"context"
	"rudy_gc/internal/types"
)

func (l *ScService) PickMovieCards(ctx context.Context, req *requestWithWeight, n int) ([]*types.MovieType, error) {

	reqs :=
		l.PickMovie(ctx)

}
