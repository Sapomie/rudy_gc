package movie

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type cardPeriodReleaseBuckets struct {
	moviex.MovieReleaseBucketStatModel
	filter moviex.MovieReleaseBucketStatListFilter
	key    string
}

func (m *cardPeriodReleaseBuckets) ListPage(_ context.Context, filter moviex.MovieReleaseBucketStatListFilter) ([]*moviex.MovieReleaseBucketStat, int64, error) {
	m.filter = filter
	key := m.key
	if key == "" {
		key = "week:2020-W53"
	}
	return []*moviex.MovieReleaseBucketStat{{ScopeKey: key, CountAll: 5, CountOwned: 2}}, 5, nil
}

type cardPeriodBirthBuckets struct {
	moviex.WMediaBirthBucketStatModel
	filter moviex.WMediaBirthBucketStatListFilter
	key    string
}

func (m *cardPeriodBirthBuckets) ListPage(_ context.Context, filter moviex.WMediaBirthBucketStatListFilter) ([]*moviex.WMediaBirthBucketStat, int64, error) {
	m.filter = filter
	key := m.key
	if key == "" {
		key = "week:2020-W53"
	}
	return []*moviex.WMediaBirthBucketStat{{ScopeKey: key, MediaCount: 5}}, 5, nil
}

type cardPeriodQuery struct {
	query      string
	args       []any
	countQuery string
	countArgs  []any
	calls      int
}

func (q *cardPeriodQuery) QueryRowNoCacheCtx(_ context.Context, dest any, query string, args ...any) error {
	q.calls++
	q.countQuery, q.countArgs = query, args
	*dest.(*int64) = 5
	return nil
}
func (q *cardPeriodQuery) QueryRowsNoCacheCtx(_ context.Context, dest any, query string, args ...any) error {
	q.calls++
	q.query, q.args = query, args
	*dest.(*[]string) = []string{"jav-a", "jav-b"}
	return nil
}

type cardPeriodMovies struct {
	moviex.AMovieModel
	*cardPeriodQuery
}

func (*cardPeriodMovies) TableName() string { return "a_movie" }

func (m *cardPeriodMovies) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.cardPeriodQuery.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *cardPeriodMovies) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.cardPeriodQuery.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

type cardPeriodMedia struct {
	moviex.WMediaModel
	*cardPeriodQuery
}

func (*cardPeriodMedia) TableName() string { return "w_media" }

func (m *cardPeriodMedia) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.cardPeriodQuery.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *cardPeriodMedia) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.cardPeriodQuery.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

type cardPeriodCache struct{ dep.MovieTypeCache }

type cardPeriodMinfo struct {
	moviex.BmMinfoModel
	query *cardPeriodQuery
}

func (*cardPeriodMinfo) TableName() string { return "bm_minfo" }
func (m *cardPeriodMinfo) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.query.QueryRowNoCacheCtx(ctx, dest, query, args...)
}
func (m *cardPeriodMinfo) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.query.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

type cardPeriodSC struct{ moviex.GScStatModel }

func (*cardPeriodSC) TableName() string { return "g_sc_stat" }

func (*cardPeriodCache) GetMovieType(_ context.Context, id string) (*types.MovieType, error) {
	return &types.MovieType{JavId: id}, nil
}

