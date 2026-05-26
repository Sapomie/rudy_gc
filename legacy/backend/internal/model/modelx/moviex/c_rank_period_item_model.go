package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankPeriodItemModel = (*customCRankPeriodItemModel)(nil)

type (
	// CRankPeriodItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankPeriodItemModel.
	CRankPeriodItemModel interface {
		cRankPeriodItemModel
		CountByPeriodId(ctx context.Context, periodId int64) (int64, error)
		ListByPeriodId(ctx context.Context, periodId int64) ([]*CRankPeriodItem, error)
		ListByPeriodIdPage(ctx context.Context, periodId, page, pageSize int64) ([]*CRankPeriodItem, error)
		FindPeakRankMapByPeriodTypeCategoryBeforeDay(ctx context.Context, periodType, category, beforeStartDay int64) (map[string]int64, error)
	}

	customCRankPeriodItemModel struct {
		*defaultCRankPeriodItemModel
	}
)

// NewCRankPeriodItemModel returns a model for the database table.
func NewCRankPeriodItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankPeriodItemModel {
	return &customCRankPeriodItemModel{
		defaultCRankPeriodItemModel: newCRankPeriodItemModel(conn, c, opts...),
	}
}

func (m *customCRankPeriodItemModel) CountByPeriodId(ctx context.Context, periodId int64) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) AS cnt FROM %s WHERE period_id = ?`, m.tableName())
	var dst struct {
		Cnt int64 `db:"cnt"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &dst, query, periodId); err != nil {
		return 0, err
	}
	return dst.Cnt, nil
}

func (m *customCRankPeriodItemModel) ListByPeriodId(ctx context.Context, periodId int64) ([]*CRankPeriodItem, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE period_id = ?
		ORDER BY rank_pos ASC, id ASC
	`, cRankPeriodItemRows, m.tableName())

	var rows []*CRankPeriodItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, periodId); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return []*CRankPeriodItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCRankPeriodItemModel) ListByPeriodIdPage(ctx context.Context, periodId, page, pageSize int64) ([]*CRankPeriodItem, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE period_id = ?
		ORDER BY rank_pos ASC, id ASC
		LIMIT ? OFFSET ?
	`, cRankPeriodItemRows, m.tableName())

	var rows []*CRankPeriodItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, periodId, pageSize, offset); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return []*CRankPeriodItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCRankPeriodItemModel) FindPeakRankMapByPeriodTypeCategoryBeforeDay(ctx context.Context, periodType, category, beforeStartDay int64) (map[string]int64, error) {
	query := `
		SELECT i.movie_jav_id AS movie_jav_id, MIN(i.rank_pos) AS peak_rank
		FROM c_rank_period_item i
		INNER JOIN c_rank_period p ON p.id = i.period_id
		WHERE p.period_type = ? AND p.category = ? AND p.start_day_number < ?
		GROUP BY i.movie_jav_id
	`

	var rows []struct {
		MovieJavId string `db:"movie_jav_id"`
		PeakRank   int64  `db:"peak_rank"`
	}
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, periodType, category, beforeStartDay); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return map[string]int64{}, nil
		}
		return nil, err
	}

	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.MovieJavId] = row.PeakRank
	}
	return out, nil
}
