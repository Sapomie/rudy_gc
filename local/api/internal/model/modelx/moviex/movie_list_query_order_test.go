package moviex

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"

	"rudy_gc/local/api/internal/consts"
)

func TestSplitOrderClauseKeepsFunctionArgumentsTogether(t *testing.T) {
	clause := "COALESCE(gss.sc_times, 0) DESC, COALESCE(NULLIF(gss.last_selected_sc_time, 0), 0) ASC, wm.movie_jav_id DESC"

	got := splitOrderClause(clause)
	want := []string{
		"COALESCE(gss.sc_times, 0) DESC",
		"COALESCE(NULLIF(gss.last_selected_sc_time, 0), 0) ASC",
		"wm.movie_jav_id DESC",
	}

	if len(got) != len(want) {
		t.Fatalf("splitOrderClause() parts = %d, want %d: %#v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("splitOrderClause()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestApplyMovieListOrderKeepsCoalesceValid(t *testing.T) {
	const clause = "COALESCE(gss.last_selected_sc_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"

	if got, want := applyMovieListOrder(clause, "asc"), "COALESCE(gss.last_selected_sc_time, 0) ASC, gss.movie_name ASC, wm.movie_jav_id ASC"; got != want {
		t.Errorf("applyMovieListOrder(asc) = %q, want %q", got, want)
	}
	if got, want := applyMovieListOrder(clause, "desc"), "COALESCE(gss.last_selected_sc_time, 0) DESC, gss.movie_name DESC, wm.movie_jav_id DESC"; got != want {
		t.Errorf("applyMovieListOrder(desc) = %q, want %q", got, want)
	}
}

func TestSortNativeMediaRowsSupportsAscendingStatExpressions(t *testing.T) {
	rows := []*NativeMediaListRow{
		{Id: 1, MovieName: "B", ScTimes: 3},
		{Id: 2, MovieName: "A", ScTimes: 1},
	}

	sortNativeMediaRows(rows, "COALESCE(gss.sc_times, 0) ASC, wm.movie_name ASC, wm.id ASC")

	if rows[0].Id != 2 || rows[1].Id != 1 {
		t.Fatalf("SC 次数正序结果错误: got ids [%d, %d], want [2, 1]", rows[0].Id, rows[1].Id)
	}
}

func TestMovieCardOrderingsSupportBothDirectionsOnEveryQueryPath(t *testing.T) {
	type orderCase struct {
		name  string
		build func(direction string) string
	}

	amovie := func(orderBy string) func(string) string {
		return func(direction string) string {
			order, _ := amovieOrdering(orderBy, direction)
			return order
		}
	}
	minfo := func(orderBy string) func(string) string {
		return func(direction string) string {
			order, _ := minfoOrdering(orderBy, direction)
			return order
		}
	}
	gscStat := func(orderBy string) func(string) string {
		return func(direction string) string {
			order, _ := gscStatOrdering(orderBy, direction)
			return order
		}
	}
	mediaMatched := func(orderBy string) func(string) string {
		return func(direction string) string {
			return mediaMatchedOrdering(orderBy, direction, "q")
		}
	}
	wmedia := func(orderBy string) func(string) string {
		return func(direction string) string {
			return wmediaOrdering(orderBy, direction)
		}
	}

	cases := []orderCase{
		{name: "a_movie/detail_update", build: amovie(consts.OrderByDetailUpdateTime)},
		{name: "a_movie/releasing", build: amovie(consts.OrderByReleasingDate)},
		{name: "a_movie/viewer", build: amovie(consts.OrderByViewerWatched)},
		{name: "a_movie/cast_age_desc", build: amovie(consts.OrderByCastAgeDesc)},
		{name: "a_movie/cast_age_asc", build: amovie(consts.OrderByCastAgeAsc)},
		{name: "minfo/rank_date", build: minfo(consts.OrderByRankDate)},
		{name: "minfo/highest_rank", build: minfo(consts.OrderByHighestRank)},
		{name: "minfo/days_in_rank", build: minfo(consts.OrderByDaysInRank)},
		{name: "minfo/releasing", build: minfo(consts.OrderByReleasingDate)},
		{name: "g_sc_stat/sc_times", build: gscStat(consts.OrderByScTimes)},
		{name: "g_sc_stat/come_times", build: gscStat(consts.OrderByComeTimes)},
		{name: "g_sc_stat/last_selected_sc_time", build: gscStat(consts.OrderByLastSelectedScTime)},
		{name: "g_sc_stat/releasing", build: gscStat(consts.OrderByReleasingDate)},
		{name: "g_sc_stat/media_birth_time", build: gscStat(consts.OrderByMediaBirthTime)},
		{name: "media_matched/birth_time", build: mediaMatched(consts.OrderByBirthTime)},
		{name: "media_matched/media_birth_time", build: mediaMatched(consts.OrderByMediaBirthTime)},
		{name: "media_matched/releasing", build: mediaMatched(consts.OrderByReleasingDate)},
		{name: "w_media/birth_time", build: wmedia(consts.OrderByBirthTime)},
		{name: "w_media/media_birth_time", build: wmedia(consts.OrderByMediaBirthTime)},
		{name: "w_media/releasing", build: wmedia(consts.OrderByReleasingDate)},
		{
			name: "album/releasing",
			build: func(direction string) string {
				return applyMovieListOrder("releasing_date DESC,movie_name DESC,movie_jav_id DESC", direction)
			},
		},
	}

	for _, testCase := range cases {
		for _, direction := range []string{"asc", "desc"} {
			t.Run(testCase.name+"/"+direction, func(t *testing.T) {
				order := testCase.build(direction)
				parts := splitOrderClause(order)
				if len(parts) == 0 {
					t.Fatal("生成的排序表达式为空")
				}

				wantSuffix := " " + strings.ToUpper(direction)
				for _, part := range parts {
					if !strings.HasSuffix(strings.ToUpper(strings.TrimSpace(part)), wantSuffix) {
						t.Errorf("排序项方向错误: %q, want suffix %q", part, wantSuffix)
					}
				}

				if strings.Contains(order, strings.ToUpper(direction)+", 0)") {
					t.Errorf("函数参数被错误拆分: %q", order)
				}
			})
		}
	}
}

func TestRankDateGuardUsesIndexedSortColumn(t *testing.T) {
	query, _, err := squirrel.
		Select("jav_id").
		From("bm_minfo").
		Where(minfoOrderGuards(consts.OrderByRankDate)).
		ToSql()
	if err != nil {
		t.Fatalf("生成榜单日期过滤 SQL 失败: %v", err)
	}

	if !strings.Contains(query, "first_rank_day_number > 0") {
		t.Fatalf("榜单日期过滤未使用排序索引首列: %q", query)
	}
	if strings.Contains(query, "days_in_rank") {
		t.Fatalf("榜单日期过滤仍依赖非排序首列: %q", query)
	}
}
