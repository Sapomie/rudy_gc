package fetchsite

import (
	"errors"
	"testing"
)

func TestConfigFromUpdateInput(t *testing.T) {
	cfg, err := configFromUpdateInput(UpdateConfigInput{
		SiteCode:       " javbus ",
		BaseURL:        " https://example.com/ ",
		TimeoutSeconds: 30,
		RequestSleepMs: 3000,
		RetrySleepMs:   3000,
		MaxRetryTimes:  45,
		Status:         1,
	}, SiteConfig{SiteName: "JavBus"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SiteCode != "javbus" || cfg.SiteName != "JavBus" || cfg.BaseURL != "https://example.com" {
		t.Fatalf("config = %+v", cfg)
	}
}

func TestConfigFromUpdateInputRejectsInvalidValues(t *testing.T) {
	tests := []UpdateConfigInput{
		{SiteCode: "javbus", SiteName: "JavBus", TimeoutSeconds: 1, MaxRetryTimes: 1},
		{SiteCode: "javbus", SiteName: "JavBus", BaseURL: "ftp://example.com", TimeoutSeconds: 1, MaxRetryTimes: 1},
		{SiteCode: "javbus", SiteName: "JavBus", BaseURL: "https://example.com", MaxRetryTimes: 1},
		{SiteCode: "javbus", SiteName: "JavBus", BaseURL: "https://example.com", TimeoutSeconds: 1},
	}
	for _, input := range tests {
		_, err := configFromUpdateInput(input, SiteConfig{})
		if !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("configFromUpdateInput(%+v) error = %v", input, err)
		}
	}
}
