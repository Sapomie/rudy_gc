package moviex

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmCastModel = (*customAmCastModel)(nil)

type (
	// AmCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmCastModel.
	AmCastModel interface {
		amCastModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAllIDs(ctx context.Context) ([]int64, error)
	}

	customAmCastModel struct {
		*defaultAmCastModel
	}
)

// NewAmCastModel returns a model for the database table.
func NewAmCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmCastModel {
	return &customAmCastModel{
		defaultAmCastModel: newAmCastModel(conn, c, opts...),
	}
}

func (m *customAmCastModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmCastModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmCastModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT movie_jav_id) FROM amr_movie_cast WHERE cast_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT amr.movie_jav_id)
		FROM amr_movie_cast amr
		JOIN v_film vf ON vf.movie_jav_id = amr.movie_jav_id AND vf.is_removed = ?
		WHERE amr.cast_id = ?) AS owned_movie_number
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

func (m *customAmCastModel) ListAllIDs(ctx context.Context) ([]int64, error) {
	const query = "SELECT id FROM am_cast"
	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
