package sc

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type updateEventScModelStub struct {
	moviex.GScModel
	row *moviex.GSc
}

func (s *updateEventScModelStub) FindOneByName(context.Context, string) (*moviex.GSc, error) {
	return s.row, nil
}

type updateEventGListModelStub struct {
	moviex.GListModel
}

func (*updateEventGListModelStub) ListByScName(context.Context, string) ([]*moviex.GList, error) {
	return []*moviex.GList{}, nil
}

type updateEventWriteModelStub struct {
	moviex.ScEventWriteModel
}

type scPersonByIDsStub struct {
	moviex.CPersonModel
	rows      []*types.Person
	requested []int64
}

func (s *scPersonByIDsStub) FindByIDs(_ context.Context, ids []int64) ([]*types.Person, error) {
	s.requested = append([]int64(nil), ids...)
	return s.rows, nil
}

func (*updateEventWriteModelStub) UpdateEventMeta(_ context.Context, event *moviex.GSc, _ int64, now int64, imagePath *string) ([]*moviex.GList, *moviex.GSc, error) {
	event.UpdatedOn = now
	if imagePath != nil {
		event.ImagePath = *imagePath
	}
	return []*moviex.GList{}, event, nil
}

func TestMapScModelsToTypesResolvesCanonicalOptionNames(t *testing.T) {
	service := NewService(scOptionTestDeps(
		&scOptionModelStub{all: []*moviex.GScOption{{ID: 1, Name: " 已改名 Kind ", IsEnabled: 0}}},
		&scOptionModelStub{all: []*moviex.GScOption{{ID: 2, Name: "FG A", IsEnabled: 1}}},
		&scOptionModelStub{all: []*moviex.GScOption{{ID: 3, Name: "Vessel A", IsEnabled: 1}}},
	))
	rows := []*moviex.GSc{
		{
			Id:             10,
			KindOptionId:   sql.NullInt64{Int64: 1, Valid: true},
			FgOptionId:     sql.NullInt64{Int64: 2, Valid: true},
			VesselOptionId: sql.NullInt64{Int64: 3, Valid: true},
		},
		{
			Id:             11,
			KindOptionId:   sql.NullInt64{Int64: 2, Valid: true},
			VesselOptionId: sql.NullInt64{Int64: 99, Valid: true},
		},
	}

	mapped, err := service.mapScModelsToTypes(context.Background(), rows)
	if err != nil {
		t.Fatalf("mapScModelsToTypes: %v", err)
	}
	if len(mapped) != 2 {
		t.Fatalf("mapped rows = %d", len(mapped))
	}
	if mapped[0].Kind != "已改名 Kind" || mapped[0].Fg != "FG A" || mapped[0].Vessel != "Vessel A" {
		t.Fatalf("canonical option names = %#v", mapped[0])
	}
	if mapped[0].KindOptionID != 1 || mapped[0].FGOptionID != 2 || mapped[0].VesselOptionID != 3 {
		t.Fatalf("option ids = %#v", mapped[0])
	}
	if mapped[1].Kind != "" || mapped[1].Fg != "" || mapped[1].Vessel != "" {
		t.Fatalf("invalid or empty option names must stay empty: %#v", mapped[1])
	}
}

func TestMapScModelsToTypesResolvesMovieCastByPersonID(t *testing.T) {
	deps := scOptionTestDeps(&scOptionModelStub{}, &scOptionModelStub{}, &scOptionModelStub{})
	persons := &scPersonByIDsStub{rows: []*types.Person{{Id: 91, Name: "Actor Name", Chinese: "演员名"}}}
	deps.PersonModel = persons
	service := NewService(deps)

	mapped, err := service.mapScModelsToTypes(context.Background(), []*moviex.GSc{{
		Id: 10, Name: "event-a", MovieCastPersonId: sql.NullInt64{Int64: 91, Valid: true},
	}})
	if err != nil {
		t.Fatalf("mapScModelsToTypes: %v", err)
	}
	if len(mapped) != 1 || mapped[0].MovieCastPersonID != 91 || mapped[0].MovieCast != "演员名" {
		t.Fatalf("mapped actor = %#v", mapped)
	}
	if len(persons.requested) != 1 || persons.requested[0] != 91 {
		t.Fatalf("requested person ids = %#v", persons.requested)
	}
	hrefs := buildScEventMovieCastHrefMap(mapped)
	if hrefs[mapped[0].Name] != "/castdetail/91" {
		t.Fatalf("actor hrefs = %#v", hrefs)
	}
}

