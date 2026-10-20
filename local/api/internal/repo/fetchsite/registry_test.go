package fetchsite

import (
	"testing"

	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

func TestRegistryCopiesAndNormalizesConfigs(t *testing.T) {
	input := map[string]fetchsitetypes.Config{
		fetchsitetypes.CodeJavbus: {
			SiteCode: fetchsitetypes.CodeJavbus,
			BaseURL:  "https://example.com",
		},
	}
	registry := NewRegistry(input)
	input[fetchsitetypes.CodeJavbus] = fetchsitetypes.Config{}

	got, ok := registry.Get(" javbus ")
	if !ok || got.BaseURL != "https://example.com" {
		t.Fatalf("Get() = (%+v, %v)", got, ok)
	}

	if !registry.Set(fetchsitetypes.Config{SiteCode: " sukebei ", BaseURL: " https://sukebei.example/ "}) {
		t.Fatal("Set() = false")
	}
	got, ok = registry.Get(fetchsitetypes.CodeSukebei)
	if !ok || got.BaseURL != "https://sukebei.example" {
		t.Fatalf("Get() after Set() = (%+v, %v)", got, ok)
	}
}
