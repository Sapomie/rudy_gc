package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankPeriodModel = (*customCRankPeriodModel)(nil)

type (
	// CRankPeriodModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankPeriodModel.
	CRankPeriodModel interface {
		cRankPeriodModel
		FindLatestByPeriodTypeCategory(ctx context.Context, periodType, category int64) (*CRankPeriod, error)
		FindPrevByPeriodTypeCategory(ctx context.Context, periodType, category, currentStartDay int64) (*CRankPeriod, error)
		FindNextByPeriodTypeCategory(ctx context.Context, periodType, category, currentStartDay int64) (*CRankPeriod, error)
	}

	customCRankPeriodModel struct {
		*defaultCRankPeriodModel
	}
)

// NewCRankPeriodModel returns a model for the database table.
func NewCRankPeriodModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankPeriodModel {
	return &customCRankPeriodModel{
		defaultCRankPeriodModel: newCRankPeriodModel(conn, c, opts...),
	}
}

func (m *customCRankPeriodModel) FindLatestByPeriodTypeCategory(ctx context.Context, periodType, category int64) (*CRankPeriod, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE period_type = ? AND category = ?
		ORDER BY start_day_number DESC, id DESC
		LIMIT 1
	`, cRankPeriodRows, m.tableName())

	var row CRankPeriod
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, periodType, category); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *customCRankPeriodModel) FindPrevByPeriodTypeCategory(ctx context.Context, periodType, category, currentStartDay int64) (*CRankPeriod, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE period_type = ? AND category = ? AND start_day_number < ?
		ORDER BY start_day_number DESC, id DESC
		LIMIT 1
	`, cRankPeriodRows, m.tableName())

	var row CRankPeriod
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, periodType, category, currentStartDay); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (m *customCRankPeriodModel) FindNextByPeriodTypeCategory(ctx context.Context, periodType, category, currentStartDay int64) (*CRankPeriod, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE period_type = ? AND category = ? AND start_day_number > ?
		ORDER BY start_day_number ASC, id ASC
		LIMIT 1
	`, cRankPeriodRows, m.tableName())

	var row CRankPeriod
	if err := m.QueryRowNoCacheCtx(ctx, &row, query, periodType, category, currentStartDay); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}
