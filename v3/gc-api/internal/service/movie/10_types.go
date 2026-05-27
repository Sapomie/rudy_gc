package movie

import (
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type RankPeriodPageRequest struct {
	PeriodType int64
	PeriodKey  string
	Category   int64
	Page       int64
	PageSize   int64
}

type RankPeriodMovieCard struct {
	Movie           *types.MovieType
	Item            *moviex.CRankPeriodItem
	PrevRankText    string
	RankChangeText  string
	RankChangeClass string
}

type RankPeriodPage struct {
	Title           string
	Period          *moviex.CRankPeriod
	PrevPeriod      *moviex.CRankPeriod
	NextPeriod      *moviex.CRankPeriod
	Cards           []*RankPeriodMovieCard
	Movies          []*types.MovieType
	Total           int64
	PeriodTypeLabel string
	CategoryLabel   string
	RangeStart      string
	RangeEnd        string
}
