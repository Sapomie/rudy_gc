package movie

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/router/handler/common"
)

type emptyCardPeriods struct {
	moviex.MovieReleaseBucketStatModel
}

func (*emptyCardPeriods) ListPage(context.Context, moviex.MovieReleaseBucketStatListFilter) ([]*moviex.MovieReleaseBucketStat, int64, error) {
	return nil, 0, nil
}

type emptyCardBirthPeriods struct {
	moviex.WMediaBirthBucketStatModel
}

func (*emptyCardBirthPeriods) ListPage(context.Context, moviex.WMediaBirthBucketStatListFilter) ([]*moviex.WMediaBirthBucketStat, int64, error) {
	return nil, 0, nil
}

func TestCardPeriodAPIValidationAndEmptyResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewMovieCardAPI(&dep.Dep{MovieReleaseBucketStatModel: &emptyCardPeriods{}, WMediaBirthBucketStatModel: &emptyCardBirthPeriods{}})
	engine.GET("/api/cardsperiodrelease", handler.ListMovieCardPeriodRelease)
	engine.GET("/api/cardsperiodbirth", handler.ListMovieCardPeriodBirth)
	for _, basis := range []string{"release", "birth"} {
		path := "/api/cardsperiod" + basis
		for _, query := range []string{"period_type=unknown", "period_type=week&key=week:2021-W53", "p=bad", "ps=0", "period_ps=-1", "order=invalid"} {
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path+"?"+query, nil))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("参数 %s?%s 状态 = %d, %s", path, query, response.Code, response.Body.String())
			}
		}
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("空目录返回 %d: %s", response.Code, response.Body.String())
		}
		var result struct {
			Basis          string
			PeriodType     string
			Periods        []any
			Movies         []any `json:"movies"`
			SelectedPeriod any
			PageInfo       common.PageInfo
			PeriodPageInfo common.PageInfo
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Periods == nil || result.Movies == nil || result.SelectedPeriod != nil || result.PageInfo.Page != 1 || result.PeriodPageInfo.Page != 1 {
			t.Fatalf("空状态错误: %+v", result)
		}
		if result.Basis != basis || result.PeriodType != "day" {
			t.Fatalf("默认周期错误: %+v", result)
		}
	}
}

func TestCardPeriodPaginationKeepsIndependentQueries(t *testing.T) {
	for _, basis := range []string{"release", "birth"} {
		path := "/listcardperiod" + basis
		periodSort := "count"
		if basis == "release" {
			periodSort = "owned"
		}
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest(http.MethodGet, path+"?period_type=month&key=month:2020-12&period_p=3&period_ps=24&p=4&ps=18&od=vw&order=asc&period_sort="+periodSort+"&period_order=desc", nil)
		cards := common.BuildPageInfo(context, 100, 4, 18, common.PageWindow)
		periods := buildCardPeriodPageInfo(context, 100, 3, 24)
		for _, tc := range []struct{ href, p, periodPage, key string }{
			{cards.NextHref, "5", "3", "month:2020-12"},
			{periods.NextHref, "1", "4", ""},
		} {
			u, err := url.Parse(tc.href)
			if err != nil {
				t.Fatal(err)
			}
			q := u.Query()
			if u.Path != path || q.Has("basis") || q.Get("p") != tc.p || q.Get("period_p") != tc.periodPage || q.Get("key") != tc.key || q.Get("period_ps") != "24" || q.Get("ps") != "18" || q.Get("od") != "vw" || q.Get("order") != "asc" || q.Get("period_sort") != periodSort || q.Get("period_order") != "desc" {
				t.Fatalf("分页互相覆盖或页面路径错误: %s", tc.href)
			}
		}
	}
}

