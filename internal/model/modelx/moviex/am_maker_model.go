package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmMakerModel = (*customAmMakerModel)(nil)

type (
	// AmMakerModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmMakerModel.
	AmMakerModel interface {
		amMakerModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAllIDs(ctx context.Context) ([]int64, error)
	}

	customAmMakerModel struct {
		*defaultAmMakerModel
	}
)

// NewAmMakerModel returns a model for the database table.
func NewAmMakerModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmMakerModel {
	return &customAmMakerModel{
		defaultAmMakerModel: newAmMakerModel(conn, c, opts...),
	}
}

func (m *customAmMakerModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmMakerModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmMakerModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT jav_id) FROM a_movie WHERE maker_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT am.jav_id)
		FROM a_movie am
		JOIN w_media vf ON vf.movie_jav_id = am.jav_id AND vf.source_type = ? AND vf.is_removed = ?
		WHERE am.maker_id = ?) AS owned_movie_number
`
	var resp struct {
		MovieNumber      int64 `db:"movie_number"`
		OwnedMovieNumber int64 `db:"owned_movie_number"`
	}
	if err := m.QueryRowNoCacheCtx(ctx, &resp, query, id, consts.WMediaSourceLegacyVFilm, ownedRemovedStatus, id); err != nil {
		return 0, 0, err
	}
	return resp.MovieNumber, resp.OwnedMovieNumber, nil
}

func (m *customAmMakerModel) ListAllIDs(ctx context.Context) ([]int64, error) {
	const query = "SELECT id FROM am_maker"
	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
