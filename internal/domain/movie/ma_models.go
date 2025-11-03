package movie

import "rudy_gc/internal/types"

// ----- 视图模型（模板键名一致） -----

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

// Top（Owned 版）
type CastStat struct {
	Name  string
	Count int
	ScSum int64
}
type TopStat struct {
	Name  string
	Count int
	ScSum int64
}

// Top（All 版）
type TopStatAll struct {
	Name       string
	CountAll   int
	CountOwned int
}

// ----- Owned/All 聚合的入参 -----

type AggParams struct {
	// mode = "release" | "birth"
	Mode string

	Year    int
	Quarter int
	Month   int

	OrderBy string // releasing_date/birth_time
	Page    int
	Size    int
	TopN    int
}

// ----- Owned/All 聚合的返回（模板用键名完全匹配你现有模板） -----
// 备注：PageInfo/SortQuery 仍在 HTTP 层计算，这里只返回 Movies/Total 等原始数据。
type AggResult struct {
	// 通用
	Title       string
	Breadcrumbs []Breadcrumb
	Level       string // "root" | "year" | "quarter" | "month"

	// 范围（非 root 时会回填）
	RangeStart string
	RangeEnd   string

	// 列表数据
	Movies []*types.MovieType
	Total  int64

	// Owned 页面 Buckets
	BucketsY []Bucket // 根页（年份入口）
	BucketsQ []Bucket // 年页：Q1~Q4
	BucketsM []Bucket // 季页：3 个月；或年页时未使用

	// All 页面 Buckets
	BucketsAll  []AllBucket // 非年页使用
	BucketsQAll []AllBucket // 年页：季度
	BucketsMAll []AllBucket // 年页：12 月

	// Top - Owned
	TopCasts     []CastStat
	TopDirectors []TopStat
	TopLabels    []TopStat
	TopPrefixes  []TopStat

	// Top - All
	TopCastsAll     []TopStatAll
	TopDirectorsAll []TopStatAll
	TopLabelsAll    []TopStatAll
	TopPrefixesAll  []TopStatAll
}
