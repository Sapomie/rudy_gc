package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CRankDayStat struct {
	Id                  int64  `db:"id"`
	DayNumber           int64  `db:"day_number"`
	RankDate            string `db:"rank_date"`
	PrevDayNumber       int64  `db:"prev_day_number"`
	RankRowCount        int64  `db:"rank_row_count"`
	MovieCount          int64  `db:"movie_count"`
	ContinuedMovieCount int64  `db:"continued_movie_count"`
	FirstSeenMovieCount int64  `db:"first_seen_movie_count"`
	ReturningMovieCount int64  `db:"returning_movie_count"`
	DroppedMovieCount   int64  `db:"dropped_movie_count"`
	SourceChecksum      string `db:"source_checksum"`
	FetchedOn           int64  `db:"fetched_on"`
	SyncedOn            int64  `db:"synced_on"`
	CalculatedOn        int64  `db:"calculated_on"`
	CreatedOn           int64  `db:"created_on"`
	UpdatedOn           int64  `db:"updated_on"`
}

type CRankDayStatListQuery struct {
	StartDayNumber    int64
	EndDayNumber      int64
	HasStartDayNumber bool
	HasEndDayNumber   bool
	Page              int64
	PageSize          int64
	Sort              string
	Order             string
}

type CRankDayStatSummary struct {
	DayCount            int64 `db:"day_count"`
	RankRowCount        int64 `db:"rank_row_count"`
	MovieCount          int64 `db:"movie_count"`
	ContinuedMovieCount int64 `db:"continued_movie_count"`
	FirstSeenMovieCount int64 `db:"first_seen_movie_count"`
	ReturningMovieCount int64 `db:"returning_movie_count"`
	DroppedMovieCount   int64 `db:"dropped_movie_count"`
	LatestCalculatedOn  int64 `db:"latest_calculated_on"`
}

type CRankDayStatPage struct {
	Items   []*CRankDayStat
	Total   int64
	Page    int64
	Summary *CRankDayStatSummary
}

type CRankDayStatModel interface {
	InsertMissing(ctx context.Context, rows []*CRankDayStat) error
	ListAll(ctx context.Context) ([]*CRankDayStat, error)
	ListPage(ctx context.Context, query CRankDayStatListQuery) (*CRankDayStatPage, error)
}

type customCRankDayStatModel struct {
	conn sqlx.SqlConn
}

func NewCRankDayStatModel(conn sqlx.SqlConn) CRankDayStatModel {
	return &customCRankDayStatModel{conn: conn}
}

func (m *customCRankDayStatModel) InsertMissing(ctx context.Context, rows []*CRankDayStat) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		const query = `
INSERT INTO c_rank_day_stat (
  day_number, rank_date, prev_day_number, rank_row_count, movie_count,
  continued_movie_count, first_seen_movie_count, returning_movie_count, dropped_movie_count,
  source_checksum, fetched_on, synced_on, calculated_on, created_on, updated_on
) SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
WHERE NOT EXISTS (
  SELECT 1 FROM c_rank_day_stat WHERE day_number = ?
)`

		for _, row := range rows {
			if row == nil {
				continue
			}
			if _, err := session.ExecCtx(ctx, query,
				row.DayNumber,
				row.RankDate,
				row.PrevDayNumber,
				row.RankRowCount,
				row.MovieCount,
				row.ContinuedMovieCount,
				row.FirstSeenMovieCount,
				row.ReturningMovieCount,
				row.DroppedMovieCount,
				row.SourceChecksum,
				row.FetchedOn,
				row.SyncedOn,
				row.CalculatedOn,
				row.CreatedOn,
				row.UpdatedOn,
				row.DayNumber,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *customCRankDayStatModel) ListAll(ctx context.Context) ([]*CRankDayStat, error) {
	var rows []*CRankDayStat
	if err := m.conn.QueryRowsCtx(ctx, &rows, `SELECT * FROM c_rank_day_stat ORDER BY day_number ASC, id ASC`); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return []*CRankDayStat{}, nil
		}
		return nil, err
	}
	return rows, nil
}

func (m *customCRankDayStatModel) ListPage(ctx context.Context, filter CRankDayStatListQuery) (*CRankDayStatPage, error) {
	where, args := rankDayStatWhere(filter)
	summaryQuery := `
SELECT
  COUNT(*) AS day_count,
  COALESCE(SUM(stat.rank_row_count), 0) AS rank_row_count,
  COALESCE(SUM(stat.movie_count), 0) AS movie_count,
  COALESCE(SUM(stat.continued_movie_count), 0) AS continued_movie_count,
  COALESCE(SUM(stat.first_seen_movie_count), 0) AS first_seen_movie_count,
  COALESCE(SUM(stat.returning_movie_count), 0) AS returning_movie_count,
  COALESCE(SUM(stat.dropped_movie_count), 0) AS dropped_movie_count,
  COALESCE(MAX(stat.calculated_on), 0) AS latest_calculated_on
FROM c_rank_day_stat AS stat` + where

	summary := new(CRankDayStatSummary)
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
	pageTotal := summary.DayCount / pageSize
	if summary.DayCount%pageSize != 0 {
		pageTotal++
	}
	if pageTotal == 0 {
		page = 1
	} else if page > pageTotal {
		page = pageTotal
	}

	orderColumn := rankDayStatOrderColumn(filter.Sort)
	orderDirection := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		orderDirection = "ASC"
	}

	query := `SELECT stat.* FROM c_rank_day_stat AS stat` + where
	query += fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDirection)
	if orderColumn != "stat.day_number" {
		query += fmt.Sprintf(", stat.day_number %s", orderDirection)
	}
	query += fmt.Sprintf(", stat.id %s LIMIT ? OFFSET ?", orderDirection)
	pageArgs := append(append([]any(nil), args...), pageSize, (page-1)*pageSize)

	var rows []*CRankDayStat
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, pageArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			rows = []*CRankDayStat{}
		} else {
			return nil, err
		}
	}
	return &CRankDayStatPage{
		Items:   rows,
		Total:   summary.DayCount,
		Page:    page,
		Summary: summary,
	}, nil
}

func rankDayStatWhere(filter CRankDayStatListQuery) (string, []any) {
	where := ` WHERE 1 = 1`
	args := make([]any, 0, 2)
	if filter.HasStartDayNumber {
		where += " AND stat.day_number >= ?"
		args = append(args, filter.StartDayNumber)
	}
	if filter.HasEndDayNumber {
		where += " AND stat.day_number <= ?"
		args = append(args, filter.EndDayNumber)
	}
	return where, args
}

func rankDayStatOrderColumn(sort string) string {
	switch strings.TrimSpace(sort) {
	case "rank_row_count":
		return "stat.rank_row_count"
	case "movie_count":
		return "stat.movie_count"
	case "continued_movie_count":
		return "stat.continued_movie_count"
	case "first_seen_movie_count":
		return "stat.first_seen_movie_count"
	case "returning_movie_count":
		return "stat.returning_movie_count"
	case "dropped_movie_count":
		return "stat.dropped_movie_count"
	case "calculated_on":
		return "stat.calculated_on"
	default:
		return "stat.day_number"
	}
}
