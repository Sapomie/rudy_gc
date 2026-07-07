package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShSehuatangFetchItemModel interface {
		Insert(ctx context.Context, data *ShSehuatangFetchItem) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangFetchItem, error)
		FindOneByRunThread(ctx context.Context, runID int64, threadTID string) (*ShSehuatangFetchItem, error)
		ListByThreadTID(ctx context.Context, threadTID string, limit int64) ([]*ShSehuatangFetchItem, error)
		Upsert(ctx context.Context, data *ShSehuatangFetchItem) (*ShSehuatangFetchItem, error)
		UpdateStatus(ctx context.Context, runID int64, threadTID string, status string, threadRowID int64, errText string, now int64) (*ShSehuatangFetchItem, error)
		TableName() string
	}

	defaultShSehuatangFetchItemModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShSehuatangFetchItem struct {
		Id            int64  `db:"id"`
		RunID         int64  `db:"run_id"`
		FetchMode     string `db:"fetch_mode"`
		SourceForumID string `db:"source_forum_id"`
		SourcePage    int64  `db:"source_page"`
		SourcePageURL string `db:"source_page_url"`
		ThreadTID     string `db:"thread_tid"`
		ThreadURL     string `db:"thread_url"`
		ThreadTitle   string `db:"thread_title"`
		AuthorName    string `db:"author_name"`
		Status        string `db:"status"`
		ThreadRowID   int64  `db:"thread_row_id"`
		Error         string `db:"error"`
		CreatedOn     int64  `db:"created_on"`
		UpdatedOn     int64  `db:"updated_on"`
	}
)

var shSehuatangFetchItemRows = "`id`,`run_id`,`fetch_mode`,`source_forum_id`,`source_page`,`source_page_url`,`thread_tid`,`thread_url`,`thread_title`,`author_name`,`status`,`thread_row_id`,`error`,`created_on`,`updated_on`"

func NewShSehuatangFetchItemModel(conn sqlx.SqlConn) ShSehuatangFetchItemModel {
	return &defaultShSehuatangFetchItemModel{
		conn:  conn,
		table: "`sh_sehuatang_fetch_item`",
	}
}

func (m *defaultShSehuatangFetchItemModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangFetchItemModel) Insert(ctx context.Context, data *ShSehuatangFetchItem) (sql.Result, error) {
	query := "insert into " + m.table + " (`run_id`,`fetch_mode`,`source_forum_id`,`source_page`,`source_page_url`,`thread_tid`,`thread_url`,`thread_title`,`author_name`,`status`,`thread_row_id`,`error`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.RunID, data.FetchMode, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.AuthorName, data.Status, data.ThreadRowID, data.Error, data.CreatedOn, data.UpdatedOn)
}

func (m *defaultShSehuatangFetchItemModel) FindOne(ctx context.Context, id int64) (*ShSehuatangFetchItem, error) {
	var row ShSehuatangFetchItem
	query := "select " + shSehuatangFetchItemRows + " from " + m.table + " where `id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangFetchItemModel) FindOneByRunThread(ctx context.Context, runID int64, threadTID string) (*ShSehuatangFetchItem, error) {
	threadTID = strings.TrimSpace(threadTID)
	if runID <= 0 || threadTID == "" {
		return nil, ErrNotFound
	}
	var row ShSehuatangFetchItem
	query := "select " + shSehuatangFetchItemRows + " from " + m.table + " where `run_id` = ? and `thread_tid` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, runID, threadTID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangFetchItemModel) ListByThreadTID(ctx context.Context, threadTID string, limit int64) ([]*ShSehuatangFetchItem, error) {
	threadTID = strings.TrimSpace(threadTID)
	if threadTID == "" {
		return []*ShSehuatangFetchItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}
	query := "select " + shSehuatangFetchItemRows + " from " + m.table + " where `thread_tid` = ? order by `run_id` desc, `id` desc limit ?"
	var rows []*ShSehuatangFetchItem
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, threadTID, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShSehuatangFetchItem{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShSehuatangFetchItemModel) Upsert(ctx context.Context, data *ShSehuatangFetchItem) (*ShSehuatangFetchItem, error) {
	query := "insert into " + m.table + " (`run_id`,`fetch_mode`,`source_forum_id`,`source_page`,`source_page_url`,`thread_tid`,`thread_url`,`thread_title`,`author_name`,`status`,`thread_row_id`,`error`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `fetch_mode` = values(`fetch_mode`), `source_forum_id` = values(`source_forum_id`), `source_page` = values(`source_page`), `source_page_url` = values(`source_page_url`), `thread_url` = values(`thread_url`), `thread_title` = values(`thread_title`), `author_name` = values(`author_name`), `status` = values(`status`), `thread_row_id` = values(`thread_row_id`), `error` = values(`error`), `updated_on` = values(`updated_on`)"
	if _, err := m.conn.ExecCtx(ctx, query, data.RunID, data.FetchMode, data.SourceForumID, data.SourcePage, data.SourcePageURL, data.ThreadTID, data.ThreadURL, data.ThreadTitle, data.AuthorName, data.Status, data.ThreadRowID, data.Error, data.CreatedOn, data.UpdatedOn); err != nil {
		return nil, err
	}
	return m.FindOneByRunThread(ctx, data.RunID, data.ThreadTID)
}

func (m *defaultShSehuatangFetchItemModel) UpdateStatus(ctx context.Context, runID int64, threadTID string, status string, threadRowID int64, errText string, now int64) (*ShSehuatangFetchItem, error) {
	threadTID = strings.TrimSpace(threadTID)
	if runID <= 0 || threadTID == "" {
		return nil, ErrNotFound
	}
	query := "update " + m.table + " set `status` = ?, `thread_row_id` = ?, `error` = ?, `updated_on` = ? where `run_id` = ? and `thread_tid` = ?"
	if _, err := m.conn.ExecCtx(ctx, query, status, threadRowID, errText, now, runID, threadTID); err != nil {
		return nil, err
	}
	return m.FindOneByRunThread(ctx, runID, threadTID)
}
