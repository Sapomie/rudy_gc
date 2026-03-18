package sc

import (
	"context"
	"fmt"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/taskctx"
)

type movieScInfo struct {
	ScTimes    int64
	ComeTimes  int64
	LastScTime int64
}

func (l *ScService) AddMovieAndCastScInfo(ctx context.Context, movieJavIdMap map[string]struct{}) error {
	movieJavIDs := mapMovieJavIDs(movieJavIdMap)
	if len(movieJavIDs) == 0 {
		return nil
	}

	if err := l.rebuildMovieScStatsByJavIDs(ctx, movieJavIDs); err != nil {
		return fmt.Errorf("rebuild movie sc stats: %w", err)
	}

	castIDs, err := l.collectCastIDsByMovieJavIDs(ctx, movieJavIDs)
	if err != nil {
		return fmt.Errorf("collect affected casts: %w", err)
	}
	if len(castIDs) == 0 {
		return nil
	}

	if err := l.rebuildCastScStatsByIDs(ctx, castIDs); err != nil {
		return fmt.Errorf("rebuild cast sc stats: %w", err)
	}

	return nil
}

func (l *ScService) rebuildMovieScStatsByJavIDs(ctx context.Context, movieJavIDs []string) error {
	movieJavIDs = uniqueStrings(movieJavIDs)
	if len(movieJavIDs) == 0 {
		return nil
	}

	movieAgg, err := l.buildMovieScInfo(ctx, movieJavIDs)
	if err != nil {
		return err
	}

	for _, movieJavID := range movieJavIDs {
		info := movieAgg[movieJavID]

		vFilm, err := l.filmFindOneByMovieJavID(ctx, movieJavID)
		if err != nil {
			return fmt.Errorf("FilmRepo.FindOneByMovieJavId %s: %w", movieJavID, err)
		}

		if vFilm.ScTimes == info.ScTimes &&
			vFilm.ComeTimes == info.ComeTimes &&
			vFilm.LastScTime == info.LastScTime {
			continue
		}

		vFilm.ScTimes = info.ScTimes
		vFilm.ComeTimes = info.ComeTimes
		vFilm.LastScTime = info.LastScTime
		if _, _, err := l.filmUpsert(ctx, vFilm); err != nil {
			return fmt.Errorf("FilmRepo.UpsertFilm %s: %w", movieJavID, err)
		}
		l.movieSvc.InvalidateMovieType(ctx, movieJavID)
	}

	return nil
}

func (l *ScService) rebuildCastScStatsByIDs(ctx context.Context, castIDs map[int64]struct{}) error {
	for castID := range castIDs {
		castRow, err := l.castFindOne(ctx, castID)
		if err != nil {
			return fmt.Errorf("CastRepo.FindOne %d: %w", castID, err)
		}

		movieJavIDs, err := l.movieCastListMovieJavIDsByCastID(ctx, castID)
		if err != nil {
			return fmt.Errorf("MovieCastRepo.ListMovieJavIDsByCastID %d: %w", castID, err)
		}

		movieAgg, err := l.buildMovieScInfo(ctx, movieJavIDs)
		if err != nil {
			return fmt.Errorf("build movie sc info for cast %d: %w", castID, err)
		}

		info := sumMovieScInfo(movieAgg)
		if castRow.ScTimes == info.ScTimes &&
			castRow.ComeTimes == info.ComeTimes &&
			castRow.LastScTime == info.LastScTime {
			continue
		}

		castRow.ScTimes = info.ScTimes
		castRow.ComeTimes = info.ComeTimes
		castRow.LastScTime = info.LastScTime
		if _, err := l.castUpsert(ctx, castRow); err != nil {
			return fmt.Errorf("CastRepo.Upsert %s: %w", castRow.Name, err)
		}
	}

	return nil
}

func (l *ScService) collectCastIDsByMovieJavIDs(ctx context.Context, movieJavIDs []string) (map[int64]struct{}, error) {
	out := make(map[int64]struct{})
	for _, movieJavID := range uniqueStrings(movieJavIDs) {
		castIDs, err := l.movieCastListCastIDsByMovieJavID(ctx, movieJavID)
		if err != nil {
			return nil, fmt.Errorf("MovieCastRepo.ListCastIDsByMovieJavId %s: %w", movieJavID, err)
		}
		for _, castID := range castIDs {
			if castID > 0 {
				out[castID] = struct{}{}
			}
		}
	}
	return out, nil
}

