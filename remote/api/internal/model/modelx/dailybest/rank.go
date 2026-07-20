package dailybestremotex

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"rudy_gc/remote/api/internal/model/modelg"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (r *Repository) ListAllRanks(ctx context.Context) ([]contracts.DailyBestRank, error) {
	var rows []modelg.Rank
	if err := r.db.WithContext(ctx).Order("day_number ASC, category ASC, rank_pos ASC, movie_jav_id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRanks(rows, nil), nil
}

func (r *Repository) MarkFetchRunNoChange(
	ctx context.Context,
	runID, finishedAt, pageCount, itemCount int64,
	checksum string,
	movies []dailybesttypes.MovieDraft,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertMovies(tx, movies, finishedAt); err != nil {
			return err
		}
		return tx.Model(&modelg.DailyBestFetchRun{}).Where("id = ?", runID).Updates(map[string]any{
			"status":        dailybesttypes.RunStatusNoChange,
			"finished_at":   finishedAt,
			"page_count":    pageCount,
			"item_count":    itemCount,
			"checksum":      checksum,
			"error_message": "",
			"updated_on":    finishedAt,
		}).Error
	})
}

func (r *Repository) PublishRanks(ctx context.Context, runID int64, draft dailybesttypes.RankDraft, projection dailybesttypes.RankProjection, finishedAt int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertMovies(tx, draft.Movies, finishedAt); err != nil {
			return err
		}
		if err := tx.Where("day_number = ? AND category = ?", draft.DayNumber, draft.Category).Delete(&modelg.Rank{}).Error; err != nil {
			return err
		}
		rows := make([]modelg.Rank, 0, len(draft.Ranks))
		for _, rank := range draft.Ranks {
			rows = append(rows, modelg.Rank{
				RankKey: rank.RankKey, MovieJavID: rank.MovieJavID, DayNumber: rank.DayNumber,
				RankPos: rank.RankPos, Category: rank.Category, CreatedOn: finishedAt, UpdatedOn: finishedAt,
			})
		}
		if len(rows) > 0 {
			if err := tx.Create(&rows).Error; err != nil {
				return fmt.Errorf("保存排行明细失败: %w", err)
			}
		}

		if err := replaceRankProjection(tx, projection); err != nil {
			return err
		}

		return tx.Model(&modelg.DailyBestFetchRun{}).Where("id = ?", runID).Updates(map[string]any{
			"status":        dailybesttypes.RunStatusSucceeded,
			"finished_at":   finishedAt,
			"page_count":    draft.PageCount,
			"item_count":    len(draft.Ranks),
			"checksum":      draft.Checksum,
			"error_message": "",
			"updated_on":    finishedAt,
		}).Error
	})
}

func upsertMovies(tx *gorm.DB, movies []dailybesttypes.MovieDraft, updatedOn int64) error {
	if len(movies) == 0 {
		return nil
	}
	rows := make([]modelg.DailyBestMovie, 0, len(movies))
	for _, movie := range movies {
		rows = append(rows, modelg.DailyBestMovie{
			MovieJavID: movie.MovieJavID,
			MovieName:  movie.MovieName,
			MovieCode:  movie.MovieCode,
			Title:      movie.Title,
			CoverURL:   movie.CoverURL,
			CreatedOn:  updatedOn,
			UpdatedOn:  updatedOn,
		})
	}
	updates := clause.Assignments(map[string]any{
		"movie_name": gorm.Expr("COALESCE(NULLIF(VALUES(movie_name), ''), movie_name)"),
		"movie_code": gorm.Expr("COALESCE(NULLIF(VALUES(movie_code), ''), movie_code)"),
		"title":      gorm.Expr("COALESCE(NULLIF(VALUES(title), ''), title)"),
		"cover_url":  gorm.Expr("COALESCE(NULLIF(VALUES(cover_url), ''), cover_url)"),
		"updated_on": updatedOn,
	})
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "movie_jav_id"}},
		DoUpdates: updates,
	}).Create(&rows).Error; err != nil {
		return fmt.Errorf("保存影片基本信息失败: %w", err)
	}
	return nil
}

