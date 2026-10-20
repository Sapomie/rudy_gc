package spider

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsiterepo "rudy_gc/local/api/internal/repo/fetchsite"
	"rudy_gc/local/api/internal/service/fetchsite"
	"rudy_gc/local/api/internal/service/spider/fetcher"
	"rudy_gc/local/api/internal/types"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type javlibRuntimeTestModel struct {
	moviex.TFetchSiteModel
	row *moviex.TFetchSite
}

func (m *javlibRuntimeTestModel) FindOneBySiteCode(_ context.Context, code string) (*moviex.TFetchSite, error) {
	if m.row == nil || m.row.SiteCode != code {
		return nil, moviex.ErrNotFound
	}
	copy := *m.row
	return &copy, nil
}

func (m *javlibRuntimeTestModel) Insert(_ context.Context, row *moviex.TFetchSite) (sql.Result, error) {
	copy := *row
	copy.Id = 1
	m.row = &copy
	return bestinvTestResult{}, nil
}

func (m *javlibRuntimeTestModel) Update(_ context.Context, row *moviex.TFetchSite) error {
	copy := *row
	m.row = &copy
	return nil
}

func TestJavlibSavedConfigUsedByAllLocalFetchPaths(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	model := &javlibRuntimeTestModel{}
	deps := &dep.Dep{
		Config: config.Config{Fetcher: config.FetcherConf{JavAddress: "stale-yaml.test"}},
		Log:    logger, FetchSiteModel: model,
		FetchSiteRegistry: fetchsiterepo.NewRegistry(map[string]fetchsitetypes.Config{
			fetchsitetypes.CodeJavlib: {SiteCode: fetchsitetypes.CodeJavlib, SiteName: "Javlib"},
		}),
		Fetcher: fetcher.NewFetcher(fetcher.Config{UserAgent: "stale-agent", Cookie: "stale-cookie"}),
	}
	service := &Service{deps: &runtimeDeps{Dep: deps}}
	site := fetchsite.NewService(deps)
	var requests []*http.Request
	var cancel context.CancelFunc
	originalTransport := http.DefaultTransport
	http.DefaultTransport = bestinvRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests = append(requests, request)
		cancel()
		return &http.Response{
			StatusCode: http.StatusOK, Header: make(http.Header), Request: request,
			Body: io.NopCloser(strings.NewReader("<html><body>空的列表</body></html>")),
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	paths := []struct {
		name string
		path string
		run  func(context.Context) error
	}{
		{"详情", "/cn/", func(ctx context.Context) error {
			return service.fetchAndSaveDetail(ctx, &types.Item{Name: "TEST-001", JavId: "jav-id"})
		}},
		{"榜单", "/cn/vl_bestrated.php", func(ctx context.Context) error {
			_, err := service.fetchBestinvContentByRated(ctx, consts.BestCategoryMonth, 2)
			return err
		}},
		{"种子列表", "/cn/vl_searchbyid.php", func(ctx context.Context) error {
			return service.fetchAndSaveInventory(ctx, nameTypePrefix, "TEST", "vl_searchbyid.php?&keyword=TEST", 2)
		}},
	}
	for _, host := range []string{"first.test", "changed.test"} {
		_, err := site.UpdateConfig(context.Background(), fetchsite.UpdateConfigInput{
			SiteCode: "javlib", BaseURL: "https://" + host + "/", UserAgent: host + "-agent",
			Cookie: "session=" + host, TimeoutSeconds: 30, RequestSleepMs: 3000,
			RetrySleepMs: 5000, MaxRetryTimes: 20, Status: 1,
		})
		if err != nil {
			t.Fatal(err)
		}
		if model.row.BaseUrl != "https://"+host {
			t.Fatalf("saved URL = %q", model.row.BaseUrl)
		}
		for _, path := range paths {
			t.Run(host+"/"+path.name, func(t *testing.T) {
				ctx, stop := context.WithCancel(context.Background())
				defer stop()
				cancel = stop
				before := len(requests)
				// 空页面加取消只检查请求目标，避免写入抓取数据及触发重试等待。
				if err := path.run(ctx); err == nil {
					t.Fatal("expected empty page or cancellation")
				}
				if len(requests) != before+1 {
					t.Fatalf("request count = %d, want 1", len(requests)-before)
				}
				request := requests[before]
				if request.URL.Host != host || request.URL.Path != path.path ||
					request.Header.Get("User-Agent") != host+"-agent" || request.Header.Get("Cookie") != "session="+host {
					t.Fatalf("saved configuration not used: %s headers=%v", request.URL, request.Header)
				}
				if path.name == "详情" && request.URL.Query().Get("v") != "jav-id" ||
					path.name != "详情" && request.URL.Query().Get("page") != "2" {
					t.Fatalf("query parameters changed: %s", request.URL)
				}
			})
		}
	}

	deps.FetchSiteRegistry = fetchsiterepo.NewRegistry(fetchsiterepo.Load(context.Background(), deps.Config, model, nil))
	if got, err := service.javlibURL("?v=restart"); err != nil || got != "https://changed.test/cn/?v=restart" {
		t.Fatalf("database config after reload = %q, err=%v", got, err)
	}
	if cfg, err := site.GetConfig("javlib"); err != nil || cfg.Timeout != 30*time.Second {
		t.Fatalf("saved timeout = %v, err=%v", cfg.Timeout, err)
	}
}
