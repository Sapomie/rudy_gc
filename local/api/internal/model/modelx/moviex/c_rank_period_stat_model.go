package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CRankPeriodStat struct {
	Id                   int64  `db:"id"`
	PeriodType           int64  `db:"period_type"`
	PeriodKey            string `db:"period_key"`
	StartDayNumber       int64  `db:"start_day_number"`
	EndDayNumber         int64  `db:"end_day_number"`
	SourceRankRowCount   int64  `db:"source_rank_row_count"`
	SourceRankDayCount   int64  `db:"source_rank_day_count"`
	SourceMovieCount     int64  `db:"source_movie_count"`
	MissingRankDayCount  int64  `db:"missing_rank_day_count"`
	PeriodMovieCount     int64  `db:"period_movie_count"`
	NewEntryMovieCount   int64  `db:"new_entry_movie_count"`
	SourceFetchedStartOn int64  `db:"source_fetched_start_on"`
	SourceFetchedEndOn   int64  `db:"source_fetched_end_on"`
	SourceChecksum       string `db:"source_checksum"`
	CalculatedOn         int64  `db:"calculated_on"`
	CreatedOn            int64  `db:"created_on"`
	UpdatedOn            int64  `db:"updated_on"`
}

type CRankPeriodStatListQuery struct {
	PeriodType        int64
	StartDayNumber    int64
	EndDayNumber      int64
	HasStartDayNumber bool
	HasEndDayNumber   bool
	Page              int64
	PageSize          int64
	Sort              string
	Order             string
}

type CRankPeriodStatSummary struct {
	PeriodCount         int64 `db:"period_count"`
	SourceRankRowCount  int64 `db:"source_rank_row_count"`
	SourceRankDayCount  int64 `db:"source_rank_day_count"`
	SourceMovieCount    int64 `db:"source_movie_count"`
	MissingRankDayCount int64 `db:"missing_rank_day_count"`
	PeriodMovieCount    int64 `db:"period_movie_count"`
	NewEntryMovieCount  int64 `db:"new_entry_movie_count"`
	LatestCalculatedOn  int64 `db:"latest_calculated_on"`
}

type CRankPeriodStatPage struct {
	Items   []*CRankPeriodStat
	Total   int64
	Page    int64
	Summary *CRankPeriodStatSummary
}

type CRankPeriodStatModel interface {
	Upsert(ctx context.Context, rows []*CRankPeriodStat) error
	ListPage(ctx context.Context, query CRankPeriodStatListQuery) (*CRankPeriodStatPage, error)
}

type customCRankPeriodStatModel struct {
	conn sqlx.SqlConn
}

func NewCRankPeriodStatModel(conn sqlx.SqlConn) CRankPeriodStatModel {
	return &customCRankPeriodStatModel{conn: conn}
}

