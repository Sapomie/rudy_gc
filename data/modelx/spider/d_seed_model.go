package spider

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DSeedModel = (*customDSeedModel)(nil)

type (
	// DSeedModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDSeedModel.
	DSeedModel interface {
		dSeedModel
		withSession(session sqlx.Session) DSeedModel
		FindQueriesActive(ctx context.Context, nameType int64) ([]*DSeed, error)
	}

	customDSeedModel struct {
		*defaultDSeedModel
	}
)

const (
	QueryInactive = 1 + iota
	QueryActive
)

const (
	QueryNamePrefix = 1 + iota
	QueryNameLabel
)

const (
	QueryByOffset = 1 + iota
	QueryByStartEnd
)

// NewDSeedModel returns a model for the database table.
func NewDSeedModel(conn sqlx.SqlConn) DSeedModel {
	return &customDSeedModel{
		defaultDSeedModel: newDSeedModel(conn),
	}
}

func (m *customDSeedModel) withSession(session sqlx.Session) DSeedModel {
	return NewDSeedModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDSeedModel) FindQueriesActive(ctx context.Context, nameType int64) ([]*DSeed, error) {
	b := squirrel.
		Select("*").
		From(m.table).
		Where("`active` = ?", QueryActive).
		Where("`name_type` = ?", nameType).
		OrderBy("`updated_on` DESC, `name` ASC").
		PlaceholderFormat(squirrel.Question)

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build SQL failed: %w", err)
	}

	var rows []*DSeed
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}
