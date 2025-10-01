package movie_repo

import "context"

type CastRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
}
