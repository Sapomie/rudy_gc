package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type MinfoRepo interface {
	// UpsertPreserve:
	// 以 jav_id 为幂等键；不存在则插入，
	// 存在则更新但保留历史字段（Chinese/FirstRankDayNumber/HighestRank/DaysInRank/NeedDownload/CreatedOn）。
	UpsertPreserve(ctx context.Context, minfo *moviex.BmMinfo) error

	// 更新排行榜数据
	UpdateRankStatsByJavId(ctx context.Context, javId string, firstDay, highestRank, daysInRank, updatedOn int64) error

	// ✅ 新增：按 jav_id 查询完整记录
	FindOneByJavId(ctx context.Context, javId string) (*moviex.BmMinfo, error)
}
