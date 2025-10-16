package film_repo

import (
	"context"
	"rudy_gc/internal/types"
)

type GListRepo interface {
	//已有不变
	FindAll(ctx context.Context) ([]*types.GList, error)
	Upsert(ctx context.Context, in *types.GList) (*types.GList, error)
	FindGList(ctx context.Context, scName string, isCome *int64, page, pageSize int) ([]*types.GList, error)

	FindGListByMovieJavIds(ctx context.Context, movieJavIds []string) ([]*types.GList, error)
	FindGListByMovieJavId(ctx context.Context, movieJavId string) ([]*types.GList, error)
}
