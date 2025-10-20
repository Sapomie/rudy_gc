package movie

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/internal/types"
)

const (
	fieldTypeName int = iota
	fieldTypeEncode
)
const defaultMovieJavId = "javli6a53m"

func (s *MovieService) GetMovieDetailByName(ctx context.Context, movieName string) (*types.MovieDetail, error) {
	movie, err := s.findOrFallbackMovie(ctx, movieName)
	if err != nil {
		return nil, err
	}
	return s.buildMovieDetail(ctx, movie)
}

func (s *MovieService) findOrFallbackMovie(ctx context.Context, movieName string) (*types.Movie, error) {
	return s.findOrFallbackMovieByField(ctx, fieldTypeName, movieName)
}

func (s *MovieService) findOrFallbackMovieByEncodeName(ctx context.Context, encodeName string) (*types.Movie, error) {
	return s.findOrFallbackMovieByField(ctx, fieldTypeEncode, encodeName)
}

// 通用的电影查找方法，参数 FieldName 用来标识查询字段（如 "Name" 或 "EncodeName"）
func (s *MovieService) findOrFallbackMovieByField(ctx context.Context, fieldType int, fieldValue string) (*types.Movie, error) {
	var movies []*types.Movie
	var err error

	// 根据不同字段查询
	switch fieldType {
	case fieldTypeName:
		movies, err = s.deps.MovieRepo.FindMoviesByName(ctx, fieldValue)
	case fieldTypeEncode:
		movies, err = s.deps.MovieRepo.FindMoviesByName(ctx, fieldValue)
	default:
		return nil, fmt.Errorf("invalid fieldTypeName: %v", fieldTypeName)
	}
	if err != nil {
		return nil, errors.New("failed to find movie: " + err.Error())
	}

	// 处理查询结果
	if len(movies) == 1 {
		return movies[0], nil
	} else if len(movies) < 1 {
		return s.deps.MovieRepo.FindOneByJavId(ctx, defaultMovieJavId)
	} else { // len(movies) > 1
		movie := movies[0]
		movie.Name += "(----1----)"
		movie.Title += "(----1----)"
		return movie, nil
	}
}
