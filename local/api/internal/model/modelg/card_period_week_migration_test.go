package modelg

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestCardPeriodDirectoryIndexes(t *testing.T) {
	for _, tc := range []struct {
		model   any
		index   string
		columns []string
	}{
		{&MovieReleaseBucketStat{}, "idx_mrbs_period_scope", []string{"agg_mode", "level", "scope_key"}},
		{&MovieReleaseBucketStat{}, "idx_mrbs_period_count", []string{"agg_mode", "level", "count_all"}},
		{&MediaBirthBucketStat{}, "idx_wmbbs_period_scope", []string{"level", "scope_key"}},
		{&MediaBirthBucketStat{}, "idx_wmbbs_period_count", []string{"level", "media_count"}},
	} {
		parsed, err := schema.Parse(tc.model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatal(err)
		}
		index := parsed.LookIndex(tc.index)
		if index == nil {
			t.Fatalf("缺少周期索引 %s", tc.index)
		}
		var columns []string
		for _, field := range index.Fields {
			columns = append(columns, field.Field.DBName)
		}
		if !reflect.DeepEqual(columns, tc.columns) {
			t.Fatalf("%s 列 = %v, want %v", tc.index, columns, tc.columns)
		}
	}
}
