package fetchsehuatang

import (
	"context"
	"strings"
)

var sehuatangThreadMovieNameOverrides = map[string]string{
	"3270076": "FC2PPV-4821506",
}

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

func resolveSehuatangThreadMovieName(threadTID string, titles ...string) string {
	movieName := ""
	for _, title := range titles {
		movieName = parseMovieName(title)
		if movieName != "" {
			break
		}
	}
	if override := strings.TrimSpace(sehuatangThreadMovieNameOverrides[strings.TrimSpace(threadTID)]); override != "" {
		return override
	}
	return movieName
}

func (s *Service) resolveSehuatangThreadMovieFields(ctx context.Context, threadTID string, titles ...string) (string, string, error) {
	movieName := resolveSehuatangThreadMovieName(threadTID, titles...)
	movieJavID, err := s.resolveMovieJavIDByMovieName(ctx, movieName)
	if err != nil {
		return movieName, "", err
	}
	return movieName, movieJavID, nil
}
