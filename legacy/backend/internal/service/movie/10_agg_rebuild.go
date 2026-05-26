package movie

import (
	"context"
	"errors"
	"strings"
	"sync"

	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/service/wmediaagg"

	"github.com/zeromicro/go-zero/core/threading"
)

func (s *Service) RebuildAggsByMovieJavIDs(ctx context.Context, flowKey string, javIDs ...string) error {
	cleanIDs := uniqueMovieJavIDs(javIDs...)
	if len(cleanIDs) == 0 {
		return nil
	}
	return s.rebuildAggs(ctx, flowKey, cleanIDs, nil)
}

func (s *Service) rebuildReleaseAggByTimes(ctx context.Context, flowKey string, releaseTimes ...int64) error {
	return moviereleaseagg.NewService(s.deps).RebuildByReleaseTimesAndLogEvent(ctx, flowKey, releaseTimes...)
}

func (s *Service) EnqueueAggRebuildByMovieJavIDs(flowKey string, javIDs ...string) {
	s.enqueueAggRebuild(flowKey, uniqueMovieJavIDs(javIDs...), nil)
}

func (s *Service) EnqueueAggRebuild(flowKey string, javIDs []string, releaseTimes []int64) {
	s.enqueueAggRebuild(flowKey, uniqueMovieJavIDs(javIDs...), uniqueReleaseTimes(releaseTimes...))
}

func (s *Service) enqueueAggRebuild(flowKey string, javIDs []string, releaseTimes []int64) {
	if len(javIDs) == 0 && len(releaseTimes) == 0 {
		return
	}
	threading.GoSafe(func() {
		if err := s.rebuildAggs(context.Background(), flowKey, javIDs, releaseTimes); err != nil {
			s.deps.Log.Errorf("async agg rebuild failed, flow=%s, movies=%d, releases=%d, err=%v", flowKey, len(javIDs), len(releaseTimes), err)
		}
	})
}

func (s *Service) rebuildAggs(ctx context.Context, flowKey string, javIDs []string, releaseTimes []int64) error {
	cleanIDs := uniqueMovieJavIDs(javIDs...)
	cleanReleaseTimes := uniqueReleaseTimes(releaseTimes...)
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	appendErr := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		errs = append(errs, err)
		mu.Unlock()
	}

	if len(cleanIDs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			appendErr(wmediaagg.NewService(s.deps).RebuildByMovieJavIDsAndLogEvent(ctx, flowKey, cleanIDs...))
		}()
	}

	if len(cleanReleaseTimes) > 0 || len(cleanIDs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			if len(cleanReleaseTimes) > 0 {
				err = moviereleaseagg.NewService(s.deps).RebuildByReleaseTimesAndLogEvent(ctx, flowKey, cleanReleaseTimes...)
			} else {
				err = moviereleaseagg.NewService(s.deps).RebuildByMovieJavIDsAndLogEvent(ctx, flowKey, cleanIDs...)
			}
			appendErr(err)
		}()
	}

	wg.Wait()
	return errors.Join(errs...)
}

func uniqueMovieJavIDs(javIDs ...string) []string {
	if len(javIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(javIDs))
	out := make([]string, 0, len(javIDs))
	for _, javID := range javIDs {
		javID = strings.TrimSpace(javID)
		if javID == "" {
			continue
		}
		if _, ok := seen[javID]; ok {
			continue
		}
		seen[javID] = struct{}{}
		out = append(out, javID)
	}
	return out
}

func uniqueReleaseTimes(releaseTimes ...int64) []int64 {
	if len(releaseTimes) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(releaseTimes))
	out := make([]int64, 0, len(releaseTimes))
	for _, releaseTime := range releaseTimes {
		if releaseTime <= 0 {
			continue
		}
		if _, ok := seen[releaseTime]; ok {
			continue
		}
		seen[releaseTime] = struct{}{}
		out = append(out, releaseTime)
	}
	return out
}
