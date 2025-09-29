// data/modelx/spiderx/e_item_model.go
package spiderx

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EItemModel = (*customEItemModel)(nil)

type (
	EItemModel interface {
		eItemModel
		withSession(session sqlx.Session) EItemModel
		ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error)
	}

	customEItemModel struct {
		*defaultEItemModel
	}
)

func NewEItemModel(conn sqlx.SqlConn) EItemModel {
	return &customEItemModel{defaultEItemModel: newEItemModel(conn)}
}

func (m *customEItemModel) withSession(session sqlx.Session) EItemModel {
	return NewEItemModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customEItemModel) ListByDetailStatus(ctx context.Context, hasDetail int64, limit int64) ([]*EItem, error) {
	if limit <= 0 {
		limit = 10000
	}

	// 用 goctl 生成的 eItemRows + m.tableName()
	builder := squirrel.
		Select(eItemRows).
		From(m.tableName()).
		Where(squirrel.Eq{"has_detail": hasDetail}).
		Limit(uint64(limit))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*EItem
	if err := m.conn.QueryRowsCtx(ctx, &items, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return items, nil
}
