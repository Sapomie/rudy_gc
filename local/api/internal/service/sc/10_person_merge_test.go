package sc

import (
	"context"
	"database/sql"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type personMergeScModelStub struct {
	moviex.GScModel
	rows    []*moviex.GSc
	updated []*moviex.GSc
}

func (s *personMergeScModelStub) FindAll(context.Context) ([]*moviex.GSc, error) {
	return s.rows, nil
}

func (s *personMergeScModelStub) Update(_ context.Context, row *moviex.GSc) error {
	copyRow := *row
	s.updated = append(s.updated, &copyRow)
	return nil
}

func TestReassignScEventMovieCastPersons(t *testing.T) {
	model := &personMergeScModelStub{rows: []*moviex.GSc{
		{Id: 1, MovieCastPersonId: sql.NullInt64{Int64: 20, Valid: true}},
		{Id: 2, MovieCastPersonId: sql.NullInt64{Int64: 30, Valid: true}},
		{Id: 3},
	}}
	service := NewService(&dep.Dep{ScModel: model})

	if err := service.reassignScEventMovieCastPersons(context.Background(), 10, []int64{20, 21}, 999); err != nil {
		t.Fatalf("reassignScEventMovieCastPersons: %v", err)
	}
	if len(model.updated) != 1 {
		t.Fatalf("updated rows = %#v", model.updated)
	}
	updated := model.updated[0]
	if updated.Id != 1 || !updated.MovieCastPersonId.Valid || updated.MovieCastPersonId.Int64 != 10 || updated.UpdatedOn != 999 {
		t.Fatalf("updated event = %#v", updated)
	}
}
