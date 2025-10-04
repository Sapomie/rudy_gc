// internal/repo/movie_repo/cafo_repo.go
package movie_repo

import (
	"context"
)

type CafoRepo interface {
	// 按演员名查询生日；不存在返回 found=false。
	FindBirthByName(ctx context.Context, name string) (birthDay int64, found bool, err error)
}