func TestCardPeriodUsesMatchingDirectoryAndIndexedMovieRange(t *testing.T) {
	for _, basis := range []string{"release", "birth"} {
		t.Run(basis, func(t *testing.T) {
			query := &cardPeriodQuery{}
			releases, births := &cardPeriodReleaseBuckets{}, &cardPeriodBirthBuckets{}
			svc := NewService(&dep.Dep{
				MovieReleaseBucketStatModel: releases, WMediaBirthBucketStatModel: births,
				MovieModel: &cardPeriodMovies{cardPeriodQuery: query}, WMediaModel: &cardPeriodMedia{cardPeriodQuery: query}, MovieTypeCache: &cardPeriodCache{},
			})
			params := CardPeriodParams{Basis: basis, PeriodType: "week", PeriodPage: 3, PeriodPageSize: 2, PeriodSort: "time", PeriodOrder: "desc", Page: 2, PageSize: 2, Order: "asc"}
			result, err := svc.BuildCardPeriod(context.Background(), params)
			if err != nil {
				t.Fatal(err)
			}
			if result.Params.Key != "week:2020-W53" || result.Params.Page != 2 || result.Params.PeriodPage != 3 || result.Total != 5 || len(result.Movies) != 2 {
				t.Fatalf("独立分页或自动选择错误: %+v", result)
			}
			if result.SelectedPeriod.RangeStart != "2020-12-28" || result.SelectedPeriod.RangeEnd != "2021-01-03" {
				t.Fatalf("跨年周范围错误: %+v", result.SelectedPeriod)
			}
			wantOwned := int64(0)
			if basis == "release" {
				wantOwned = 2
			}
			if result.Periods[0].CountOwned != wantOwned || result.SelectedPeriod.CountOwned != wantOwned {
				t.Fatalf("已拥有统计未传递: 目录=%+v, 选中=%+v", result.Periods[0], result.SelectedPeriod)
			}
			if !strings.Contains(query.query, "LIMIT 2 OFFSET 2") || strings.Contains(query.query, " IN ") {
				t.Fatalf("卡片未在数据库直接分页: %s", query.query)
			}
			period, _ := types.ParseCalendarPeriod("week", result.Params.Key, time.Local)
			wantArgs := []any{period.Start.Unix(), period.End.Unix()}
			// Release dates are stored at midnight; download ranges include the whole final day.
			if basis == "release" {
				wantArgs[1] = time.Date(2021, 1, 3, 0, 0, 0, 0, time.Local).Unix()
				if releases.filter.AggMode != "all" || !releases.filter.PositiveOnly || releases.filter.Level != "week" || releases.filter.Page != 3 {
					t.Fatalf("上映目录筛选错误: %+v", releases.filter)
				}
				if !strings.Contains(query.query, "releasing_date >= ?") || !strings.Contains(query.query, "releasing_date ASC") {
					t.Fatalf("上映范围或排序错误: %s", query.query)
				}
			} else {
				if !births.filter.PositiveOnly || births.filter.Level != "week" || births.filter.Page != 3 {
					t.Fatalf("下载目录筛选错误: %+v", births.filter)
				}
				if !strings.Contains(query.query, "birth_time >= ?") || !strings.Contains(query.query, "birth_time ASC") {
					t.Fatalf("下载范围或排序错误: %s", query.query)
				}
				wantArgs = []any{consts.WMediaSourceNative, consts.FilmIsNotRemoved, period.Start.Unix(), period.End.Unix()}
			}
			if !reflect.DeepEqual(query.args, wantArgs) {
				t.Fatalf("范围参数 = %#v, want %#v", query.args, wantArgs)
			}
		})
	}
}

func TestCardPeriodOwnedDirectorySort(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		t.Run(direction, func(t *testing.T) {
			query := &cardPeriodQuery{}
			releases := &cardPeriodReleaseBuckets{key: "month:2026-10"}
			svc := NewService(&dep.Dep{
				MovieReleaseBucketStatModel: releases, MovieModel: &cardPeriodMovies{cardPeriodQuery: query}, MovieTypeCache: &cardPeriodCache{},
			})
			params := CardPeriodParams{Basis: "release", PeriodType: "month", PeriodPage: 3, PeriodPageSize: 2, PeriodSort: "owned", PeriodOrder: direction, Page: 1, PageSize: 2, OrderBy: consts.OrderByReleasingDate, Order: "desc"}
			result, err := svc.BuildCardPeriod(context.Background(), params)
			if err != nil {
				t.Fatal(err)
			}
			if releases.filter.Sort != "owned" || releases.filter.Dir != direction || releases.filter.AggMode != "all" || releases.filter.Level != "month" || !releases.filter.PositiveOnly || releases.filter.Page != 3 || releases.filter.PageSize != 2 {
				t.Fatalf("已拥有排序未传递给数据库分页查询: %+v", releases.filter)
			}
			if result.Params.PeriodSort != "owned" || result.Params.PeriodOrder != direction || result.SelectedPeriod.CountOwned != 2 || result.Params.Key != "month:2026-10" {
				t.Fatalf("已拥有排序、统计或自动选择错误: %+v", result)
			}
			if result.OrderBy != consts.OrderByReleasingDate || !strings.Contains(query.query, "releasing_date DESC") {
				t.Fatalf("目录排序覆盖了卡片排序: %s", query.query)
			}
			params.Basis = "birth"
			if _, err := svc.BuildCardPeriod(context.Background(), params); !errors.Is(err, ErrCardPeriodParams) {
				t.Fatalf("下载周期接受了已拥有排序: %v", err)
			}
		})
	}
}

