package wdir

import (
	"context"
	"time"

	"rudy_gc/internal/service/wfoldertree"
)

func (s *Service) ensureFolderTreeNormalized(ctx context.Context) error {
	s.normalizeMu.Lock()
	defer s.normalizeMu.Unlock()

	if s.normalizedAll {
		return nil
	}

	if err := wfoldertree.NormalizeAll(ctx, s.deps.WFolderModel, time.Now().Unix()); err != nil {
		return err
	}

	s.normalizedAll = true
	return nil
}
