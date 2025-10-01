package movie_repo

import "context"

type LabelRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
}
