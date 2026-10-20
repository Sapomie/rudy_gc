package wmediaagg

import (
	"reflect"
	"testing"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func TestBucketDaysFromMediaRowsIncludesBirthAndRemoveDays(t *testing.T) {
	birthTime := time.Date(2026, time.July, 1, 14, 30, 0, 0, time.Local)
	removeTime := time.Date(2026, time.July, 3, 8, 15, 0, 0, time.Local)

	got := BucketDaysFromMediaRows(
		nil,
		&moviex.WMedia{BirthTime: birthTime.Unix(), RemoveTime: removeTime.Unix()},
		&moviex.WMedia{BirthTime: birthTime.Add(time.Hour).Unix()},
	)
	want := []int64{
		time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local).Unix(),
		time.Date(2026, time.July, 3, 0, 0, 0, 0, time.Local).Unix(),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BucketDaysFromMediaRows() = %v, want %v", got, want)
	}
}

func TestAppendUniqueBucketDaysKeepsPreviousRemoveDayForRestore(t *testing.T) {
	previous := &moviex.WMedia{
		BirthTime:  time.Date(2026, time.July, 1, 10, 0, 0, 0, time.Local).Unix(),
		RemoveTime: time.Date(2026, time.July, 3, 10, 0, 0, 0, time.Local).Unix(),
	}
	current := &moviex.WMedia{
		BirthTime: time.Date(2026, time.July, 5, 10, 0, 0, 0, time.Local).Unix(),
	}

	got := appendUniqueBucketDays(BucketDaysFromMediaRows(previous), BucketDaysFromMediaRows(current)...)
	want := []int64{
		time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local).Unix(),
		time.Date(2026, time.July, 3, 0, 0, 0, 0, time.Local).Unix(),
		time.Date(2026, time.July, 5, 0, 0, 0, 0, time.Local).Unix(),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("appendUniqueBucketDays() = %v, want %v", got, want)
	}
}