func TestCardPeriodReleaseOwnershipFilters(t *testing.T) {
	for _, ownership := range []int64{0, consts.MovieAll, consts.OwnedAllNotRemoved, consts.OwnedNotOwned} {
		for _, sort := range []string{consts.OrderByReleasingDate, consts.OrderByRankDate, consts.OrderByBirthTime, consts.OrderByScTimes} {
			query := &cardPeriodQuery{}
			releases := &cardPeriodReleaseBuckets{key: "month:2026-10"}
			svc := NewService(&dep.Dep{
				MovieReleaseBucketStatModel: releases,
				MovieModel:                  &cardPeriodMovies{cardPeriodQuery: query}, WMediaModel: &cardPeriodMedia{cardPeriodQuery: query},
				MinfoModel: &cardPeriodMinfo{query: query}, GScStatModel: &cardPeriodSC{}, MovieTypeCache: &cardPeriodCache{},
			})
			params := CardPeriodParams{Basis: "release", PeriodType: "month", PeriodPage: 3, PeriodPageSize: 2, PeriodSort: "owned", PeriodOrder: "asc", Page: 2, PageSize: 2, OrderBy: sort, Order: "desc", MediaOwned: ownership}
			result, err := svc.BuildCardPeriod(context.Background(), params)
			if err != nil {
				t.Fatal(err)
			}
			wantOwned := ownership
			if wantOwned == 0 {
				wantOwned = consts.MovieAll
			}
			if result.Params.MediaOwned != wantOwned || result.OrderBy != sort || result.Params.Key != "month:2026-10" || result.Params.PeriodPage != 3 {
				t.Fatalf("拥有筛选覆盖了周期、分页或排序: %+v", result.Params)
			}
			if result.Periods[0].Count != 5 || result.Periods[0].CountOwned != 2 || releases.filter.AggMode != "all" {
				t.Fatalf("卡片筛选改变了左侧桶统计: %+v", result.Periods)
			}
			mediaSource := sort == consts.OrderByBirthTime || sort == consts.OrderByScTimes
			if wantOwned == consts.OwnedNotOwned && mediaSource {
				if result.Total != 0 || len(result.Movies) != 0 || result.Params.Page != 1 || query.calls != 0 {
					t.Fatalf("互斥筛选未直接返回空结果: %+v, 数据库查询=%d", result, query.calls)
				}
				continue
			}
			if result.Params.Page != 2 {
				t.Fatalf("拥有筛选覆盖了卡片页码: %+v", result.Params)
			}
			if query.calls != 2 || !reflect.DeepEqual(query.args, query.countArgs) || !strings.Contains(query.query, "LIMIT 2 OFFSET 2") || strings.Contains(query.query, " IN ") {
				t.Fatalf("筛选计数与数据库分页不一致: %s / %s", query.countQuery, query.query)
			}
			outerID := "a_movie.jav_id"
			if sort == consts.OrderByRankDate {
				outerID = "bm_minfo.jav_id"
			} else if mediaSource {
				outerID = "wm.movie_jav_id"
			}
			for _, sql := range []string{query.countQuery, query.query} {
				if !strings.Contains(sql, "releasing_date >= ?") || !strings.Contains(sql, "releasing_date <= ?") {
					t.Fatalf("拥有筛选丢失上映日期范围: %s", sql)
				}
				switch wantOwned {
				case consts.MovieAll:
					if strings.Contains(sql, "period_unowned") || strings.Contains(sql, "period_media") || strings.Contains(sql, "is_removed = ?") {
						t.Fatalf("All 增加了拥有限制: %s", sql)
					}
				case consts.OwnedAllNotRemoved:
					alias := "period_media"
					if mediaSource {
						alias = "wm"
					} else if !strings.Contains(sql, "period_media.movie_jav_id = "+outerID) {
						t.Fatalf("Owned 未关联排序影片: %s", sql)
					}
					if !strings.Contains(sql, alias+".source_type = ?") || !strings.Contains(sql, alias+".is_removed = ?") {
						t.Fatalf("Owned 未复用原生未移除媒体条件: %s", sql)
					}
				case consts.OwnedNotOwned:
					if !strings.Contains(sql, "NOT EXISTS (SELECT 1 FROM w_media period_unowned") || !strings.Contains(sql, "period_unowned.movie_jav_id = "+outerID) || !strings.Contains(sql, "period_unowned.source_type = 2") {
						t.Fatalf("NotOwned 未按当前影片排除原生媒体: %s", sql)
					}
				}
			}
		}
	}
	params := CardPeriodParams{Basis: "release", PeriodType: "month", PeriodPage: 1, PeriodPageSize: 2, PeriodSort: "time", PeriodOrder: "desc", Page: 1, PageSize: 2, Order: "desc", MediaOwned: 8}
	if err := params.Validate(); !errors.Is(err, ErrCardPeriodParams) {
		t.Fatalf("无效拥有筛选被接受: %v", err)
	}
}

