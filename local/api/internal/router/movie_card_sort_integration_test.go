package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/core/logx"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/router"
	"rudy_gc/local/api/internal/types"
)

func TestMovieCardSortDirectionsIntegration(t *testing.T) {
	configPath := strings.TrimSpace(os.Getenv("RUDY_GC_INTEGRATION_CONFIG"))
	if configPath == "" {
		t.Skip("设置 RUDY_GC_INTEGRATION_CONFIG 后执行真实数据库排序集成测试")
	}

	gin.SetMode(gin.TestMode)
	logx.DisableStat()
	logx.Disable()

	cfg := config.MustLoad(configPath)
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		t.Fatalf("构建依赖失败: %v", err)
	}
	engine := router.New(deps)

	orderByValues := []string{
		consts.OrderByDetailUpdateTime,
		consts.OrderByCastAgeDesc,
		consts.OrderByCastAgeAsc,
		consts.OrderByViewerWatched,
		consts.OrderByReleasingDate,
		consts.OrderByRankDate,
		consts.OrderByHighestRank,
		consts.OrderByDaysInRank,
		consts.OrderByBirthTime,
		consts.OrderByMediaBirthTime,
		consts.OrderByScTimes,
		consts.OrderByComeTimes,
		consts.OrderByLastSelectedScTime,
	}

	for _, orderBy := range orderByValues {
		for _, order := range []string{"asc", "desc"} {
			t.Run("cardstoday/"+orderBy+"/"+order, func(t *testing.T) {
				assertMovieCardRequestOK(t, engine, "/api/cardstoday", url.Values{
					"od":    {orderBy},
					"order": {order},
					"ps":    {"1"},
				})
			})
		}
	}

	for _, orderBy := range []string{consts.OrderByReleasingDate, consts.OrderByMediaBirthTime} {
		for _, order := range []string{"asc", "desc"} {
			t.Run("g_sc_stat/"+orderBy+"/"+order, func(t *testing.T) {
				assertMovieCardRequestOK(t, engine, "/api/cardstoday", url.Values{
					"od":    {orderBy},
					"order": {order},
					"scmin": {"1"},
					"ps":    {"1"},
				})
			})
		}
	}

	endpoints := []string{
		"/api/cards",
		"/api/cardstoday",
		"/api/cardshasrank",
		"/api/cardsmediamowned",
		"/api/cardsneeddownload",
		"/api/cardsrandom",
	}
	for _, endpoint := range endpoints {
		for _, order := range []string{"asc", "desc"} {
			t.Run(strings.TrimPrefix(endpoint, "/api/")+"/"+order, func(t *testing.T) {
				assertMovieCardRequestOK(t, engine, endpoint, url.Values{
					"od":    {consts.OrderByLastSelectedScTime},
					"order": {order},
					"ps":    {"1"},
					"n":     {"1"},
				})
			})
		}
	}
}

func TestMovieCardSortOrderValuesIntegration(t *testing.T) {
	configPath := strings.TrimSpace(os.Getenv("RUDY_GC_INTEGRATION_CONFIG"))
	if configPath == "" {
		t.Skip("设置 RUDY_GC_INTEGRATION_CONFIG 后执行真实数据库排序值集成测试")
	}

	gin.SetMode(gin.TestMode)
	logx.DisableStat()
	logx.Disable()

	cfg := config.MustLoad(configPath)
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		t.Fatalf("构建依赖失败: %v", err)
	}
	engine := router.New(deps)
	orderByValues := []string{
		consts.OrderByDetailUpdateTime,
		consts.OrderByCastAgeDesc,
		consts.OrderByCastAgeAsc,
		consts.OrderByViewerWatched,
		consts.OrderByReleasingDate,
		consts.OrderByRankDate,
		consts.OrderByHighestRank,
		consts.OrderByDaysInRank,
		consts.OrderByBirthTime,
		consts.OrderByMediaBirthTime,
		consts.OrderByScTimes,
		consts.OrderByComeTimes,
		consts.OrderByLastSelectedScTime,
	}

	for _, orderBy := range orderByValues {
		for _, order := range []string{"asc", "desc"} {
			t.Run(orderBy+"/"+order, func(t *testing.T) {
				movies := loadMovieCardResponse(t, engine, "/api/cardstoday", url.Values{
					"od":    {orderBy},
					"order": {order},
					"ps":    {"18"},
				})
				for i := 1; i < len(movies); i++ {
					previous := movieCardSortPrimaryValue(movies[i-1], orderBy)
					current := movieCardSortPrimaryValue(movies[i], orderBy)
					if order == "asc" && previous > current {
						t.Fatalf("第 %d 项未按正序排列: %d > %d", i, previous, current)
					}
					if order == "desc" && previous < current {
						t.Fatalf("第 %d 项未按倒序排列: %d < %d", i, previous, current)
					}
				}
			})
		}
	}
}

func assertMovieCardRequestOK(t *testing.T, handler http.Handler, path string, query url.Values) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path+"?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code == http.StatusOK {
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("%s 返回 HTTP %d，响应解析失败: %v", path, response.Code, err)
	}
	t.Fatalf("%s 返回 HTTP %d: %v", path, response.Code, payload["error"])
}

func loadMovieCardResponse(t *testing.T, handler http.Handler, path string, query url.Values) []*types.MovieType {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path+"?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("%s 返回 HTTP %d: %s", path, response.Code, response.Body.String())
	}
	var payload struct {
		Movies []*types.MovieType `json:"movies"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("%s 响应解析失败: %v", path, err)
	}
	return payload.Movies
}

func movieCardSortPrimaryValue(movie *types.MovieType, orderBy string) int64 {
	if movie == nil {
		return 0
	}
	switch orderBy {
	case consts.OrderByDetailUpdateTime:
		if movie.AMovie != nil {
			return movie.AMovie.DetailUpdateTime
		}
	case consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc:
		if movie.AMovie != nil {
			return movie.AMovie.CastAverageAge
		}
	case consts.OrderByViewerWatched:
		return movie.ViewersNumberWatched
	case consts.OrderByReleasingDate:
		if movie.AMovie != nil {
			return movie.AMovie.ReleasingDate
		}
	case consts.OrderByRankDate:
		if movie.BmMinfo != nil {
			return movie.BmMinfo.FirstRankDayNumber
		}
	case consts.OrderByHighestRank:
		return movie.HighestRank
	case consts.OrderByDaysInRank:
		if movie.BmMinfo != nil {
			return movie.BmMinfo.DaysInRank
		}
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		if movie.WMedia != nil {
			return movie.WMedia.BirthTime
		}
	case consts.OrderByScTimes:
		return movie.ScTimes
	case consts.OrderByComeTimes:
		return movie.ComeTimes
	case consts.OrderByLastSelectedScTime:
		return movie.LastSelectedScTime
	}
	return 0
}
