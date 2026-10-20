package sc

import (
	"context"
	"errors"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type scOptionModelStub struct {
	moviex.GScOptionModel
	items      map[int64]*moviex.GScOption
	enabled    []*moviex.GScOption
	all        []*moviex.GScOption
	inserted   *moviex.GScOption
	updated    *moviex.GScOption
	disabledID int64
	writeErr   error
	nextID     int64
}

func scOptionTestDeps(kindModel, fgModel, vesselModel moviex.GScOptionModel) *dep.Dep {
	if kindModel == nil {
		kindModel = &scOptionModelStub{}
	}
	if fgModel == nil {
		fgModel = &scOptionModelStub{}
	}
	if vesselModel == nil {
		vesselModel = &scOptionModelStub{}
	}
	return &dep.Dep{
		ScKindOptionModel:   kindModel,
		ScFGOptionModel:     fgModel,
		ScVesselOptionModel: vesselModel,
	}
}

func (s *scOptionModelStub) FindOne(_ context.Context, id int64) (*moviex.GScOption, error) {
	row := s.items[id]
	if row == nil {
		return nil, moviex.ErrNotFound
	}
	return row, nil
}

func (s *scOptionModelStub) ListEnabled(_ context.Context) ([]*moviex.GScOption, error) {
	return s.enabled, nil
}

func (s *scOptionModelStub) ListAll(_ context.Context) ([]*moviex.GScOption, error) {
	return s.all, nil
}

func (s *scOptionModelStub) Insert(_ context.Context, row *moviex.GScOption) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	if s.nextID <= 0 {
		s.nextID = 100
	}
	row.ID = s.nextID
	copyRow := *row
	s.inserted = &copyRow
	return nil
}

func (s *scOptionModelStub) Update(_ context.Context, row *moviex.GScOption) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	copyRow := *row
	s.updated = &copyRow
	return nil
}

func (s *scOptionModelStub) Disable(_ context.Context, id, updatedOn int64) error {
	if s.writeErr != nil {
		return s.writeErr
	}
	s.disabledID = id
	if row := s.items[id]; row != nil {
		row.IsEnabled = 0
		row.UpdatedOn = updatedOn
	}
	return nil
}

func TestResolveScOptionValidatesTypeAndEnabledState(t *testing.T) {
	kindModel := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		1: {ID: 1, Name: "类型 A", IsEnabled: 1},
	}}
	fgModel := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		2: {ID: 2, Name: "FG A", IsEnabled: 1},
	}}
	vesselModel := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		3: {ID: 3, Name: "Vessel A", IsEnabled: 0},
	}}
	service := NewService(scOptionTestDeps(kindModel, fgModel, vesselModel))

	selected, err := service.resolveScOption(context.Background(), ScOptionTypeKind, 1, 0)
	if err != nil || selected.ID != 1 || selected.Name != "类型 A" {
		t.Fatalf("selected = %#v, err = %v", selected, err)
	}

	if _, err := service.resolveScOption(context.Background(), ScOptionTypeKind, 2, 0); err == nil {
		t.Fatal("wrong option type accepted")
	} else {
		var badRequest *types.BadRequestError
		if !errors.As(err, &badRequest) {
			t.Fatalf("wrong type error = %T %v", err, err)
		}
	}

	if _, err := service.resolveScOption(context.Background(), ScOptionTypeVessel, 3, 0); err == nil {
		t.Fatal("disabled option accepted as a new selection")
	}
	if selected, err := service.resolveScOption(context.Background(), ScOptionTypeVessel, 3, 3); err != nil || selected.ID != 3 {
		t.Fatalf("current disabled selection = %#v, err = %v", selected, err)
	}
}

