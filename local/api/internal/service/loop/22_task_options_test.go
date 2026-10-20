package loop

import "testing"

func TestAutoFetchSiteDefault(t *testing.T) {
	tests := []struct {
		name     string
		taskType string
		want     bool
	}{
		{name: "seeds", taskType: TaskSpiderSeeds, want: false},
		{name: "seed by name", taskType: TaskSpiderSeedByName, want: false},
		{name: "daily best", taskType: TaskSpiderDailyBest, want: true},
		{name: "daily best sync", taskType: TaskSpiderDailyBestSync, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := autoFetchSiteDefault(tt.taskType); got != tt.want {
				t.Fatalf("autoFetchSiteDefault(%q) = %v, want %v", tt.taskType, got, tt.want)
			}
		})
	}
}

func TestParseAutoFetchSiteEnabledExplicitValueOverridesDefault(t *testing.T) {
	enabled, err := parseAutoFetchSiteEnabled("1", false)
	if err != nil {
		t.Fatalf("parse enabled value: %v", err)
	}
	if !enabled {
		t.Fatal("explicit enabled value was ignored")
	}

	disabled, err := parseAutoFetchSiteEnabled("0", true)
	if err != nil {
		t.Fatalf("parse disabled value: %v", err)
	}
	if disabled {
		t.Fatal("explicit disabled value was ignored")
	}
}