func (r *Repository) RankProjectionComplete(ctx context.Context) (bool, error) {
	type counts struct {
		RankCount        int64 `gorm:"column:rank_count"`
		RankDayCount     int64 `gorm:"column:rank_day_count"`
		DayStatCount     int64 `gorm:"column:day_stat_count"`
		PeriodCount      int64 `gorm:"column:period_count"`
		PeriodStatCount  int64 `gorm:"column:period_stat_count"`
		EmptyPeriodCount int64 `gorm:"column:empty_period_count"`
	}
	var result counts
	if err := r.db.WithContext(ctx).Raw(`
SELECT
	(SELECT COUNT(*) FROM c_rank) AS rank_count,
	(SELECT COUNT(DISTINCT day_number) FROM c_rank) AS rank_day_count,
	(SELECT COUNT(*) FROM c_rank_day_stat) AS day_stat_count,
	(SELECT COUNT(*) FROM c_rank_period) AS period_count,
	(SELECT COUNT(*) FROM c_rank_period_stat) AS period_stat_count,
	(SELECT COUNT(*) FROM c_rank_period AS period
		LEFT JOIN c_rank_period_item AS item ON item.period_id = period.id
		WHERE item.id IS NULL) AS empty_period_count`).Scan(&result).Error; err != nil {
		return false, err
	}
	if result.RankCount == 0 {
		return true, nil
	}
	return result.RankDayCount == result.DayStatCount && result.PeriodCount > 0 &&
		result.PeriodCount == result.PeriodStatCount && result.EmptyPeriodCount == 0, nil
}

func (r *Repository) SyncRankProjection(ctx context.Context, projection dailybesttypes.RankProjection) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return replaceRankProjection(tx, projection)
	})
}

func replaceRankProjection(tx *gorm.DB, projection dailybesttypes.RankProjection) error {
	for _, model := range []any{&modelg.RankPeriodStat{}, &modelg.RankPeriodItem{}, &modelg.RankPeriod{}, &modelg.RankDayStat{}} {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error; err != nil {
			return fmt.Errorf("清理排行派生数据失败: %w", err)
		}
	}
	dayStats := make([]modelg.RankDayStat, 0, len(projection.DayStats))
	for _, stat := range projection.DayStats {
		dayStats = append(dayStats, rankDayStatModel(stat))
	}
	if len(dayStats) > 0 {
		if err := tx.Create(&dayStats).Error; err != nil {
			return fmt.Errorf("保存每日排行统计失败: %w", err)
		}
	}
	periodIDs := make(map[string]int64, len(projection.Periods))
	for _, item := range projection.Periods {
		period := rankPeriodModel(item.Period)
		if err := tx.Create(&period).Error; err != nil {
			return fmt.Errorf("保存周期排行 %s 失败: %w", item.Key, err)
		}
		periodIDs[item.Key] = period.Id
		periodItems := make([]modelg.RankPeriodItem, 0, len(item.Items))
		for _, source := range item.Items {
			row := rankPeriodItemModel(source)
			row.PeriodID = period.Id
			periodItems = append(periodItems, row)
		}
		if len(periodItems) > 0 {
			if err := tx.Create(&periodItems).Error; err != nil {
				return fmt.Errorf("保存周期排行条目 %s 失败: %w", item.Key, err)
			}
		}
		stat := rankPeriodStatModel(item.Stat)
		stat.PeriodID = period.Id
		stat.PrevPeriodID = periodIDs[item.PrevKey]
		if err := tx.Create(&stat).Error; err != nil {
			return fmt.Errorf("保存周期排行统计 %s 失败: %w", item.Key, err)
		}
	}
	return nil
}

func (r *Repository) ListRankSyncManifests(ctx context.Context) ([]contracts.DailyBestRankSyncManifest, error) {
	type row struct {
		FetchRunID   int64  `gorm:"column:fetch_run_id"`
		BusinessDate string `gorm:"column:business_date"`
		Category     int64  `gorm:"column:category"`
		PageCount    int64  `gorm:"column:page_count"`
		ItemCount    int64  `gorm:"column:item_count"`
		Checksum     string `gorm:"column:checksum"`
		FinishedAt   int64  `gorm:"column:finished_at"`
	}
	var rows []row
	if err := r.db.WithContext(ctx).Raw(`
SELECT run.id AS fetch_run_id, run.business_date, run.category, run.page_count,
	run.item_count, run.checksum, run.finished_at
FROM d_dailybest_fetch_run AS run
JOIN (
	SELECT business_date, category, MAX(id) AS fetch_run_id
	FROM d_dailybest_fetch_run
	WHERE status IN (?, ?) AND TRIM(checksum) <> ''
	GROUP BY business_date, category
) AS latest ON latest.fetch_run_id = run.id
ORDER BY run.business_date ASC, run.category ASC`, dailybesttypes.RunStatusSucceeded, dailybesttypes.RunStatusNoChange).Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]contracts.DailyBestRankSyncManifest, 0, len(rows))
	for _, source := range rows {
		dayNumber, err := contracts.RankDayNumber(source.BusinessDate)
		if err != nil {
			return nil, fmt.Errorf("抓取记录 %d 的业务日期无效: %w", source.FetchRunID, err)
		}
		out = append(out, contracts.DailyBestRankSyncManifest{
			FetchRunID: source.FetchRunID, BusinessDate: source.BusinessDate, DayNumber: dayNumber,
			Category: source.Category, PageCount: source.PageCount, ItemCount: source.ItemCount,
			Checksum: source.Checksum, FinishedAt: source.FinishedAt,
		})
	}
	return out, nil
}

