package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmPrefixModel = (*customAmPrefixModel)(nil)

type (
	// AmPrefixModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmPrefixModel.
	AmPrefixModel interface {
		amPrefixModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customAmPrefixModel struct {
		*defaultAmPrefixModel
	}
)

// NewAmPrefixModel returns a model for the database table.
func NewAmPrefixModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmPrefixModel {
	return &customAmPrefixModel{
		defaultAmPrefixModel: newAmPrefixModel(conn, c, opts...),
	}
}

func (m *customAmPrefixModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmPrefixModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT jav_id) FROM a_movie WHERE prefix_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT am.jav_id)
		FROM a_movie am
		JOIN v_film vf ON vf.movie_jav_id = am.jav_id AND vf.is_removed = ?
		WHERE am.prefix_id = ?) AS owned_movie_number
`
	var resp struct {
		MovieNumber      int64 `db:"movie_number"`
		OwnedMovieNumber int64 `db:"owned_movie_number"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, id, ownedRemovedStatus, id); err != nil {
		return 0, 0, err
	}
	return resp.MovieNumber, resp.OwnedMovieNumber, nil
}
