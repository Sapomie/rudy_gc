package spiderx

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DBestinvModel = (*customDBestinvModel)(nil)

type (
	// DBestinvModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDBestinvModel.
	DBestinvModel interface {
		dBestinvModel
		withSession(session sqlx.Session) DBestinvModel

		ListNeedScanIDs(ctx context.Context, flag int64, limit int64) ([]int64, error)
		ListIDsByRankCheck(ctx context.Context, flag int64, limit int64) ([]int64, error)
		LatestDayNumber(ctx context.Context) (int64, error)
	}

	customDBestinvModel struct {
		*defaultDBestinvModel
	}
)

// NewDBestinvModel returns a model for the database table.
func NewDBestinvModel(conn sqlx.SqlConn) DBestinvModel {
	return &customDBestinvModel{
		defaultDBestinvModel: newDBestinvModel(conn),
	}
}

func (m *customDBestinvModel) withSession(session sqlx.Session) DBestinvModel {
	return NewDBestinvModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDBestinvModel) ListNeedScanIDs(ctx context.Context, flag int64, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 100000
	}

	query, args, err := squirrel.
		Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", flag).
		OrderBy("`id` ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (m *customDBestinvModel) ListIDsByRankCheck(ctx context.Context, flag int64, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 100000
	}

	query, args, err := squirrel.
		Select("`id`").
		From(m.tableName()).
		Where("`need_rank_check` = ?", flag).
		OrderBy("`id` ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (m *customDBestinvModel) LatestDayNumber(ctx context.Context) (int64, error) {
	query, values, err := squirrel.Select("`day_number`").
		From(m.tableName()).
		OrderBy("day_number desc").
		Limit(1).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result int64
	if err = m.conn.QueryRowCtx(ctx, &result, query, values...); err != nil {
		return 0, err
	}

	return result, nil
}
