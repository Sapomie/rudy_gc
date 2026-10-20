package wmediaagg

import (
	"context"
	"testing"
	"time"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type weekBirthMedia struct {
	moviex.WMediaModel
	ranges map[string]int
}

func (m *weekBirthMedia) CalcBirthBucket(_ context.Context, start, end int64) (*moviex.WMediaBirthBucketCalc, error) {
	m.ranges[time.Unix(start, 0).In(time.Local).Format(time.DateOnly)+":"+time.Unix(end, 0).In(time.Local).Format(time.DateOnly)]++
	return &moviex.WMediaBirthBucketCalc{AddedCount: 1, MediaCount: 1}, nil
}
func (*weekBirthMedia) CalcTopCastsByBirthRange(context.Context, int64, int64, int) ([]*moviex.WMediaBirthTopCalc, error) {
	return nil, nil
}
func (*weekBirthMedia) CalcTopDirectorsByBirthRange(context.Context, int64, int64, int) ([]*moviex.WMediaBirthTopCalc, error) {
	return nil, nil
}
func (*weekBirthMedia) CalcTopLabelsByBirthRange(context.Context, int64, int64, int) ([]*moviex.WMediaBirthTopCalc, error) {
	return nil, nil
}
func (*weekBirthMedia) CalcTopPrefixesByBirthRange(context.Context, int64, int64, int) ([]*moviex.WMediaBirthTopCalc, error) {
	return nil, nil
}

type weekBirthTops struct{ moviex.WMediaBirthTopStatModel }

func (*weekBirthTops) DeleteByScopeAggType(context.Context, string, string) error { return nil }

func TestBirthWeeksDeduplicateAndIncludeWholeCrossYearWeek(t *testing.T) {
	media := &weekBirthMedia{ranges: map[string]int{}}
	svc := NewService(&dep.Dep{WMediaModel: media, WMediaBirthBucketStatModel: &birthChangesBucketModel{}, WMediaBirthTopStatModel: &weekBirthTops{}})
	days := []int64{time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local).Unix(), time.Date(2021, 1, 3, 0, 0, 0, 0, time.Local).Unix(), time.Date(2021, 1, 4, 0, 0, 0, 0, time.Local).Unix()}
	if _, err := svc.rebuildByBucketDays(context.Background(), days, 0); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"2020-12-28:2021-01-03", "2021-01-04:2021-01-10"} {
		if got := media.ranges[key]; got != 1 {
			t.Fatalf("周范围 %s 重建次数 = %d, want 1", key, got)
		}
	}
	scope := weekScopeFromBucketDay(days[0])
	if scope.Key != "week:2020-W53" || scope.Year != 2020 {
		t.Fatalf("ISO 周归属错误: %+v", scope)
	}
}
