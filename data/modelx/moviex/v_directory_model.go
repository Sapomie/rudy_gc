// data/modelx/moviex/vdirectory_model_custom.go
package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VDirectoryModel = (*customVDirectoryModel)(nil)

type (
	// VDirectoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customVDirectoryModel.
	VDirectoryModel interface {
		vDirectoryModel
		// 新增：按名称精确查找一条目录记录（和以前一样用 squirrel 构造 SQL）
		FindOneByName(ctx context.Context, name string) (*VDirectory, error)
	}

	customVDirectoryModel struct {
		*defaultVDirectoryModel
	}
)

// NewVDirectoryModel returns a model for the database table.
func NewVDirectoryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) VDirectoryModel {
	return &customVDirectoryModel{
		defaultVDirectoryModel: newVDirectoryModel(conn, c, opts...),
	}
}

// FindOneByName 按目录名称精确查找一条记录（不走缓存）
func (m *customVDirectoryModel) FindOneByName(ctx context.Context, name string) (*VDirectory, error) {
	query, args, err := squirrel.
		Select(vDirectoryRows).
		From(m.table).
		Where(squirrel.Eq{"name": name}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, err
	}

	var resp VDirectory
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, args...); err != nil {
		// 和 goctl 生成代码保持一致的错误处理
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}
