package movie

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

var ErrCardPeriodParams = errors.New("周期卡片参数错误")

// Period pagination and movie pagination have independent query parameters.
type CardPeriodParams struct {
	Basis          string `form:"basis"`
	PeriodType     string `form:"period_type"`
	Key            string `form:"key"`
	PeriodPage     int64  `form:"period_p"`
	PeriodPageSize int64  `form:"period_ps"`
	PeriodSort     string `form:"period_sort"`
	PeriodOrder    string `form:"period_order"`
	Page           int64  `form:"p"`
	PageSize       int64  `form:"ps"`
	OrderBy        string `form:"od"`
	Order          string `form:"order"`
	MediaOwned     int64  `form:"mowned"`
}

type CardPeriodRow struct {
	Key        string
	Label      string
	RangeStart string
	RangeEnd   string
	Count      int64
	CountOwned int64
}

type CardPeriodResult struct {
	Params         CardPeriodParams
	Periods        []CardPeriodRow
	PeriodTotal    int64
	SelectedPeriod *CardPeriodRow
	Movies         []*types.MovieType
	Total          int64
	OrderBy        string
}

func (p CardPeriodParams) Validate() error {
	bad := func(message string) error { return fmt.Errorf("%w: %s", ErrCardPeriodParams, message) }
	if p.Basis != "release" && p.Basis != "birth" {
		return bad("无效时间依据")
	}
	if _, err := types.CalendarPeriodAt(p.PeriodType, time.Now()); err != nil {
		return bad(err.Error())
	}
	if p.Key != "" {
		if _, err := types.ParseCalendarPeriod(p.PeriodType, p.Key, time.Local); err != nil {
			return bad(err.Error())
		}
	}
	if p.PeriodSort != "time" && p.PeriodSort != "count" && (p.PeriodSort != "owned" || p.Basis != "release") {
		return bad("无效周期排序")
	}
	if (p.PeriodOrder != "asc" && p.PeriodOrder != "desc") || (p.Order != "asc" && p.Order != "desc") {
		return bad("无效排序方向")
	}
	if p.Basis == "release" {
		switch p.MediaOwned {
		case 0, consts.MovieAll, consts.OwnedAllNotRemoved, consts.OwnedNotOwned:
		default:
			return bad("无效拥有筛选")
		}
	}
	for _, pair := range [][2]int64{{p.PeriodPage, p.PeriodPageSize}, {p.Page, p.PageSize}} {
		if pair[0] <= 0 || pair[1] <= 0 || pair[0]-1 > math.MaxInt64/pair[1] {
			return bad("无效页码或每页条数")
		}
	}
	return nil
}

func (s *Service) BuildCardPeriod(ctx context.Context, params CardPeriodParams) (*CardPeriodResult, error) {
	if params.Basis == "release" && params.MediaOwned == 0 {
		params.MediaOwned = consts.MovieAll
	}
	if params.OrderBy == "" {
		params.OrderBy = consts.OrderByReleasingDate
		if params.Basis == "birth" {
			params.OrderBy = consts.OrderByBirthTime
		}
	}
	if params.Order == "" {
		params.Order = "desc"
		if params.OrderBy == consts.OrderByHighestRank || params.OrderBy == consts.OrderByCastAgeAsc {
			params.Order = "asc"
		}
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	rows, periodTotal, err := s.cardPeriodDirectory(ctx, params)
	if err != nil {
		return nil, err
	}
	if page := cardPeriodPage(params.PeriodPage, periodTotal, params.PeriodPageSize); page != params.PeriodPage {
		params.PeriodPage = page
		rows, periodTotal, err = s.cardPeriodDirectory(ctx, params)
		if err != nil {
			return nil, err
		}
	}
	result := &CardPeriodResult{Params: params, Periods: rows, PeriodTotal: periodTotal, Movies: []*types.MovieType{}, OrderBy: params.OrderBy}
	if params.Key == "" && len(rows) > 0 {
		params.Key = rows[0].Key
	}
	if params.Key == "" {
		result.Params.Page = 1
		return result, nil
	}
	period, err := types.ParseCalendarPeriod(params.PeriodType, params.Key, time.Local)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCardPeriodParams, err)
	}
	selected := cardPeriodRow(period, 0)
	for _, row := range rows {
		if row.Key == params.Key {
			selected = row
			break
		}
	}
	result.SelectedPeriod = &selected
	request := cardPeriodMovieRequest(params, period)
	movies, err := s.listCardPeriodMovies(ctx, &request)
	if err != nil {
		return nil, err
	}
	if page := cardPeriodPage(params.Page, movies.Total, params.PageSize); page != params.Page {
		params.Page, request.Page = page, page
		movies, err = s.listCardPeriodMovies(ctx, &request)
		if err != nil {
			return nil, err
		}
	}
	result.Params, result.Movies, result.Total = params, movies.List, movies.Total
	return result, nil
}

