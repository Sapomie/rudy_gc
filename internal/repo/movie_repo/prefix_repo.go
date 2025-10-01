package movie_repo

import "context"

type PrefixRepo interface {
	// 按 name 查询，不存在则插入，返回主键 id
	GetOrCreateByName(ctx context.Context, name string) (int64, error)
}
