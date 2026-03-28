package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSukebeiTorrentModel = (*customTSukebeiTorrentModel)(nil)

type (
	// TSukebeiTorrentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSukebeiTorrentModel.
	TSukebeiTorrentModel interface {
		tSukebeiTorrentModel
		SummarizeByMovieJavId(ctx context.Context, movieJavId string) (int64, int64, error)
		ListByMovieJavId(ctx context.Context, movieJavId string) ([]*TSukebeiTorrent, error)
	}

	customTSukebeiTorrentModel struct {
		*defaultTSukebeiTorrentModel
	}
)

// NewTSukebeiTorrentModel returns a model for the database table.
func NewTSukebeiTorrentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSukebeiTorrentModel {
	return &customTSukebeiTorrentModel{
		defaultTSukebeiTorrentModel: newTSukebeiTorrentModel(conn, c, opts...),
	}
}

func (m *customTSukebeiTorrentModel) SummarizeByMovieJavId(ctx context.Context, movieJavId string) (int64, int64, error) {
	var resp struct {
		TotalCount        int64 `db:"total_count"`
		LatestPublishTime int64 `db:"latest_publish_time"`
	}

	query := "select count(*) as total_count, ifnull(max(`publish_time`), 0) as latest_publish_time from `t_sukebei_torrent` where `movie_jav_id` = ?"
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, movieJavId); err != nil {
		return 0, 0, err
	}
	return resp.TotalCount, resp.LatestPublishTime, nil
}

func (m *customTSukebeiTorrentModel) ListByMovieJavId(ctx context.Context, movieJavId string) ([]*TSukebeiTorrent, error) {
	query, args, err := squirrel.
		Select(tSukebeiTorrentRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("`publish_time` DESC", "`seeders` DESC", "`id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TSukebeiTorrent
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSukebeiTorrent{}, nil
		}
		return nil, err
	}
	return rows, nil
}
