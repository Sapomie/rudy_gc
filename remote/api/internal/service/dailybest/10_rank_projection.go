package dailybest

import (
	"context"
	"time"
)

func (s *Service) EnsureRankProjection(ctx context.Context) error {
	complete, err := s.repository.RankProjectionComplete(ctx)
	if err != nil || complete {
		return err
	}
	ranks, err := s.repository.ListAllRanks(ctx)
	if err != nil {
		return err
	}
	return s.repository.SyncRankProjection(ctx, buildRankProjection(ranks, time.Now().Unix(), s.loc))
}
