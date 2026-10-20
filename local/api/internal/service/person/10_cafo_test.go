package person

import (
	"context"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type cafoServicePersonModelStub struct {
	moviex.CPersonModel
	row *moviex.CPerson
	err error
}

func (s *cafoServicePersonModelStub) FindOne(context.Context, int64) (*moviex.CPerson, error) {
	return s.row, s.err
}

type cafoServiceCafoModelStub struct {
	moviex.CCafoModel
	rows          []*moviex.CCafoProfile
	updatedCafo   *moviex.CCafoProfile
	updatedPerson *moviex.CPerson
	personID      int64
	cafoID        int64
}

func (s *cafoServiceCafoModelStub) ListByPersonID(context.Context, int64) ([]*moviex.CCafoProfile, error) {
	return s.rows, nil
}

func (s *cafoServiceCafoModelStub) UpdateProfileForPerson(
	_ context.Context,
	personID, cafoID int64,
	data *moviex.CCafoProfileUpdate,
) (*moviex.CCafoProfile, *moviex.CPerson, error) {
	s.personID = personID
	s.cafoID = cafoID
	s.updatedCafo.Name = data.Name
	s.updatedCafo.Biography = data.Biography
	s.updatedPerson.Name = data.Name
	return s.updatedCafo, s.updatedPerson, nil
}

func TestNormalizeCafoProfileWrite(t *testing.T) {
	got, err := normalizeCafoProfileWrite(&types.CafoProfileWriteForm{
		Name:      "  Actor Name  ",
		Alias:     " Alias ",
		Chinese:   " 中文名 ",
		BirthDay:  -86400,
		Height:    168,
		Cup:       " C ",
		Bwh:       " 90-60-88 ",
		Avatar:    " /images/actor.jpg ",
		Biography: " 人物简介 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Actor Name" || got.Alias != "Alias" || got.Chinese != "中文名" {
		t.Fatalf("文本字段未正确整理：%+v", got)
	}
	if got.BirthDay != -86400 || got.Height != 168 || got.Cup != "C" || got.Bwh != "90-60-88" || got.Avatar != "/images/actor.jpg" {
		t.Fatalf("资料字段未正确保留：%+v", got)
	}
	if got.Biography != "人物简介" {
		t.Fatalf("人物简介未正确整理：%+v", got)
	}
}

func TestNormalizeCafoProfileWriteRequiresName(t *testing.T) {
	if _, err := normalizeCafoProfileWrite(&types.CafoProfileWriteForm{Name: "  "}); err == nil {
		t.Fatal("空 Name 应返回参数错误")
	}
}

func TestListCafoProfilesKeepsEveryLinkedRow(t *testing.T) {
	cafoModel := &cafoServiceCafoModelStub{rows: []*moviex.CCafoProfile{
		{Id: 7, PersonId: 11, Name: "Actor A"},
		{Id: 8, PersonId: 11, Name: "Actor B"},
	}}
	service := NewService(&dep.Dep{
		PersonModel: &cafoServicePersonModelStub{row: &moviex.CPerson{Id: 11}},
		CafoModel:   cafoModel,
	})

	rows, err := service.ListCafoProfiles(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Id != 7 || rows[1].Id != 8 {
		t.Fatalf("关联资料列表不完整：%+v", rows)
	}
}

func TestUpdateCafoProfileReturnsSynchronizedPerson(t *testing.T) {
	cafoModel := &cafoServiceCafoModelStub{
		updatedCafo:   &moviex.CCafoProfile{Id: 7, PersonId: 11},
		updatedPerson: &moviex.CPerson{Id: 11, MovieNumber: 9},
	}
	service := NewService(&dep.Dep{CafoModel: cafoModel})

	cafo, person, err := service.UpdateCafoProfile(context.Background(), 11, 7, &types.CafoProfileWriteForm{
		Name:      "Actor",
		Biography: "人物简介",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cafoModel.personID != 11 || cafoModel.cafoID != 7 {
		t.Fatalf("更新范围错误：person=%d cafo=%d", cafoModel.personID, cafoModel.cafoID)
	}
	if cafo.Name != "Actor" || person.Name != "Actor" || person.MovieNumber != 9 {
		t.Fatalf("同步结果错误：cafo=%+v person=%+v", cafo, person)
	}
	if cafo.Biography != "人物简介" {
		t.Fatalf("人物简介返回错误：cafo=%+v", cafo)
	}
}
