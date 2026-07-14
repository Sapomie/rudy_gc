package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/service/loop"
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
}
