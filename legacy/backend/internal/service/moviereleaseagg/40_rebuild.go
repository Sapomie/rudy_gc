package moviereleaseagg

import (
	"context"
	"errors"
	"sort"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	aggKeyMovieRelease  = "movie_release"
	aggEventStatusRun   = "running"
	aggEventStatusOK    = "success"
	aggEventStatusError = "failed"
)

type rebuildSummary struct {
	ScopeCount  int64
	BucketCount int64
	TopCount    int64
}

func (s *Service) rebuildByBucketMonthsAndLogEvent(ctx context.Context, flowKey string, scopeCount int64, bucketMonths []int64) error {
	if len(bucketMonths) == 0 {
		return nil
	}
	if flowKey == "" {
		flowKey = "manual"
	}

	startedAt := time.Now()
	eventRow, err := s.startAggEvent(ctx, flowKey, startedAt.Unix())
	if err != nil {
		return err
	}

	summary, rebuildErr := s.rebuildByBucketMonths(ctx, bucketMonths, scopeCount)
	if finishErr := s.finishAggEvent(ctx, eventRow, startedAt, summary, rebuildErr); finishErr != nil {
		if rebuildErr != nil {
			return errors.Join(rebuildErr, finishErr)
		}
		return finishErr
	}
	return rebuildErr
}

func (s *Service) rebuildByBucketMonths(ctx context.Context, bucketMonths []int64, scopeCount int64) (rebuildSummary, error) {
	if scopeCount <= 0 {
		scopeCount = int64(len(bucketMonths))
	}
	summary := rebuildSummary{ScopeCount: scopeCount}
	dayScopes := make(map[string]scope)
	monthScopes := make(map[string]scope)
	quarterScopes := make(map[string]scope)
	yearScopes := make(map[string]scope)

	for _, bucketMonth := range bucketMonths {
		if bucketMonth <= 0 {
			continue
		}
		monthScope := scopeFromBucketMonth(bucketMonth)
		for _, aggMode := range []string{AggModeAll, AggModeOwned} {
			dayBuckets, err := s.deps.MovieModel.ListDistinctReleaseDaysByRangeMode(ctx, monthScope.StartUnix, monthScope.EndUnix, aggMode)
			if err != nil {
				return summary, err
			}
			for _, bucketDay := range dayBuckets {
				dayScope := scopeFromBucketDay(bucketDay)
				dayScopes[dayScope.Key] = dayScope
			}
		}
		monthScopes[monthScope.Key] = monthScope
		quarterScope := buildScope(monthScope.Year, monthScope.Quarter, 0, 0)
		yearScope := buildScope(monthScope.Year, 0, 0, 0)
		quarterScopes[quarterScope.Key] = quarterScope
		yearScopes[yearScope.Key] = yearScope
	}

	for _, aggMode := range []string{AggModeAll, AggModeOwned} {
		for _, sc := range sortedScopes(dayScopes) {
			if err := s.rebuildBucketScope(ctx, aggMode, sc); err != nil {
				return summary, err
			}
			summary.BucketCount++
		}
		for _, sc := range sortedScopes(monthScopes) {
			if err := s.rebuildBucketScope(ctx, aggMode, sc); err != nil {
				return summary, err
			}
			summary.BucketCount++
			topCount, err := s.rebuildTopScope(ctx, aggMode, sc)
			if err != nil {
				return summary, err
			}
			summary.TopCount += topCount
		}
		for _, sc := range sortedScopes(quarterScopes) {
			if err := s.rebuildBucketScope(ctx, aggMode, sc); err != nil {
				return summary, err
			}
			summary.BucketCount++
			topCount, err := s.rebuildTopScope(ctx, aggMode, sc)
			if err != nil {
				return summary, err
			}
			summary.TopCount += topCount
		}
		for _, sc := range sortedScopes(yearScopes) {
			if err := s.rebuildBucketScope(ctx, aggMode, sc); err != nil {
				return summary, err
			}
			summary.BucketCount++
			topCount, err := s.rebuildTopScope(ctx, aggMode, sc)
			if err != nil {
				return summary, err
			}
			summary.TopCount += topCount
		}
		topCount, err := s.rebuildTopScope(ctx, aggMode, buildScope(0, 0, 0, 0))
		if err != nil {
			return summary, err
		}
		summary.TopCount += topCount
	}
	return summary, nil
}

