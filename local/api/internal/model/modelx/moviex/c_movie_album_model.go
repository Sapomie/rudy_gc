package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CMovieAlbumModel = (*customCMovieAlbumModel)(nil)

var ErrCMovieAlbumNotEmpty = errors.New("电影相册中还有数据")

type (
	CMovieAlbumModel interface {
		cMovieAlbumModel
		ListAll(ctx context.Context) ([]*CMovieAlbum, error)
		DeleteIfEmpty(ctx context.Context, id int64) error
	}

	customCMovieAlbumModel struct {
		*defaultCMovieAlbumModel
	}
)

func NewCMovieAlbumModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CMovieAlbumModel {
	return &customCMovieAlbumModel{
		defaultCMovieAlbumModel: newCMovieAlbumModel(conn, c, opts...),
	}
}

func (m *customCMovieAlbumModel) ListAll(ctx context.Context) ([]*CMovieAlbum, error) {
	query := "SELECT " + cMovieAlbumRows + " FROM " + strings.Trim(m.table, "`") + " ORDER BY `id` ASC"
	var rows []*CMovieAlbum
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*CMovieAlbum{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCMovieAlbumModel) DeleteIfEmpty(ctx context.Context, id int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	idKey := fmt.Sprintf("%s%v", cacheRudyGcCMovieAlbumIdPrefix, id)
	nameKey := fmt.Sprintf("%s%v", cacheRudyGcCMovieAlbumNamePrefix, row.Name)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		var albumResult sql.Result
		transactionErr := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			var itemCount int64
			if countErr := session.QueryRowCtx(ctx, &itemCount, "SELECT COUNT(1) FROM `c_movie_album_item` WHERE `album_id` = ? FOR UPDATE", id); countErr != nil {
				return countErr
			}
			if itemCount > 0 {
				return ErrCMovieAlbumNotEmpty
			}

			var deleteErr error
			albumResult, deleteErr = session.ExecCtx(ctx, "DELETE FROM `c_movie_album` WHERE `id` = ?", id)
			if deleteErr != nil {
				return deleteErr
			}
			removedAlbums, rowsErr := albumResult.RowsAffected()
			if rowsErr != nil {
				return rowsErr
			}
			if removedAlbums != 1 {
				return ErrNotFound
			}
			return nil
		})
		return albumResult, transactionErr
	}, idKey, nameKey)
	return err
}
