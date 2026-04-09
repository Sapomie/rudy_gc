package moviex

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AmLabelModel = (*customAmLabelModel)(nil)

type (
	// AmLabelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAmLabelModel.
	AmLabelModel interface {
		amLabelModel
		GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error)
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListAllIDs(ctx context.Context) ([]int64, error)
	}

	customAmLabelModel struct {
		*defaultAmLabelModel
	}
)

// NewAmLabelModel returns a model for the database table.
func NewAmLabelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AmLabelModel {
	return &customAmLabelModel{
		defaultAmLabelModel: newAmLabelModel(conn, c, opts...),
	}
}

func (m *customAmLabelModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmLabelModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customAmLabelModel) GetMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64) (int64, int64, error) {
	const query = `
SELECT
	(SELECT COUNT(DISTINCT jav_id) FROM a_movie WHERE label_id = ?) AS movie_number,
	(SELECT COUNT(DISTINCT am.jav_id)
		FROM a_movie am
		JOIN w_media vf ON vf.movie_jav_id = am.jav_id AND vf.source_type = ? AND vf.is_removed = ?
		WHERE am.label_id = ?) AS owned_movie_number
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

func (m *customAmLabelModel) ListAllIDs(ctx context.Context) ([]int64, error) {
	const query = "SELECT id FROM am_label"
	var ids []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &ids, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}
