package fetchsite

import (
	"context"
	"strings"
)

func (s *Service) FindJavbusFetchTask(ctx context.Context, movieJavID string) (*JavbusFetchTask, error) {
	row, err := s.deps.JavbusMagnetFetchModel.FindOneByMovieJavId(ctx, strings.TrimSpace(movieJavID))
	if err != nil {
		return nil, err
	}
	return &JavbusFetchTask{
		MovieJavID: row.MovieJavId,
		MovieName:  row.MovieName,
		Row:        row,
	}, nil
}

func (s *Service) FindSukebeiFetchTask(ctx context.Context, movieJavID string) (*SukebeiFetchTask, error) {
	row, err := s.deps.SukebeiTorrentFetchModel.FindOneByMovieJavId(ctx, strings.TrimSpace(movieJavID))
	if err != nil {
		return nil, err
	}
	return &SukebeiFetchTask{
		MovieJavID: row.MovieJavId,
		MovieName:  row.MovieName,
		Row:        row,
	}, nil
}
