package modelx

import (
	"testing"

	"rudy-gc-api/internal/types"
)

func TestPageSummariesCoverLegacyPages(t *testing.T) {
	summaries := PageSummaries()
	if len(summaries) < 34 {
		t.Fatalf("expected legacy page coverage, got %d summaries", len(summaries))
	}

	required := []string{
		"casts",
		"cast-detail",
		"crawl-records",
		"sc-events",
		"medias",
		"w-agg-events",
		"e-items",
		"torrent-albums",
		"movie-albums",
		"d-seeds",
		"fetch-site-javbus-list",
		"fetch-site-sukebei-list",
		"fetch-site-sehuatang-list",
		"wdir",
		"w-media-agg-birth",
		"w-media-agg-bucket-list",
		"movie-agg-all-release",
		"movie-release-bucket-list",
		"triggers",
		"crawler-tasks",
		"triggers-dailybest",
		"triggers-seeds",
		"triggers-post-process",
		"triggers-backfill",
		"triggers-fetch-site",
		"triggers-fetch-sehuatang",
		"triggers-media",
		"triggers-media-rescan",
		"triggers-media-rollback",
		"triggers-sc-media",
		"sc-pick-smart-media",
		"triggers-sc",
		"triggers-agg",
	}

	for _, key := range required {
		if _, ok := FindPageConfig(key); !ok {
			t.Fatalf("missing page config %q", key)
		}
	}
}

func TestPageConfigColumnsHaveSafeSQLExpressions(t *testing.T) {
	for _, cfg := range pageConfigs() {
		if cfg.Key == "" || cfg.Title == "" || cfg.LegacyPath == "" {
			t.Fatalf("page config has empty identity: %#v", cfg)
		}
		if cfg.Kind == pageKindOperation {
			continue
		}
		if cfg.BaseSQL == "" {
			t.Fatalf("%s has empty base sql", cfg.Key)
		}
		if cfg.DefaultOrderBy != "" && cfg.SortColumns[cfg.DefaultOrderBy] == "" {
			t.Fatalf("%s default sort %q has no sql expression", cfg.Key, cfg.DefaultOrderBy)
		}
		for _, column := range cfg.Columns {
			if column == nil {
				t.Fatalf("%s has nil column", cfg.Key)
			}
			if cfg.SortColumns[column.Key] == "" {
				t.Fatalf("%s column %q has no sql expression", cfg.Key, column.Key)
			}
		}
	}
}

func TestPageConfigsDoNotExposeGenericPendingActions(t *testing.T) {
	for _, cfg := range pageConfigs() {
		for _, action := range cfg.Actions {
			if action == nil {
				t.Fatalf("%s has nil action", cfg.Key)
			}
			if action.Path == "/api/gc/v2/page-actions" || action.Path == "/api/page-actions" {
				t.Fatalf("%s action %q uses forbidden generic page action endpoint", cfg.Key, action.Key)
			}
			if action.Disabled {
				t.Fatalf("%s action %q is disabled; hidden pending actions must not be exposed", cfg.Key, action.Key)
			}
		}
	}
}

func TestBuildWhereKeepsExplicitZeroID(t *testing.T) {
	zero := int64(0)
	cfg, ok := FindPageConfig("casts")
	if !ok {
		t.Fatal("missing casts page config")
	}
	whereSQL, args := buildWhere(cfg, &types.PageListRequest{ID: &zero})
	if whereSQL == "" {
		t.Fatal("expected explicit id=0 to produce a where clause")
	}
	if len(args) != 1 || args[0] != zero {
		t.Fatalf("expected id arg 0, got %#v", args)
	}
}

func TestNormalizeOrder(t *testing.T) {
	if got := normalizeOrder("asc"); got != "ASC" {
		t.Fatalf("asc normalized to %q", got)
	}
	if got := normalizeOrder("desc"); got != "DESC" {
		t.Fatalf("desc normalized to %q", got)
	}
	if got := normalizeOrder("drop table"); got != "DESC" {
		t.Fatalf("unsafe order normalized to %q", got)
	}
}
