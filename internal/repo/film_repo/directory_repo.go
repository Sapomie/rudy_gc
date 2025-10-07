// internal/repo/film_repo/directory_repo.go
package film_repo

import "context"

type DirectoryRepo interface {
	// 逐级 GetOrCreate，返回叶子目录ID
	GetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error)
}
