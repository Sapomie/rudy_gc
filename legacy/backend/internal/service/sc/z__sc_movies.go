package sc

import (
	"context"
	"rudy_gc/internal/types"
)

type scMovies struct {
	MovieType   *types.MovieType
	WatchedTime int64
}

func (l *ScService) allScMovies(ctx context.Context) ([]*scMovies, error) {
	glists, err := l.glFindAll(ctx)
	if err != nil {
		return nil, err
	}
	scs := make([]*scMovies, 0, len(glists))
	for _, gl := range glists {
		movieType, err := l.movieSvc.GetMovieType(ctx, gl.MovieJavId)
		if err != nil {
			return nil, err
		}
		gSc, err := l.scFindOneByName(ctx, gl.ScName)
		if err != nil {
			return nil, err
		}
		scs = append(scs, &scMovies{
			MovieType:   movieType,
			WatchedTime: gSc.ScTime,
		})
	}
	return scs, nil
}
