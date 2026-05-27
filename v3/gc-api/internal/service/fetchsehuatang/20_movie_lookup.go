package fetchsehuatang

import (
	"context"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) resolveMovieJavIDByMovieName(ctx context.Context, movieName string) (string, error) {
	movieName = strings.TrimSpace(movieName)
	if movieName == "" {
		return "", nil
	}

	rows, err := s.deps.MovieModel.FindMoviesByName(ctx, movieName)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", nil
	}

	unique := make(map[string]struct{}, len(rows))
	first := ""
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.JavId)
		if javID == "" {
			continue
		}
		if first == "" {
			first = javID
		}
		unique[javID] = struct{}{}
	}

	if len(unique) == 1 {
		return first, nil
	}
	return "", nil
}

func (s *Service) repairSehuatangRowMovieFields(ctx context.Context, row *moviex.TSehuatangMagnet) error {
	if row == nil {
		return nil
	}

	movieName := strings.TrimSpace(row.MovieName)
	if movieName == "" {
		movieName = parseMovieName(row.ThreadTitle)
	}
	if movieName == "" {
		return nil
	}

	movieJavID, err := s.resolveMovieJavIDByMovieName(ctx, movieName)
	if err != nil {
		return err
	}

	row.MovieName = movieName
	row.MovieJavId = movieJavID
	return nil
}
