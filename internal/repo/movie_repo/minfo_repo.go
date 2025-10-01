// internal/repo/movie_repo/minfo_repo.go
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
}
