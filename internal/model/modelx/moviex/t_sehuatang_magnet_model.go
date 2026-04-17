package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TSehuatangMagnetModel = (*customTSehuatangMagnetModel)(nil)

type (
	// TSehuatangMagnetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTSehuatangMagnetModel.
	TSehuatangMagnetModel interface {
		tSehuatangMagnetModel
		ListPage(ctx context.Context, offset, limit int64, orderBy string, filter SehuatangMagnetListFilter) ([]*TSehuatangMagnet, error)
		CountAll(ctx context.Context, filter SehuatangMagnetListFilter) (int64, error)
		ListByMovieKey(ctx context.Context, movieJavId string, movieName string) ([]*TSehuatangMagnet, error)
		ListMissingMovieJavIDByMovieName(ctx context.Context, movieName string) ([]*TSehuatangMagnet, error)
		ExistsByThreadURL(ctx context.Context, threadURL string) (bool, error)
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		TableName() string
	}

	customTSehuatangMagnetModel struct {
		*defaultTSehuatangMagnetModel
	}

	SehuatangMagnetListFilter struct {
		Keyword         string
		MovieJavID      string
		InfoHash        string
		Tag             string
		PostTimeFrom    int64
		HasPostTimeFrom bool
		PostTimeTo      int64
		HasPostTimeTo   bool
	}
)

// NewTSehuatangMagnetModel returns a model for the database table.
func NewTSehuatangMagnetModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TSehuatangMagnetModel {
	return &customTSehuatangMagnetModel{
		defaultTSehuatangMagnetModel: newTSehuatangMagnetModel(conn, c, opts...),
	}
}

func (m *customTSehuatangMagnetModel) TableName() string {
	return m.table
}

func (m *customTSehuatangMagnetModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTSehuatangMagnetModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customTSehuatangMagnetModel) ListPage(ctx context.Context, offset, limit int64, orderBy string, filter SehuatangMagnetListFilter) ([]*TSehuatangMagnet, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "`last_seen_time` DESC, `id` DESC"
	}

	builder := applySehuatangMagnetListFilter(squirrel.
		Select(tSehuatangMagnetRows).
		From(m.table), filter)

	sqlStr, args, err := builder.
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TSehuatangMagnet
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*TSehuatangMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTSehuatangMagnetModel) CountAll(ctx context.Context, filter SehuatangMagnetListFilter) (int64, error) {
	builder := applySehuatangMagnetListFilter(squirrel.
		Select("COUNT(*)").
		From(m.table), filter)

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (m *customTSehuatangMagnetModel) ListByMovieKey(ctx context.Context, movieJavId string, movieName string) ([]*TSehuatangMagnet, error) {
	builder := squirrel.
		Select(tSehuatangMagnetRows).
		From(strings.Trim(m.table, "`"))

	hasCondition := false
	movieJavId = strings.TrimSpace(movieJavId)
	movieName = strings.TrimSpace(movieName)
	if movieJavId != "" && movieName != "" {
		builder = builder.Where(squirrel.Or{
			squirrel.Eq{"movie_jav_id": movieJavId},
			squirrel.Eq{"movie_name": movieName},
		})
		hasCondition = true
	} else if movieJavId != "" {
		builder = builder.Where(squirrel.Eq{"movie_jav_id": movieJavId})
		hasCondition = true
	} else if movieName != "" {
		builder = builder.Where(squirrel.Eq{"movie_name": movieName})
		hasCondition = true
	}
	if !hasCondition {
		return []*TSehuatangMagnet{}, nil
	}

	query, args, err := builder.
		OrderBy("`post_time` DESC", "`last_seen_time` DESC", "`id` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TSehuatangMagnet
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSehuatangMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTSehuatangMagnetModel) ListMissingMovieJavIDByMovieName(ctx context.Context, movieName string) ([]*TSehuatangMagnet, error) {
	movieName = strings.TrimSpace(movieName)
	if movieName == "" {
		return []*TSehuatangMagnet{}, nil
	}

	query, args, err := squirrel.
		Select(tSehuatangMagnetRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"movie_name": movieName}).
		Where(squirrel.Eq{"movie_jav_id": ""}).
		OrderBy("`id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TSehuatangMagnet
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TSehuatangMagnet{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTSehuatangMagnetModel) ExistsByThreadURL(ctx context.Context, threadURL string) (bool, error) {
	threadURL = strings.TrimSpace(threadURL)
	if threadURL == "" {
		return false, nil
	}

	sqlStr, args, err := squirrel.
		Select("COUNT(*)").
		From(m.table).
		Where(squirrel.Eq{"thread_url": threadURL}).
		ToSql()
	if err != nil {
		return false, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, sqlStr, args...); err != nil {
		return false, err
	}
	return total > 0, nil
}

func applySehuatangMagnetListFilter(builder squirrel.SelectBuilder, filter SehuatangMagnetListFilter) squirrel.SelectBuilder {
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		builder = builder.Where(squirrel.Or{
			squirrel.Like{"movie_name": like},
			squirrel.Like{"movie_jav_id": like},
			squirrel.Like{"tag": like},
			squirrel.Like{"thread_title": like},
		})
	}
	if movieJavID := strings.TrimSpace(filter.MovieJavID); movieJavID != "" {
		builder = builder.Where(squirrel.Like{"movie_jav_id": "%" + movieJavID + "%"})
	}
	if infoHash := strings.TrimSpace(filter.InfoHash); infoHash != "" {
		builder = builder.Where(squirrel.Like{"info_hash": "%" + infoHash + "%"})
	}
	if tag := strings.TrimSpace(filter.Tag); tag != "" {
		builder = builder.Where(squirrel.Like{"tag": "%" + tag + "%"})
	}
	if filter.HasPostTimeFrom {
		builder = builder.Where(squirrel.GtOrEq{"post_time": filter.PostTimeFrom})
	}
	if filter.HasPostTimeTo {
		builder = builder.Where(squirrel.LtOrEq{"post_time": filter.PostTimeTo})
	}
	return builder
}
