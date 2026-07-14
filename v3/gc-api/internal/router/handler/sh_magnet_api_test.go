package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeSHMagnetDownloadState(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "all", raw: "", want: ""},
		{name: "missing", raw: "0", want: "0"},
		{name: "downloaded", raw: "1", want: "1"},
		{name: "removed", raw: "2", want: "2"},
		{name: "trimmed", raw: " 1 ", want: "1"},
		{name: "invalid", raw: "3", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSHMagnetDownloadState(tt.raw); got != tt.want {
				t.Fatalf("normalizeSHMagnetDownloadState(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSHMagnetBatchAddToAlbumValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing album", body: `{"album_name":"","source_row_ids":[1]}`, want: "请选择目标相册"},
		{name: "missing rows", body: `{"album_name":"下载中","source_row_ids":[]}`, want: "请选择要加入相册的条目"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/sh-magnets/batch-add-album", strings.NewReader(tt.body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			(&SHMagnetAPI{}).BatchAddToAlbum(ctx)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			if !strings.Contains(recorder.Body.String(), tt.want) {
				t.Fatalf("body = %s, want %q", recorder.Body.String(), tt.want)
			}
		})
	}
}
