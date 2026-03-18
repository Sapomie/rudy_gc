package types

type ScAggBreadcrumb struct {
	Title string
	Href  string
}

type ScAggBucket struct {
	Label           string
	Href            string
	EventCount      int
	AvgCooldownDays float64
	ComeRate        float64
}

type ScAggTrend struct {
	Label           string
	EventCount      int
	AvgCooldownDays float64
	ComeRate        float64
}

type ScAggTopStat struct {
	Name       string
	EventCount int
	MovieCount int
	Href       string
}

type ScAggEventItem struct {
	Event *GSc
	Href  string
}

type ScAggResult struct {
	Title       string
	Level       string
	Breadcrumbs []ScAggBreadcrumb

	TotalEvents           int
	TotalMovieAppearances int
	TotalUniqueMovies     int

	BucketsY []ScAggBucket
	BucketsQ []ScAggBucket
	BucketsM []ScAggBucket

	RecentTrend []ScAggTrend

	TopCasts    []ScAggTopStat
	TopLabels   []ScAggTopStat
	TopPrefixes []ScAggTopStat

	Events []*ScAggEventItem
	Movies []*MovieType
}

type MovieScEvent struct {
	ScName   string
	ScTime   int64
	Cooldown int64
	IsCome   bool
	Href     string
}