func (r *Repository) ListRanks(ctx context.Context, filter dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.Rank{})
	if filter.DayNumber > 0 {
		query = query.Where("day_number = ?", filter.DayNumber)
	}
	if filter.Category > 0 {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.RankPos > 0 {
		query = query.Where("rank_pos = ?", filter.RankPos)
	}
	if filter.MovieJavID != "" {
		query = query.Where("movie_jav_id LIKE ?", "%"+filter.MovieJavID+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = paginate(query, filter.PageQuery)
	var rows []modelg.Rank
	if err := query.Order("day_number DESC, category ASC, rank_pos ASC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	movies, err := r.findMovies(ctx, rankMovieJavIDs(rows))
	if err != nil {
		return nil, 0, err
	}
	return mapRanks(rows, movies), total, nil
}

func (r *Repository) ListRankDayStats(ctx context.Context, filter dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.RankDayStat{})
	if filter.DayNumber > 0 {
		query = query.Where("day_number = ?", filter.DayNumber)
	}
	if filter.RankDate != "" {
		query = query.Where("rank_date = ?", filter.RankDate)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = paginate(query, filter.PageQuery)
	var rows []modelg.RankDayStat
	if err := query.Order("day_number DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]contracts.DailyBestRankDayStat, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRankDayStat(row))
	}
	return out, total, nil
}

func (r *Repository) ListRankPeriods(ctx context.Context, filter dailybesttypes.RankPeriodFilter) ([]contracts.DailyBestRankPeriod, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.RankPeriod{})
	if filter.PeriodType > 0 {
		query = query.Where("period_type = ?", filter.PeriodType)
	}
	if filter.PeriodKey != "" {
		query = query.Where("period_key = ?", filter.PeriodKey)
	}
	if filter.Category > 0 {
		query = query.Where("category = ?", filter.Category)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = paginate(query, filter.PageQuery)
	var rows []modelg.RankPeriod
	if err := query.Order("end_day_number DESC, period_type ASC, category ASC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]contracts.DailyBestRankPeriod, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRankPeriod(row))
	}
	return out, total, nil
}

func (r *Repository) ListRankPeriodItems(ctx context.Context, filter dailybesttypes.RankPeriodItemFilter) ([]contracts.DailyBestRankPeriodItem, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.RankPeriodItem{})
	if filter.PeriodID > 0 {
		query = query.Where("period_id = ?", filter.PeriodID)
	}
	if filter.MovieJavID != "" {
		query = query.Where("movie_jav_id LIKE ?", "%"+filter.MovieJavID+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = paginate(query, filter.PageQuery)
	var rows []modelg.RankPeriodItem
	if err := query.Order("period_id DESC, rank_pos ASC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	movies, err := r.findMovies(ctx, periodItemMovieJavIDs(rows))
	if err != nil {
		return nil, 0, err
	}
	out := make([]contracts.DailyBestRankPeriodItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRankPeriodItem(row, movies[row.MovieJavID]))
	}
	return out, total, nil
}

func (r *Repository) ListRankPeriodStats(ctx context.Context, filter dailybesttypes.RankPeriodStatFilter) ([]contracts.DailyBestRankPeriodStat, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.RankPeriodStat{})
	if filter.PeriodID > 0 {
		query = query.Where("period_id = ?", filter.PeriodID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = paginate(query, filter.PageQuery)
	var rows []modelg.RankPeriodStat
	if err := query.Order("period_id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	movies, err := r.findMovies(ctx, periodStatMovieJavIDs(rows))
	if err != nil {
		return nil, 0, err
	}
	out := make([]contracts.DailyBestRankPeriodStat, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRankPeriodStat(row, movies[row.RankOneMovieJavID]))
	}
	return out, total, nil
}

func paginate(query *gorm.DB, page dailybesttypes.PageQuery) *gorm.DB {
	if page.Enabled {
		return query.Offset(page.Offset()).Limit(page.Limit())
	}
	return query
}

func (r *Repository) findMovies(ctx context.Context, movieJavIDs []string) (map[string]modelg.DailyBestMovie, error) {
	movies := make(map[string]modelg.DailyBestMovie)
	if len(movieJavIDs) == 0 {
		return movies, nil
	}
	var rows []modelg.DailyBestMovie
	if err := r.db.WithContext(ctx).Where("movie_jav_id IN ?", movieJavIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		movies[row.MovieJavID] = row
	}
	return movies, nil
}

func rankMovieJavIDs(rows []modelg.Rank) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MovieJavID)
	}
	return ids
}

func periodItemMovieJavIDs(rows []modelg.RankPeriodItem) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.MovieJavID)
	}
	return ids
}