func TestCardPeriodAPIOwnedSort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewMovieCardAPI(&dep.Dep{MovieReleaseBucketStatModel: &emptyCardPeriods{}, WMediaBirthBucketStatModel: &emptyCardBirthPeriods{}})
	engine.GET("/api/cardsperiodrelease", handler.ListMovieCardPeriodRelease)
	engine.GET("/api/cardsperiodbirth", handler.ListMovieCardPeriodBirth)
	for _, direction := range []string{"asc", "desc"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/cardsperiodrelease?period_sort=owned&period_order="+direction, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("已拥有排序返回 %d: %s", response.Code, response.Body.String())
		}
		var result struct {
			Query          url.Values
			SortQuery      map[string]common.SortLink `json:"sortQuery"`
			PeriodPageInfo common.PageInfo
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Query.Get("period_sort") != "owned" || result.Query.Get("period_order") != direction || !result.SortQuery["ByReleasing"].Active {
			t.Fatalf("已拥有排序或默认卡片排序错误: %+v", result)
		}
		for _, href := range []string{result.SortQuery["ByViewer"].Href, result.PeriodPageInfo.NextHref} {
			u, err := url.Parse(href)
			if err != nil {
				t.Fatal(err)
			}
			if u.Path != "/listcardperiodrelease" || u.Query().Get("period_sort") != "owned" || u.Query().Get("period_order") != direction {
				t.Fatalf("排序条件未保留: %s", href)
			}
		}
		birth := httptest.NewRecorder()
		engine.ServeHTTP(birth, httptest.NewRequest(http.MethodGet, "/api/cardsperiodbirth?period_sort=owned&period_order="+direction, nil))
		if birth.Code != http.StatusBadRequest {
			t.Fatalf("下载周期接受了已拥有排序: %d", birth.Code)
		}
	}
}

func TestCardPeriodAPIReleaseOwnershipFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewMovieCardAPI(&dep.Dep{MovieReleaseBucketStatModel: &emptyCardPeriods{}, WMediaBirthBucketStatModel: &emptyCardBirthPeriods{}})
	engine.GET("/api/cardsperiodrelease", handler.ListMovieCardPeriodRelease)
	engine.GET("/api/cardsperiodbirth", handler.ListMovieCardPeriodBirth)
	for _, tc := range []struct{ value, active string }{
		{"", "MediaAll"}, {"1", "MediaAll"}, {"3", "MediaOwned"}, {"7", "MediaNotOwned"},
	} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/cardsperiodrelease?mowned="+tc.value+"&period_type=month&period_sort=owned&period_order=asc&period_ps=24&ps=36&od=vw&order=asc", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("拥有筛选返回 %d: %s", response.Code, response.Body.String())
		}
		var result struct {
			Query      url.Values
			OwnedQuery map[string]common.OwnedLink `json:"ownedQuery"`
			SortQuery  map[string]common.SortLink  `json:"sortQuery"`
			PageInfo   common.PageInfo
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		wantValue := tc.value
		if wantValue == "" {
			wantValue = "1"
		}
		if result.Query.Get("mowned") != wantValue || !result.SortQuery["ByViewer"].Active {
			t.Fatalf("拥有筛选或当前卡片排序错误: %+v", result)
		}
		for key, value := range map[string]string{"MediaAll": "1", "MediaOwned": "3", "MediaNotOwned": "7"} {
			link, ok := result.OwnedQuery[key]
			if !ok || link.Active != (key == tc.active) {
				t.Fatalf("拥有筛选选中状态错误: %+v", result.OwnedQuery)
			}
			u, err := url.Parse(link.Href)
			if err != nil {
				t.Fatal(err)
			}
			q := u.Query()
			if u.Path != "/listcardperiodrelease" || q.Get("mowned") != value || q.Get("p") != "1" || q.Get("period_type") != "month" || q.Get("period_sort") != "owned" || q.Get("period_order") != "asc" || q.Get("period_ps") != "24" || q.Get("ps") != "36" || q.Get("od") != "vw" || q.Get("order") != "asc" {
				t.Fatalf("拥有筛选覆盖了周期、排序或条数: %s", link.Href)
			}
		}
		for _, href := range []string{result.SortQuery["ByReleasing"].Href, result.PageInfo.NextHref} {
			u, err := url.Parse(href)
			if err != nil || u.Query().Get("mowned") != wantValue {
				t.Fatalf("卡片排序或分页丢失拥有筛选: %s, %v", href, err)
			}
		}
	}
	for _, value := range []string{"bad", "8"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/cardsperiodrelease?mowned="+value, nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("无效拥有筛选返回 %d", response.Code)
		}
	}
	birth := httptest.NewRecorder()
	engine.ServeHTTP(birth, httptest.NewRequest(http.MethodGet, "/api/cardsperiodbirth", nil))
	var result map[string]json.RawMessage
	if err := json.Unmarshal(birth.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if _, exists := result["ownedQuery"]; exists {
		t.Fatal("下载周期新增了拥有筛选")
	}
}

