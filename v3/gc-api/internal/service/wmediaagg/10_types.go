package wmediaagg

type Bucket struct {
	Label  string
	Href   string
	Count  int64
	SizeGB float64
}

type Breadcrumb struct {
	Title string
	Href  string
}

type TopStat struct {
	AggID     int64
	Name      string
	Count     int64
	SizeBytes int64
}

type AggParams struct {
	Year    int
	Quarter int
	Month   int
	Day     int
	TopN    int
}

type BackfillResult struct {
	ClearedBucketRows int   `json:"cleared_bucket_rows"`
	ClearedTopRows    int   `json:"cleared_top_rows"`
	ClearedDirtyRows  int   `json:"cleared_dirty_rows"`
	DirtyDays         int   `json:"dirty_days"`
	YearBuckets       int   `json:"year_buckets"`
	QuarterBuckets    int   `json:"quarter_buckets"`
	MonthBuckets      int   `json:"month_buckets"`
	DayBuckets        int   `json:"day_buckets"`
	TopRows           int   `json:"top_rows"`
	ElapsedMs         int64 `json:"elapsed_ms"`
}

type AggResult struct {
	Title        string
	Breadcrumbs  []Breadcrumb
	Level        string
	RangeStart   string
	RangeEnd     string
	BucketsY     []Bucket
	BucketsQ     []Bucket
	BucketsM     []Bucket
	BucketsD     []Bucket
	TopCasts     []TopStat
	TopDirectors []TopStat
	TopLabels    []TopStat
	TopPrefixes  []TopStat
}
