package wmediaagg

import (
	"context"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BackfillAll(ctx context.Context) (*BackfillResult, error) {
	start := time.Now()
	result := &BackfillResult{}

	bucketRows, err := s.deps.WMediaBirthBucketStatModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result.ClearedBucketRows = len(bucketRows)
	for _, row := range bucketRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := s.deps.WMediaBirthBucketStatModel.Delete(ctx, row.Id); err != nil {
			return nil, err
		}
	}

	topRows, err := s.deps.WMediaBirthTopStatModel.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result.ClearedTopRows = len(topRows)
	for _, row := range topRows {
		if row == nil || row.Id <= 0 {
			continue
		}
		if err := s.deps.WMediaBirthTopStatModel.Delete(ctx, row.Id); err != nil {
			return nil, err
		}
	}

	days, err := s.deps.WMediaModel.ListDistinctBirthDays(ctx)
	if err != nil {
		return nil, err
	}
	result.DirtyDays = len(days)

	if len(days) > 0 {
		if err := s.RebuildByBirthDaysAndLogEvent(ctx, "backfill", days...); err != nil {
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
	rows, err := s.deps.WMediaBirthBucketStatModel.ListByLevel(ctx, level, 0, 0, 0, false)
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

func (s *Service) countAllTopRows(ctx context.Context) (int, error) {
	total := 0
	scopeKeys := []string{levelRoot}

	appendScopeKeys := func(rows []*moviex.WMediaBirthBucketStat) {
		for _, row := range rows {
			if row == nil || row.ScopeKey == "" {
				continue
			}
			scopeKeys = append(scopeKeys, row.ScopeKey)
		}
	}

	yearRows, err := s.deps.WMediaBirthBucketStatModel.ListByLevel(ctx, levelYear, 0, 0, 0, false)
	if err != nil {
		return 0, err
	}
	appendScopeKeys(yearRows)

	quarterRows, err := s.deps.WMediaBirthBucketStatModel.ListByLevel(ctx, levelQuarter, 0, 0, 0, false)
	if err != nil {
		return 0, err
	}
	appendScopeKeys(quarterRows)

	monthRows, err := s.deps.WMediaBirthBucketStatModel.ListByLevel(ctx, levelMonth, 0, 0, 0, false)
	if err != nil {
		return 0, err
	}
	appendScopeKeys(monthRows)

	seen := make(map[string]struct{}, len(scopeKeys))
	for _, key := range scopeKeys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		for _, aggType := range []string{aggTypeCast, aggTypeDirector, aggTypeLabel, aggTypePrefix} {
			rows, err := s.deps.WMediaBirthTopStatModel.ListByScopeAggType(ctx, key, aggType)
			if err != nil {
				return 0, err
			}
			total += len(rows)
		}
	}
	return total, nil
}
