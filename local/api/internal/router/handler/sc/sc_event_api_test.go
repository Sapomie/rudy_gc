package sc

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type scEventDeleteAPIModelStub struct {
	moviex.ScEventWriteModel
	result *moviex.ScEventDeleteResult
	err    error
}

type scEventImageAPIScModel struct {
	moviex.GScModel
	row *moviex.GSc
}

func (s *scEventImageAPIScModel) FindOneByName(context.Context, string) (*moviex.GSc, error) {
	return s.row, nil
}

type scEventImageAPIGListModel struct{ moviex.GListModel }

func (*scEventImageAPIGListModel) ListByScName(context.Context, string) ([]*moviex.GList, error) {
	return nil, nil
}

type scEventImageAPIOptions struct{ moviex.GScOptionModel }

func (*scEventImageAPIOptions) ListAll(context.Context) ([]*moviex.GScOption, error) {
	return nil, nil
}

type scEventImageAPIWriter struct{ moviex.ScEventWriteModel }

func (*scEventImageAPIWriter) UpdateEventMeta(_ context.Context, row *moviex.GSc, _ int64, now int64, imagePath *string) ([]*moviex.GList, *moviex.GSc, error) {
	updated := *row
	updated.UpdatedOn = now
	if imagePath != nil {
		updated.ImagePath = *imagePath
	}
	return nil, &updated, nil
}

func TestScEventAPIUpdateImageAndJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	for _, withImage := range []bool{false, true} {
		t.Run(map[bool]string{false: "json", true: "multipart"}[withImage], func(t *testing.T) {
			deps := &dep.Dep{
				ScModel:    &scEventImageAPIScModel{row: &moviex.GSc{Id: 42, Name: "event-a", ImagePath: "original.png"}},
				GListModel: &scEventImageAPIGListModel{}, ScEventWriteModel: &scEventImageAPIWriter{},
				ScKindOptionModel: &scEventImageAPIOptions{}, ScFGOptionModel: &scEventImageAPIOptions{}, ScVesselOptionModel: &scEventImageAPIOptions{},
			}
			deps.Config.SCEvent.ActiveImageDir = t.TempDir()
			engine := gin.New()
			engine.POST("/api/sc-events/:name", NewScEventAPI(deps).Update)
			metadata := `{"duration":12,"remarks":"补图备注"}`
			request := httptest.NewRequest(http.MethodPost, "/api/sc-events/event-a", strings.NewReader(metadata))
			request.Header.Set("Content-Type", "application/json")
			if withImage {
				var body bytes.Buffer
				form := multipart.NewWriter(&body)
				if err := form.WriteField("data", metadata); err != nil {
					t.Fatal(err)
				}
				part, err := form.CreateFormFile("image", "photo.PNG")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := part.Write(imageData.Bytes()); err != nil {
					t.Fatal(err)
				}
				if err := form.Close(); err != nil {
					t.Fatal(err)
				}
				request = httptest.NewRequest(http.MethodPost, "/api/sc-events/event-a", &body)
				request.Header.Set("Content-Type", form.FormDataContentType())
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
			var response struct {
				Item struct {
					ImagePath       string
					Remarks         string
					DurationMinutes int64
				}
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Item.Remarks != "补图备注" || response.Item.DurationMinutes != 12 {
				t.Fatalf("metadata = %#v", response.Item)
			}
			if !withImage {
				if response.Item.ImagePath != "original.png" {
					t.Fatalf("JSON edit changed image: %q", response.Item.ImagePath)
				}
				return
			}
			if filepath.Dir(response.Item.ImagePath) != deps.Config.SCEvent.ActiveImageDir {
				t.Fatalf("image path = %q", response.Item.ImagePath)
			}
			data, err := os.ReadFile(response.Item.ImagePath)
			if err != nil || !bytes.Equal(data, imageData.Bytes()) {
				t.Fatalf("image = %v, err = %v", data, err)
			}
		})
	}
}

func TestScEventAPIRejectsInvalidImageRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, scenario := range []string{"invalid-metadata", "multiple-images"} {
		t.Run(scenario, func(t *testing.T) {
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			metadata := `{"duration":12}`
			if scenario == "invalid-metadata" {
				metadata = "invalid JSON"
			}
			if err := form.WriteField("data", metadata); err != nil {
				t.Fatal(err)
			}
			if scenario == "multiple-images" {
				for _, name := range []string{"one.png", "two.png"} {
					if _, err := form.CreateFormFile("image", name); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}
			engine := gin.New()
			engine.POST("/api/sc-events/:name", NewScEventAPI(&dep.Dep{}).Update)
			request := httptest.NewRequest(http.MethodPost, "/api/sc-events/event-a", &body)
			request.Header.Set("Content-Type", form.FormDataContentType())
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func (s *scEventDeleteAPIModelStub) DeleteEvent(_ context.Context, _ string, _ int64) (*moviex.ScEventDeleteResult, error) {
	return s.result, s.err
}

func TestScEventAPIDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model := &scEventDeleteAPIModelStub{
		result: &moviex.ScEventDeleteResult{
			Event:  &moviex.GSc{Name: "2026-07-28-12-00"},
			Movies: []*moviex.GList{{Id: 1}, {Id: 2}},
		},
	}
	handler := NewScEventAPI(&dep.Dep{ScEventWriteModel: model})
	engine := gin.New()
	engine.DELETE("/api/sc-events/:name", handler.Delete)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/sc-events/2026-07-28-12-00", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		OK                bool
		Deleted           bool
		Name              string
		DeletedMovieCount int64
		DerivedDataFresh  bool
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || !response.Deleted || response.Name != "2026-07-28-12-00" || response.DeletedMovieCount != 2 || !response.DerivedDataFresh {
		t.Fatalf("response = %#v", response)
	}
}

func TestScEventAPIDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewScEventAPI(&dep.Dep{ScEventWriteModel: &scEventDeleteAPIModelStub{err: moviex.ErrNotFound}})
	engine := gin.New()
	engine.DELETE("/api/sc-events/:name", handler.Delete)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/sc-events/missing", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}
