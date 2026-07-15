package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	moviedomain "rudy_gc/internal/domain/movie"
)

var _ CMovieAlbumItemModel = (*customCMovieAlbumItemModel)(nil)

type (
	CMovieAlbumItemModel interface {
		cMovieAlbumItemModel
		QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error
		ListByAlbumId(ctx context.Context, albumId int64) ([]*CMovieAlbumItem, error)
		ListByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) ([]*CMovieAlbumItem, error)
		ListByMovieJavId(ctx context.Context, movieJavId string) ([]*CMovieAlbumItem, error)
		ListByFC2Key(ctx context.Context, fc2Key string) ([]*CMovieAlbumItem, error)
		DeleteByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) (bool, error)
		CountPageRows(ctx context.Context, albumId int64, filter CMovieAlbumItemPageFilter) (int64, error)
		ListPageRows(ctx context.Context, albumId int64, offset int64, limit int64, orderBy string, filter CMovieAlbumItemPageFilter) ([]*CMovieAlbumItem, error)
		CountByAlbumNameMovieJavId(ctx context.Context, albumName string, movieJavId string) (int64, error)
	}

	CMovieAlbumItemPageFilter struct {
		Keyword              string
		ReleasingDateFrom    int64
		HasReleasingDateFrom bool
		ReleasingDateTo      int64
		HasReleasingDateTo   bool
	}

	customCMovieAlbumItemModel struct {
		*defaultCMovieAlbumItemModel
	}
)

func NewCMovieAlbumItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CMovieAlbumItemModel {
	return &customCMovieAlbumItemModel{
		defaultCMovieAlbumItemModel: newCMovieAlbumItemModel(conn, c, opts...),
	}
}

func (m *customCMovieAlbumItemModel) QueryRowsNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCMovieAlbumItemModel) QueryRowNoCacheCtx(ctx context.Context, dest any, query string, args ...any) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, dest, query, args...)
}

func (m *customCMovieAlbumItemModel) ListByAlbumId(ctx context.Context, albumId int64) ([]*CMovieAlbumItem, error) {
	query, args, err := squirrel.
		Select(cMovieAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"album_id": albumId}).
		OrderBy("`created_on` DESC", "`id` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*CMovieAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumItemModel) ListByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) ([]*CMovieAlbumItem, error) {
	query, args, err := squirrel.
		Select(cMovieAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"album_id": albumId, "movie_jav_id": movieJavId}).
		Where(squirrel.Eq{"movie_type": moviedomain.SehuatangThreadMovieTypeJAV, "fc2_key": ""}).
		OrderBy("`id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*CMovieAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumItemModel) ListByMovieJavId(ctx context.Context, movieJavId string) ([]*CMovieAlbumItem, error) {
	query, args, err := squirrel.
		Select(cMovieAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		Where(squirrel.Eq{"movie_type": moviedomain.SehuatangThreadMovieTypeJAV, "fc2_key": ""}).
		OrderBy("`created_on` DESC", "`id` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*CMovieAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumItemModel) ListByFC2Key(ctx context.Context, fc2Key string) ([]*CMovieAlbumItem, error) {
	query, args, err := squirrel.
		Select(cMovieAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{
			"movie_type":   moviedomain.SehuatangThreadMovieTypeFC2PPV,
			"movie_jav_id": "",
			"fc2_key":      strings.TrimSpace(fc2Key),
		}).
		OrderBy("`created_on` DESC", "`id` DESC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*CMovieAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumItemModel) DeleteByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) (bool, error) {
	row, err := m.FindOneByIdentity(ctx, albumId, moviedomain.SehuatangThreadMovieTypeJAV, strings.TrimSpace(movieJavId), "")
	if err != nil {
		if err == ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if err := m.Delete(ctx, row.Id); err != nil {
		return false, err
	}
	return true, nil
}

func (m *customCMovieAlbumItemModel) CountPageRows(ctx context.Context, albumId int64, filter CMovieAlbumItemPageFilter) (int64, error) {
	queryBuilder := squirrel.Select("COUNT(1)").From(strings.Trim(m.table, "`")).Where(squirrel.Eq{"album_id": albumId})
	queryBuilder = applyCMovieAlbumPageFilter(queryBuilder, filter)
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func (m *customCMovieAlbumItemModel) ListPageRows(ctx context.Context, albumId int64, offset int64, limit int64, orderBy string, filter CMovieAlbumItemPageFilter) ([]*CMovieAlbumItem, error) {
	queryBuilder := squirrel.Select(cMovieAlbumItemRows).From(strings.Trim(m.table, "`")).Where(squirrel.Eq{"album_id": albumId})
	queryBuilder = applyCMovieAlbumPageFilter(queryBuilder, filter)
	if strings.TrimSpace(orderBy) != "" {
		queryBuilder = queryBuilder.OrderBy(orderBy)
	}
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		queryBuilder = queryBuilder.Offset(uint64(offset))
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	var rows []*CMovieAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumItemModel) CountByAlbumNameMovieJavId(ctx context.Context, albumName string, movieJavId string) (int64, error) {
	query, args, err := squirrel.
		Select("COUNT(1)").
		From(strings.Trim(m.table, "`") + " cai").
		Join("c_movie_album ca ON ca.id = cai.album_id").
		Where(squirrel.Eq{
			"ca.name":          strings.TrimSpace(albumName),
			"cai.movie_type":   moviedomain.SehuatangThreadMovieTypeJAV,
			"cai.movie_jav_id": strings.TrimSpace(movieJavId),
			"cai.fc2_key":      "",
		}).
		ToSql()
	if err != nil {
		return 0, err
	}
	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func applyCMovieAlbumKeywordFilter(queryBuilder squirrel.SelectBuilder, keyword string) squirrel.SelectBuilder {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return queryBuilder
	}
	like := "%" + keyword + "%"
	return queryBuilder.Where(squirrel.Or{
		squirrel.Like{"movie_name": like},
		squirrel.Like{"movie_jav_id": like},
		squirrel.Like{"fc2_key": like},
	})
}

func applyCMovieAlbumPageFilter(queryBuilder squirrel.SelectBuilder, filter CMovieAlbumItemPageFilter) squirrel.SelectBuilder {
	queryBuilder = applyCMovieAlbumKeywordFilter(queryBuilder, filter.Keyword)
	if filter.HasReleasingDateFrom {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"releasing_date": filter.ReleasingDateFrom})
	}
	if filter.HasReleasingDateTo {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"releasing_date": filter.ReleasingDateTo})
	}
	return queryBuilder
}
