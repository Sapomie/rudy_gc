package moviex

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type (
	ShThreadBucketStatListFilter struct {
		AggScope      string
		Level         string
		BucketEndFrom *int64
		BucketStartTo *int64
		Sort          string
		Dir           string
		Page          int64
		PageSize      int64
	}

	ShThreadBucketStatModel interface {
		Upsert(ctx context.Context, data *ShThreadBucketStat) error
		DeleteByAggScopeScopeKey(ctx context.Context, aggScope, scopeKey string) error
		DeleteAll(ctx context.Context) (int64, error)
		CountByLevel(ctx context.Context, level string) (int64, error)
		ListPage(ctx context.Context, filter ShThreadBucketStatListFilter) ([]*ShThreadBucketStat, int64, error)
		TableName() string
	}

	defaultShThreadBucketStatModel struct {
		conn  sqlx.SqlConn
		table string
	}

	ShThreadBucketStat struct {
		Id                int64  `db:"id"`
		AggScope          string `db:"agg_scope"`
		ScopeKey          string `db:"scope_key"`
		Level             string `db:"level"`
		Year              int64  `db:"year"`
		Quarter           int64  `db:"quarter"`
		Month             int64  `db:"month"`
		Week              int64  `db:"week"`
		Day               int64  `db:"day"`
		BucketStart       int64  `db:"bucket_start"`
		BucketEnd         int64  `db:"bucket_end"`
		CountAll          int64  `db:"count_all"`
		CountPending      int64  `db:"count_pending"`
		CountFetched      int64  `db:"count_fetched"`
		CountFailed       int64  `db:"count_failed"`
		CountWithMagnet   int64  `db:"count_with_magnet"`
		CountWithInfoHash int64  `db:"count_with_info_hash"`
		LatestPostTime    int64  `db:"latest_post_time"`
		CreatedOn         int64  `db:"created_on"`
		UpdatedOn         int64  `db:"updated_on"`
	}
)

const shThreadBucketStatRows = "`id`,`agg_scope`,`scope_key`,`level`,`year`,`quarter`,`month`,`week`,`day`,`bucket_start`,`bucket_end`,`count_all`,`count_pending`,`count_fetched`,`count_failed`,`count_with_magnet`,`count_with_info_hash`,`latest_post_time`,`created_on`,`updated_on`"

func NewShThreadBucketStatModel(conn sqlx.SqlConn) ShThreadBucketStatModel {
	return &defaultShThreadBucketStatModel{conn: conn, table: "`sh_bucket_stat`"}
}

func (m *defaultShThreadBucketStatModel) TableName() string {
	return m.table
}

func (m *defaultShThreadBucketStatModel) Upsert(ctx context.Context, data *ShThreadBucketStat) error {
	query := "insert into " + m.table + " (`agg_scope`,`scope_key`,`level`,`year`,`quarter`,`month`,`week`,`day`,`bucket_start`,`bucket_end`,`count_all`,`count_pending`,`count_fetched`,`count_failed`,`count_with_magnet`,`count_with_info_hash`,`latest_post_time`,`created_on`,`updated_on`) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) on duplicate key update `level` = values(`level`), `year` = values(`year`), `quarter` = values(`quarter`), `month` = values(`month`), `week` = values(`week`), `day` = values(`day`), `bucket_start` = values(`bucket_start`), `bucket_end` = values(`bucket_end`), `count_all` = values(`count_all`), `count_pending` = values(`count_pending`), `count_fetched` = values(`count_fetched`), `count_failed` = values(`count_failed`), `count_with_magnet` = values(`count_with_magnet`), `count_with_info_hash` = values(`count_with_info_hash`), `latest_post_time` = values(`latest_post_time`), `updated_on` = values(`updated_on`)"
	_, err := m.conn.ExecCtx(ctx, query, data.AggScope, data.ScopeKey, data.Level, data.Year, data.Quarter, data.Month, data.Week, data.Day, data.BucketStart, data.BucketEnd, data.CountAll, data.CountPending, data.CountFetched, data.CountFailed, data.CountWithMagnet, data.CountWithInfoHash, data.LatestPostTime, data.CreatedOn, data.UpdatedOn)
	return err
}

func (m *defaultShThreadBucketStatModel) DeleteByAggScopeScopeKey(ctx context.Context, aggScope, scopeKey string) error {
	_, err := m.conn.ExecCtx(ctx, "delete from "+m.table+" where `agg_scope` = ? and `scope_key` = ?", aggScope, scopeKey)
	return err
}

func (m *defaultShThreadBucketStatModel) DeleteAll(ctx context.Context) (int64, error) {
	result, err := m.conn.ExecCtx(ctx, "delete from "+m.table)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return count, nil
}

func (m *defaultShThreadBucketStatModel) CountByLevel(ctx context.Context, level string) (int64, error) {
	var count int64
	if err := m.conn.QueryRowCtx(ctx, &count, "select count(*) from "+m.table+" where `level` = ?", level); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

func (m *defaultShThreadBucketStatModel) ListPage(ctx context.Context, filter ShThreadBucketStatListFilter) ([]*ShThreadBucketStat, int64, error) {
	builder := squirrel.Select(shThreadBucketStatRows).From(m.table)
	countBuilder := squirrel.Select("count(*)").From(m.table)
	where := squirrel.Eq{
		"agg_scope": strings.TrimSpace(filter.AggScope),
		"level":     strings.TrimSpace(filter.Level),
	}
	builder = builder.Where(where)
	countBuilder = countBuilder.Where(where)
	if filter.BucketEndFrom != nil {
		builder = builder.Where(squirrel.GtOrEq{"bucket_end": *filter.BucketEndFrom})
		countBuilder = countBuilder.Where(squirrel.GtOrEq{"bucket_end": *filter.BucketEndFrom})
	}
	if filter.BucketStartTo != nil {
		builder = builder.Where(squirrel.LtOrEq{"bucket_start": *filter.BucketStartTo})
		countBuilder = countBuilder.Where(squirrel.LtOrEq{"bucket_start": *filter.BucketStartTo})
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, countArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShThreadBucketStat{}, 0, nil
		}
		return nil, 0, err
	}

	dir := "DESC"
	if strings.EqualFold(strings.TrimSpace(filter.Dir), "asc") {
		dir = "ASC"
	}
	orderBy := []string{"`scope_key` " + dir, "`id` " + dir}
	switch strings.TrimSpace(filter.Sort) {
	case "updated":
		orderBy = []string{"`updated_on` " + dir, "`id` " + dir}
	case "all":
		orderBy = []string{"`count_all` " + dir, "`scope_key` DESC"}
	case "pending":
		orderBy = []string{"`count_pending` " + dir, "`scope_key` DESC"}
	case "fetched":
		orderBy = []string{"`count_fetched` " + dir, "`scope_key` DESC"}
	case "failed":
		orderBy = []string{"`count_failed` " + dir, "`scope_key` DESC"}
	case "magnet":
		orderBy = []string{"`count_with_magnet` " + dir, "`scope_key` DESC"}
	case "info_hash":
		orderBy = []string{"`count_with_info_hash` " + dir, "`scope_key` DESC"}
	case "latest_post":
		orderBy = []string{"`latest_post_time` " + dir, "`scope_key` DESC"}
	}
	if filter.PageSize > 0 {
		page := filter.Page
		if page <= 0 {
			page = 1
		}
		builder = builder.Limit(uint64(filter.PageSize)).Offset(uint64((page - 1) * filter.PageSize))
	}
	query, args, err := builder.OrderBy(orderBy...).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var rows []*ShThreadBucketStat
	if err := m.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ShThreadBucketStat{}, total, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}
