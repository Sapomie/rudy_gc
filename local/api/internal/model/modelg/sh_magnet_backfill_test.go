package modelg

import "testing"

func TestBuildSHMagnetBackfillResources(t *testing.T) {
	resources := buildSHMagnetBackfillResources(
		"abcdef0123456789abcdef0123456789abcdef01\n1111111111111111111111111111111111111111\nABCDEF0123456789ABCDEF0123456789ABCDEF01",
		"magnet:?xt=urn:btih:1111111111111111111111111111111111111111&dn=one\nmagnet:?xt=urn:btih:abcdef0123456789abcdef0123456789abcdef01&dn=two\nmagnet:?xt=urn:btih:2222222222222222222222222222222222222222&dn=three",
	)
	if len(resources) != 3 {
		t.Fatalf("resources length = %d, want 3", len(resources))
	}
	if resources[0].InfoHash != "ABCDEF0123456789ABCDEF0123456789ABCDEF01" || resources[0].MagnetURL == "" {
		t.Fatalf("first resource = %#v", resources[0])
	}
	if resources[1].InfoHash != "1111111111111111111111111111111111111111" || resources[1].MagnetURL == "" {
		t.Fatalf("second resource = %#v", resources[1])
	}
	if resources[2].InfoHash != "2222222222222222222222222222222222222222" || resources[2].MagnetURL == "" {
		t.Fatalf("third resource = %#v", resources[2])
	}
}
