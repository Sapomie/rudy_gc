package moviereleaseagg

type Bucket struct {
	Label      string
	Href       string
	CountAll   int64
	CountOwned int64
	SizeAllGB  float64
}

type Breadcrumb struct {
	Title string
	Href  string
}

type TopStat struct {
	PersonId   int64
	Name       string
	CountAll   int64
	CountOwned int64
}

type AggParams struct {
	Mode    string
	Year    int
	Quarter int
	Month   int
	Day     int
	TopN    int
}

type AggResult struct {
	AggMode         string
	AggModeLabel    string
	CardFilterQuery string
	Title           string
	Breadcrumbs     []Breadcrumb
	Level           string
	RangeStart      string
	RangeEnd        string
	BucketsAll      []Bucket
	BucketsQAll     []Bucket
	BucketsMAll     []Bucket
	BucketsDAll     []Bucket
	TopCastsAll     []TopStat
	TopDirectorsAll []TopStat
	TopLabelsAll    []TopStat
	TopPrefixesAll  []TopStat
}

type BackfillResult struct {
	ClearedBucketRows int   `json:"cleared_bucket_rows"`
	ClearedTopRows    int   `json:"cleared_top_rows"`
	ClearedDirtyRows  int   `json:"cleared_dirty_rows"`
	DirtyMonths       int   `json:"dirty_months"`
	YearBuckets       int   `json:"year_buckets"`
	QuarterBuckets    int   `json:"quarter_buckets"`
	MonthBuckets      int   `json:"month_buckets"`
	DayBuckets        int   `json:"day_buckets"`
	TopRows           int   `json:"top_rows"`
	ElapsedMs         int64 `json:"elapsed_ms"`
}
