package movie

import (
	"context"
	"rudy_gc/internal/types"
)

func (s *MovieService) ListRecords(ctx context.Context, startFrom int64, typ string, limit int) ([]*types.Record, error) {
	return s.deps.RecordRepo.Find(ctx, startFrom, typ, limit)
}
