package types

import "fmt"

type Rank struct {
	RankKey    string // 唯一键，例如 "2024-09-30_001"
	MovieJavId string
	DayNumber  int64 // 第几天（与老项目算法一致）
	RankPos    int64 // 当天名次（1 开始）
	Category   int64 // 榜单类别（沿用你的常量）
}

func BuildRankKey(date string, pos int64) string {
	return fmt.Sprintf("%s_%03d", date, pos)
}
