package moviereleaseagg

import (
	"context"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BackfillAll(ctx context.Context) (*BackfillResult, error) {
	start := time.Now()
	result := &BackfillResult{}

	bucketRows, err := s.deps.MovieReleaseBucketStatModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result.ClearedBucketRows = len(bucketRows)
	for _, row := range bucketRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := s.deps.MovieReleaseBucketStatModel.Delete(ctx, row.Id); err != nil {
			return nil, err
		}
	}

	topRows, err := s.deps.MovieReleaseTopStatModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result.ClearedTopRows = len(topRows)
	for _, row := range topRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := s.deps.MovieReleaseTopStatModel.Delete(ctx, row.Id); err != nil {
			return nil, err
		}
	}

	dirtyRows, err := s.deps.MovieReleaseAggDirtyModel.ListAll(ctx, 0)
	if err != nil {
		return nil, err
	}
	result.ClearedDirtyRows = len(dirtyRows)
	for _, row := range dirtyRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := s.deps.MovieReleaseAggDirtyModel.Delete(ctx, row.Id); err != nil {
			return nil, err
		}
	}

	months, err := s.deps.MovieModel.ListDistinctReleaseMonths(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	result.DirtyMonths = len(months)
	for _, bucketMonth := range months {
		sc := scopeFromBucketMonth(bucketMonth)
		if err := s.deps.MovieReleaseAggDirtyModel.TouchMonth(ctx, bucketMonth, sc.Key, now); err != nil {
			return nil, err
		}
	}

	if len(months) > 0 {
		if err := s.RebuildDirtyAndLogEvent(ctx, "backfill"); err != nil {
			return nil, err
		}
	}

	if result.YearBuckets, err = s.countBucketLevel(ctx, levelYear); err != nil {
		return nil, err
	}
	if result.QuarterBuckets, err = s.countBucketLevel(ctx, levelQuarter); err != nil {
		return nil, err
	}
	if result.MonthBuckets, err = s.countBucketLevel(ctx, levelMonth); err != nil {
		return nil, err
	}
	if result.DayBuckets, err = s.countBucketLevel(ctx, levelDay); err != nil {
		return nil, err
	}
	if result.TopRows, err = s.countAllTopRows(ctx); err != nil {
		return nil, err
	}

	result.ElapsedMs = time.Since(start).Milliseconds()
	return result, nil
}

func (s *Service) countBucketLevel(ctx context.Context, level string) (int, error) {
	total := 0
	for _, aggMode := range []string{AggModeAll, AggModeOwned} {
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, aggMode, level, 0, 0, 0, false)
		if err != nil {
			return 0, err
		}
		total += len(rows)
	}
	return total, nil
}

func (s *Service) countAllTopRows(ctx context.Context) (int, error) {
	total := 0
	scopeKeysByMode := map[string][]string{
		AggModeAll:   {levelRoot},
		AggModeOwned: {levelRoot},
	}

	appendScopeKeys := func(rows []*moviex.MovieReleaseBucketStat) {
		for _, row := range rows {
			if row == nil || row.ScopeKey == "" {
				continue
			}
			scopeKeysByMode[row.AggMode] = append(scopeKeysByMode[row.AggMode], row.ScopeKey)
		}
	}

	for _, aggMode := range []string{AggModeAll, AggModeOwned} {
		yearRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, aggMode, levelYear, 0, 0, 0, false)
		if err != nil {
			return 0, err
		}
		appendScopeKeys(yearRows)

		quarterRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, aggMode, levelQuarter, 0, 0, 0, false)
		if err != nil {
			return 0, err
		}
		appendScopeKeys(quarterRows)

		monthRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, aggMode, levelMonth, 0, 0, 0, false)
		if err != nil {
			return 0, err
		}
		appendScopeKeys(monthRows)
	}

	for _, aggMode := range []string{AggModeAll, AggModeOwned} {
		seen := make(map[string]struct{}, len(scopeKeysByMode[aggMode]))
		for _, key := range scopeKeysByMode[aggMode] {
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			for _, aggType := range []string{aggTypeCast, aggTypeDirector, aggTypeLabel, aggTypePrefix} {
				rows, err := s.deps.MovieReleaseTopStatModel.ListByAggModeScopeAggType(ctx, aggMode, key, aggType)
				if err != nil {
					return 0, err
				}
				total += len(rows)
			}
		}
	}
	return total, nil
}
