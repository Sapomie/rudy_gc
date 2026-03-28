package spiderx

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DDetailModel = (*customDDetailModel)(nil)

type (
	// DDetailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDDetailModel.
	DDetailModel interface {
		dDetailModel
		withSession(session sqlx.Session) DDetailModel
		ListAllJavIds(ctx context.Context) ([]string, error)
	}

	customDDetailModel struct {
		*defaultDDetailModel
	}
)

// NewDDetailModel returns a model for the database table.
func NewDDetailModel(conn sqlx.SqlConn) DDetailModel {
	return &customDDetailModel{
		defaultDDetailModel: newDDetailModel(conn),
	}
}

func (m *customDDetailModel) withSession(session sqlx.Session) DDetailModel {
	return NewDDetailModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDDetailModel) ListAllJavIds(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT `jav_id` FROM %s", m.table)
	var ids []string
	if err := m.conn.QueryRowsCtx(ctx, &ids, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []string{}, nil
		}
		return nil, err
	}
	return ids, nil
}
