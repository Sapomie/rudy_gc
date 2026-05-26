package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/dep"

	"github.com/gin-gonic/gin"
)

func TestRouterHealthAndPagesWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(dep.Build(config.Config{}))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "health", method: http.MethodGet, path: "/api/gc/v2/healthz"},
		{name: "summaries", method: http.MethodGet, path: "/api/gc/v2/pages"},
		{name: "generic page", method: http.MethodGet, path: "/api/gc/v2/pages/casts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s %s returned status %d: %s", tt.method, tt.path, rec.Code, rec.Body.String())
			}
			var envelope struct {
				Data  any `json:"data"`
				Error any `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("response is not json: %v", err)
			}
			if envelope.Data == nil || envelope.Error != nil {
				t.Fatalf("response is not a successful envelope: %s", rec.Body.String())
			}
		})
	}
}

func TestRouterDoesNotExposeGenericPageActions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(dep.Build(config.Config{}))

	req := httptest.NewRequest(http.MethodPost, "/api/gc/v2/page-actions", strings.NewReader(`{"action":"triggers-dailybest-start"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("generic page action endpoint should be absent, got status %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRouterRegistersLegacyVolumeStaticRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(dep.Build(config.Config{}))

	expected := []string{
		"/Volumes/Expansion/*filepath",
		"/Volumes/Getea/*filepath",
		"/Volumes/movie-un/*filepath",
		"/Volumes/T7/data/*filepath",
	}
	for _, path := range expected {
		if !hasRoute(engine, http.MethodGet, path) {
			t.Fatalf("static route %s is not registered", path)
		}
		if !hasRoute(engine, http.MethodHead, path) {
			t.Fatalf("static HEAD route %s is not registered", path)
		}
	}
}

func hasRoute(engine *gin.Engine, method string, path string) bool {
	for _, route := range engine.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
