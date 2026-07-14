package shthreadagg

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/repository/moviex"
)

const (
	aggKeySHThread    = "sh_thread"
	aggEventStatusRun = "running"
	aggEventStatusOK  = "success"
	aggEventStatusErr = "failed"
)

type rebuildSummary struct {
	ScopeCount  int64
	BucketCount int64
}

type bucketWork struct {
	AggScope      string
	SourceForumID string
	Scope         bucketScope
}

func (s *Service) RebuildByTargetsAndLogEvent(ctx context.Context, flowKey string, targets ...moviex.ShThreadAggTarget) error {
	works := buildBucketWorks(targets)
	if len(works) == 0 {
		return nil
	}
	if strings.TrimSpace(flowKey) == "" {
		flowKey = "manual"
	}
	startedAt := time.Now()
	eventRow, err := s.startAggEvent(ctx, flowKey, startedAt.Unix())
	if err != nil {
		return err
	}
	summary := rebuildSummary{ScopeCount: int64(len(targets))}
	for _, work := range works {
		if err := s.rebuildBucket(ctx, work); err != nil {
			if finishErr := s.finishAggEvent(context.Background(), eventRow, startedAt, summary, err); finishErr != nil {
				return errors.Join(err, finishErr)
			}
			return err
		}
		summary.BucketCount++
	}
	return s.finishAggEvent(ctx, eventRow, startedAt, summary, nil)
}

func buildBucketWorks(targets []moviex.ShThreadAggTarget) []bucketWork {
	items := make(map[string]bucketWork)
	for _, target := range targets {
		if target.PostTime <= 0 {
			continue
		}
		forumID := strings.TrimSpace(target.SourceForumID)
		for _, scope := range buildScopes(target.PostTime) {
			allKey := AggScopeAll + "|" + scope.Key
			items[allKey] = bucketWork{AggScope: AggScopeAll, Scope: scope}
			if forumID != "" {
				aggScope := "forum:" + forumID
				forumKey := aggScope + "|" + scope.Key
				items[forumKey] = bucketWork{AggScope: aggScope, SourceForumID: forumID, Scope: scope}
			}
		}
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]bucketWork, 0, len(keys))
	for _, key := range keys {
		out = append(out, items[key])
	}
	return out
}

func (s *Service) rebuildBucket(ctx context.Context, work bucketWork) error {
	calc, err := s.deps.ShSehuatangThreadModel.CalcBucket(ctx, work.Scope.StartUnix, work.Scope.EndUnix, work.SourceForumID)
	if err != nil {
		return err
	}
	if calc == nil || calc.CountAll == 0 {
		return s.deps.ShThreadBucketStatModel.DeleteByAggScopeScopeKey(ctx, work.AggScope, work.Scope.Key)
	}
	now := time.Now().Unix()
	return s.deps.ShThreadBucketStatModel.Upsert(ctx, &moviex.ShThreadBucketStat{
		AggScope:          work.AggScope,
		ScopeKey:          work.Scope.Key,
		Level:             work.Scope.Level,
		Year:              int64(work.Scope.Year),
		Quarter:           int64(work.Scope.Quarter),
		Month:             int64(work.Scope.Month),
		Week:              int64(work.Scope.Week),
		Day:               int64(work.Scope.Day),
		BucketStart:       work.Scope.StartUnix,
		BucketEnd:         work.Scope.EndUnix,
		CountAll:          calc.CountAll,
		CountPending:      calc.CountPending,
		CountFetched:      calc.CountFetched,
		CountFailed:       calc.CountFailed,
		CountWithMagnet:   calc.CountWithMagnet,
		CountWithInfoHash: calc.CountWithInfoHash,
		LatestPostTime:    calc.LatestPostTime,
		CreatedOn:         now,
		UpdatedOn:         now,
	})
}

func (s *Service) startAggEvent(ctx context.Context, flowKey string, startedTime int64) (*moviex.WAggEvent, error) {
	row := &moviex.WAggEvent{
		AggKey:      aggKeySHThread,
		FlowKey:     flowKey,
		Status:      aggEventStatusRun,
		StartedTime: startedTime,
		CreatedTime: startedTime,
		UpdatedTime: startedTime,
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
	row.TopCount = 0
	row.FinishedTime = finishedTime
	row.DurationMs = time.Since(startedAt).Milliseconds()
	row.UpdatedTime = finishedTime
	if rebuildErr != nil {
		row.Status = aggEventStatusErr
		row.ErrorMessage = rebuildErr.Error()
	} else {
		row.Status = aggEventStatusOK
		row.ErrorMessage = ""
	}
	return s.deps.WAggEventModel.Update(ctx, row)
}
