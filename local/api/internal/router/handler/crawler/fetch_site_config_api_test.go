package crawler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsiterepo "rudy_gc/local/api/internal/repo/fetchsite"
	"rudy_gc/local/api/internal/service/spider/fetcher"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type fetchSiteConfigTestModel struct {
	moviex.TFetchSiteModel
	rows map[string]*moviex.TFetchSite
}

func (m *fetchSiteConfigTestModel) ListAll(context.Context) ([]*moviex.TFetchSite, error) {
	rows := make([]*moviex.TFetchSite, 0, len(m.rows))
	for _, row := range m.rows {
		copy := *row
		rows = append(rows, &copy)
	}
	return rows, nil
}

func (m *fetchSiteConfigTestModel) FindOneBySiteCode(_ context.Context, code string) (*moviex.TFetchSite, error) {
	if row := m.rows[code]; row != nil {
		copy := *row
		return &copy, nil
	}
	return nil, moviex.ErrNotFound
}

func (m *fetchSiteConfigTestModel) Update(_ context.Context, row *moviex.TFetchSite) error {
	copy := *row
	m.rows[row.SiteCode] = &copy
	return nil
}

func TestFetchSiteJavlibConfigAPIUpdatesSavedAndRuntimeConfig(t *testing.T) {
	other := &moviex.TFetchSite{Id: 2, SiteCode: "javbus", SiteName: "JavBus", BaseUrl: "https://javbus.test"}
	model := &fetchSiteConfigTestModel{rows: map[string]*moviex.TFetchSite{
		"javlib": {Id: 1, SiteCode: "javlib", SiteName: "Javlib", BaseUrl: "https://old.test", CreatedOn: 10},
		"javbus": other,
	}}
	deps := &dep.Dep{
		FetchSiteModel: model,
		FetchSiteRegistry: fetchsiterepo.NewRegistry(map[string]fetchsitetypes.Config{
			"javlib": {SiteCode: "javlib", SiteName: "Javlib", BaseURL: "https://old.test"},
			"javbus": {SiteCode: "javbus", SiteName: "JavBus", BaseURL: "https://javbus.test"},
		}),
		Fetcher: fetcher.NewFetcher(fetcher.Config{}),
	}
	api := NewFetchSiteConfigAPI(deps)
	engine := gin.New()
	engine.GET("/api/fetch-sites", api.List)
	engine.PUT("/api/fetch-sites/:site_code", api.Update)
	var proxyRequests atomic.Int64
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyRequests.Add(1)
		if r.URL.Host != "changed.test" || r.URL.Path != "/cn/" || r.Header.Get("User-Agent") != "saved-agent" || r.Header.Get("Cookie") != "session=saved" {
			t.Errorf("saved connection parameters not used: %s headers=%v", r.URL, r.Header)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer proxy.Close()

	body, err := json.Marshal(updateFetchSiteConfigRequest{
		SiteName: "Javlib", BaseURL: "http://changed.test/", UserAgent: "saved-agent", Cookie: "session=saved",
		Proxy: proxy.URL, TimeoutSeconds: 30, RequestSleepMs: 3000, RetrySleepMs: 5000, MaxRetryTimes: 20, Status: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/fetch-sites/javlib", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("save status=%d body=%s", response.Code, response.Body.String())
	}
	if row := model.rows["javlib"]; row.BaseUrl != "http://changed.test" || row.CreatedOn != 10 || row.UpdatedOn == 0 {
		t.Fatalf("saved row=%+v", row)
	}
	if model.rows["javbus"] != other {
		t.Fatal("unrelated site was modified")
	}
	cfg, ok := deps.FetchSiteRegistry.Get("javlib")
	if !ok || cfg.BaseURL != "http://changed.test" || cfg.Timeout != 30*time.Second || cfg.Proxy != proxy.URL {
		t.Fatalf("runtime config=%+v", cfg)
	}
	if _, err := deps.Fetcher.GetBySite(context.Background(), "javlib", cfg.BaseURL+"/cn/"); err != nil || proxyRequests.Load() != 1 {
		t.Fatalf("proxy requests=%d err=%v", proxyRequests.Load(), err)
	}

	response = httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/fetch-sites", nil))
	var payload struct {
		Items []fetchSiteConfigResponse `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil || response.Code != http.StatusOK || len(payload.Items) != 2 {
		t.Fatalf("list status=%d body=%s err=%v", response.Code, response.Body.String(), err)
	}
	for _, item := range payload.Items {
		if item.SiteCode == "javlib" && (!item.Persisted || item.BaseURL != "http://changed.test") {
			t.Fatalf("siteconfiglist did not receive saved Javlib row: %+v", item)
		}
	}
}
