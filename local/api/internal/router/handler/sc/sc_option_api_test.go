package sc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type scOptionAPIModelStub struct {
	moviex.GScOptionModel
	items  []*moviex.GScOption
	all    []*moviex.GScOption
	byID   map[int64]*moviex.GScOption
	nextID int64
}

func scOptionAPIDeps(kindModel, fgModel, vesselModel moviex.GScOptionModel) *dep.Dep {
	if kindModel == nil {
		kindModel = &scOptionAPIModelStub{}
	}
	if fgModel == nil {
		fgModel = &scOptionAPIModelStub{}
	}
	if vesselModel == nil {
		vesselModel = &scOptionAPIModelStub{}
	}
	return &dep.Dep{
		ScKindOptionModel:   kindModel,
		ScFGOptionModel:     fgModel,
		ScVesselOptionModel: vesselModel,
	}
}

func (s *scOptionAPIModelStub) ListEnabled(_ context.Context) ([]*moviex.GScOption, error) {
	return s.items, nil
}

func (s *scOptionAPIModelStub) ListAll(_ context.Context) ([]*moviex.GScOption, error) {
	return s.all, nil
}

func (s *scOptionAPIModelStub) FindOne(_ context.Context, id int64) (*moviex.GScOption, error) {
	row := s.byID[id]
	if row == nil {
		return nil, moviex.ErrNotFound
	}
	copyRow := *row
	return &copyRow, nil
}

func (s *scOptionAPIModelStub) Insert(_ context.Context, row *moviex.GScOption) error {
	if s.nextID <= 0 {
		s.nextID = 10
	}
	row.ID = s.nextID
	if s.byID == nil {
		s.byID = make(map[int64]*moviex.GScOption)
	}
	copyRow := *row
	s.byID[row.ID] = &copyRow
	return nil
}

func (s *scOptionAPIModelStub) Update(_ context.Context, row *moviex.GScOption) error {
	copyRow := *row
	s.byID[row.ID] = &copyRow
	return nil
}

func (s *scOptionAPIModelStub) Disable(_ context.Context, id, updatedOn int64) error {
	row := s.byID[id]
	row.IsEnabled = 0
	row.UpdatedOn = updatedOn
	return nil
}

func TestScOptionAPIListGroupsOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewScOptionAPI(scOptionAPIDeps(
		&scOptionAPIModelStub{items: []*moviex.GScOption{{ID: 1, Name: "类型 A", IsEnabled: 1, ScEventTimes: 12, LastScEventTime: 345}}},
		&scOptionAPIModelStub{items: []*moviex.GScOption{{ID: 1, Name: "FG A", IsEnabled: 1}}},
		&scOptionAPIModelStub{items: []*moviex.GScOption{{ID: 1, Name: "Vessel A", IsEnabled: 1}}},
	))
	engine := gin.New()
	engine.GET("/api/sc-options", handler.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/sc-options", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		OK     bool              `json:"ok"`
		Kind   []*types.ScOption `json:"kind"`
		FG     []*types.ScOption `json:"fg"`
		Vessel []*types.ScOption `json:"vessel"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.OK || len(response.Kind) != 1 || len(response.FG) != 1 || len(response.Vessel) != 1 {
		t.Fatalf("response = %#v", response)
	}
	if response.Kind[0].ID != 1 || response.FG[0].ID != 1 || response.Vessel[0].ID != 1 || response.Kind[0].OptionType != "kind" || response.FG[0].OptionType != "fg" || response.Vessel[0].OptionType != "vessel" {
		t.Fatalf("type-specific identities = %#v", response)
	}
	if response.Kind[0].ScEventTimes != 12 || response.Kind[0].LastScEventTime != 345 {
		t.Fatalf("kind aggregate = %#v", response.Kind[0])
	}
	if strings.Contains(recorder.Body.String(), "optionCode") {
		t.Fatalf("response still contains optionCode: %s", recorder.Body.String())
	}
}

func TestScOptionAPIListIncludesDisabledRowsWhenRequested(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewScOptionAPI(scOptionAPIDeps(&scOptionAPIModelStub{all: []*moviex.GScOption{
		{ID: 1, Name: "类型 A", IsEnabled: 1},
		{ID: 2, Name: "类型 B", IsEnabled: 0},
	}}, nil, nil))
	engine := gin.New()
	engine.GET("/api/sc-options", handler.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/sc-options?includeDisabled=1", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Kind []*types.ScOption `json:"kind"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(response.Kind) != 2 || response.Kind[1].IsEnabled != 0 {
		t.Fatalf("response = %#v", response)
	}
}

func TestScOptionAPICreateUpdateAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	kindModel := &scOptionAPIModelStub{}
	fgModel := &scOptionAPIModelStub{byID: map[int64]*moviex.GScOption{
		7: {ID: 7, Name: "旧名称", IsEnabled: 1},
	}}
	handler := NewScOptionAPI(scOptionAPIDeps(kindModel, fgModel, nil))
	engine := gin.New()
	engine.POST("/api/sc-options/:type", handler.Create)
	engine.PUT("/api/sc-options/:type/:id", handler.Update)
	engine.DELETE("/api/sc-options/:type/:id", handler.Delete)

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/sc-options/kind", bytes.NewBufferString(`{"optionName":"新类型","sortOrder":2,"isEnabled":1}`))
	createRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRecorder.Code, createRecorder.Body.String())
	}

	updateRecorder := httptest.NewRecorder()
	updateRequest := httptest.NewRequest(http.MethodPut, "/api/sc-options/fg/7", bytes.NewBufferString(`{"optionName":"新名称","sortOrder":4,"isEnabled":1}`))
	updateRequest.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(updateRecorder, updateRequest)
	if updateRecorder.Code != http.StatusOK || fgModel.byID[7].Name != "新名称" {
		t.Fatalf("update status = %d, body = %s, row = %#v", updateRecorder.Code, updateRecorder.Body.String(), fgModel.byID[7])
	}

	deleteRecorder := httptest.NewRecorder()
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/sc-options/fg/7", nil)
	engine.ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusOK || fgModel.byID[7].IsEnabled != 0 {
		t.Fatalf("delete status = %d, body = %s, row = %#v", deleteRecorder.Code, deleteRecorder.Body.String(), fgModel.byID[7])
	}
}
