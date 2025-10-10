// internal/repo/movie_repo/murl_repo.go
package movie_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type MurlRepo interface {
	FindOneByJavId(ctx context.Context, javId string) (*types.Murl, error)
	UpsertByJavIdPreserveLocal(ctx context.Context, murl *types.Murl) error
	// 新增：
	UpdatePartialByJavId(ctx context.Context, javId string, patch types.MurlPatch) error
}
