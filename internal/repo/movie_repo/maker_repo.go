package movie_repo

import "context"

type MakerRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
}
