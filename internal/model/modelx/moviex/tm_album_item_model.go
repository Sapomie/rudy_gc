package moviex

import (
	"context"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TmAlbumItemModel = (*customTmAlbumItemModel)(nil)

type (
	// TmAlbumItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTmAlbumItemModel.
	TmAlbumItemModel interface {
		tmAlbumItemModel
		ListByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) ([]*TmAlbumItem, error)
		DeleteByAlbumIdSourceTypeSourceRowId(ctx context.Context, albumId int64, sourceType string, sourceRowId int64) (bool, error)
		CountPageRows(ctx context.Context, albumId int64, filter AlbumItemPageFilter) (int64, error)
		ListPageRows(ctx context.Context, albumId int64, offset int64, limit int64, orderBy string, filter AlbumItemPageFilter) ([]*TmAlbumItem, error)
	}

	customTmAlbumItemModel struct {
		*defaultTmAlbumItemModel
	}
)

type AlbumItemPageFilter struct {
	Keyword            string
	SourceType         string
	SourceTypeSet      bool
	InfoHash           string
	PublishTimeFrom    int64
	PublishTimeTo      int64
	HasPublishTimeFrom bool
	HasPublishTimeTo   bool
	CreatedOnFrom      int64
	CreatedOnTo        int64
	HasCreatedOnFrom   bool
	HasCreatedOnTo     bool
}

// NewTmAlbumItemModel returns a model for the database table.
func NewTmAlbumItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TmAlbumItemModel {
	return &customTmAlbumItemModel{
		defaultTmAlbumItemModel: newTmAlbumItemModel(conn, c, opts...),
	}
}

func (m *customTmAlbumItemModel) ListByAlbumIdMovieJavId(ctx context.Context, albumId int64, movieJavId string) ([]*TmAlbumItem, error) {
	query, args, err := squirrel.
		Select(tmAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{
			"album_id":     albumId,
			"movie_jav_id": movieJavId,
		}).
		OrderBy("`id` ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var rows []*TmAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TmAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTmAlbumItemModel) DeleteByAlbumIdSourceTypeSourceRowId(ctx context.Context, albumId int64, sourceType string, sourceRowId int64) (bool, error) {
	row, err := m.FindOneByAlbumIdSourceTypeSourceRowId(ctx, albumId, sourceType, sourceRowId)
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

func (m *customTmAlbumItemModel) CountPageRows(ctx context.Context, albumId int64, filter AlbumItemPageFilter) (int64, error) {
	queryBuilder := squirrel.
		Select("COUNT(1)").
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"album_id": albumId})
	queryBuilder = applyAlbumItemPageFilter(queryBuilder, filter)

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

func (m *customTmAlbumItemModel) ListPageRows(ctx context.Context, albumId int64, offset int64, limit int64, orderBy string, filter AlbumItemPageFilter) ([]*TmAlbumItem, error) {
	queryBuilder := squirrel.
		Select(tmAlbumItemRows).
		From(strings.Trim(m.table, "`")).
		Where(squirrel.Eq{"album_id": albumId})
	queryBuilder = applyAlbumItemPageFilter(queryBuilder, filter)
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

	var rows []*TmAlbumItem
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TmAlbumItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func applyAlbumItemPageFilter(queryBuilder squirrel.SelectBuilder, filter AlbumItemPageFilter) squirrel.SelectBuilder {
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword != "" {
		like := "%" + keyword + "%"
		queryBuilder = queryBuilder.Where(squirrel.Or{
			squirrel.Like{"movie_name": like},
			squirrel.Like{"movie_jav_id": like},
		})
	}

	if filter.SourceTypeSet {
		queryBuilder = queryBuilder.Where(squirrel.Eq{"source_type": strings.TrimSpace(filter.SourceType)})
	}

	infoHash := strings.ToUpper(strings.TrimSpace(filter.InfoHash))
	if infoHash != "" {
		queryBuilder = queryBuilder.Where("UPPER(`info_hash`) LIKE ?", "%"+infoHash+"%")
	}

	if filter.HasPublishTimeFrom {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"publish_time": filter.PublishTimeFrom})
	}
	if filter.HasPublishTimeTo {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"publish_time": filter.PublishTimeTo})
	}
	if filter.HasCreatedOnFrom {
		queryBuilder = queryBuilder.Where(squirrel.GtOrEq{"created_on": filter.CreatedOnFrom})
	}
	if filter.HasCreatedOnTo {
		queryBuilder = queryBuilder.Where(squirrel.LtOrEq{"created_on": filter.CreatedOnTo})
	}
	return queryBuilder
}
