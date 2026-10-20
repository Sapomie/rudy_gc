package movie

import (
	"context"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type RankDayStatListRequest struct {
	StartDayNumber    int64
	EndDayNumber      int64
	HasStartDayNumber bool
	HasEndDayNumber   bool
	Page              int64
	PageSize          int64
	Sort              string
	Order             string
}

type RankPeriodStatListRequest struct {
	PeriodType        int64
	StartDayNumber    int64
	EndDayNumber      int64
	HasStartDayNumber bool
	HasEndDayNumber   bool
	Page              int64
	PageSize          int64
	Sort              string
	Order             string
}

func (s *Service) ListRankDayStats(ctx context.Context, req RankDayStatListRequest) (*moviex.CRankDayStatPage, error) {
	return s.deps.RankDayStatModel.ListPage(ctx, moviex.CRankDayStatListQuery{
		StartDayNumber:    req.StartDayNumber,
		EndDayNumber:      req.EndDayNumber,
		HasStartDayNumber: req.HasStartDayNumber,
		HasEndDayNumber:   req.HasEndDayNumber,
		Page:              req.Page,
		PageSize:          req.PageSize,
		Sort:              req.Sort,
		Order:             req.Order,
	})
}

func (s *Service) ListRankPeriodStats(ctx context.Context, req RankPeriodStatListRequest) (*moviex.CRankPeriodStatPage, error) {
	return s.deps.RankPeriodStatModel.ListPage(ctx, moviex.CRankPeriodStatListQuery{
		PeriodType:        req.PeriodType,
		StartDayNumber:    req.StartDayNumber,
		EndDayNumber:      req.EndDayNumber,
		HasStartDayNumber: req.HasStartDayNumber,
		HasEndDayNumber:   req.HasEndDayNumber,
		Page:              req.Page,
		PageSize:          req.PageSize,
		Sort:              req.Sort,
		Order:             req.Order,
	})
}
