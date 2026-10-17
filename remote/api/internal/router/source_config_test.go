package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	dailybestservice "rudy_gc/remote/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type sourceConfigRepository struct {
	fakeRepository
	config dailybesttypes.SourceConfig
}

func (r *sourceConfigRepository) FindSourceConfig(context.Context) (dailybesttypes.SourceConfig, error) {
	return r.config, nil
}
func (r *sourceConfigRepository) InsertSourceConfigIfMissing(context.Context, string) error {
	return nil
}
func (r *sourceConfigRepository) SaveSourceConfig(_ context.Context, address string) (dailybesttypes.SourceConfig, error) {
	r.config = dailybesttypes.SourceConfig{JavAddress: address, CreatedOn: 10, UpdatedOn: 20}
	return r.config, nil
}

func TestSourceConfigAPIReadsAndSavesSameConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repository := &sourceConfigRepository{config: dailybesttypes.SourceConfig{JavAddress: "old.example.com"}}
	engine := New(dailybestservice.New(context.Background(), repository, nil, nil))
	for _, test := range []struct {
		method, body, address string
		status                int
	}{
		{http.MethodGet, "", "old.example.com", 200},
		{http.MethodPut, `{"jav_address":" NEW.EXAMPLE.COM "}`, "new.example.com", 200},
		{http.MethodGet, "", "new.example.com", 200},
		{http.MethodPut, `{"jav_address":"https://invalid.example.com"}`, "", 400},
		{http.MethodPut, `{}`, "", 400},
		{http.MethodPut, `invalid`, "", 400},
		{http.MethodGet, "", "new.example.com", 200},
	} {
		request := httptest.NewRequest(test.method, "/api/v1/dailybest/source-config", strings.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
		if test.status == 200 {
			var config dailybesttypes.SourceConfig
			if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil {
				t.Fatal(err)
			}
			if config.JavAddress != test.address || config.TargetProbeURL != dailybesttypes.SourcePageURL(test.address, 1) {
				t.Fatalf("config=%+v", config)
			}
		} else if !strings.Contains(response.Body.String(), `"code":"invalid_source_config"`) {
			t.Fatalf("body=%s", response.Body.String())
		}
	}
}
