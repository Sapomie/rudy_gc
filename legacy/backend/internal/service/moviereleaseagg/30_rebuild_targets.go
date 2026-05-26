package moviereleaseagg

import (
	"context"
	"errors"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) RebuildByMovieJavIDsAndLogEvent(ctx context.Context, flowKey string, javIDs ...string) error {
	normalized := make([]string, 0, len(javIDs))
	seen := make(map[string]struct{}, len(javIDs))
	for _, javID := range javIDs {
		javID = strings.TrimSpace(javID)
		if javID == "" {
			continue
		}
		if _, ok := seen[javID]; ok {
			continue
		}
		seen[javID] = struct{}{}
		normalized = append(normalized, javID)
	}
	if len(normalized) == 0 {
		return nil
	}

	rows := make([]*moviex.AMovie, 0, len(normalized))
	for _, javID := range normalized {
		row, err := s.deps.MovieModel.FindOneByJavId(ctx, javID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return err
		}
		rows = append(rows, row)
	}
	return s.RebuildByMovieRowsAndLogEvent(ctx, flowKey, rows...)
}

func (s *Service) RebuildByMovieRowsAndLogEvent(ctx context.Context, flowKey string, rows ...*moviex.AMovie) error {
	releaseTimes := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ReleasingDate <= 0 {
			continue
		}
		releaseTimes = append(releaseTimes, row.ReleasingDate)
	}
	return s.RebuildByReleaseTimesAndLogEvent(ctx, flowKey, releaseTimes...)
}

func (s *Service) RebuildByReleaseTimesAndLogEvent(ctx context.Context, flowKey string, releaseTimes ...int64) error {
	bucketMonths := make([]int64, 0, len(releaseTimes))
	seen := make(map[int64]struct{}, len(releaseTimes))
	for _, releaseTime := range releaseTimes {
		bucketMonth := bucketMonthFromReleaseTime(releaseTime)
		if bucketMonth <= 0 {
			continue
		}
		if _, ok := seen[bucketMonth]; ok {
			continue
		}
		seen[bucketMonth] = struct{}{}
		bucketMonths = append(bucketMonths, bucketMonth)
	}
	return s.rebuildByBucketMonthsAndLogEvent(ctx, flowKey, int64(len(releaseTimes)), bucketMonths)
}

func (s *Service) RebuildByReleaseMonthsAndLogEvent(ctx context.Context, flowKey string, bucketMonths ...int64) error {
	return s.rebuildByBucketMonthsAndLogEvent(ctx, flowKey, int64(len(bucketMonths)), bucketMonths)
}
