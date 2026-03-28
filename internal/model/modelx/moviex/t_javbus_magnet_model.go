package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TJavbusMagnetModel = (*customTJavbusMagnetModel)(nil)

type (
	// TJavbusMagnetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTJavbusMagnetModel.
	TJavbusMagnetModel interface {
		tJavbusMagnetModel
		SummarizeByMovieJavId(ctx context.Context, movieJavId string) (int64, int64, error)
		ListByMovieJavId(ctx context.Context, movieJavId string) ([]*TJavbusMagnet, error)
	}

	customTJavbusMagnetModel struct {
		*defaultTJavbusMagnetModel
	}
)

// NewTJavbusMagnetModel returns a model for the database table.
func NewTJavbusMagnetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TJavbusMagnetModel {
	return &customTJavbusMagnetModel{
		defaultTJavbusMagnetModel: newTJavbusMagnetModel(conn, c, opts...),
	}
}

func (m *customTJavbusMagnetModel) SummarizeByMovieJavId(ctx context.Context, movieJavId string) (int64, int64, error) {
	var resp struct {
		TotalCount        int64 `db:"total_count"`
		LatestPublishTime int64 `db:"latest_publish_time"`
	}

	query := "select count(*) as total_count, ifnull(max(`share_date`), 0) as latest_publish_time from `t_javbus_magnet` where `movie_jav_id` = ?"
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, movieJavId); err != nil {
		return 0, 0, err
	}
	return resp.TotalCount, resp.LatestPublishTime, nil
}

func (m *customTJavbusMagnetModel) ListByMovieJavId(ctx context.Context, movieJavId string) ([]*TJavbusMagnet, error) {
	query, args, err := squirrel.
		Select(tJavbusMagnetRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("`share_date` DESC", "`row_sort` ASC", "`id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TJavbusMagnet
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TJavbusMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}