func TestCardPeriodAPIAppliesCardSortAndBuildsAllSortLinks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handler := NewMovieCardAPI(&dep.Dep{MovieReleaseBucketStatModel: &emptyCardPeriods{}, WMediaBirthBucketStatModel: &emptyCardBirthPeriods{}})
	engine.GET("/api/cardsperiodrelease", handler.ListMovieCardPeriodRelease)
	engine.GET("/api/cardsperiodbirth", handler.ListMovieCardPeriodBirth)
	for _, tc := range []struct{ query, active, order, basis string }{
		{"", "ByBirthTime", "desc", "birth"},
		{"", "ByReleasing", "desc", "release"},
		{"od=vw&order=asc", "ByViewer", "asc", "release"},
		{"od=sc&order=desc", "ByScTimes", "desc", "birth"},
		{"od=hrk", "ByHighestRank", "asc", "release"},
		{"od=invalid", "ByReleasing", "desc", "release"},
		{"basis=birth", "ByReleasing", "desc", "release"},
		{"basis=release", "ByBirthTime", "desc", "birth"},
		{"basis=rank", "ByBirthTime", "desc", "birth"},
	} {
		t.Run(tc.basis+"/"+tc.query, func(t *testing.T) {
			response := httptest.NewRecorder()
			engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/cardsperiod"+tc.basis+"?"+tc.query, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("排序请求返回 %d: %s", response.Code, response.Body.String())
			}
			var result struct {
				Basis     string
				Title     string
				Query     url.Values
				SortQuery map[string]common.SortLink `json:"sortQuery"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			title := "上映周期卡片"
			if tc.basis == "birth" {
				title = "下载周期卡片"
			}
			if result.Basis != tc.basis || result.Title != title || result.Query.Has("basis") || result.Query.Get("order") != tc.order || !result.SortQuery[tc.active].Active || result.SortQuery[tc.active].Desc != (tc.order == "desc") {
				t.Fatalf("当前排序与方向错误: %+v", result)
			}
			codes := map[string]string{"ByDetailUpdate": consts.OrderByDetailUpdateTime, "ByReleasing": consts.OrderByReleasingDate, "ByViewer": consts.OrderByViewerWatched, "ByRankDate": consts.OrderByRankDate, "ByDaysInRank": consts.OrderByDaysInRank, "ByHighestRank": consts.OrderByHighestRank, "ByBirthTime": consts.OrderByBirthTime, "ByLastSelectedScTime": consts.OrderByLastSelectedScTime, "ByScTimes": consts.OrderByScTimes, "ByComeTimes": consts.OrderByComeTimes}
			for key, code := range codes {
				link := result.SortQuery[key]
				u, err := url.Parse(link.Href)
				if err != nil {
					t.Fatal(err)
				}
				q := u.Query()
				if u.Path != "/listcardperiod"+tc.basis || q.Get("od") != code || q.Get("p") != "1" || q.Has("basis") {
					t.Fatalf("排序链接错误: %s", link.Href)
				}
				if key == tc.active && q.Get("order") == tc.order {
					t.Fatalf("当前排序未切换方向: %s", link.Href)
				}
			}
		})
	}
}
