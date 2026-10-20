package moviex

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShForumModel interface {
		ListEnabled(ctx context.Context) ([]*ShForum, error)
		TableName() string
	}

	defaultShForumModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShForum struct {
		ForumID   string `db:"forum_id" json:"forum_id"`
		ForumName string `db:"forum_name" json:"forum_name"`
		SortOrder int64  `db:"sort_order" json:"sort_order"`
	}
)

func NewShForumModel(conn sqlx.SqlConn) ShForumModel {
	return &defaultShForumModel{conn: conn, table: "`sh_forum`"}
}

func (m *defaultShForumModel) TableName() string { return m.table }

func (m *defaultShForumModel) ListEnabled(ctx context.Context) ([]*ShForum, error) {
	var rows []*ShForum
	query := "select `forum_id`,`forum_name`,`sort_order` from " + m.table + " where `is_enabled` = 1 order by `sort_order` asc, `forum_id` asc"
	if err := m.conn.QueryRowsCtx(ctx, &rows, query); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShForum{}, nil
		}
		return nil, err
	}
	return rows, nil
}
