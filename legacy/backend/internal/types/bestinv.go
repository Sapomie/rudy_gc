// internal/types/bestinv.go
package types

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
