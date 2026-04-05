package movie

import (
	"context"
	"errors"
	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/moviereleaseagg"
	"rudy_gc/internal/service/wmediaagg"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}

func (s *Service) rebuildAggsAfterFlow(ctx context.Context, flowKey string) error {
	var err error
	if rebuildErr := wmediaagg.NewService(s.deps).RebuildDirtyAndLogEvent(ctx, flowKey); rebuildErr != nil {
		err = errors.Join(err, rebuildErr)
	}
	if rebuildErr := moviereleaseagg.NewService(s.deps).RebuildDirtyAndLogEvent(ctx, flowKey); rebuildErr != nil {
		err = errors.Join(err, rebuildErr)
	}
	return err
}
