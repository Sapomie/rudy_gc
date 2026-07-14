package sc

import (
	"context"
	"errors"
	"fmt"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
)

type movieScInfo struct {
	ScEventTimes    int64
	ScTimes         int64
	ComeTimes       int64
	LastScTime      int64
	LastScEventTime int64
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

func (l *ScService) AddFC2MovieScInfo(ctx context.Context, movieIDMap map[int64]struct{}) error {
	movieIDs := make([]int64, 0, len(movieIDMap))
	for movieID := range movieIDMap {
		if movieID > 0 {
			movieIDs = append(movieIDs, movieID)
		}
	}
	if len(movieIDs) == 0 {
		return nil
	}
	if err := l.rebuildFC2MovieScStatsByIDs(ctx, movieIDs); err != nil {
		return fmt.Errorf("rebuild fc2 movie sc stats: %w", err)
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

		movieRow, err := l.movieFindOneByJavID(ctx, movieJavID)
		if err != nil {
			return fmt.Errorf("MovieRepo.FindOneByJavId %s: %w", movieJavID, err)
		}
		releasingDate := movieRow.ReleasingDate
		mediaBirthTime := int64(0)
		mediaRow, err := l.wMediaFindOneByMovieJavID(ctx, movieJavID)
		if err == nil && mediaRow != nil && mediaRow.IsRemoved != int64(consts.FilmIsRemoved) {
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

func (l *ScService) rebuildFC2MovieScStatsByIDs(ctx context.Context, movieIDs []int64) error {
	movieIDs = uniqueFC2MovieIDs(movieIDs)
	if len(movieIDs) == 0 {
		return nil
	}
	movieAgg, err := l.buildFC2MovieScInfo(ctx, movieIDs)
	if err != nil {
		return err
	}
	movieRows, err := l.deps.FC2MovieModel.ListByIDs(ctx, movieIDs)
	if err != nil {
		return fmt.Errorf("FC2MovieModel.ListByIDs: %w", err)
	}
	mediaRows, err := l.deps.FC2MediaModel.ListByMovieIDs(ctx, movieIDs)
	if err != nil {
		return fmt.Errorf("FC2MediaModel.ListByMovieIDs: %w", err)
	}
	mediaBirthByMovieID := make(map[int64]int64, len(mediaRows))
	for _, row := range mediaRows {
		if row == nil || row.FC2MovieID <= 0 {
			continue
		}
		if row.IsRemoved != int64(consts.FilmIsRemoved) {
			mediaBirthByMovieID[row.FC2MovieID] = row.BirthTime
		}
	}
	movieByID := make(map[int64]*moviex.FC2Movie, len(movieRows))
	for _, row := range movieRows {
		if row != nil && row.ID > 0 {
			movieByID[row.ID] = row
		}
	}
	for _, movieID := range movieIDs {
		movie := movieByID[movieID]
		if movie == nil {
			continue
		}
		if _, _, err := l.fc2ScStatUpsert(ctx, movie, mediaBirthByMovieID[movieID], movieAgg[movieID]); err != nil {
			return fmt.Errorf("FC2ScStat.Upsert %d: %w", movieID, err)
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
			castRow.LastScTime == info.LastScTime &&
			castRow.LastScEventTime == info.LastScEventTime {
			continue
		}

		castRow.ScTimes = info.ScTimes
		castRow.ComeTimes = info.ComeTimes
		castRow.LastScTime = info.LastScTime
		castRow.LastScEventTime = info.LastScEventTime
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

	for _, movieJavID := range movieJavIDs {
		result[movieJavID] = movieScInfo{}
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
		if scTime := scTimeMap[gl.ScName]; scTime > info.LastScEventTime {
			info.LastScEventTime = scTime
		}
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

func (l *ScService) buildFC2MovieScInfo(ctx context.Context, movieIDs []int64) (map[int64]movieScInfo, error) {
	movieIDs = uniqueFC2MovieIDs(movieIDs)
	if len(movieIDs) == 0 {
		return map[int64]movieScInfo{}, nil
	}
	gls, err := l.glFindAllByFC2MovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, fmt.Errorf("find g_list by fc2 movies: %w", err)
	}
	scNameSet := make(map[string]struct{}, len(gls))
	for _, gl := range gls {
		if gl == nil || gl.FC2MovieId <= 0 || gl.ScName == "" {
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
	return aggregateFC2MovieScInfo(movieIDs, gls, scTimeMap), nil
}

func aggregateFC2MovieScInfo(movieIDs []int64, gls []*types.GList, scTimeMap map[string]int64) map[int64]movieScInfo {
	result := make(map[int64]movieScInfo)
	for _, movieID := range uniqueFC2MovieIDs(movieIDs) {
		result[movieID] = movieScInfo{}
	}
	for _, gl := range gls {
		if gl == nil || gl.FC2MovieId <= 0 || gl.ScName == "" {
			continue
		}
		info := result[gl.FC2MovieId]
		info.ScEventTimes++
		if scTime := scTimeMap[gl.ScName]; scTime > info.LastScEventTime {
			info.LastScEventTime = scTime
		}
		if gl.IsSc == consts.GListIsSc {
			info.ScTimes++
			if gl.IsCome == consts.GListIsCome {
				info.ComeTimes++
			}
			if scTime := scTimeMap[gl.ScName]; scTime > info.LastScTime {
				info.LastScTime = scTime
			}
		}
		result[gl.FC2MovieId] = info
	}
	return result
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
		if info.LastScEventTime > out.LastScEventTime {
			out.LastScEventTime = info.LastScEventTime
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

func uniqueFC2MovieIDs(items []int64) []int64 {
	seen := make(map[int64]struct{}, len(items))
	out := make([]int64, 0, len(items))
	for _, item := range items {
		if item <= 0 {
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
	movieJavIDs, err := l.glListDistinctAllMovieJavIDs(ctx)
	if err != nil {
		return fmt.Errorf("GListRepo.ListDistinctMovieJavIds: %w", err)
	}
	existingMovieJavIDs, err := l.deps.GScStatModel.ListAllMovieJavIDs(ctx)
	if err != nil {
		return fmt.Errorf("GScStatModel.ListAllMovieJavIDs: %w", err)
	}
	movieJavIDs = uniqueStrings(append(movieJavIDs, existingMovieJavIDs...))
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

	fc2MovieIDs, err := l.glListDistinctAllFC2MovieIDs(ctx)
	if err != nil {
		return fmt.Errorf("GListRepo.ListDistinctFC2MovieIDs: %w", err)
	}
	existingFC2MovieIDs, err := l.deps.FC2ScStatModel.ListAllMovieIDs(ctx)
	if err != nil {
		return fmt.Errorf("FC2ScStatModel.ListAllMovieIDs: %w", err)
	}
	fc2MovieIDs = uniqueFC2MovieIDs(append(fc2MovieIDs, existingFC2MovieIDs...))
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:   "sc_fc2_movie_stats_begin",
		Message: fmt.Sprintf("开始回填 FC2 影片 SC 统计，影片数=%d", len(fc2MovieIDs)),
	})
	if err := l.rebuildFC2MovieScStatsByIDs(ctx, fc2MovieIDs); err != nil {
		return fmt.Errorf("rebuild fc2 movie sc stats: %w", err)
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "sc_fc2_movie_stats_done",
		Message:      "FC2 影片 SC 统计回填完成",
		HandledCount: len(fc2MovieIDs),
		SuccessCount: len(fc2MovieIDs),
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