func cardPeriodMovieRequest(params CardPeriodParams, period types.CalendarPeriod) types.ListMovieFullRequest {
	request := types.ListMovieFullRequest{Page: params.Page, PageSize: params.PageSize, Order: params.Order, OrderBy: params.OrderBy}
	start, end := period.Start.Format(time.DateOnly), period.End.Format(time.DateOnly)
	if params.Basis == "birth" {
		request.MediaBirthTimeStart, request.MediaBirthTimeEnd = start, end
		request.MediaOwned = consts.OwnedAllNotRemoved
	} else {
		request.ReleasingDateStart, request.ReleasingDateEnd = start, end
		request.MediaOwned = params.MediaOwned
	}
	return request
}

func (s *Service) listCardPeriodMovies(ctx context.Context, request *types.ListMovieFullRequest) (*types.ListMovieResponse, error) {
	ids, total, err := s.movieListQuery().ListCardPeriod(ctx, request)
	if err != nil {
		return nil, err
	}
	rows, err := s.getMovieTypesByJavIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return &types.ListMovieResponse{List: rows, Total: total, JavIds: ids}, nil
}

func (s *Service) cardPeriodDirectory(ctx context.Context, params CardPeriodParams) ([]CardPeriodRow, int64, error) {
	rows := make([]CardPeriodRow, 0)
	sort := "scope"
	if params.Basis == "birth" {
		if params.PeriodSort == "count" {
			sort = "media"
		}
		buckets, total, err := s.deps.WMediaBirthBucketStatModel.ListPage(ctx, moviex.WMediaBirthBucketStatListFilter{
			Level: params.PeriodType, Sort: sort, Dir: params.PeriodOrder, Page: params.PeriodPage, PageSize: params.PeriodPageSize, PositiveOnly: true,
		})
		if err != nil {
			return nil, 0, err
		}
		for _, bucket := range buckets {
			if bucket == nil {
				continue
			}
			period, err := types.ParseCalendarPeriod(params.PeriodType, bucket.ScopeKey, time.Local)
			if err != nil {
				return nil, 0, err
			}
			rows = append(rows, cardPeriodRow(period, bucket.MediaCount))
		}
		return rows, total, nil
	}
	switch params.PeriodSort {
	case "count":
		sort = "all"
	case "owned":
		sort = "owned"
	}
	buckets, total, err := s.deps.MovieReleaseBucketStatModel.ListPage(ctx, moviex.MovieReleaseBucketStatListFilter{
		AggMode: "all", Level: params.PeriodType, Sort: sort, Dir: params.PeriodOrder, Page: params.PeriodPage, PageSize: params.PeriodPageSize, PositiveOnly: true,
	})
	if err != nil {
		return nil, 0, err
	}
	for _, bucket := range buckets {
		if bucket == nil {
			continue
		}
		period, err := types.ParseCalendarPeriod(params.PeriodType, bucket.ScopeKey, time.Local)
		if err != nil {
			return nil, 0, err
		}
		row := cardPeriodRow(period, bucket.CountAll)
		row.CountOwned = bucket.CountOwned
		rows = append(rows, row)
	}
	return rows, total, nil
}

func cardPeriodRow(period types.CalendarPeriod, count int64) CardPeriodRow {
	return CardPeriodRow{Key: period.Key, Label: strings.TrimPrefix(period.Key, period.Kind+":"), RangeStart: period.Start.Format(time.DateOnly), RangeEnd: period.End.Format(time.DateOnly), Count: count}
}

func cardPeriodPage(page, total, size int64) int64 {
	pages := total / size
	if total%size != 0 {
		pages++
	}
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		return pages
	}
	return page
}
