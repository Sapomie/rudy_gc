package moviex

import (
	"context"
	"database/sql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShSehuatangThreadListFetchRunModel interface {
		Insert(ctx context.Context, data *ShSehuatangThreadListFetchRun) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ShSehuatangThreadListFetchRun, error)
		FindOneByJobID(ctx context.Context, jobID string) (*ShSehuatangThreadListFetchRun, error)
		Update(ctx context.Context, data *ShSehuatangThreadListFetchRun) error
		TableName() string
	}

	defaultShSehuatangThreadListFetchRunModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShSehuatangThreadListFetchRun struct {
		Id                   int64  `db:"id"`
		JobID                string `db:"job_id"`
		TargetForumIDs       string `db:"target_forum_ids"`
		StartPage            int64  `db:"start_page"`
		EndPage              int64  `db:"end_page"`
		CurrentForumID       string `db:"current_forum_id"`
		CurrentPage          int64  `db:"current_page"`
		CurrentURL           string `db:"current_url"`
		TotalPageCount       int64  `db:"total_page_count"`
		CompletedPageCount   int64  `db:"completed_page_count"`
		SuccessPageCount     int64  `db:"success_page_count"`
		FailedPageCount      int64  `db:"failed_page_count"`
		SeenCount            int64  `db:"seen_count"`
		InsertedCount        int64  `db:"inserted_count"`
		ExistingCount        int64  `db:"existing_count"`
		SkippedTopCount      int64  `db:"skipped_top_count"`
		FailedCount          int64  `db:"failed_count"`
		CoveragePostTimeFrom int64  `db:"coverage_post_time_from"`
		CoveragePostTimeTo   int64  `db:"coverage_post_time_to"`
		CoverageComplete     bool   `db:"coverage_complete"`
		PageCoverage         string `db:"page_coverage"`
		Status               string `db:"status"`
		StopReason           string `db:"stop_reason"`
		StartedOn            int64  `db:"started_on"`
		FinishedOn           int64  `db:"finished_on"`
		LastError            string `db:"last_error"`
		CreatedOn            int64  `db:"created_on"`
		UpdatedOn            int64  `db:"updated_on"`
	}
)

var shSehuatangThreadListFetchRunRows = "`id`,`job_id`,`target_forum_ids`,`start_page`,`end_page`,`current_forum_id`,`current_page`,`current_url`,`total_page_count`,`completed_page_count`,`success_page_count`,`failed_page_count`,`seen_count`,`inserted_count`,`existing_count`,`skipped_top_count`,`failed_count`,`coverage_post_time_from`,`coverage_post_time_to`,`coverage_complete`,`page_coverage`,`status`,`stop_reason`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`"

func NewShSehuatangThreadListFetchRunModel(conn sqlx.SqlConn) ShSehuatangThreadListFetchRunModel {
	return &defaultShSehuatangThreadListFetchRunModel{conn: conn, table: "`sh_list_run`"}
}

func (m *defaultShSehuatangThreadListFetchRunModel) TableName() string {
	return m.table
}

func (m *defaultShSehuatangThreadListFetchRunModel) Insert(ctx context.Context, data *ShSehuatangThreadListFetchRun) (sql.Result, error) {
	query := "insert into " + m.table + " (`job_id`,`target_forum_ids`,`start_page`,`end_page`,`current_forum_id`,`current_page`,`current_url`,`total_page_count`,`completed_page_count`,`success_page_count`,`failed_page_count`,`seen_count`,`inserted_count`,`existing_count`,`skipped_top_count`,`failed_count`,`coverage_post_time_from`,`coverage_post_time_to`,`coverage_complete`,`page_coverage`,`status`,`stop_reason`,`started_on`,`finished_on`,`last_error`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return m.conn.ExecCtx(ctx, query, data.JobID, data.TargetForumIDs, data.StartPage, data.EndPage, data.CurrentForumID, data.CurrentPage, data.CurrentURL, data.TotalPageCount, data.CompletedPageCount, data.SuccessPageCount, data.FailedPageCount, data.SeenCount, data.InsertedCount, data.ExistingCount, data.SkippedTopCount, data.FailedCount, data.CoveragePostTimeFrom, data.CoveragePostTimeTo, data.CoverageComplete, data.PageCoverage, data.Status, data.StopReason, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn)
}

func (m *defaultShSehuatangThreadListFetchRunModel) FindOne(ctx context.Context, id int64) (*ShSehuatangThreadListFetchRun, error) {
	var row ShSehuatangThreadListFetchRun
	query := "select " + shSehuatangThreadListFetchRunRows + " from " + m.table + " where `id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, id); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadListFetchRunModel) FindOneByJobID(ctx context.Context, jobID string) (*ShSehuatangThreadListFetchRun, error) {
	var row ShSehuatangThreadListFetchRun
	query := "select " + shSehuatangThreadListFetchRunRows + " from " + m.table + " where `job_id` = ? limit 1"
	if err := m.conn.QueryRowCtx(ctx, &row, query, jobID); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (m *defaultShSehuatangThreadListFetchRunModel) Update(ctx context.Context, data *ShSehuatangThreadListFetchRun) error {
	query := "update " + m.table + " set `job_id` = ?, `target_forum_ids` = ?, `start_page` = ?, `end_page` = ?, `current_forum_id` = ?, `current_page` = ?, `current_url` = ?, `total_page_count` = ?, `completed_page_count` = ?, `success_page_count` = ?, `failed_page_count` = ?, `seen_count` = ?, `inserted_count` = ?, `existing_count` = ?, `skipped_top_count` = ?, `failed_count` = ?, `coverage_post_time_from` = ?, `coverage_post_time_to` = ?, `coverage_complete` = ?, `page_coverage` = ?, `status` = ?, `stop_reason` = ?, `started_on` = ?, `finished_on` = ?, `last_error` = ?, `created_on` = ?, `updated_on` = ? where `id` = ?"
	_, err := m.conn.ExecCtx(ctx, query, data.JobID, data.TargetForumIDs, data.StartPage, data.EndPage, data.CurrentForumID, data.CurrentPage, data.CurrentURL, data.TotalPageCount, data.CompletedPageCount, data.SuccessPageCount, data.FailedPageCount, data.SeenCount, data.InsertedCount, data.ExistingCount, data.SkippedTopCount, data.FailedCount, data.CoveragePostTimeFrom, data.CoveragePostTimeTo, data.CoverageComplete, data.PageCoverage, data.Status, data.StopReason, data.StartedOn, data.FinishedOn, data.LastError, data.CreatedOn, data.UpdatedOn, data.Id)
	return err
}
