package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ERecordModel = (*customERecordModel)(nil)

type (
	// ERecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customERecordModel.
	ERecordModel interface {
		eRecordModel
		FindByStartTimeAndType(ctx context.Context, startFrom int64, typ string, limit int) ([]*ERecord, error)
	}

	customERecordModel struct {
		*defaultERecordModel
	}
)

// NewERecordModel returns a model for the database table.
func NewERecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ERecordModel {
	return &customERecordModel{
		defaultERecordModel: newERecordModel(conn, c, opts...),
	}
}

func (m *customERecordModel) FindByStartTimeAndType(ctx context.Context, startFrom int64, typ string, limit int) ([]*ERecord, error) {
	b := squirrel.
		Select(eRecordRows).
		From(m.table).
		Where("`start_time` >= ?", startFrom).
		OrderBy("`start_time` DESC").
		PlaceholderFormat(squirrel.Question)

	if typ != "" {
		b = b.Where("`type` = ?", typ)
	}
	if limit > 0 {
		b = b.Limit(uint64(limit))
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql failed: %w", err)
	}

	// 不走缓存，直接查
	var list []*ERecord
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*ERecord{}, nil
		}
		return nil, err
	}
	return list, nil
}
