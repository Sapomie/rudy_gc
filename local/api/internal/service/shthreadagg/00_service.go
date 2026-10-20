package shthreadagg

import "rudy_gc/local/api/internal/dep"

const (
	AggScopeAll = "all"

	levelYear    = "year"
	levelQuarter = "quarter"
	levelMonth   = "month"
	levelWeek    = "week"
	levelDay     = "day"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{deps: d}
}

type BackfillResult struct {
	ClearedBucketRows int64 `json:"cleared_bucket_rows"`
	DirtyDays         int64 `json:"dirty_days"`
	YearBuckets       int64 `json:"year_buckets"`
	QuarterBuckets    int64 `json:"quarter_buckets"`
	MonthBuckets      int64 `json:"month_buckets"`
	WeekBuckets       int64 `json:"week_buckets"`
	DayBuckets        int64 `json:"day_buckets"`
	ElapsedMs         int64 `json:"elapsed_ms"`
}
