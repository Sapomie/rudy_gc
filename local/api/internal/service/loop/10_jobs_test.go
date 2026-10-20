package loop

import (
	"reflect"
	"testing"

	"rudy_gc/local/api/internal/service/fetchsehuatang"
)

func TestNormalizeSHFetchAllExistingPolicy(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "default", raw: "", want: fetchsehuatang.SHExistingPolicySkip},
		{name: "blank", raw: "  ", want: fetchsehuatang.SHExistingPolicySkip},
		{name: "skip", raw: " SKIP ", want: fetchsehuatang.SHExistingPolicySkip},
		{name: "upsert", raw: " UPSERT ", want: fetchsehuatang.SHExistingPolicyUpsert},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeSHFetchAllExistingPolicy(test.raw); got != test.want {
				t.Fatalf("normalizeSHFetchAllExistingPolicy(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

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
