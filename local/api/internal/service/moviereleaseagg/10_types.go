package moviereleaseagg

type BackfillResult struct {
	ClearedBucketRows int   `json:"cleared_bucket_rows"`
	ClearedTopRows    int   `json:"cleared_top_rows"`
	ClearedDirtyRows  int   `json:"cleared_dirty_rows"`
	DirtyMonths       int   `json:"dirty_months"`
	YearBuckets       int   `json:"year_buckets"`
	QuarterBuckets    int   `json:"quarter_buckets"`
	MonthBuckets      int   `json:"month_buckets"`
	WeekBuckets       int   `json:"week_buckets"`
	DayBuckets        int   `json:"day_buckets"`
	TopRows           int   `json:"top_rows"`
	ElapsedMs         int64 `json:"elapsed_ms"`
}
