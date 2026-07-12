package moviex

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShSehuatangThreadDetailFetchRunModel interface {
		Insert(ctx context.Context, data *ShSehuatangThreadDetailFetchRun) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangThreadDetailFetchRun, error)
		FindOneByJobID(ctx context.Context, jobID string) (*ShSehuatangThreadDetailFetchRun, error)
		Update(ctx context.Context, data *ShSehuatangThreadDetailFetchRun) error
		TableName() string
	}

	defaultShSehuatangThreadDetailFetchRunModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShSehuatangThreadDetailFetchRun struct {
		Id                 int64  `db:"id"`
		JobID              string `db:"job_id"`
		TriggerSource      string `db:"trigger_source"`
		TargetForumIDs     string `db:"target_forum_ids"`
		SelectionCriteria  string `db:"selection_criteria"`
		SelectedThreadTIDs string `db:"selected_thread_tids"`
		SelectedCount      int64  `db:"selected_count"`
		CurrentIndex       int64  `db:"current_index"`
		CurrentThreadTID   string `db:"current_thread_tid"`
		CurrentThreadURL   string `db:"current_thread_url"`
		CompletedCount     int64  `db:"completed_count"`
		FetchedCount       int64  `db:"fetched_count"`
		FailedCount        int64  `db:"failed_count"`
		Status             string `db:"status"`
		StopReason         string `db:"stop_reason"`
		StartedOn          int64  `db:"started_on"`
		FinishedOn         int64  `db:"finished_on"`
		LastError          string `db:"last_error"`
		CreatedOn          int64  `db:"created_on"`
		UpdatedOn          int64  `db:"updated_on"`
	}
)

var shSehuatangThreadDetailFetchRunRows = "`id`,`job_id`,`trigger_source`,`target_forum_ids`,`selection_criteria`,`selected_thread_tids`,`selected_count`,`current_index`,`current_thread_tid`,`current_thread_url`,`completed_count`,`fetched_count`,`failed_count`,`status`,`stop_reason`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`"

func NewShSehuatangThreadDetailFetchRunModel(conn sqlx.SqlConn) ShSehuatangThreadDetailFetchRunModel {
	return &defaultShSehuatangThreadDetailFetchRunModel{conn: conn, table: "`sh_detail_run`"}
}

func (m *defaultShSehuatangThreadDetailFetchRunModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangThreadDetailFetchRunModel) Insert(ctx context.Context, data *ShSehuatangThreadDetailFetchRun) (sql.Result, error) {
	query := "insert into " + m.table + " (`job_id`,`trigger_source`,`target_forum_ids`,`selection_criteria`,`selected_thread_tids`,`selected_count`,`current_index`,`current_thread_tid`,`current_thread_url`,`completed_count`,`fetched_count`,`failed_count`,`status`,`stop_reason`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.JobID, data.TriggerSource, data.TargetForumIDs, data.SelectionCriteria, data.SelectedThreadTIDs, data.SelectedCount, data.CurrentIndex, data.CurrentThreadTID, data.CurrentThreadURL, data.CompletedCount, data.FetchedCount, data.FailedCount, data.Status, data.StopReason, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn)
}

func (m *defaultShSehuatangThreadDetailFetchRunModel) FindOne(ctx context.Context, id int64) (*ShSehuatangThreadDetailFetchRun, error) {
	var row ShSehuatangThreadDetailFetchRun
	query := "select " + shSehuatangThreadDetailFetchRunRows + " from " + m.table + " where `id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadDetailFetchRunModel) FindOneByJobID(ctx context.Context, jobID string) (*ShSehuatangThreadDetailFetchRun, error) {
	var row ShSehuatangThreadDetailFetchRun
	query := "select " + shSehuatangThreadDetailFetchRunRows + " from " + m.table + " where `job_id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, jobID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadDetailFetchRunModel) Update(ctx context.Context, data *ShSehuatangThreadDetailFetchRun) error {
	query := "update " + m.table + " set `job_id` = ?, `trigger_source` = ?, `target_forum_ids` = ?, `selection_criteria` = ?, `selected_thread_tids` = ?, `selected_count` = ?, `current_index` = ?, `current_thread_tid` = ?, `current_thread_url` = ?, `completed_count` = ?, `fetched_count` = ?, `failed_count` = ?, `status` = ?, `stop_reason` = ?, `started_on` = ?, `finished_on` = ?, `last_error` = ?, `created_on` = ?, `updated_on` = ? where `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, data.JobID, data.TriggerSource, data.TargetForumIDs, data.SelectionCriteria, data.SelectedThreadTIDs, data.SelectedCount, data.CurrentIndex, data.CurrentThreadTID, data.CurrentThreadURL, data.CompletedCount, data.FetchedCount, data.FailedCount, data.Status, data.StopReason, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn, data.Id)
	return err
}
