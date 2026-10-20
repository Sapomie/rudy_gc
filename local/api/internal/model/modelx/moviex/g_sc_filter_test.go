package moviex

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"

	"rudy_gc/local/api/internal/types"
)

func TestApplyGScListFilterCombinesOptionIDs(t *testing.T) {
	query, args, err := applyGScListFilter(
		squirrel.Select("id").From("g_sc"),
		types.ScEventListFilter{
			StartTime:      100,
			EndTime:        200,
			KindOptionID:   3,
			FGOptionID:     4,
			VesselOptionID: 5,
		},
	).ToSql()
	if err != nil {
		t.Fatalf("ToSql: %v", err)
	}
	for _, condition := range []string{
		"sc_time >= ?",
		"sc_time < ?",
		"kind_option_id = ?",
		"fg_option_id = ?",
		"vessel_option_id = ?",
	} {
		if !strings.Contains(query, condition) {
			t.Fatalf("query %q missing %q", query, condition)
		}
	}
	if want := []any{int64(100), int64(200), int64(3), int64(4), int64(5)}; !reflect.DeepEqual(args, want) {
		t.Fatalf("args = %#v, want %#v", args, want)
	}
}

func TestApplyGScListFilterIgnoresEmptyOptionIDs(t *testing.T) {
	query, args, err := applyGScListFilter(
		squirrel.Select("id").From("g_sc"),
		types.ScEventListFilter{},
	).ToSql()
	if err != nil {
		t.Fatalf("ToSql: %v", err)
	}
	if strings.Contains(query, " WHERE ") || len(args) != 0 {
		t.Fatalf("empty filter query = %q, args = %#v", query, args)
	}
}
