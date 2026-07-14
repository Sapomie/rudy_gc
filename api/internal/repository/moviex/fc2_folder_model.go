package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
)

type FC2FolderModel interface {
	Insert(ctx context.Context, data *FC2Folder) (*FC2Folder, error)
	FindOne(ctx context.Context, id int64) (*FC2Folder, error)
	FindOneByPath(ctx context.Context, path string) (*FC2Folder, error)
	FindOneByPathSourceType(ctx context.Context, path string, sourceType int64) (*FC2Folder, error)
	Update(ctx context.Context, data *FC2Folder) error
	Delete(ctx context.Context, id int64) error
}

type FC2Folder struct {
	ID         int64  `db:"id"`
	ParentID   int64  `db:"parent_id"`
	Name       string `db:"name"`
	SourceType int64  `db:"source_type"`
	Depth      int64  `db:"depth"`
	Path       string `db:"path"`
	PathHash   []byte `db:"path_hash"`
	CreatedOn  int64  `db:"created_on"`
	UpdatedOn  int64  `db:"updated_on"`
}

type defaultFC2FolderModel struct {
	conn sqlx.SqlConn
}

const fc2FolderRows = "`id`,`parent_id`,`name`,`source_type`,`depth`,`path`,`path_hash`,`created_on`,`updated_on`"

func NewFC2FolderModel(conn sqlx.SqlConn) FC2FolderModel {
	return &defaultFC2FolderModel{conn: conn}
}

func (m *defaultFC2FolderModel) Insert(ctx context.Context, data *FC2Folder) (*FC2Folder, error) {
	result, err := m.conn.ExecCtx(ctx, `INSERT INTO fc2_folder
(`+"`parent_id`,`name`,`source_type`,`depth`,`path`,`path_hash`,`created_on`,`updated_on`"+`)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, data.ParentID, data.Name, data.SourceType, data.Depth, data.Path, data.PathHash, data.CreatedOn, data.UpdatedOn)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return m.FindOne(ctx, id)
}

func (m *defaultFC2FolderModel) FindOne(ctx context.Context, id int64) (*FC2Folder, error) {
	var row FC2Folder
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2FolderRows+" FROM `fc2_folder` WHERE `id` = ? LIMIT 1", id); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2FolderModel) FindOneByPath(ctx context.Context, path string) (*FC2Folder, error) {
	return m.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
}

func (m *defaultFC2FolderModel) FindOneByPathSourceType(ctx context.Context, path string, sourceType int64) (*FC2Folder, error) {
	var row FC2Folder
	if err := m.conn.QueryRowCtx(ctx, &row, "SELECT "+fc2FolderRows+" FROM `fc2_folder` WHERE `path` = ? AND `source_type` = ? LIMIT 1", strings.TrimSpace(path), sourceType); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultFC2FolderModel) Update(ctx context.Context, data *FC2Folder) error {
	_, err := m.conn.ExecCtx(ctx, `UPDATE fc2_folder SET
parent_id = ?, name = ?, source_type = ?, depth = ?, path = ?, path_hash = ?, created_on = ?, updated_on = ?
WHERE id = ?`, data.ParentID, data.Name, data.SourceType, data.Depth, data.Path, data.PathHash, data.CreatedOn, data.UpdatedOn, data.ID)
	return err
}

func (m *defaultFC2FolderModel) Delete(ctx context.Context, id int64) error {
	_, err := m.conn.ExecCtx(ctx, "DELETE FROM `fc2_folder` WHERE `id` = ?", id)
	return err
}
