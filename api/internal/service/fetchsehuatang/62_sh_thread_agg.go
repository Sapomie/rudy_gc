package fetchsehuatang

import (
	"context"
	"fmt"
	"strings"

	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/shthreadagg"
)

func markSHThreadAggTarget(result *SHFetchResult, target moviex.ShThreadAggTarget) {
	if result == nil || target.PostTime <= 0 {
		return
	}
	if result.aggTargets == nil {
		result.aggTargets = make(map[string]moviex.ShThreadAggTarget)
	}
	target.SourceForumID = strings.TrimSpace(target.SourceForumID)
	key := fmt.Sprintf("%d|%s", target.PostTime, target.SourceForumID)
	result.aggTargets[key] = target
}

func markSHThreadAggRow(result *SHFetchResult, row *moviex.ShSehuatangThread) {
	if row == nil {
		return
	}
	markSHThreadAggTarget(result, moviex.ShThreadAggTarget{
		ThreadTID:     row.ThreadTID,
		PostTime:      row.PostTime,
		SourceForumID: row.SourceForumID,
	})
}

func (s *Service) flushSHThreadAgg(ctx context.Context, flowKey string, result *SHFetchResult) error {
	if result == nil || len(result.aggTargets) == 0 {
		return nil
	}
	targets := make([]moviex.ShThreadAggTarget, 0, len(result.aggTargets))
	for _, target := range result.aggTargets {
		targets = append(targets, target)
	}
	result.aggTargets = nil
	return shthreadagg.NewService(s.deps).RebuildByTargetsAndLogEvent(ctx, flowKey, targets...)
}