func (s *Service) rebuildBucketScope(ctx context.Context, aggMode string, sc scope) error {
	if sc.Level == levelRoot {
		return nil
	}

	calc, err := s.deps.MovieModel.CalcReleaseBucketByMode(ctx, sc.StartUnix, sc.EndUnix, aggMode)
	if err != nil {
		return err
	}
	if calc == nil {
		calc = &moviex.MovieReleaseBucketCalc{}
	}

	existing, err := s.deps.MovieReleaseBucketStatModel.FindOneByAggModeScopeKey(ctx, aggMode, sc.Key)
	if err != nil && err != moviex.ErrNotFound {
		return err
	}

	if calc.CountAll == 0 && calc.CountOwned == 0 && calc.SizeBytes == 0 {
		if existing != nil {
			return s.deps.MovieReleaseBucketStatModel.Delete(ctx, existing.Id)
		}
		return nil
	}

	now := time.Now().Unix()
	if existing == nil {
		_, err = s.deps.MovieReleaseBucketStatModel.Insert(ctx, &moviex.MovieReleaseBucketStat{
			AggMode:             aggMode,
			ScopeKey:            sc.Key,
			Level:               sc.Level,
			Year:                int64(sc.Year),
			Quarter:             int64(sc.Quarter),
			Month:               int64(sc.Month),
			Day:                 int64(sc.Day),
			CountAll:            calc.CountAll,
			CountOwned:          calc.CountOwned,
			SizeBytes:           calc.SizeBytes,
			LatestReleasingDate: calc.LatestReleasingDate,
			CreatedOn:           now,
			UpdatedOn:           now,
		})
		return err
	}

	existing.AggMode = aggMode
	existing.Level = sc.Level
	existing.Year = int64(sc.Year)
	existing.Quarter = int64(sc.Quarter)
	existing.Month = int64(sc.Month)
	existing.Day = int64(sc.Day)
	existing.CountAll = calc.CountAll
	existing.CountOwned = calc.CountOwned
	existing.SizeBytes = calc.SizeBytes
	existing.LatestReleasingDate = calc.LatestReleasingDate
	existing.UpdatedOn = now
	return s.deps.MovieReleaseBucketStatModel.Update(ctx, existing)
}

func (s *Service) rebuildTopScope(ctx context.Context, aggMode string, sc scope) (int64, error) {
	var total int64
	castRows, err := s.deps.MovieModel.CalcTopCastsByReleaseRangeMode(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit, aggMode)
	if err != nil {
		return total, err
	}
	inserted, err := s.replaceTopRows(ctx, aggMode, sc, aggTypeCast, castRows)
	if err != nil {
		return total, err
	}
	total += inserted

	directorRows, err := s.deps.MovieModel.CalcTopDirectorsByReleaseRangeMode(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit, aggMode)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, aggMode, sc, aggTypeDirector, directorRows)
	if err != nil {
		return total, err
	}
	total += inserted

	labelRows, err := s.deps.MovieModel.CalcTopLabelsByReleaseRangeMode(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit, aggMode)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, aggMode, sc, aggTypeLabel, labelRows)
	if err != nil {
		return total, err
	}
	total += inserted

	prefixRows, err := s.deps.MovieModel.CalcTopPrefixesByReleaseRangeMode(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit, aggMode)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, aggMode, sc, aggTypePrefix, prefixRows)
	if err != nil {
		return total, err
	}
	total += inserted
	return total, nil
}

func (s *Service) replaceTopRows(ctx context.Context, aggMode string, sc scope, aggType string, rows []*moviex.MovieReleaseTopCalc) (int64, error) {
	if err := s.deps.MovieReleaseTopStatModel.DeleteByAggModeScopeAggType(ctx, aggMode, sc.Key, aggType); err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	var inserted int64
	for idx, row := range rows {
		if row == nil || row.AggName == "" {
			continue
		}
		_, err := s.deps.MovieReleaseTopStatModel.Insert(ctx, &moviex.MovieReleaseTopStat{
			AggMode:    aggMode,
			ScopeKey:   sc.Key,
			Level:      sc.Level,
			Year:       int64(sc.Year),
			Quarter:    int64(sc.Quarter),
			Month:      int64(sc.Month),
			AggType:    aggType,
			AggKey:     row.AggKey,
			AggId:      row.AggID,
			AggName:    row.AggName,
			CountAll:   row.CountAll,
			CountOwned: row.CountOwned,
			RankNo:     int64(idx + 1),
			CreatedOn:  now,
			UpdatedOn:  now,
		})
		if err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

func (s *Service) startAggEvent(ctx context.Context, flowKey string, startedTime int64) (*moviex.WAggEvent, error) {
	row := &moviex.WAggEvent{
		AggKey:       aggKeyMovieRelease,
		FlowKey:      flowKey,
		Status:       aggEventStatusRun,
		ScopeCount:   0,
		BucketCount:  0,
		TopCount:     0,
		StartedTime:  startedTime,
		FinishedTime: 0,
		DurationMs:   0,
		ErrorMessage: "",
		CreatedTime:  startedTime,
		UpdatedTime:  startedTime,
	}
	result, err := s.deps.WAggEventModel.Insert(ctx, row)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	row.Id = id
	return row, nil
}

func (s *Service) finishAggEvent(ctx context.Context, row *moviex.WAggEvent, startedAt time.Time, summary rebuildSummary, rebuildErr error) error {
	if row == nil {
		return nil
	}
	finishedTime := time.Now().Unix()
	row.ScopeCount = summary.ScopeCount
	row.BucketCount = summary.BucketCount
	row.TopCount = summary.TopCount
	row.FinishedTime = finishedTime
	row.DurationMs = time.Since(startedAt).Milliseconds()
	row.UpdatedTime = finishedTime
	if rebuildErr != nil {
		row.Status = aggEventStatusError
		row.ErrorMessage = rebuildErr.Error()
	} else {
		row.Status = aggEventStatusOK
		row.ErrorMessage = ""
	}
	return s.deps.WAggEventModel.Update(ctx, row)
}

func sortedScopes(items map[string]scope) []scope {
	out := make([]scope, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}
