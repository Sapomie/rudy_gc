package handler

import (
	"reflect"
	"testing"
)

func TestNormalizeSHThreadDeleteTIDs(t *testing.T) {
	input := []string{" 3619676 ", "", "3619677", "3619676", "  "}
	want := []string{"3619676", "3619677"}
	if got := normalizeSHThreadDeleteTIDs(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeSHThreadDeleteTIDs() = %#v, want %#v", got, want)
	}
}

func TestNormalizeSHFillTargetForumIDs(t *testing.T) {
	tests := []struct {
		name      string
		values    []string
		want      []string
		wantError bool
	}{
		{name: "empty", wantError: true},
		{name: "single", values: []string{" 36 ", "36"}, want: []string{"36"}},
		{name: "multiple", values: []string{"36", "103", "36"}, want: []string{"36", "103"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSHFillTargetForumIDs(tt.values)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %v", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
