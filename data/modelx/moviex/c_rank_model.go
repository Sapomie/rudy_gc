// data/modelx/moviex/c_rank_model.go
package moviex

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CRankModel = (*customCRankModel)(nil)

type (
	// CRankModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCRankModel.
	CRankModel interface {
		cRankModel

		// AggregateByJavId 统计某影片的排行聚合：
		// - firstDay:   首次上榜日(MIN(day_number))
		// - bestRank:   最佳名次(MIN(rank_pos))
		// - daysInRank: 上榜天数(COUNT(*))
		AggregateByJavId(ctx context.Context, javId string) (firstDay int64, bestRank int64, daysInRank int64, err error)
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
