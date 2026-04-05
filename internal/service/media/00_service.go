package media

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/service/wmediaagg"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps: d,
	}
}

func (s *Service) mediaRoots() []string {
	raw := s.deps.Config.Media.RootDirs
	if len(raw) == 0 {
		return nil
	}

	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, root := range raw {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		cleaned := filepath.Clean(root)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}

func (s *Service) markMediaAggDirty(ctx context.Context, rows ...*moviex.WMedia) {
	if len(rows) == 0 {
		return
	}
	if err := wmediaagg.NewService(s.deps).MarkMediaRowsDirty(ctx, rows...); err != nil {
		s.deps.Log.WithContext(ctx).Errorf("mark w_media agg dirty failed: %v", err)
	}
	releaseTimes := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.ReleasingDate <= 0 {
			continue
		}
		releaseTimes = append(releaseTimes, row.ReleasingDate)
	}
	if err := moviereleaseagg.NewService(s.deps).MarkReleaseTimesDirty(ctx, releaseTimes...); err != nil {
		s.deps.Log.WithContext(ctx).Errorf("mark movie release agg dirty failed: %v", err)
	}
}

func (s *Service) rebuildMediaAggsAfterFlow(ctx context.Context, flowKey string) error {
	var err error
	if rebuildErr := wmediaagg.NewService(s.deps).RebuildDirtyAndLogEvent(ctx, flowKey); rebuildErr != nil {
		err = joinFlowErr(err, rebuildErr)
	}
	if rebuildErr := moviereleaseagg.NewService(s.deps).RebuildDirtyAndLogEvent(ctx, flowKey); rebuildErr != nil {
		err = joinFlowErr(err, rebuildErr)
	}
	return err
}

func joinFlowErr(origin error, rebuildErr error) error {
	if rebuildErr == nil {
		return origin
	}
	if origin == nil {
		return rebuildErr
	}
	return errors.Join(origin, rebuildErr)
}
