package modelg

import "testing"

func TestNormalizeLegacyGScMovieCastName(t *testing.T) {
	tests := map[string]string{
		" 演员甲(26.9) ": "演员甲",
		"演员乙（30）":     "演员乙",
		"演员丙(备注)":     "演员丙(备注)",
		"演员丁":         "演员丁",
	}
	for input, want := range tests {
		if got := normalizeLegacyGScMovieCastName(input); got != want {
			t.Fatalf("normalizeLegacyGScMovieCastName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGScMovieCastResolverUsesStablePrecedenceAndRejectsAmbiguity(t *testing.T) {
	resolver := &gScMovieCastResolver{
		castNames:    map[string]int64{"actor": 10},
		chineseNames: map[string]int64{"actor": 20, "中文名": 21},
		personNames:  map[string]int64{"person": 30},
		aliases:      map[string]int64{"alias": 40, "ambiguous": 0},
	}
	if got := resolver.resolve(" Actor "); got != 10 {
		t.Fatalf("cast precedence person_id = %d, want 10", got)
	}
	if got := resolver.resolve("中文名"); got != 21 {
		t.Fatalf("chinese person_id = %d, want 21", got)
	}
	if got := resolver.resolve("person"); got != 30 {
		t.Fatalf("name person_id = %d, want 30", got)
	}
	if got := resolver.resolve("alias"); got != 40 {
		t.Fatalf("alias person_id = %d, want 40", got)
	}
	if got := resolver.resolve("ambiguous"); got != 0 {
		t.Fatalf("ambiguous person_id = %d, want 0", got)
	}
}
