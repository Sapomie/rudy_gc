package forum

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
