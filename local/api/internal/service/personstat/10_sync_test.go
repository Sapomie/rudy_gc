package personstat

import (
	"reflect"
	"testing"
)

func TestUniquePositiveInt64s(t *testing.T) {
	got := uniquePositiveInt64s([]int64{3, 0, 2, 3, -1, 2, 1})
	want := []int64{3, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("uniquePositiveInt64s() = %v, want %v", got, want)
	}
}

func TestCalcOwnedWMediaRatio(t *testing.T) {
	tests := []struct {
		ownedMovieNumber  int64
		ownedWMediaNumber int64
		want              int64
	}{
		{ownedMovieNumber: 0, ownedWMediaNumber: 0, want: 0},
		{ownedMovieNumber: 4, ownedWMediaNumber: 3, want: 7500},
	}
	for _, tt := range tests {
		if got := calcOwnedWMediaRatio(tt.ownedMovieNumber, tt.ownedWMediaNumber); got != tt.want {
			t.Fatalf("calcOwnedWMediaRatio(%d, %d) = %d, want %d", tt.ownedMovieNumber, tt.ownedWMediaNumber, got, tt.want)
		}
	}
}
