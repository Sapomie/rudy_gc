// data/modelx/moviex/c_rank_model.go
package moviex

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankModel = (*customCRankModel)(nil)

type (
	// CRankModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankModel.
	CRankModel interface {
		cRankModel
		All(ctx context.Context) ([]*CRank, error)
		AggregateByJavId(ctx context.Context, javId string) (firstDay int64, bestRank int64, daysInRank int64, err error)
		FindHighestRank(ctx context.Context, movieJavId string, limit uint64) ([]*CRank, error)
	}

	customCRankModel struct {
		*defaultCRankModel
	}
)

// NewCRankModel returns a model for the database table.
func NewCRankModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankModel {
	return &customCRankModel{
		defaultCRankModel: newCRankModel(conn, c, opts...),
	}
}

// AggregateByJavId 使用无缓存查询做一次性聚合，避免全量加载
func (m *customCRankModel) AggregateByJavId(ctx context.Context, javId string) (int64, int64, int64, error) {
	query := fmt.Sprintf(`
		SELECT
			COALESCE(MIN(day_number), 0)  AS first_day,
			COALESCE(MIN(rank_pos),  0)  AS best_rank,
			COUNT(*)                      AS days_in_rank
		FROM %s
		WHERE movie_jav_id = ?
	`, m.tableName())

	var dst struct {
		FirstDay   int64 `db:"first_day"`
		BestRank   int64 `db:"best_rank"`
		DaysInRank int64 `db:"days_in_rank"`
	}

	// 用 go-zero 的无缓存查询接口，避免污染缓存键空间
	if err := m.QueryRowNoCacheCtx(ctx, &dst, query, javId); err != nil {
		return 0, 0, 0, err
	}
	return dst.FirstDay, dst.BestRank, dst.DaysInRank, nil
}

func (m *defaultCRankModel) FindHighestRank(ctx context.Context, movieJavId string, limit uint64) ([]*CRank, error) {
	var ranks []*CRank

	// 使用 SQL 构建器来创建查询语句
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("number ASC, day_number ASC").
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	// 执行查询并填充 ranks 切片
	if err := m.QueryRowsNoCacheCtx(ctx, &ranks, query, args...); err != nil {
		return nil, err
	}

	return ranks, nil
}

func (m *customCRankModel) All(ctx context.Context) ([]*CRank, error) {
	query := fmt.Sprintf(`SELECT * FROM %s`, m.tableName())

	var rows []*CRank
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		return nil, err
	}
	return rows, nil
}
