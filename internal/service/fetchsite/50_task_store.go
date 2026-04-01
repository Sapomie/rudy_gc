package fetchsite

import (
	"context"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) ListPendingJavbusFetchTasks(ctx context.Context, limit int64) ([]*JavbusFetchTask, error) {
	statuses := []int64{FetchStatusPending, FetchStatusFailed}
	rows, err := s.deps.JavbusMagnetFetchModel.ListByFetchStatuses(ctx, statuses, limit)
	if err != nil {
		return nil, err
	}

	out := make([]*JavbusFetchTask, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" || row.MovieName == "" {
			continue
		}
		out = append(out, &JavbusFetchTask{
			MovieJavID: row.MovieJavId,
			MovieName:  row.MovieName,
			Row:        row,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		left := pickRowOrderTimeJavbus(out[i].Row)
		right := pickRowOrderTimeJavbus(out[j].Row)
		if left == right {
			return out[i].MovieJavID < out[j].MovieJavID
		}
		return left < right
	})
	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Service) ListPendingSukebeiFetchTasks(ctx context.Context, limit int64) ([]*SukebeiFetchTask, error) {
	statuses := []int64{FetchStatusPending, FetchStatusFailed}
	rows, err := s.deps.SukebeiTorrentFetchModel.ListByFetchStatuses(ctx, statuses, limit)
	if err != nil {
		return nil, err
	}

	out := make([]*SukebeiFetchTask, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.MovieJavId == "" || row.MovieName == "" {
			continue
		}
		out = append(out, &SukebeiFetchTask{
			MovieJavID: row.MovieJavId,
			MovieName:  row.MovieName,
			Row:        row,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		left := pickRowOrderTimeSukebei(out[i].Row)
		right := pickRowOrderTimeSukebei(out[j].Row)
		if left == right {
			return out[i].MovieJavID < out[j].MovieJavID
		}
		return left < right
	})
	if limit > 0 && int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Service) BuildJavbusFetchTasksByMovies(ctx context.Context, movies []*types.MovieType) ([]*JavbusFetchTask, error) {
	now := time.Now().Unix()
	out := make([]*JavbusFetchTask, 0, len(movies))
	seen := make(map[string]struct{}, len(movies))

	for _, mv := range movies {
		if mv == nil {
			continue
		}
		movieJavID := strings.TrimSpace(mv.JavId)
		movieName := strings.TrimSpace(mv.Name)
		if movieJavID == "" || movieName == "" {
			continue
		}
		if _, ok := seen[movieJavID]; ok {
			continue
		}
		seen[movieJavID] = struct{}{}

		if err := s.ensureJavbusFetchTask(ctx, movieJavID, movieName, pickMovieReleaseDate(mv), now); err != nil {
			return nil, err
		}
		task, err := s.FindJavbusFetchTask(ctx, movieJavID)
		if err != nil {
			return nil, err
		}
		if task == nil {
			continue
		}
		task.MovieName = movieName
		out = append(out, task)
	}
	return out, nil
}

func (s *Service) BuildSukebeiFetchTasksByMovies(ctx context.Context, movies []*types.MovieType) ([]*SukebeiFetchTask, error) {
	now := time.Now().Unix()
	out := make([]*SukebeiFetchTask, 0, len(movies))
	seen := make(map[string]struct{}, len(movies))

	for _, mv := range movies {
		if mv == nil {
			continue
		}
		movieJavID := strings.TrimSpace(mv.JavId)
		movieName := strings.TrimSpace(mv.Name)
		if movieJavID == "" || movieName == "" {
			continue
		}
		if _, ok := seen[movieJavID]; ok {
			continue
		}
		seen[movieJavID] = struct{}{}

		if err := s.ensureSukebeiFetchTask(ctx, movieJavID, movieName, pickMovieReleaseDate(mv), now); err != nil {
			return nil, err
		}
		task, err := s.FindSukebeiFetchTask(ctx, movieJavID)
		if err != nil {
			return nil, err
		}
		if task == nil {
			continue
		}
		task.MovieName = movieName
		out = append(out, task)
	}
	return out, nil
}

func (s *Service) BuildSukebeiFetchTasksByRows(rows []*moviex.TSukebeiTorrentFetch) []*SukebeiFetchTask {
	out := make([]*SukebeiFetchTask, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		movieJavID := strings.TrimSpace(row.MovieJavId)
		movieName := strings.TrimSpace(row.MovieName)
		if movieJavID == "" || movieName == "" {
			continue
		}
		if _, ok := seen[movieJavID]; ok {
			continue
		}
		seen[movieJavID] = struct{}{}
		out = append(out, &SukebeiFetchTask{
			MovieJavID: movieJavID,
			MovieName:  movieName,
			Row:        row,
		})
	}
	return out
}

func (s *Service) MarkJavbusRunning(ctx context.Context, row *moviex.TJavbusMagnetFetch) error {
	if row == nil {
		return nil
	}
	now := time.Now().Unix()
	row.FetchStatus = FetchStatusRunning
	row.UpdatedOn = now
	return s.deps.JavbusMagnetFetchModel.Update(ctx, row)
}

func (s *Service) MarkSukebeiRunning(ctx context.Context, row *moviex.TSukebeiTorrentFetch) error {
	if row == nil {
		return nil
	}
	now := time.Now().Unix()
	row.FetchStatus = FetchStatusRunning
	row.UpdatedOn = now
	return s.deps.SukebeiTorrentFetchModel.Update(ctx, row)
}

func ShouldFetchJavbus(row *moviex.TJavbusMagnetFetch) bool {
	return row != nil && (row.FetchStatus == FetchStatusPending || row.FetchStatus == FetchStatusFailed)
}

func ShouldFetchSukebei(row *moviex.TSukebeiTorrentFetch) bool {
	return row != nil && (row.FetchStatus == FetchStatusPending || row.FetchStatus == FetchStatusFailed)
}

func pickRowOrderTimeJavbus(row *moviex.TJavbusMagnetFetch) int64 {
	if row == nil {
		return 0
	}
	if row.LastFetchTime > 0 {
		return row.LastFetchTime
	}
	return row.CreatedOn
}

func pickRowOrderTimeSukebei(row *moviex.TSukebeiTorrentFetch) int64 {
	if row == nil {
		return 0
	}
	if row.LastFetchTime > 0 {
		return row.LastFetchTime
	}
	return row.CreatedOn
}

func pickMovieReleaseDate(movie *types.MovieType) int64 {
	if movie == nil {
		return 0
	}
	if movie.AMovie != nil && movie.AMovie.ReleasingDate > 0 {
		return movie.AMovie.ReleasingDate
	}
	return 0
}
