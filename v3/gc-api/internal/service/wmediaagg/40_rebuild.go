package wmediaagg

import (
	"context"
	"errors"
	"sort"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	aggKeyWMediaBirth    = "w_media_birth"
	aggEventStatusRun    = "running"
	aggEventStatusOK     = "success"
	aggEventStatusFailed = "failed"
)

type rebuildSummary struct {
	ScopeCount  int64
	BucketCount int64
	TopCount    int64
}

func (s *Service) rebuildByBucketDaysAndLogEvent(ctx context.Context, flowKey string, scopeCount int64, bucketDays []int64) error {
	if len(bucketDays) == 0 {
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

	summary, rebuildErr := s.rebuildByBucketDays(ctx, bucketDays, scopeCount)
	if finishErr := s.finishAggEvent(ctx, eventRow, startedAt, summary, rebuildErr); finishErr != nil {
		if rebuildErr != nil {
			return errors.Join(rebuildErr, finishErr)
		}
		return finishErr
	}
	return rebuildErr
}

func (s *Service) rebuildByBucketDays(ctx context.Context, bucketDays []int64, scopeCount int64) (rebuildSummary, error) {
	if scopeCount <= 0 {
		scopeCount = int64(len(bucketDays))
	}
	summary := rebuildSummary{ScopeCount: scopeCount}
	dayScopes := make(map[string]scope)
	monthScopes := make(map[string]scope)
	quarterScopes := make(map[string]scope)
	yearScopes := make(map[string]scope)

	for _, bucketDay := range bucketDays {
		if bucketDay <= 0 {
			continue
		}
		dayScope := scopeFromBucketDay(bucketDay)
		dayScopes[dayScope.Key] = dayScope
		monthScope := buildScope(dayScope.Year, 0, dayScope.Month, 0)
		quarterScope := buildScope(dayScope.Year, dayScope.Quarter, 0, 0)
		yearScope := buildScope(dayScope.Year, 0, 0, 0)
		monthScopes[monthScope.Key] = monthScope
		quarterScopes[quarterScope.Key] = quarterScope
		yearScopes[yearScope.Key] = yearScope
	}

	for _, sc := range sortedScopes(dayScopes) {
		if err := s.rebuildBucketScope(ctx, sc); err != nil {
			return summary, err
		}
		summary.BucketCount++
	}
	for _, sc := range sortedScopes(monthScopes) {
		if err := s.rebuildBucketScope(ctx, sc); err != nil {
			return summary, err
		}
		summary.BucketCount++
		topCount, err := s.rebuildTopScope(ctx, sc)
		if err != nil {
			return summary, err
		}
		summary.TopCount += topCount
	}
	for _, sc := range sortedScopes(quarterScopes) {
		if err := s.rebuildBucketScope(ctx, sc); err != nil {
			return summary, err
		}
		summary.BucketCount++
		topCount, err := s.rebuildTopScope(ctx, sc)
		if err != nil {
			return summary, err
		}
		summary.TopCount += topCount
	}
	for _, sc := range sortedScopes(yearScopes) {
		if err := s.rebuildBucketScope(ctx, sc); err != nil {
			return summary, err
		}
		summary.BucketCount++
		topCount, err := s.rebuildTopScope(ctx, sc)
		if err != nil {
			return summary, err
		}
		summary.TopCount += topCount
	}
	topCount, err := s.rebuildTopScope(ctx, buildScope(0, 0, 0, 0))
	if err != nil {
		return summary, err
	}
	summary.TopCount += topCount
	return summary, nil
}

func (s *Service) rebuildBucketScope(ctx context.Context, sc scope) error {
	if sc.Level == levelRoot {
		return nil
	}

	calc, err := s.deps.WMediaModel.CalcBirthBucket(ctx, sc.StartUnix, sc.EndUnix)
	if err != nil {
		return err
	}

	existing, err := s.deps.WMediaBirthBucketStatModel.FindOneByScopeKey(ctx, sc.Key)
	if err != nil && err != moviex.ErrNotFound {
		return err
	}
	if calc == nil {
		calc = &moviex.WMediaBirthBucketCalc{}
	}

	if calc.MediaCount == 0 && calc.RemovedCount == 0 {
		if existing != nil {
			return s.deps.WMediaBirthBucketStatModel.Delete(ctx, existing.Id)
		}
		return nil
	}

	now := time.Now().Unix()
	if existing == nil {
		_, err = s.deps.WMediaBirthBucketStatModel.Insert(ctx, &moviex.WMediaBirthBucketStat{
			ScopeKey:        sc.Key,
			Level:           sc.Level,
			Year:            int64(sc.Year),
			Quarter:         int64(sc.Quarter),
			Month:           int64(sc.Month),
			Day:             int64(sc.Day),
			MediaCount:      calc.MediaCount,
			RemovedCount:    calc.RemovedCount,
			SizeBytes:       calc.SizeBytes,
			HasSubCount:     calc.HasSubCount,
			LatestBirthTime: calc.LatestBirthTime,
			CreatedOn:       now,
			UpdatedOn:       now,
		})
		return err
	}

	existing.Level = sc.Level
	existing.Year = int64(sc.Year)
	existing.Quarter = int64(sc.Quarter)
	existing.Month = int64(sc.Month)
	existing.Day = int64(sc.Day)
	existing.MediaCount = calc.MediaCount
	existing.RemovedCount = calc.RemovedCount
	existing.SizeBytes = calc.SizeBytes
	existing.HasSubCount = calc.HasSubCount
	existing.LatestBirthTime = calc.LatestBirthTime
	existing.UpdatedOn = now
	return s.deps.WMediaBirthBucketStatModel.Update(ctx, existing)
}

func (s *Service) rebuildTopScope(ctx context.Context, sc scope) (int64, error) {
	var total int64
	castRows, err := s.deps.WMediaModel.CalcTopCastsByBirthRange(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit)
	if err != nil {
		return total, err
	}
	inserted, err := s.replaceTopRows(ctx, sc, aggTypeCast, castRows)
	if err != nil {
		return total, err
	}
	total += inserted

	directorRows, err := s.deps.WMediaModel.CalcTopDirectorsByBirthRange(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, sc, aggTypeDirector, directorRows)
	if err != nil {
		return total, err
	}
	total += inserted

	labelRows, err := s.deps.WMediaModel.CalcTopLabelsByBirthRange(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, sc, aggTypeLabel, labelRows)
	if err != nil {
		return total, err
	}
	total += inserted

	prefixRows, err := s.deps.WMediaModel.CalcTopPrefixesByBirthRange(ctx, sc.StartUnix, sc.EndUnix, topPersistLimit)
	if err != nil {
		return total, err
	}
	inserted, err = s.replaceTopRows(ctx, sc, aggTypePrefix, prefixRows)
	if err != nil {
		return total, err
	}
	total += inserted
	return total, nil
}

func (s *Service) replaceTopRows(ctx context.Context, sc scope, aggType string, rows []*moviex.WMediaBirthTopCalc) (int64, error) {
	if err := s.deps.WMediaBirthTopStatModel.DeleteByScopeAggType(ctx, sc.Key, aggType); err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	var inserted int64
	for idx, row := range rows {
		if row == nil || row.AggName == "" {
			continue
		}
		_, err := s.deps.WMediaBirthTopStatModel.Insert(ctx, &moviex.WMediaBirthTopStat{
			ScopeKey:   sc.Key,
			Level:      sc.Level,
			Year:       int64(sc.Year),
			Quarter:    int64(sc.Quarter),
			Month:      int64(sc.Month),
			Day:        int64(sc.Day),
			AggType:    aggType,
			AggKey:     row.AggKey,
			AggId:      row.AggId,
			AggName:    row.AggName,
			MediaCount: row.MediaCount,
			SizeBytes:  row.SizeBytes,
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
		AggKey:       aggKeyWMediaBirth,
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
		row.Status = aggEventStatusFailed
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