func periodStatMovieJavIDs(rows []modelg.RankPeriodStat) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.RankOneMovieJavID != "" {
			ids = append(ids, row.RankOneMovieJavID)
		}
	}
	return ids
}

func movieBasic(row modelg.DailyBestMovie) contracts.DailyBestMovieBasic {
	return contracts.DailyBestMovieBasic{
		MovieName: row.MovieName,
		MovieCode: row.MovieCode,
		Title:     row.Title,
		CoverURL:  row.CoverURL,
	}
}

func mapRanks(rows []modelg.Rank, movies map[string]modelg.DailyBestMovie) []contracts.DailyBestRank {
	out := make([]contracts.DailyBestRank, 0, len(rows))
	for _, row := range rows {
		out = append(out, contracts.DailyBestRank{
			ID: row.Id, RankKey: row.RankKey, MovieJavID: row.MovieJavID, DayNumber: row.DayNumber,
			RankPos: row.RankPos, Category: row.Category, Movie: movieBasic(movies[row.MovieJavID]),
			CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
		})
	}
	return out
}

func rankDayStatModel(source contracts.DailyBestRankDayStat) modelg.RankDayStat {
	return modelg.RankDayStat{
		DayNumber: source.DayNumber, RankDate: source.RankDate, PrevDayNumber: source.PrevDayNumber,
		RankRowCount: source.RankRowCount, MovieCount: source.MovieCount,
		ContinuedMovieCount: source.ContinuedMovieCount, FirstSeenMovieCount: source.FirstSeenMovieCount,
		ReturningMovieCount: source.ReturningMovieCount, DroppedMovieCount: source.DroppedMovieCount,
		RiseMovieCount: source.RiseMovieCount, FallMovieCount: source.FallMovieCount,
		UnchangedMovieCount: source.UnchangedMovieCount, SourceChecksum: source.SourceChecksum,
		CalculatedOn: source.CalculatedOn, CreatedOn: source.CreatedOn, UpdatedOn: source.UpdatedOn,
	}
}

