package moviex

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
)

func TestApplySHMagnetListFilterDownloadState(t *testing.T) {
	query, args, err := applyShMagnetListFilter(squirrel.Select("m.id").From("sh_magnet m"), ShMagnetListFilter{
		DownloadState: "1",
	}).ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "m.download_state = ?") {
		t.Fatalf("query missing download_state condition: %s", query)
	}
	if !reflect.DeepEqual(args, []any{"1"}) {
		t.Fatalf("args = %#v, want %#v", args, []any{"1"})
	}
}

func TestApplySHMagnetListFilterIgnoresInvalidDownloadState(t *testing.T) {
	query, args, err := applyShMagnetListFilter(squirrel.Select("m.id").From("sh_magnet m"), ShMagnetListFilter{
		DownloadState: "invalid",
	}).ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(query, "download_state") {
		t.Fatalf("query contains invalid download_state condition: %s", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v, want empty", args)
	}
}
