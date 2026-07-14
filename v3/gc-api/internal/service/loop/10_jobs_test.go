package loop

import (
	"reflect"
	"testing"
)

func TestNormalizeSHDetailTaskTargetForumIDs(t *testing.T) {
	tests := []struct {
		name          string
		triggerSource string
		values        []string
		want          []string
		wantError     bool
	}{
		{name: "shfill empty", triggerSource: "shfill", wantError: true},
		{name: "shfill single", triggerSource: "shfill", values: []string{" 103 ", "103"}, want: []string{"103"}},
		{name: "shfill multiple", triggerSource: "shfill", values: []string{"36", "103", "36"}, want: []string{"36", "103"}},
		{name: "empty source multiple", values: []string{"36", "103"}, want: []string{"36", "103"}},
		{name: "shpool multiple", triggerSource: "shpool", values: []string{"36", "103", "36"}, want: []string{"36", "103"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeSHDetailTaskTargetForumIDs(tt.triggerSource, tt.values)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %v", err, tt.wantError)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