func (m *customCRankPeriodStatModel) Upsert(ctx context.Context, rows []*CRankPeriodStat) error {
	const query = `
INSERT INTO c_rank_period_stat (
  period_type, period_key, start_day_number, end_day_number,
  source_rank_row_count, source_rank_day_count, source_movie_count, missing_rank_day_count,
  period_movie_count, new_entry_movie_count,
  source_fetched_start_on, source_fetched_end_on, source_checksum,
  calculated_on, created_on, updated_on
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  start_day_number = IF(source_checksum <> VALUES(source_checksum), VALUES(start_day_number), start_day_number),
  end_day_number = IF(source_checksum <> VALUES(source_checksum), VALUES(end_day_number), end_day_number),
  source_rank_row_count = IF(source_checksum <> VALUES(source_checksum), VALUES(source_rank_row_count), source_rank_row_count),
  source_rank_day_count = IF(source_checksum <> VALUES(source_checksum), VALUES(source_rank_day_count), source_rank_day_count),
  source_movie_count = IF(source_checksum <> VALUES(source_checksum), VALUES(source_movie_count), source_movie_count),
  missing_rank_day_count = IF(source_checksum <> VALUES(source_checksum), VALUES(missing_rank_day_count), missing_rank_day_count),
  period_movie_count = IF(source_checksum <> VALUES(source_checksum), VALUES(period_movie_count), period_movie_count),
  new_entry_movie_count = IF(source_checksum <> VALUES(source_checksum), VALUES(new_entry_movie_count), new_entry_movie_count),
  source_fetched_start_on = IF(source_checksum <> VALUES(source_checksum), VALUES(source_fetched_start_on), source_fetched_start_on),
  source_fetched_end_on = IF(source_checksum <> VALUES(source_checksum), VALUES(source_fetched_end_on), source_fetched_end_on),
  calculated_on = IF(source_checksum <> VALUES(source_checksum), VALUES(calculated_on), calculated_on),
  updated_on = IF(source_checksum <> VALUES(source_checksum), VALUES(updated_on), updated_on),
  source_checksum = VALUES(source_checksum)`

	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, row := range rows {
			if row == nil {
				continue
			}
			if _, err := session.ExecCtx(ctx, query,
				row.PeriodType,
				row.PeriodKey,
				row.StartDayNumber,
				row.EndDayNumber,
				row.SourceRankRowCount,
				row.SourceRankDayCount,
				row.SourceMovieCount,
				row.MissingRankDayCount,
				row.PeriodMovieCount,
				row.NewEntryMovieCount,
				row.SourceFetchedStartOn,
				row.SourceFetchedEndOn,
				row.SourceChecksum,
				row.CalculatedOn,
				row.CreatedOn,
				row.UpdatedOn,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *customCRankPeriodStatModel) ListPage(ctx context.Context, filter CRankPeriodStatListQuery) (*CRankPeriodStatPage, error) {
	where, args := rankPeriodStatWhere(filter)
	summaryQuery := `
SELECT
  COUNT(*) AS period_count,
  COALESCE(SUM(stat.source_rank_row_count), 0) AS source_rank_row_count,
  COALESCE(SUM(stat.source_rank_day_count), 0) AS source_rank_day_count,
  COALESCE(SUM(stat.source_movie_count), 0) AS source_movie_count,
  COALESCE(SUM(stat.missing_rank_day_count), 0) AS missing_rank_day_count,
  COALESCE(SUM(stat.period_movie_count), 0) AS period_movie_count,
  COALESCE(SUM(stat.new_entry_movie_count), 0) AS new_entry_movie_count,
  COALESCE(MAX(stat.calculated_on), 0) AS latest_calculated_on
FROM c_rank_period_stat AS stat` + where

	summary := new(CRankPeriodStatSummary)
	if err := m.conn.QueryRowCtx(ctx, summary, summaryQuery, args...); err != nil {
		return nil, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 24
	}
	pageTotal := summary.PeriodCount / pageSize
	if summary.PeriodCount%pageSize != 0 {
		pageTotal++
	}
	if pageTotal == 0 {
		page = 1
	} else if page > pageTotal {
		page = pageTotal
	}

	query := `SELECT stat.* FROM c_rank_period_stat AS stat` + where

	orderDirection := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		orderDirection = "ASC"
	}
	orderColumn := rankPeriodStatOrderColumn(filter.Sort)
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
	if orderColumn != "stat.start_day_number" {
		query += fmt.Sprintf(", stat.start_day_number %s", orderDirection)
	}
	query += fmt.Sprintf(", stat.period_type ASC, stat.id %s LIMIT ? OFFSET ?", orderDirection)
	pageArgs := append(append([]any(nil), args...), pageSize, (page-1)*pageSize)

	var rows []*CRankPeriodStat
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, pageArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			rows = []*CRankPeriodStat{}
		} else {
			return nil, err
		}
	}
	return &CRankPeriodStatPage{
		Items:   rows,
		Total:   summary.PeriodCount,
		Page:    page,
		Summary: summary,
	}, nil
}

func rankPeriodStatWhere(filter CRankPeriodStatListQuery) (string, []any) {
	where := ` WHERE 1 = 1`
	args := make([]any, 0, 4)
	if filter.PeriodType > 0 {
		where += " AND stat.period_type = ?"
		args = append(args, filter.PeriodType)
	}
	if filter.HasStartDayNumber {
		where += " AND stat.end_day_number >= ?"
		args = append(args, filter.StartDayNumber)
	}
	if filter.HasEndDayNumber {
		where += " AND stat.start_day_number <= ?"
		args = append(args, filter.EndDayNumber)
	}
	return where, args
}

func rankPeriodStatOrderColumn(sort string) string {
	switch strings.TrimSpace(sort) {
	case "source_rank_row_count":
		return "stat.source_rank_row_count"
	case "source_rank_day_count":
		return "stat.source_rank_day_count"
	case "source_movie_count":
		return "stat.source_movie_count"
	case "missing_rank_day_count":
		return "stat.missing_rank_day_count"
	case "period_movie_count":
		return "stat.period_movie_count"
	case "new_entry_movie_count":
		return "stat.new_entry_movie_count"
	case "source_fetched_start_on":
		return "stat.source_fetched_start_on"
	case "source_fetched_end_on":
		return "stat.source_fetched_end_on"
	case "calculated_on":
		return "stat.calculated_on"
	default:
		return "stat.start_day_number"
	}
}
