package movie_repo

import (
	"context"
	"rudy_gc/data/modelx/moviex"
)

type MurlRepo interface {
	FindOneByJavId(ctx context.Context, javId string) (*moviex.BmMurl, error)
	UpsertByJavIdPreserveLocal(ctx context.Context, murl *moviex.BmMurl) error
}
