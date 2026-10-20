package wmediaagg

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type birthChangesMediaModel struct {
	moviex.WMediaModel
	calc *moviex.WMediaBirthBucketCalc
}

func (m *birthChangesMediaModel) CalcBirthBucket(context.Context, int64, int64) (*moviex.WMediaBirthBucketCalc, error) {
	return m.calc, nil
}

type birthChangesBucketModel struct {
	moviex.WMediaBirthBucketStatModel
	existing *moviex.WMediaBirthBucketStat
	written  *moviex.WMediaBirthBucketStat
	deleted  bool
	filter   moviex.WMediaBirthBucketStatListFilter
}

func (m *birthChangesBucketModel) FindOneByScopeKey(context.Context, string) (*moviex.WMediaBirthBucketStat, error) {
	if m.existing == nil {
		return nil, moviex.ErrNotFound
	}
	return m.existing, nil
}

func (m *birthChangesBucketModel) Insert(_ context.Context, row *moviex.WMediaBirthBucketStat) (sql.Result, error) {
	m.written = row
	return nil, nil
}

func (m *birthChangesBucketModel) Update(_ context.Context, row *moviex.WMediaBirthBucketStat) error {
	m.written = row
	return nil
}

func (m *birthChangesBucketModel) Delete(context.Context, int64) error {
	m.deleted = true
	return nil
}

func (m *birthChangesBucketModel) ListPage(_ context.Context, filter moviex.WMediaBirthBucketStatListFilter) ([]*moviex.WMediaBirthBucketStat, int64, error) {
	m.filter = filter
	return []*moviex.WMediaBirthBucketStat{nil, m.existing}, 1, nil
}

func TestRebuildBirthBucketRetainsHistoricalAdditions(t *testing.T) {
	for _, existing := range []*moviex.WMediaBirthBucketStat{nil, {Id: 10, ScopeKey: "month:2026-07", CreatedOn: 123}} {
		buckets := &birthChangesBucketModel{existing: existing}
		svc := NewService(&dep.Dep{
			WMediaModel: &birthChangesMediaModel{calc: &moviex.WMediaBirthBucketCalc{
				AddedCount: 2, AddedSizeBytes: 500, NetSizeBytes: 500,
			}},
			WMediaBirthBucketStatModel: buckets,
		})
		if err := svc.rebuildBucketScope(context.Background(), buildScope(2026, 0, 7, 0)); err != nil {
			t.Fatal(err)
		}
		row := buckets.written
		if buckets.deleted || row == nil || row.AddedCount != 2 || row.AddedSizeBytes != 500 || row.NetSizeBytes != 500 {
			t.Fatalf("historical additions lost: written=%+v deleted=%v", row, buckets.deleted)
		}
		if row.MediaCount != 0 || row.SizeBytes != 0 || row.ScopeKey != "month:2026-07" {
			t.Fatalf("unexpected owned statistics or scope: %+v", row)
		}
		if existing != nil && row.CreatedOn != 123 {
			t.Fatalf("created_on changed: %d", row.CreatedOn)
		}
	}
}

func TestRebuildBirthBucketPersistsDeletionAndNegativeNetSize(t *testing.T) {
	for _, existing := range []*moviex.WMediaBirthBucketStat{nil, {Id: 11, AddedCount: 9, AddedSizeBytes: 900}} {
		buckets := &birthChangesBucketModel{existing: existing}
		svc := NewService(&dep.Dep{
			WMediaModel: &birthChangesMediaModel{calc: &moviex.WMediaBirthBucketCalc{
				AddedCount: 1, AddedSizeBytes: 100, RemovedCount: 2,
				RemovedSizeBytes: 400, NetSizeBytes: -300,
				MediaCount: 1, SizeBytes: 100, HasSubCount: 1, LatestBirthTime: 456,
			}},
			WMediaBirthBucketStatModel: buckets,
		})
		if err := svc.rebuildBucketScope(context.Background(), buildScope(2026, 0, 8, 0)); err != nil {
			t.Fatal(err)
		}
		row := buckets.written
		if row == nil || row.AddedCount != 1 || row.AddedSizeBytes != 100 || row.RemovedCount != 2 || row.RemovedSizeBytes != 400 || row.NetSizeBytes != -300 {
			t.Fatalf("unexpected change statistics: %+v", row)
		}
		if row.MediaCount != 1 || row.SizeBytes != 100 || row.HasSubCount != 1 || row.LatestBirthTime != 456 {
			t.Fatalf("owned statistics changed: %+v", row)
		}
	}
}

func TestBuildBucketListReturnsChangeMetricsAndSort(t *testing.T) {
	model := &birthChangesBucketModel{existing: &moviex.WMediaBirthBucketStat{
		ScopeKey: "month:2026-08", Level: levelMonth, Year: 2026, Month: 8,
		AddedCount: 2, AddedSizeBytes: 300, RemovedCount: 3,
		RemovedSizeBytes: 500, NetSizeBytes: -200,
	}}
	svc := NewService(&dep.Dep{WMediaBirthBucketStatModel: model})
	result, err := svc.BuildBucketList(context.Background(), BucketListParams{Level: "month", Sort: "net_size", Dir: "asc", Page: 2, PageSize: 24})
	if err != nil {
		t.Fatal(err)
	}
	want := []BucketListRow{{
		BucketDisplay: "2026-M08", AddedCount: 2, AddedSizeBytes: 300,
		RemovedCount: 3, RemovedSizeBytes: 500, NetSizeBytes: -200,
	}}
	if result.Total != 1 || !reflect.DeepEqual(result.Rows, want) {
		t.Fatalf("bucket list = %+v, want %+v", result, want)
	}
	if model.filter.Sort != "net_size" || model.filter.Dir != "asc" || model.filter.Page != 2 || model.filter.PageSize != 24 {
		t.Fatalf("filter = %+v", model.filter)
	}
}