func TestCardPeriodParamsAndExplicitSelection(t *testing.T) {
	params := CardPeriodParams{Basis: "release", PeriodType: "week", Key: "week:2021-W01", PeriodPage: 1, PeriodPageSize: 18, PeriodSort: "count", PeriodOrder: "asc", Page: 1, PageSize: 20001, Order: "desc"}
	if err := params.Validate(); err != nil {
		t.Fatalf("每页条数被限制: %v", err)
	}
	query := &cardPeriodQuery{}
	svc := NewService(&dep.Dep{MovieReleaseBucketStatModel: &cardPeriodReleaseBuckets{}, MovieModel: &cardPeriodMovies{cardPeriodQuery: query}, MovieTypeCache: &cardPeriodCache{}})
	result, err := svc.BuildCardPeriod(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if result.Params.Key != params.Key || result.SelectedPeriod.RangeStart != "2021-01-04" || !strings.Contains(query.query, "LIMIT 20001") {
		t.Fatalf("显式周期选择或条数错误: %+v, %s", result, query.query)
	}
	params.Key = "week:2021-W53"
	if _, err := svc.BuildCardPeriod(context.Background(), params); !errors.Is(err, ErrCardPeriodParams) {
		t.Fatalf("无效周应在查询前拒绝: %v", err)
	}
}

func TestCardPeriodAllCardSortsKeepPeriodRangeAndSQLPagination(t *testing.T) {
	sorts := []struct{ key, column string }{
		{consts.OrderByDetailUpdateTime, "detail_update_time"},
		{consts.OrderByReleasingDate, "releasing_date"},
		{consts.OrderByViewerWatched, "viewers_number_watched"},
		{consts.OrderByRankDate, "first_rank_day_number"},
		{consts.OrderByDaysInRank, "days_in_rank"},
		{consts.OrderByHighestRank, "highest_rank"},
		{consts.OrderByBirthTime, "birth_time"},
		{consts.OrderByLastSelectedScTime, "COALESCE(gss.last_selected_sc_time, 0)"},
		{consts.OrderByScTimes, "COALESCE(gss.sc_times, 0)"},
		{consts.OrderByComeTimes, "COALESCE(gss.come_times, 0)"},
	}
	for _, basis := range []string{"release", "birth"} {
		for _, sort := range sorts {
			for _, direction := range []string{"asc", "desc"} {
				t.Run(basis+"/"+sort.key+"/"+direction, func(t *testing.T) {
					query := &cardPeriodQuery{}
					svc := NewService(&dep.Dep{
						MovieReleaseBucketStatModel: &cardPeriodReleaseBuckets{key: "month:2026-10"}, WMediaBirthBucketStatModel: &cardPeriodBirthBuckets{key: "month:2026-10"},
						MovieModel: &cardPeriodMovies{cardPeriodQuery: query}, WMediaModel: &cardPeriodMedia{cardPeriodQuery: query},
						MinfoModel: &cardPeriodMinfo{query: query}, GScStatModel: &cardPeriodSC{}, MovieTypeCache: &cardPeriodCache{},
					})
					params := CardPeriodParams{Basis: basis, PeriodType: "month", Key: "month:2026-10", PeriodPage: 3, PeriodPageSize: 2, PeriodSort: "count", PeriodOrder: "desc", Page: 2, PageSize: 2, OrderBy: sort.key, Order: direction}
					result, err := svc.BuildCardPeriod(context.Background(), params)
					if err != nil {
						t.Fatal(err)
					}
					if result.OrderBy != sort.key || result.Params.OrderBy != sort.key || result.Params.Key != params.Key || result.Params.PeriodPage != 3 || len(result.Movies) != 2 {
						t.Fatalf("排序或周期被覆盖: %+v", result)
					}
					if !strings.Contains(query.query, sort.column+" "+strings.ToUpper(direction)) || !strings.Contains(query.query, "LIMIT 2 OFFSET 2") {
						t.Fatalf("排序或数据库分页错误: %s", query.query)
					}
					if query.calls != 2 || !strings.Contains(query.countQuery, "COUNT(*)") || strings.Contains(query.query, " IN ") {
						t.Fatalf("未直接计数和分页: %d, %s", query.calls, query.query)
					}
					if !reflect.DeepEqual(query.args, query.countArgs) {
						t.Fatalf("计数和分页条件不一致: %#v / %#v", query.args, query.countArgs)
					}
					if basis == "birth" {
						if !strings.Contains(query.query, "birth_time >= ?") || !strings.Contains(query.query, "is_removed = ?") || !strings.Contains(query.query, "source_type = ?") {
							t.Fatalf("丢失原生媒体下载范围: %s", query.query)
						}
					} else if !strings.Contains(query.query, "releasing_date >= ?") || !strings.Contains(query.query, "releasing_date <= ?") {
						t.Fatalf("丢失上映范围: %s", query.query)
					}
					if strings.Contains(query.query, "FROM bm_minfo") && basis == "release" && !strings.Contains(query.query, "period_movie.jav_id = bm_minfo.jav_id") {
						t.Fatalf("上映范围未关联排序影片: %s", query.query)
					}
				})
			}
		}
	}
}
