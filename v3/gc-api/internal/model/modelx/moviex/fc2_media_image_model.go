package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type FC2MediaImageModel interface {
	Insert(ctx context.Context, data *FC2MediaImage) (*FC2MediaImage, error)
	FindOneByMediaIDAndID(ctx context.Context, mediaID, id int64) (*FC2MediaImage, error)
	ListByMediaID(ctx context.Context, mediaID int64) ([]*FC2MediaImage, error)
	ListByMediaIDs(ctx context.Context, mediaIDs []int64) ([]*FC2MediaImage, error)
	DeleteByMediaIDAndID(ctx context.Context, mediaID, id int64) error
}

type FC2MediaImage struct {
	ID         int64  `db:"id" json:"ID"`
	FC2MediaID int64  `db:"fc2_media_id" json:"FC2MediaID"`
	ImageType  string `db:"image_type" json:"ImageType"`
	FilePath   string `db:"file_path" json:"FilePath"`
	CreatedOn  int64  `db:"created_on" json:"CreatedOn"`
}

type defaultFC2MediaImageModel struct {
	conn sqlx.SqlConn
}

const fc2MediaImageRows = "`id`,`fc2_media_id`,`image_type`,`file_path`,`created_on`"

func NewFC2MediaImageModel(conn sqlx.SqlConn) FC2MediaImageModel {
	return &defaultFC2MediaImageModel{conn: conn}
}

func (m *defaultFC2MediaImageModel) Insert(ctx context.Context, data *FC2MediaImage) (*FC2MediaImage, error) {
	result, err := m.conn.ExecCtx(ctx, `INSERT INTO fc2_media_image
(`+"`fc2_media_id`,`image_type`,`file_path`,`created_on`"+`)
VALUES (?, ?, ?, ?)`, data.FC2MediaID, data.ImageType, data.FilePath, data.CreatedOn)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	row := *data
	row.ID = id
	return &row, nil
}

func (m *defaultFC2MediaImageModel) FindOneByMediaIDAndID(ctx context.Context, mediaID, id int64) (*FC2MediaImage, error) {
	var row FC2MediaImage
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2MediaImageRows+" FROM `fc2_media_image` WHERE `fc2_media_id` = ? AND `id` = ? LIMIT 1", mediaID, id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2MediaImageModel) ListByMediaID(ctx context.Context, mediaID int64) ([]*FC2MediaImage, error) {
	if mediaID <= 0 {
		return []*FC2MediaImage{}, nil
	}
	rows := make([]*FC2MediaImage, 0)
	err := m.conn.QueryRowsCtx(ctx, &rows, "SELECT "+fc2MediaImageRows+" FROM `fc2_media_image` WHERE `fc2_media_id` = ? ORDER BY `created_on` DESC, `id` DESC", mediaID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2MediaImage{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaImageModel) ListByMediaIDs(ctx context.Context, mediaIDs []int64) ([]*FC2MediaImage, error) {
	ids := uniquePositiveInt64s(mediaIDs)
	if len(ids) == 0 {
		return []*FC2MediaImage{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows := make([]*FC2MediaImage, 0)
	query := "SELECT " + fc2MediaImageRows + " FROM `fc2_media_image` WHERE `fc2_media_id` IN (" + placeholders + ") ORDER BY `fc2_media_id` ASC, `created_on` DESC, `id` DESC"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*FC2MediaImage{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultFC2MediaImageModel) DeleteByMediaIDAndID(ctx context.Context, mediaID, id int64) error {
	result, err := m.conn.ExecCtx(ctx, "DELETE FROM `fc2_media_image` WHERE `fc2_media_id` = ? AND `id` = ?", mediaID, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
