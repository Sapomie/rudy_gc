package moviereleaseagg

import (
	"context"
	"testing"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type releaseBucketListModel struct {
	moviex.MovieReleaseBucketStatModel
	rows  []*moviex.MovieReleaseBucketStat
	total int64
}

func (m *releaseBucketListModel) ListPage(context.Context, moviex.MovieReleaseBucketStatListFilter) ([]*moviex.MovieReleaseBucketStat, int64, error) {
	return m.rows, m.total, nil
}

func TestBuildBucketListMapsRemovedCount(t *testing.T) {
	model := &releaseBucketListModel{
		rows: []*moviex.MovieReleaseBucketStat{{
			ScopeKey:     "month:2026-08",
			Level:        levelMonth,
			Year:         2026,
			Month:        8,
			CountAll:     45,
			CountOwned:   25,
			RemovedCount: 15,
		}},
		total: 1,
	}
	service := NewService(&dep.Dep{MovieReleaseBucketStatModel: model})

	result, err := service.BuildBucketList(context.Background(), BucketListParams{AggMode: AggModeAll})
	if err != nil {
		t.Fatalf("BuildBucketList() error = %v", err)
	}
	if len(result.Rows) != 1 || result.Rows[0].RemovedCount != 15 {
		t.Fatalf("BuildBucketList() rows = %+v", result.Rows)
	}
}