func TestUpdateEventMetaReturnsCanonicalOptionNames(t *testing.T) {
	kindOption := &moviex.GScOption{ID: 1, Name: "Kind A", IsEnabled: 1}
	fgOption := &moviex.GScOption{ID: 2, Name: "FG A", IsEnabled: 1}
	vesselOption := &moviex.GScOption{ID: 3, Name: "Vessel A", IsEnabled: 1}
	deps := scOptionTestDeps(
		&scOptionModelStub{items: map[int64]*moviex.GScOption{1: kindOption}, all: []*moviex.GScOption{kindOption}},
		&scOptionModelStub{items: map[int64]*moviex.GScOption{2: fgOption}, all: []*moviex.GScOption{fgOption}},
		&scOptionModelStub{items: map[int64]*moviex.GScOption{3: vesselOption}, all: []*moviex.GScOption{vesselOption}},
	)
	deps.ScModel = &updateEventScModelStub{row: &moviex.GSc{Name: "event-a"}}
	deps.GListModel = &updateEventGListModelStub{}
	deps.ScEventWriteModel = &updateEventWriteModelStub{}
	service := NewService(deps)

	updated, err := service.UpdateEventMeta(context.Background(), &types.ScEventEditForm{
		Name:           "event-a",
		KindOptionID:   1,
		FGOptionID:     2,
		VesselOptionID: 3,
	})
	if err != nil {
		t.Fatalf("UpdateEventMeta: %v", err)
	}
	if updated.Kind != "Kind A" || updated.Fg != "FG A" || updated.Vessel != "Vessel A" {
		t.Fatalf("canonical option names = %#v", updated)
	}
	if updated.KindOptionID != 1 || updated.FGOptionID != 2 || updated.VesselOptionID != 3 {
		t.Fatalf("option ids = %#v", updated)
	}
}

func TestBuildScOrderByUsesCanonicalOptionNames(t *testing.T) {
	tests := []struct {
		field    string
		idColumn string
		table    string
	}{
		{field: "kind", idColumn: "kind_option_id", table: "g_sc_option_kind"},
		{field: "fg", idColumn: "fg_option_id", table: "g_sc_option_fg"},
		{field: "vessel", idColumn: "vessel_option_id", table: "g_sc_option_vessl"},
	}
	for _, test := range tests {
		t.Run(test.field, func(t *testing.T) {
			orderBy := buildScOrderBy(test.field, "asc")
			if !strings.Contains(orderBy, "FROM `"+test.table+"`") || !strings.Contains(orderBy, "`g_sc`.`"+test.idColumn+"`") || !strings.Contains(orderBy, "SELECT `name`") {
				t.Fatalf("orderBy = %q", orderBy)
			}
			if !strings.HasSuffix(orderBy, "ASC, `g_sc`.`sc_time` DESC, `g_sc`.`id` DESC") {
				t.Fatalf("orderBy tie-breakers = %q", orderBy)
			}
		})
	}
}

func TestBuildScOrderByUsesCanonicalPersonName(t *testing.T) {
	orderBy := buildScOrderBy("movie_cast", "asc")
	if !strings.Contains(orderBy, "FROM `c_person`") || !strings.Contains(orderBy, "`g_sc`.`movie_cast_person_id`") {
		t.Fatalf("orderBy = %q", orderBy)
	}
	if strings.Contains(orderBy, "`g_sc`.`movie_cast`") {
		t.Fatalf("orderBy still uses legacy actor name column: %q", orderBy)
	}
}
