package sc

import (
	"context"
	"errors"
	"fmt"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/taskctx"
)

type movieScInfo struct {
	ScEventTimes int64
	ScTimes      int64
	ComeTimes    int64
	LastScTime   int64
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
		info, ok := movieAgg[movieJavID]
		if !ok || info.ScTimes <= 0 {
			continue
		}

		movieRow, err := l.movieFindOneByJavID(ctx, movieJavID)
		if err != nil {
			return fmt.Errorf("MovieRepo.FindOneByJavId %s: %w", movieJavID, err)
		}
		releasingDate := movieRow.ReleasingDate
		mediaBirthTime := int64(0)
		mediaRow, err := l.wMediaFindOneByMovieJavID(ctx, movieJavID)
		if err == nil && mediaRow != nil {
			mediaBirthTime = mediaRow.BirthTime
		} else if err != nil && !errors.Is(err, moviex.ErrNotFound) {
			return fmt.Errorf("WMediaModel.FindOneByMovieJavId %s: %w", movieJavID, err)
		}

		needInvalidateMovieType := false
		if _, statStatus, err := l.gScStatUpsert(ctx, movieJavID, movieRow.Name, releasingDate, mediaBirthTime, info); err != nil {
			return fmt.Errorf("GScStat.Upsert %s: %w", movieJavID, err)
		} else if statStatus != consts.UpsertUnchanged {
			needInvalidateMovieType = true
		}

		if needInvalidateMovieType {
			l.movieSvc.InvalidateMovieType(ctx, movieJavID)
		}
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

	gls, err := l.glFindAllByMovieJavIDs(ctx, movieJavIDs)
	if err != nil {
		return nil, fmt.Errorf("find g_list by movies: %w", err)
	}
	if len(gls) == 0 {
		return result, nil
	}

	scNameSet := make(map[string]struct{}, len(gls))
	for _, gl := range gls {
		if gl == nil || gl.MovieJavId == "" || gl.ScName == "" {
			continue
		}
		scNameSet[gl.ScName] = struct{}{}
	}

	scNames := make([]string, 0, len(scNameSet))
	for scName := range scNameSet {
		scNames = append(scNames, scName)
	}

	scTimeMap, err := l.loadScTimes(ctx, scNames)
	if err != nil {
		return nil, err
	}

	for _, gl := range gls {
		if gl == nil || gl.MovieJavId == "" || gl.ScName == "" {
			continue
		}
		info := result[gl.MovieJavId]
		info.ScEventTimes++
		if gl.IsSc != consts.GListIsSc {
			result[gl.MovieJavId] = info
			continue
		}
		info.ScTimes++
		if gl.IsCome == consts.GListIsCome {
			info.ComeTimes++
		}
		if scTime := scTimeMap[gl.ScName]; scTime > info.LastScTime {
			info.LastScTime = scTime
		}
		result[gl.MovieJavId] = info
	}

	return result, nil
}

func (l *ScService) loadScTimes(ctx context.Context, names []string) (map[string]int64, error) {
	names = uniqueStrings(names)
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

func (l *ScService) RebuildAllScStats(ctx context.Context) error {
	movieJavIDs, err := l.glListDistinctMovieJavIDs(ctx)
	if err != nil {
		return fmt.Errorf("GListRepo.ListDistinctMovieJavIds: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:   "sc_movie_stats_begin",
		Message: fmt.Sprintf("开始回填影片 SC 统计，影片数=%d", len(movieJavIDs)),
	})

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
