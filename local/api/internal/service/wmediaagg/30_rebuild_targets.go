package wmediaagg

import (
	"context"
	"strings"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func (s *Service) RebuildByMovieJavIDsAndLogEvent(ctx context.Context, flowKey string, javIDs ...string) error {
	return s.RebuildByMovieJavIDsAndBucketDaysAndLogEvent(ctx, flowKey, nil, javIDs...)
}

func (s *Service) RebuildByMovieJavIDsAndBucketDaysAndLogEvent(ctx context.Context, flowKey string, extraBucketDays []int64, javIDs ...string) error {
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
	if len(normalized) == 0 && len(extraBucketDays) == 0 {
		return nil
	}

	var rows []*moviex.WMedia
	if len(normalized) > 0 {
		var err error
		rows, err = s.deps.WMediaModel.ListByMovieJavIds(ctx, normalized)
		if err != nil {
			return err
		}
	}
	bucketDays := appendUniqueBucketDays(extraBucketDays, BucketDaysFromMediaRows(rows...)...)
	return s.rebuildByBucketDaysAndLogEvent(ctx, flowKey, int64(len(normalized)), bucketDays)
}

func (s *Service) RebuildByMediaRowsAndLogEvent(ctx context.Context, flowKey string, rows ...*moviex.WMedia) error {
	return s.rebuildByBucketDaysAndLogEvent(ctx, flowKey, int64(len(rows)), BucketDaysFromMediaRows(rows...))
}

func BucketDaysFromMediaRows(rows ...*moviex.WMedia) []int64 {
	bucketDays := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		if row == nil {
			continue
		}
		bucketDays = appendUniqueBucketDays(bucketDays, bucketDayFromBirthTime(row.BirthTime), bucketDayFromBirthTime(row.RemoveTime))
	}
	return bucketDays
}

func appendUniqueBucketDays(current []int64, candidates ...int64) []int64 {
	out := make([]int64, 0, len(current)+len(candidates))
	seen := make(map[int64]struct{}, len(current)+len(candidates))
	appendDay := func(bucketDay int64) {
		if bucketDay <= 0 {
			return
		}
		if _, ok := seen[bucketDay]; ok {
			return
		}
		seen[bucketDay] = struct{}{}
		out = append(out, bucketDay)
	}
	for _, bucketDay := range current {
		appendDay(bucketDay)
	}
	for _, bucketDay := range candidates {
		appendDay(bucketDay)
	}
	return out
}

func (s *Service) RebuildByBucketDaysAndLogEvent(ctx context.Context, flowKey string, bucketDays ...int64) error {
	return s.rebuildByBucketDaysAndLogEvent(ctx, flowKey, int64(len(bucketDays)), bucketDays)
}
