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

var _ TAlbumModel = (*customTAlbumModel)(nil)

var ErrTAlbumNotEmpty = errors.New("相册中还有数据")

type (
	// TAlbumModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTAlbumModel.
	TAlbumModel interface {
		tAlbumModel
		ListAll(ctx context.Context) ([]*TAlbum, error)
		DeleteIfEmpty(ctx context.Context, id int64) error
	}

	customTAlbumModel struct {
		*defaultTAlbumModel
	}
)

// NewTAlbumModel returns a model for the database table.
func NewTAlbumModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TAlbumModel {
	return &customTAlbumModel{
		defaultTAlbumModel: newTAlbumModel(conn, c, opts...),
	}
}

func (m *customTAlbumModel) ListAll(ctx context.Context) ([]*TAlbum, error) {
	query := "SELECT " + tAlbumRows + " FROM " + strings.Trim(m.table, "`") + " ORDER BY `id` ASC"
	var rows []*TAlbum
	if err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &rows, query); err != nil {
		if err == sqlx.ErrNotFound {
			return []*TAlbum{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customTAlbumModel) DeleteIfEmpty(ctx context.Context, id int64) error {
	row, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	idKey := fmt.Sprintf("%s%v", cacheRudyGcTAlbumIdPrefix, id)
	nameKey := fmt.Sprintf("%s%v", cacheRudyGcTAlbumNamePrefix, row.Name)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		var albumResult sql.Result
		transactionErr := conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			var itemCount int64
			if countErr := session.QueryRowCtx(ctx, &itemCount, "SELECT COUNT(1) FROM `tm_album_item` WHERE `album_id` = ? FOR UPDATE", id); countErr != nil {
				return countErr
			}
			if itemCount > 0 {
				return ErrTAlbumNotEmpty
			}

			var deleteItemsErr error
			albumResult, deleteItemsErr = session.ExecCtx(ctx, "DELETE FROM `t_album` WHERE `id` = ?", id)
			if deleteItemsErr != nil {
				return deleteItemsErr
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
