package moviex

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmDirectorModel = (*customAmDirectorModel)(nil)

type (
	// AmDirectorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmDirectorModel.
	AmDirectorModel interface {
		amDirectorModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
	}

	customAmDirectorModel struct {
		*defaultAmDirectorModel
	}
)

// NewAmDirectorModel returns a model for the database table.
func NewAmDirectorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmDirectorModel {
	return &customAmDirectorModel{
		defaultAmDirectorModel: newAmDirectorModel(conn, c, opts...),
	}
}

func (m *customAmDirectorModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmDirectorModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT jav_id) FROM a_movie WHERE director_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT am.jav_id)
		FROM a_movie am
		JOIN v_film vf ON vf.movie_jav_id = am.jav_id AND vf.is_removed = ?
		WHERE am.director_id = ?) AS owned_movie_number
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