func (l *ScService) buildMovieScInfo(ctx context.Context, movieJavIDs []string) (map[string]movieScInfo, error) {
	result := make(map[string]movieScInfo)
	movieJavIDs = uniqueStrings(movieJavIDs)
	if len(movieJavIDs) == 0 {
		return result, nil
	}

	gls, err := l.glFindByMovieJavIDs(ctx, movieJavIDs)
	if err != nil {
		return nil, fmt.Errorf("find g_list by movies: %w", err)
	}
	if len(gls) == 0 {
		return result, nil
	}

	pairs := make(map[string]int64, len(gls))
	for _, gl := range gls {
		if gl == nil || gl.MovieJavId == "" || gl.ScName == "" {
			continue
		}
		key := gl.MovieJavId + "\x00" + gl.ScName
		if old, ok := pairs[key]; ok {
			if old == consts.GListIsCome || gl.IsCome != consts.GListIsCome {
				continue
			}
		}
		pairs[key] = gl.IsCome
	}

	scTimeMap, err := l.loadScTimes(ctx, pairs)
	if err != nil {
		return nil, err
	}

	for key, isCome := range pairs {
		movieJavID, scName := splitMovieScPair(key)
		info := result[movieJavID]
		info.ScTimes++
		if isCome == consts.GListIsCome {
			info.ComeTimes++
		}
		if scTime := scTimeMap[scName]; scTime > info.LastScTime {
			info.LastScTime = scTime
		}
		result[movieJavID] = info
	}

	return result, nil
}

func (l *ScService) loadScTimes(ctx context.Context, pairs map[string]int64) (map[string]int64, error) {
	scNames := make(map[string]struct{}, len(pairs))
	for key := range pairs {
		_, scName := splitMovieScPair(key)
		if scName != "" {
			scNames[scName] = struct{}{}
		}
	}

	names := make([]string, 0, len(scNames))
	for scName := range scNames {
		names = append(names, scName)
	}

	scRows, err := l.scFindByNames(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("ScRepo.FindByNames: %w", err)
	}

	out := make(map[string]int64, len(names))
	for _, row := range scRows {
		if row == nil || row.Name == "" {
			continue
		}
		out[row.Name] = row.ScTime
	}

	for _, name := range names {
		if _, ok := out[name]; ok {
			continue
		}
		scTime, err := getScTime(name)
		if err != nil {
			return nil, fmt.Errorf("getScTime %q: %w", name, err)
		}
		out[name] = scTime
	}

	return out, nil
}

func sumMovieScInfo(movieAgg map[string]movieScInfo) movieScInfo {
	var out movieScInfo
	for _, info := range movieAgg {
		out.ScTimes += info.ScTimes
		out.ComeTimes += info.ComeTimes
		if info.LastScTime > out.LastScTime {
			out.LastScTime = info.LastScTime
		}
	}
	return out
}

func mapMovieJavIDs(movieJavIdMap map[string]struct{}) []string {
	out := make([]string, 0, len(movieJavIdMap))
	for movieJavID := range movieJavIdMap {
		if movieJavID != "" {
			out = append(out, movieJavID)
		}
	}
	return out
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func splitMovieScPair(key string) (movieJavID string, scName string) {
	for i := 0; i < len(key); i++ {
		if key[i] != 0 {
			continue
		}
		return key[:i], key[i+1:]
	}
	return key, ""
}

func (l *ScService) RebuildAllScStats(ctx context.Context) error {
	films, err := l.filmFindAll(ctx, consts.OwnedAll)
	if err != nil {
		return fmt.Errorf("FilmRepo.FindAll: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:   "sc_movie_stats_begin",
		Message: fmt.Sprintf("开始回填影片 SC 统计，影片数=%d", len(films)),
	})

	movieJavIDs := make([]string, 0, len(films))
	for _, film := range films {
		if film == nil || film.MovieJavId == "" {
			continue
		}
		movieJavIDs = append(movieJavIDs, film.MovieJavId)
	}

	if err := l.rebuildMovieScStatsByJavIDs(ctx, movieJavIDs); err != nil {
		return fmt.Errorf("rebuild movie sc stats: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "sc_movie_stats_done",
		Message:      "影片 SC 统计回填完成",
		HandledCount: len(movieJavIDs),
		SuccessCount: len(movieJavIDs),
	})

	castIDs, err := l.castListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("CastRepo.ListAllIDs: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:   "sc_cast_stats_begin",
		Message: fmt.Sprintf("开始回填演员 SC 统计，演员数=%d", len(castIDs)),
	})

	castIDSet := make(map[int64]struct{}, len(castIDs))
	for _, castID := range castIDs {
		if castID > 0 {
			castIDSet[castID] = struct{}{}
		}
	}

	if err := l.rebuildCastScStatsByIDs(ctx, castIDSet); err != nil {
		return fmt.Errorf("rebuild cast sc stats: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "sc_cast_stats_done",
		Message:      "演员 SC 统计回填完成",
		HandledCount: len(castIDSet),
		SuccessCount: len(castIDSet),
	})

	return nil
}