func TestListScOptionsReturnsEnabledRows(t *testing.T) {
	service := NewService(scOptionTestDeps(&scOptionModelStub{enabled: []*moviex.GScOption{
		{ID: 1, Name: "类型 A", SortOrder: 10, IsEnabled: 1, ScEventTimes: 12, LastScEventTime: 345},
	}}, nil, nil))

	items, err := service.ListScOptions(context.Background(), false)
	if err != nil {
		t.Fatalf("ListScOptions: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 || items[0].OptionType != ScOptionTypeKind || items[0].OptionName != "类型 A" || items[0].ScEventTimes != 12 || items[0].LastScEventTime != 345 {
		t.Fatalf("items = %#v", items)
	}
}

func TestListScOptionsIncludesDisabledRowsForManagement(t *testing.T) {
	service := NewService(scOptionTestDeps(&scOptionModelStub{all: []*moviex.GScOption{
		{ID: 1, Name: "类型 A", IsEnabled: 1},
		{ID: 2, Name: "类型 B", IsEnabled: 0},
	}}, nil, nil))

	items, err := service.ListScOptions(context.Background(), true)
	if err != nil {
		t.Fatalf("ListScOptions: %v", err)
	}
	if len(items) != 2 || items[1].ID != 2 || items[1].IsEnabled != 0 {
		t.Fatalf("items = %#v", items)
	}
}

func TestScOptionCreateUpdateAndDisable(t *testing.T) {
	model := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		7: {ID: 7, Name: "旧名称", IsEnabled: 1, ScEventTimes: 8, LastScEventTime: 99, CreatedOn: 10},
	}}
	service := NewService(scOptionTestDeps(nil, model, nil))

	created, err := service.CreateScOption(context.Background(), &types.ScOptionWriteForm{
		OptionType: " fg ", OptionName: " 名称 ", SortOrder: 3, IsEnabled: 1,
	})
	if err != nil {
		t.Fatalf("CreateScOption: %v", err)
	}
	if created.ID != 100 || model.inserted == nil || model.inserted.Name != "名称" {
		t.Fatalf("created = %#v, inserted = %#v", created, model.inserted)
	}

	updated, err := service.UpdateScOption(context.Background(), 7, &types.ScOptionWriteForm{
		OptionType: ScOptionTypeFG, OptionName: "新名称", SortOrder: 8, IsEnabled: 1,
	})
	if err != nil {
		t.Fatalf("UpdateScOption: %v", err)
	}
	if updated.ID != 7 || model.updated == nil || model.updated.CreatedOn != 10 || model.updated.Name != "新名称" || updated.ScEventTimes != 8 || updated.LastScEventTime != 99 {
		t.Fatalf("updated = %#v, stored = %#v", updated, model.updated)
	}

	disabled, err := service.DisableScOption(context.Background(), ScOptionTypeFG, 7)
	if err != nil {
		t.Fatalf("DisableScOption: %v", err)
	}
	if disabled.IsEnabled != 0 || model.disabledID != 7 {
		t.Fatalf("disabled = %#v, disabledID = %d", disabled, model.disabledID)
	}
}

func TestUpdateScOptionUsesTypeSpecificModel(t *testing.T) {
	fgModel := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		7: {ID: 7, Name: "FG 旧名称", IsEnabled: 1},
	}}
	kindModel := &scOptionModelStub{items: map[int64]*moviex.GScOption{
		7: {ID: 7, Name: "Kind 旧名称", IsEnabled: 1},
	}}
	service := NewService(scOptionTestDeps(kindModel, fgModel, nil))

	updated, err := service.UpdateScOption(context.Background(), 7, &types.ScOptionWriteForm{
		OptionType: ScOptionTypeKind, OptionName: "新名称", IsEnabled: 1,
	})
	if err != nil || updated.OptionType != ScOptionTypeKind || kindModel.updated == nil || kindModel.updated.Name != "新名称" {
		t.Fatalf("updated = %#v, kind = %#v, err = %v", updated, kindModel.updated, err)
	}
	if fgModel.updated != nil || fgModel.items[7].Name != "FG 旧名称" {
		t.Fatalf("FG model changed: %#v", fgModel)
	}
}

func TestCreateScOptionMapsDuplicateError(t *testing.T) {
	model := &scOptionModelStub{writeErr: errors.New("Error 1062: Duplicate entry")}
	service := NewService(scOptionTestDeps(model, nil, nil))

	_, err := service.CreateScOption(context.Background(), &types.ScOptionWriteForm{
		OptionType: ScOptionTypeKind, OptionName: "重复", IsEnabled: 1,
	})
	var badRequest *types.BadRequestError
	if !errors.As(err, &badRequest) {
		t.Fatalf("error = %T %v", err, err)
	}
}
