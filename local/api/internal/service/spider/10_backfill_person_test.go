package spider

import (
	"context"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type personBackfillCafoModelStub struct {
	moviex.CCafoModel
	cafoID   int64
	personID int64
}

func (s *personBackfillCafoModelStub) SetPersonID(_ context.Context, cafoID, personID int64) error {
	s.cafoID = cafoID
	s.personID = personID
	return nil
}

func TestPersonBackfillLookupMatchesUniqueCafoPerson(t *testing.T) {
	person := &moviex.CPerson{Id: 10, Name: "Actor Name", Alias: "演员别名", Chinese: "中文名"}
	lookup := newPersonBackfillLookup([]*moviex.CPerson{person})

	got, ambiguous := lookup.matchCafo(&moviex.CCafo{Name: " actor name ", Chinese: "中文名"})
	if ambiguous {
		t.Fatal("唯一人物被判定为存在歧义")
	}
	if got == nil || got.Id != person.Id {
		t.Fatalf("匹配人物 = %#v，期望 person_id=%d", got, person.Id)
	}
}

func TestPersonBackfillLookupLeavesDuplicateNameUnresolved(t *testing.T) {
	lookup := newPersonBackfillLookup([]*moviex.CPerson{
		{Id: 10, Name: "同名演员", Chinese: "中文甲"},
		{Id: 20, Name: "同名演员", Chinese: "中文乙"},
	})

	got, ambiguous := lookup.matchCafo(&moviex.CCafo{Name: "同名演员", Chinese: "中文甲"})
	if !ambiguous {
		t.Fatal("同名人物未被判定为存在歧义")
	}
	if got != nil {
		t.Fatalf("歧义匹配返回了人物：%#v", got)
	}
}

func TestPersonBackfillLookupRejectsConflictingExactKeys(t *testing.T) {
	lookup := newPersonBackfillLookup([]*moviex.CPerson{
		{Id: 10, Name: "演员甲", Chinese: "中文甲"},
		{Id: 20, Name: "演员乙", Chinese: "中文乙"},
	})

	got, ambiguous := lookup.matchCafo(&moviex.CCafo{Name: "演员甲", Chinese: "中文乙"})
	if !ambiguous {
		t.Fatal("冲突的姓名和中文名未被判定为存在歧义")
	}
	if got != nil {
		t.Fatalf("冲突匹配返回了人物：%#v", got)
	}
}

func TestBackfillPersonFromCafoPersistsUniqueMatch(t *testing.T) {
	person := &moviex.CPerson{Id: 10, Name: "演员名", Alias: "演员别名", Chinese: "中文名", BirthDay: 123}
	lookup := newPersonBackfillLookup([]*moviex.CPerson{person})
	cafoModel := &personBackfillCafoModelStub{}
	service := &Service{deps: &runtimeDeps{Dep: &dep.Dep{CafoModel: cafoModel}}}

	created, updated, linked, ambiguous, err := service.backfillPersonFromCafo(
		context.Background(),
		lookup,
		&moviex.CCafo{Id: 7, Name: person.Name, Alias: person.Alias, Chinese: person.Chinese, BirthDay: person.BirthDay},
		0,
		999,
	)
	if err != nil {
		t.Fatalf("backfillPersonFromCafo: %v", err)
	}
	if created || updated || !linked || ambiguous {
		t.Fatalf("created=%v, updated=%v, linked=%v, ambiguous=%v", created, updated, linked, ambiguous)
	}
	if cafoModel.cafoID != 7 || cafoModel.personID != person.Id {
		t.Fatalf("persisted link = %d -> %d", cafoModel.cafoID, cafoModel.personID)
	}
}