func mapRankDayStat(row modelg.RankDayStat) contracts.DailyBestRankDayStat {
	return contracts.DailyBestRankDayStat{
		ID: row.Id, DayNumber: row.DayNumber, RankDate: row.RankDate, PrevDayNumber: row.PrevDayNumber,
		RankRowCount: row.RankRowCount, MovieCount: row.MovieCount,
		ContinuedMovieCount: row.ContinuedMovieCount, FirstSeenMovieCount: row.FirstSeenMovieCount,
		ReturningMovieCount: row.ReturningMovieCount, DroppedMovieCount: row.DroppedMovieCount,
		RiseMovieCount: row.RiseMovieCount, FallMovieCount: row.FallMovieCount,
		UnchangedMovieCount: row.UnchangedMovieCount, SourceChecksum: row.SourceChecksum,
		CalculatedOn: row.CalculatedOn, CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}

func rankPeriodModel(source contracts.DailyBestRankPeriod) modelg.RankPeriod {
	return modelg.RankPeriod{
		PeriodType: source.PeriodType, PeriodKey: source.PeriodKey, Category: source.Category,
		StartDayNumber: source.StartDayNumber, EndDayNumber: source.EndDayNumber,
		PickDays: source.PickDays, TopN: source.TopN, Status: source.Status,
		CreatedOn: source.CreatedOn, UpdatedOn: source.UpdatedOn,
	}
}

func mapRankPeriod(row modelg.RankPeriod) contracts.DailyBestRankPeriod {
	return contracts.DailyBestRankPeriod{
		ID: row.Id, PeriodType: row.PeriodType, PeriodKey: row.PeriodKey, Category: row.Category,
		StartDayNumber: row.StartDayNumber, EndDayNumber: row.EndDayNumber,
		PickDays: row.PickDays, TopN: row.TopN, Status: row.Status,
		CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}

func rankPeriodItemModel(source contracts.DailyBestRankPeriodItem) modelg.RankPeriodItem {
	return modelg.RankPeriodItem{
		MovieJavID: source.MovieJavID, RankPos: source.RankPos, Score: source.Score,
		DaysInRank: source.DaysInRank, UsedPickDays: source.UsedPickDays,
		FirstRankDayNumber: source.FirstRankDayNumber, LastRankDayNumber: source.LastRankDayNumber,
		BestRank: source.BestRank, WorstPickedRank: source.WorstPickedRank,
		PrevRank: source.PrevRank, RankChange: source.RankChange, PeakRank: source.PeakRank,
		CreatedOn: source.CreatedOn, UpdatedOn: source.UpdatedOn,
	}
}

func mapRankPeriodItem(row modelg.RankPeriodItem, movie modelg.DailyBestMovie) contracts.DailyBestRankPeriodItem {
	return contracts.DailyBestRankPeriodItem{
		ID: row.Id, PeriodID: row.PeriodID, MovieJavID: row.MovieJavID, Movie: movieBasic(movie),
		RankPos: row.RankPos, Score: row.Score, DaysInRank: row.DaysInRank, UsedPickDays: row.UsedPickDays,
		FirstRankDayNumber: row.FirstRankDayNumber, LastRankDayNumber: row.LastRankDayNumber,
		BestRank: row.BestRank, WorstPickedRank: row.WorstPickedRank,
		PrevRank: row.PrevRank, RankChange: row.RankChange, PeakRank: row.PeakRank,
		CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}

func rankPeriodStatModel(source contracts.DailyBestRankPeriodStat) modelg.RankPeriodStat {
	return modelg.RankPeriodStat{
		SourceRankRowCount: source.SourceRankRowCount, SourceRankDayCount: source.SourceRankDayCount,
		SourceMovieCount: source.SourceMovieCount, MissingRankDayCount: source.MissingRankDayCount,
		PeriodMovieCount: source.PeriodMovieCount, ContinuedMovieCount: source.ContinuedMovieCount,
		FirstSeenMovieCount: source.FirstSeenMovieCount, ReturningMovieCount: source.ReturningMovieCount,
		DroppedMovieCount: source.DroppedMovieCount, RiseMovieCount: source.RiseMovieCount,
		FallMovieCount: source.FallMovieCount, UnchangedMovieCount: source.UnchangedMovieCount,
		PeakMovieCount: source.PeakMovieCount, RankOneMovieJavID: source.RankOneMovieJavID,
		AverageScore: source.AverageScore, AverageDaysInRank: source.AverageDaysInRank,
		SourceFetchedStartOn: source.SourceFetchedStartOn, SourceFetchedEndOn: source.SourceFetchedEndOn,
		SourceChecksum: source.SourceChecksum, CalculatedOn: source.CalculatedOn,
		CreatedOn: source.CreatedOn, UpdatedOn: source.UpdatedOn,
	}
}

func mapRankPeriodStat(row modelg.RankPeriodStat, rankOneMovie modelg.DailyBestMovie) contracts.DailyBestRankPeriodStat {
	return contracts.DailyBestRankPeriodStat{
		ID: row.Id, PeriodID: row.PeriodID, PrevPeriodID: row.PrevPeriodID,
		SourceRankRowCount: row.SourceRankRowCount, SourceRankDayCount: row.SourceRankDayCount,
		SourceMovieCount: row.SourceMovieCount, MissingRankDayCount: row.MissingRankDayCount,
		PeriodMovieCount: row.PeriodMovieCount, ContinuedMovieCount: row.ContinuedMovieCount,
		FirstSeenMovieCount: row.FirstSeenMovieCount, ReturningMovieCount: row.ReturningMovieCount,
		DroppedMovieCount: row.DroppedMovieCount, RiseMovieCount: row.RiseMovieCount,
		FallMovieCount: row.FallMovieCount, UnchangedMovieCount: row.UnchangedMovieCount,
		PeakMovieCount: row.PeakMovieCount, RankOneMovieJavID: row.RankOneMovieJavID,
		RankOneMovie: movieBasic(rankOneMovie),
		AverageScore: row.AverageScore, AverageDaysInRank: row.AverageDaysInRank,
		SourceFetchedStartOn: row.SourceFetchedStartOn, SourceFetchedEndOn: row.SourceFetchedEndOn,
		SourceChecksum: row.SourceChecksum, CalculatedOn: row.CalculatedOn,
		CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}
