package page

import (
	"context"

	"rudy-gc-api/internal/model/modelx"
	"rudy-gc-api/internal/types"
)

func (s *Service) Summaries() []*types.PageSummary {
	return modelx.PageSummaries()
}

func (s *Service) Load(ctx context.Context, key string, req *types.PageListRequest) (*types.PageListResponse, error) {
	return s.repo.Load(ctx, key, req)
}
