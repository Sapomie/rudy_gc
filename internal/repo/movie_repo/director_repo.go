package movie_repo

import (
	"context"
)

type DirectorRepo interface {
	GetOrCreateByName(ctx context.Context, name, javId string) (int64, error)
}
