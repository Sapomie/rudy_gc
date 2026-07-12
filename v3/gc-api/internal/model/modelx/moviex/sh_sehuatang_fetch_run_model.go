package moviex

import (
	"context"
	"database/sql"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShSehuatangFetchRunModel interface {
		Insert(ctx context.Context, data *ShSehuatangFetchRun) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangFetchRun, error)
		ListRecent(ctx context.Context, limit int64) ([]*ShSehuatangFetchRun, error)
		Update(ctx context.Context, data *ShSehuatangFetchRun) error
		TableName() string
	}

	defaultShSehuatangFetchRunModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShSehuatangFetchRun struct {
		Id                   int64  `db:"id"`
		FetchMode            string `db:"fetch_mode"`
		Keyword              string `db:"keyword"`
		TargetForumIDs       string `db:"target_forum_ids"`
		StartURL             string `db:"start_url"`
		CurrentURL           string `db:"current_url"`
		CurrentPage          int64  `db:"current_page"`
		Status               string `db:"status"`
		SeenCount            int64  `db:"seen_count"`
		MatchedCount         int64  `db:"matched_count"`
		SkippedExistingCount int64  `db:"skipped_existing_count"`
		SkippedForumCount    int64  `db:"skipped_forum_count"`
		FetchedCount         int64  `db:"fetched_count"`
		FailedCount          int64  `db:"failed_count"`
		StartedOn            int64  `db:"started_on"`
		FinishedOn           int64  `db:"finished_on"`
		LastError            string `db:"last_error"`
		CreatedOn            int64  `db:"created_on"`
		UpdatedOn            int64  `db:"updated_on"`
	}
)

var shSehuatangFetchRunRows = "`id`,`fetch_mode`,`keyword`,`target_forum_ids`,`start_url`,`current_url`,`current_page`,`status`,`seen_count`,`matched_count`,`skipped_existing_count`,`skipped_forum_count`,`fetched_count`,`failed_count`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`"

func NewShSehuatangFetchRunModel(conn sqlx.SqlConn) ShSehuatangFetchRunModel {
	return &defaultShSehuatangFetchRunModel{
		conn:  conn,
		table: "`sh_fetch_run`",
	}
}

func (m *defaultShSehuatangFetchRunModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangFetchRunModel) Insert(ctx context.Context, data *ShSehuatangFetchRun) (sql.Result, error) {
	query := "insert into " + m.table + " (`fetch_mode`,`keyword`,`target_forum_ids`,`start_url`,`current_url`,`current_page`,`status`,`seen_count`,`matched_count`,`skipped_existing_count`,`skipped_forum_count`,`fetched_count`,`failed_count`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.FetchMode, data.Keyword, data.TargetForumIDs, data.StartURL, data.CurrentURL, data.CurrentPage, data.Status, data.SeenCount, data.MatchedCount, data.SkippedExistingCount, data.SkippedForumCount, data.FetchedCount, data.FailedCount, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn)
}

func (m *defaultShSehuatangFetchRunModel) FindOne(ctx context.Context, id int64) (*ShSehuatangFetchRun, error) {
	var row ShSehuatangFetchRun
	query := "select " + shSehuatangFetchRunRows + " from " + m.table + " where `id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangFetchRunModel) ListRecent(ctx context.Context, limit int64) ([]*ShSehuatangFetchRun, error) {
	if limit <= 0 {
		limit = 10
	}
	query := "select " + shSehuatangFetchRunRows + " from " + m.table + " order by `id` desc limit ?"
	var rows []*ShSehuatangFetchRun
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, limit); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShSehuatangFetchRun{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *defaultShSehuatangFetchRunModel) Update(ctx context.Context, data *ShSehuatangFetchRun) error {
	query := "update " + m.table + " set `fetch_mode` = ?, `keyword` = ?, `target_forum_ids` = ?, `start_url` = ?, `current_url` = ?, `current_page` = ?, `status` = ?, `seen_count` = ?, `matched_count` = ?, `skipped_existing_count` = ?, `skipped_forum_count` = ?, `fetched_count` = ?, `failed_count` = ?, `started_on` = ?, `finished_on` = ?, `last_error` = ?, `created_on` = ?, `updated_on` = ? where `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, data.FetchMode, data.Keyword, data.TargetForumIDs, data.StartURL, data.CurrentURL, data.CurrentPage, data.Status, data.SeenCount, data.MatchedCount, data.SkippedExistingCount, data.SkippedForumCount, data.FetchedCount, data.FailedCount, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn, data.Id)
	return err
}
