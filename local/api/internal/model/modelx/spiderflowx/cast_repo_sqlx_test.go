package spiderflowx

import (
	"context"
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type castBirthCastModelStub struct {
	moviex.AmCastModel
	row *moviex.AmCast
	err error
}

func (s *castBirthCastModelStub) FindOneByName(context.Context, string) (*moviex.AmCast, error) {
	return s.row, s.err
}

type castBirthPersonModelStub struct {
	moviex.CPersonModel
	row *moviex.CPerson
	err error
}

func (s *castBirthPersonModelStub) FindOne(context.Context, int64) (*moviex.CPerson, error) {
	return s.row, s.err
}

func TestFindPersonBirthByNameUsesBoundPerson(t *testing.T) {
	repo := &CastRepoSqlx{
		m:  &castBirthCastModelStub{row: &moviex.AmCast{Id: 1, PersonId: 9, Name: "演员"}},
		pm: &castBirthPersonModelStub{row: &moviex.CPerson{Id: 9, BirthDay: 123456}},
	}

	birth, found, err := repo.FindPersonBirthByName(context.Background(), "演员")
	if err != nil {
		t.Fatalf("FindPersonBirthByName: %v", err)
	}
	if !found || birth != 123456 {
		t.Fatalf("birth=%d, found=%v", birth, found)
	}
}

func TestFindPersonBirthByNameHandlesMissingCast(t *testing.T) {
	repo := &CastRepoSqlx{
		m:  &castBirthCastModelStub{err: moviex.ErrNotFound},
		pm: &castBirthPersonModelStub{},
	}

	birth, found, err := repo.FindPersonBirthByName(context.Background(), "未知演员")
	if err != nil {
		t.Fatalf("FindPersonBirthByName: %v", err)
	}
	if found || birth != 0 {
		t.Fatalf("birth=%d, found=%v", birth, found)
	}
}
