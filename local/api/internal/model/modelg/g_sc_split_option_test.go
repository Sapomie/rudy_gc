package modelg

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestGScSplitOptionModelSchemas(t *testing.T) {
	tests := []struct {
		name        string
		model       any
		tableName   string
		uniqueIndex string
		listIndex   string
	}{
		{
			name:        "kind",
			model:       &GScKindOption{},
			tableName:   GScKindOptionTableName,
			uniqueIndex: "uk_g_sc_option_kind_name",
			listIndex:   "idx_g_sc_option_kind_list",
		},
		{
			name:        "fg",
			model:       &GScFGOption{},
			tableName:   GScFGOptionTableName,
			uniqueIndex: "uk_g_sc_option_fg_name",
			listIndex:   "idx_g_sc_option_fg_list",
		},
		{
			name:        "vessel",
			model:       &GScVesselOption{},
			tableName:   GScVesselOptionTableName,
			uniqueIndex: "uk_g_sc_option_vessl_name",
			listIndex:   "idx_g_sc_option_vessl_list",
		},
	}
	wantColumns := []string{"id", "name", "sort_order", "is_enabled", "sc_event_times", "last_sc_event_time", "created_on", "updated_on"}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := schema.Parse(test.model, &sync.Map{}, schema.NamingStrategy{})
			if err != nil {
				t.Fatalf("parse schema: %v", err)
			}
			if parsed.Table != test.tableName {
				t.Fatalf("table = %q, want %q", parsed.Table, test.tableName)
			}
			if !reflect.DeepEqual(parsed.DBNames, wantColumns) {
				t.Fatalf("columns = %#v, want %#v", parsed.DBNames, wantColumns)
			}
			if len(parsed.PrimaryFields) != 1 || parsed.PrimaryFields[0].DBName != "id" {
				t.Fatalf("primary fields = %#v", parsed.PrimaryFields)
			}
			if len(parsed.Relationships.Relations) != 0 {
				t.Fatalf("unexpected relationships = %#v", parsed.Relationships.Relations)
			}

			uniqueIndex := parsed.LookIndex(test.uniqueIndex)
			if uniqueIndex == nil || uniqueIndex.Class != "UNIQUE" || len(uniqueIndex.Fields) != 1 || uniqueIndex.Fields[0].DBName != "name" {
				t.Fatalf("unique index = %#v", uniqueIndex)
			}
			listIndex := parsed.LookIndex(test.listIndex)
			if listIndex == nil || len(listIndex.Fields) != 2 {
				t.Fatalf("list index = %#v", listIndex)
			}
			gotListColumns := []string{listIndex.Fields[0].DBName, listIndex.Fields[1].DBName}
			wantListColumns := []string{"is_enabled", "sort_order"}
			if !reflect.DeepEqual(gotListColumns, wantListColumns) {
				t.Fatalf("list index columns = %#v, want %#v", gotListColumns, wantListColumns)
			}
		})
	}
}

func TestGScOptionTableRenameSpecs(t *testing.T) {
	want := []gScOptionTableRenameSpec{
		{
			oldTable: "g_sc_kind_option", newTable: "g_sc_option_kind",
			oldUniqueIndex: "uk_g_sc_kind_option_name", newUniqueIndex: "uk_g_sc_option_kind_name",
			oldListIndex: "idx_g_sc_kind_option_list", newListIndex: "idx_g_sc_option_kind_list",
		},
		{
			oldTable: "g_sc_fg_option", newTable: "g_sc_option_fg",
			oldUniqueIndex: "uk_g_sc_fg_option_name", newUniqueIndex: "uk_g_sc_option_fg_name",
			oldListIndex: "idx_g_sc_fg_option_list", newListIndex: "idx_g_sc_option_fg_list",
		},
		{
			oldTable: "g_sc_vessel_option", newTable: "g_sc_option_vessl",
			oldUniqueIndex: "uk_g_sc_vessel_option_name", newUniqueIndex: "uk_g_sc_option_vessl_name",
			oldListIndex: "idx_g_sc_vessel_option_list", newListIndex: "idx_g_sc_option_vessl_list",
		},
	}
	if got := gScOptionTableRenameSpecs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("rename specs = %#v, want %#v", got, want)
	}
}

func TestGScOptionSplitSpecs(t *testing.T) {
	want := []gScOptionSplitSpec{
		{optionType: "kind", tableName: GScKindOptionTableName, idColumn: "kind_option_id"},
		{optionType: "fg", tableName: GScFGOptionTableName, idColumn: "fg_option_id"},
		{optionType: "vessel", tableName: GScVesselOptionTableName, idColumn: "vessel_option_id"},
	}
	if got := gScOptionSplitSpecs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("specs = %#v, want %#v", got, want)
	}
}

func TestQuoteGScOptionIdentifier(t *testing.T) {
	if got, want := quoteGScOptionIdentifier("name`part"), "`name``part`"; got != want {
		t.Fatalf("quoted identifier = %q, want %q", got, want)
	}
}
