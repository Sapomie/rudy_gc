package crawler

import (
	"reflect"
	"testing"
)

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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeSHFillTargetForumIDs(test.values)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError %v", err, test.wantError)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %#v, want %#v", got, test.want)
			}
		})
	}
}
