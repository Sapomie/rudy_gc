package shthreadagg

import (
	"context"
	"time"
)

func (s *Service) BackfillAll(ctx context.Context) (*BackfillResult, error) {
	startedAt := time.Now()
	result := &BackfillResult{}

	cleared, err := s.deps.ShThreadBucketStatModel.DeleteAll(ctx)
	if err != nil {
		return nil, err
	}
	result.ClearedBucketRows = cleared

	targets, err := s.deps.ShSehuatangThreadModel.ListDistinctPostDayForumTargets(ctx)
	if err != nil {
		return nil, err
	}
	days := make(map[int64]struct{}, len(targets))
	for _, target := range targets {
		if target.PostTime > 0 {
			days[target.PostTime] = struct{}{}
		}
	}
	result.DirtyDays = int64(len(days))
	if err := s.RebuildByTargetsAndLogEvent(ctx, "backfill", targets...); err != nil {
		return nil, err
	}

	if result.YearBuckets, err = s.deps.ShThreadBucketStatModel.CountByLevel(ctx, levelYear); err != nil {
		return nil, err
	}
	if result.QuarterBuckets, err = s.deps.ShThreadBucketStatModel.CountByLevel(ctx, levelQuarter); err != nil {
		return nil, err
	}
	if result.MonthBuckets, err = s.deps.ShThreadBucketStatModel.CountByLevel(ctx, levelMonth); err != nil {
		return nil, err
	}
	if result.WeekBuckets, err = s.deps.ShThreadBucketStatModel.CountByLevel(ctx, levelWeek); err != nil {
		return nil, err
	}
	if result.DayBuckets, err = s.deps.ShThreadBucketStatModel.CountByLevel(ctx, levelDay); err != nil {
		return nil, err
	}
	result.ElapsedMs = time.Since(startedAt).Milliseconds()
	return result, nil
}
