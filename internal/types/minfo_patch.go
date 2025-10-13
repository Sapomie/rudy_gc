package types

// MinfoPatch：按需局部更新（nil 表示不改）
type MinfoPatch struct {
	Chinese            *string
	FirstRankDayNumber *int64
	HighestRank        *int64
	DaysInRank         *int64
	NeedDownload       *int64
	UpdatedOn          *int64 // 若为 nil，仓储层会用 now 兜底
}
