package movie_repo

import (
	"context"

	"rudy_gc/data/modelx/moviex"
)

type MurlRepo interface {
	// UpsertByJavIdPreserveLocal 幂等保存 bm_murl
	// - 按 jav_id 唯一键
	// - 不存在则 Insert
	// - 存在则 Update，但保留 JacketImgLocal / SmallImgLocal
	UpsertByJavIdPreserveLocal(ctx context.Context, murl *moviex.BmMurl) error
}
