// internal/types/bestinv.go
package types

// Bestinv 分类：月榜 / 总榜
const (
	BestCategoryMonth   = 1 + iota // 1
	BestCategoryAllTime            // 2
)

// 是否需要排名检查
const (
	BestinvNeedRankCheck   = 1 + iota // 1
	BestinvNoNeedRankCheck            // 2
)

// 是否需要扫描（旧逻辑 Inventory 用过，可以沿用）
const (
	BestinvNeedScan   = 1 + iota // 1
	BestinvNoNeedScan            // 2
)

type Bestinv struct {
	Id            int64
	Name          string // 唯一名 (query+date+page)
	NeedScan      int64
	NeedRankCheck int64
	Category      int64
	Page          int64
	DayNumber     int64
	Content       string
	LastQueryTime int64
	Date          string

	CreatedOn int64
	UpdatedOn int64
}
