package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/router"
)

type shMovieLinkResponse struct {
	Rows []struct {
		MovieName string
		MovieHref string
	}
}

func TestSHListsResolveMovieLinksFromCurrentMoviesIntegration(t *testing.T) {
	configPath := strings.TrimSpace(os.Getenv("RUDY_GC_INTEGRATION_CONFIG"))
	if configPath == "" {
		t.Skip("设置 RUDY_GC_INTEGRATION_CONFIG 后执行 SH 影片链接真实数据库测试")
	}

	gin.SetMode(gin.TestMode)
	cfg := config.MustLoad(configPath)
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		t.Fatalf("构建依赖失败: %v", err)
	}
	engine := router.New(deps)

	tests := []struct {
		endpoint  string
		movieName string
	}{
		{endpoint: "/api/sh-magnets?keyword=OFES-029&ps=50", movieName: "OFES-029"},
		{endpoint: "/api/fetch-site-sehuatang-sh-list?keyword=OFES-029&ps=50", movieName: "OFES-029"},
		{endpoint: "/api/fetch-site-sehuatang-sh-list?keyword=PRST-001&ps=50", movieName: "PRST-001"},
	}
	for _, testCase := range tests {
		endpoint := testCase.endpoint
		movieName := testCase.movieName
		t.Run(endpoint, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, endpoint, nil)
			engine.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("%s status = %d, body = %s", endpoint, recorder.Code, recorder.Body.String())
			}

			var response shMovieLinkResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("%s 响应解析失败: %v", endpoint, err)
			}
			found := false
			for _, row := range response.Rows {
				if row.MovieName != movieName {
					continue
				}
				found = true
				wantHref := "/moviedetail/" + movieName
				if row.MovieHref != wantHref {
					t.Fatalf("%s MovieHref = %q, want %q", endpoint, row.MovieHref, wantHref)
				}
			}
			if !found {
				t.Fatalf("%s 未返回 %s", endpoint, movieName)
			}
		})
	}
}
