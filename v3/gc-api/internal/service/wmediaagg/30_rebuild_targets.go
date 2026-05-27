package wmediaagg

import (
	"context"
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

	rows, err := s.deps.WMediaModel.ListByMovieJavIds(ctx, normalized)
	if err != nil {
		return err
	}
	return s.RebuildByMediaRowsAndLogEvent(ctx, flowKey, rows...)
}

func (s *Service) RebuildByMediaRowsAndLogEvent(ctx context.Context, flowKey string, rows ...*moviex.WMedia) error {
	bucketDays := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || row.BirthTime <= 0 {
			continue
		}
		bucketDay := bucketDayFromBirthTime(row.BirthTime)
		if bucketDay <= 0 {
			continue
		}
		if _, ok := seen[bucketDay]; ok {
			continue
		}
		seen[bucketDay] = struct{}{}
		bucketDays = append(bucketDays, bucketDay)
	}
	return s.rebuildByBucketDaysAndLogEvent(ctx, flowKey, int64(len(rows)), bucketDays)
}

func (s *Service) RebuildByBirthDaysAndLogEvent(ctx context.Context, flowKey string, bucketDays ...int64) error {
	return s.rebuildByBucketDaysAndLogEvent(ctx, flowKey, int64(len(bucketDays)), bucketDays)
}
