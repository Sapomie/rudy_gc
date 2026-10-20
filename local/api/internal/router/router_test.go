package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/service/loop"
)

func TestDailyBestTriggerRoutes(t *testing.T) {
	engine := New(&dep.Dep{})

	pageResp := httptest.NewRecorder()
	pageReq := httptest.NewRequest(http.MethodGet, "/api/triggers/dailybest?job_id=test_job", nil)
	engine.ServeHTTP(pageResp, pageReq)
	if pageResp.Code != http.StatusOK {
		t.Fatalf("GET /api/triggers/dailybest status = %d, want %d", pageResp.Code, http.StatusOK)
	}

	var pageData map[string]any
	if err := json.Unmarshal(pageResp.Body.Bytes(), &pageData); err != nil {
		t.Fatalf("decode dailybest response: %v", err)
	}
	if pageData["ViewType"] != "crawler-jobs" {
		t.Fatalf("ViewType = %v, want crawler-jobs", pageData["ViewType"])
	}
	if pageData["DefaultTaskType"] != loop.TaskSpiderDailyBest {
		t.Fatalf("DefaultTaskType = %v, want %s", pageData["DefaultTaskType"], loop.TaskSpiderDailyBest)
	}
	if pageData["JobID"] != "test_job" {
		t.Fatalf("JobID = %v, want test_job", pageData["JobID"])
	}

	jobsResp := httptest.NewRecorder()
	jobsReq := httptest.NewRequest(http.MethodGet, "/api/crawler/jobs", nil)
	engine.ServeHTTP(jobsResp, jobsReq)
	if jobsResp.Code != http.StatusOK {
		t.Fatalf("GET /api/crawler/jobs status = %d, want %d", jobsResp.Code, http.StatusOK)
	}
}

func TestCardPeriodRoutesAreSeparateAndLegacyRouteRemoved(t *testing.T) {
	engine := New(&dep.Dep{})
	for _, path := range []string{"/api/cardsperiodrelease", "/api/cardsperiodbirth"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path+"?period_type=unknown", nil))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("周期卡片路由 %s 状态 = %d, want 400: %s", path, response.Code, response.Body.String())
		}
	}
	legacy := httptest.NewRecorder()
	engine.ServeHTTP(legacy, httptest.NewRequest(http.MethodGet, "/api/cardsperiod", nil))
	if legacy.Code != http.StatusNotFound {
		t.Fatalf("旧统一路由状态 = %d, want 404", legacy.Code)
	}
}

func TestAggregatePagesRemovedAndBucketRoutesRemain(t *testing.T) {
	engine := New(&dep.Dep{})
	for _, path := range []string{"/api/movie-agg-all/release", "/api/w-media-agg/birth"} {
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("旧聚合页面 API %s 状态 = %d, want 404", path, response.Code)
		}
	}
	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, path := range []string{"/api/movie-agg-all/release-bucket-list", "/api/w-media-agg/bucket-list", "/api/cardsperiodrelease", "/api/cardsperiodbirth", "/api/triggers/agg"} {
		if !routes[http.MethodGet+" "+path] {
			t.Fatalf("缺少时间桶、周期卡片或回填入口: %s", path)
		}
	}
}

func TestDailyBestRankDayRoutesReplaceLegacySyncRoute(t *testing.T) {
	engine := New(&dep.Dep{})
	routes := make(map[string]struct{})
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	for _, expected := range []string{
		http.MethodGet + " /api/dailybest/rank-days",
		http.MethodPost + " /api/dailybest/rank-days/sync",
	} {
		if _, ok := routes[expected]; !ok {
			t.Fatalf("缺少 DailyBest 日期路由: %s", expected)
		}
	}
	if _, ok := routes[http.MethodPost+" /api/triggers/daily-best-sync"]; ok {
		t.Fatal("旧 DailyBest 全量同步路由仍然存在")
	}
}

func TestSeedsTriggerRoutes(t *testing.T) {
	engine := New(&dep.Dep{})

	pageResp := httptest.NewRecorder()
	pageReq := httptest.NewRequest(http.MethodGet, "/api/triggers/seeds?job_id=seed_job", nil)
	engine.ServeHTTP(pageResp, pageReq)
	if pageResp.Code != http.StatusOK {
		t.Fatalf("GET /api/triggers/seeds status = %d, want %d", pageResp.Code, http.StatusOK)
	}

	var pageData map[string]any
	if err := json.Unmarshal(pageResp.Body.Bytes(), &pageData); err != nil {
		t.Fatalf("decode seeds response: %v", err)
	}
	if pageData["ViewType"] != "crawler-jobs" {
		t.Fatalf("ViewType = %v, want crawler-jobs", pageData["ViewType"])
	}
	if pageData["DefaultTaskType"] != loop.TaskSpiderSeeds {
		t.Fatalf("DefaultTaskType = %v, want %s", pageData["DefaultTaskType"], loop.TaskSpiderSeeds)
	}
	if pageData["OverviewExtraMode"] != "seeds_stages" {
		t.Fatalf("OverviewExtraMode = %v, want seeds_stages", pageData["OverviewExtraMode"])
	}
	if pageData["JobID"] != "seed_job" {
		t.Fatalf("JobID = %v, want seed_job", pageData["JobID"])
	}
	taskButtons, ok := pageData["TaskButtons"].([]any)
	if !ok || len(taskButtons) != 3 {
		t.Fatalf("TaskButtons length = %v, want 3", pageData["TaskButtons"])
	}
}

func TestFC2ImageGenerationRoutesAreReducedToDefaultCoverTask(t *testing.T) {
	engine := New(&dep.Dep{})
	routes := make(map[string]struct{})
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}
	for _, removed := range []string{
		http.MethodPost + " /api/fc2/movies/:fc2_key/screenshot",
		http.MethodPost + " /api/fc2/movies/:fc2_key/contact-sheet",
	} {
		if _, ok := routes[removed]; ok {
			t.Fatalf("旧图片生成路由仍然存在: %s", removed)
		}
	}
	defaultCoverRoute := http.MethodPost + " /api/triggers/fc2-media/default-covers"
	if _, ok := routes[defaultCoverRoute]; !ok {
		t.Fatalf("默认联系表封面任务路由缺失: %s", defaultCoverRoute)
	}
	movieAlbumRoute := http.MethodPost + " /api/fc2/movies/:fc2_key/movie-album-item"
	if _, ok := routes[movieAlbumRoute]; !ok {
		t.Fatalf("FC2电影相册路由缺失: %s", movieAlbumRoute)
	}
	for _, resourceAlbumRoute := range []string{
		http.MethodPost + " /api/fc2/movies/:fc2_key/resource-album-item",
		http.MethodDelete + " /api/fc2/movies/:fc2_key/resource-album-item",
	} {
		if _, ok := routes[resourceAlbumRoute]; !ok {
			t.Fatalf("FC2磁力相册路由缺失: %s", resourceAlbumRoute)
		}
	}
}
