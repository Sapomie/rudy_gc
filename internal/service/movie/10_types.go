package movie

import (
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/types"
)

type Bucket struct {
	Label  string
	Href   string
	Count  int
	SizeGB float64
}

type AllBucket struct {
	Label      string
	Href       string
	CountAll   int
	CountOwned int
	SizeAllGB  float64
}

type Breadcrumb struct {
	Title string
	Href  string
}

type CastStat struct {
	PersonId int64
	Name     string
	Count    int
	ScSum    int64
}

type TopStat struct {
	Name  string
	Count int
	ScSum int64
}

type TopStatAll struct {
	PersonId   int64
	Name       string
	CountAll   int
	CountOwned int
}

type AggParams struct {
	Mode    string
	Year    int
	Quarter int
	Month   int
	OrderBy string
	Page    int
	Size    int
	TopN    int
}

type AggResult struct {
	Title           string
	Breadcrumbs     []Breadcrumb
	Level           string
	RangeStart      string
	RangeEnd        string
	Movies          []*types.MovieType
	Total           int64
	BucketsY        []Bucket
	BucketsQ        []Bucket
	BucketsM        []Bucket
	BucketsAll      []AllBucket
	BucketsQAll     []AllBucket
	BucketsMAll     []AllBucket
	TopCasts        []CastStat
	TopDirectors    []TopStat
	TopLabels       []TopStat
	TopPrefixes     []TopStat
	TopCastsAll     []TopStatAll
	TopDirectorsAll []TopStatAll
	TopLabelsAll    []TopStatAll
	TopPrefixesAll  []TopStatAll
}

type RankPeriodPageRequest struct {
	PeriodType int64
	PeriodKey  string
	Category   int64
	Page       int64
	PageSize   int64
}

type RankPeriodMovieCard struct {
	Movie           *types.MovieType
	Item            *moviex.CRankPeriodItem
	PrevRankText    string
	RankChangeText  string
	RankChangeClass string
}

type RankPeriodPage struct {
	Title           string
	Period          *moviex.CRankPeriod
	PrevPeriod      *moviex.CRankPeriod
	NextPeriod      *moviex.CRankPeriod
	Cards           []*RankPeriodMovieCard
	Movies          []*types.MovieType
	Total           int64
	PeriodTypeLabel string
	CategoryLabel   string
	RangeStart      string
	RangeEnd        string
}
