package moviereleaseagg

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type weekReleaseMovies struct {
	moviex.AMovieModel
	ranges map[string]int
}

func (*weekReleaseMovies) ListDistinctReleaseDaysByRangeMode(context.Context, int64, int64, string) ([]int64, error) {
	return nil, nil
}
func (m *weekReleaseMovies) CalcReleaseBucketByMode(_ context.Context, start, end int64, mode string) (*moviex.MovieReleaseBucketCalc, error) {
	m.ranges[mode+":"+time.Unix(start, 0).In(time.Local).Format(time.DateOnly)+":"+time.Unix(end, 0).In(time.Local).Format(time.DateOnly)]++
	return &moviex.MovieReleaseBucketCalc{}, nil
}
func (*weekReleaseMovies) CalcTopCastsByReleaseRangeMode(context.Context, int64, int64, int64, string) ([]*moviex.MovieReleaseTopCalc, error) {
	return nil, nil
}
func (*weekReleaseMovies) CalcTopDirectorsByReleaseRangeMode(context.Context, int64, int64, int64, string) ([]*moviex.MovieReleaseTopCalc, error) {
	return nil, nil
}
func (*weekReleaseMovies) CalcTopLabelsByReleaseRangeMode(context.Context, int64, int64, int64, string) ([]*moviex.MovieReleaseTopCalc, error) {
	return nil, nil
}
func (*weekReleaseMovies) CalcTopPrefixesByReleaseRangeMode(context.Context, int64, int64, int64, string) ([]*moviex.MovieReleaseTopCalc, error) {
	return nil, nil
}

type weekReleaseBuckets struct {
	moviex.MovieReleaseBucketStatModel
	deleted int
}

func (*weekReleaseBuckets) FindOneByAggModeScopeKey(_ context.Context, mode, key string) (*moviex.MovieReleaseBucketStat, error) {
	if key == "week:2020-W53" {
		return &moviex.MovieReleaseBucketStat{Id: 1, AggMode: mode, ScopeKey: key}, nil
	}
	return nil, moviex.ErrNotFound
}
func (m *weekReleaseBuckets) Delete(context.Context, int64) error { m.deleted++; return nil }
func (*weekReleaseBuckets) Insert(context.Context, *moviex.MovieReleaseBucketStat) (sql.Result, error) {
	panic("空周不应插入")
}

type weekReleaseTops struct {
	moviex.MovieReleaseTopStatModel
}

func (*weekReleaseTops) DeleteByAggModeScopeAggType(context.Context, string, string, string) error {
	return nil
}

func TestReleaseWeeksRebuildAcrossMonthsAndDeleteEmptyWeeks(t *testing.T) {
	movies, buckets := &weekReleaseMovies{ranges: map[string]int{}}, &weekReleaseBuckets{}
	svc := NewService(&dep.Dep{MovieModel: movies, MovieReleaseBucketStatModel: buckets, MovieReleaseTopStatModel: &weekReleaseTops{}})
	months := []int64{time.Date(2020, 12, 1, 0, 0, 0, 0, time.Local).Unix(), time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local).Unix()}
	if _, err := svc.rebuildByBucketMonths(context.Background(), months, 0); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{AggModeAll, AggModeOwned} {
		if got := movies.ranges[mode+":2020-12-28:2021-01-03"]; got != 1 {
			t.Fatalf("跨年周重建次数 = %d, want 1 (%s)", got, mode)
		}
	}
	if buckets.deleted != 2 {
		t.Fatalf("最后一部影片移出后应删除两种统计模式的空周，实际 %d", buckets.deleted)
	}
}
