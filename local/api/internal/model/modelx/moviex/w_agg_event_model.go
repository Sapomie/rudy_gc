package moviex

import (
	"context"
	"errors"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ WAggEventModel = (*customWAggEventModel)(nil)

type (
	WAggEventListFilter struct {
		AggKey   string
		FlowKey  string
		Status   string
		Sort     string
		Order    string
		Page     int64
		PageSize int64
	}

	WAggEventModel interface {
		wAggEventModel
		ListPage(ctx context.Context, filter WAggEventListFilter) ([]*WAggEvent, int64, error)
		TableName() string
	}

	customWAggEventModel struct {
		*defaultWAggEventModel
	}
)

func NewWAggEventModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) WAggEventModel {
	return &customWAggEventModel{
		defaultWAggEventModel: newWAggEventModel(conn, c, opts...),
	}
}

func (m *customWAggEventModel) TableName() string {
	return m.table
}

func (m *customWAggEventModel) ListPage(ctx context.Context, filter WAggEventListFilter) ([]*WAggEvent, int64, error) {
	builder := squirrel.
		Select(wAggEventRows).
		From(buildWAggEventSelectTable(m.table, filter))
	countBuilder := squirrel.
		Select("COUNT(*)").
		From(m.table)

	if aggKey := strings.TrimSpace(filter.AggKey); aggKey != "" {
		builder = builder.Where(squirrel.Eq{"agg_key": aggKey})
		countBuilder = countBuilder.Where(squirrel.Eq{"agg_key": aggKey})
	}
	if flowKey := strings.TrimSpace(filter.FlowKey); flowKey != "" {
		builder = builder.Where(squirrel.Eq{"flow_key": flowKey})
		countBuilder = countBuilder.Where(squirrel.Eq{"flow_key": flowKey})
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		builder = builder.Where(squirrel.Eq{"status": status})
		countBuilder = countBuilder.Where(squirrel.Eq{"status": status})
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize
	builder = builder.Offset(uint64(offset)).Limit(uint64(pageSize))

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, countArgs...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WAggEvent{}, 0, nil
		}
		return nil, 0, err
	}

	query, args, err := builder.OrderBy(buildWAggEventOrderBy(filter.Sort, filter.Order)...).ToSql()
	if err != nil {
		return nil, 0, err
	}

	var rows []*WAggEvent
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*WAggEvent{}, total, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}

func buildWAggEventOrderBy(sortField string, sortOrder string) []string {
	dir := "DESC"
	idDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		dir = "ASC"
		idDir = "ASC"
	}

	switch strings.TrimSpace(sortField) {
	case "agg_key":
		return []string{"`agg_key` " + dir, "`id` " + idDir}
	case "flow_key":
		return []string{"`flow_key` " + dir, "`id` " + idDir}
	case "scope_count":
		return []string{"`scope_count` " + dir, "`id` " + idDir}
	case "bucket_count":
		return []string{"`bucket_count` " + dir, "`id` " + idDir}
	case "top_count":
		return []string{"`top_count` " + dir, "`id` " + idDir}
	case "finished_time":
		return []string{"`finished_time` " + dir, "`id` " + idDir}
	case "duration_ms":
		return []string{"`duration_ms` " + dir, "`id` " + idDir}
	default:
		return []string{"`started_time` " + dir, "`id` " + idDir}
	}
}

func buildWAggEventSelectTable(table string, filter WAggEventListFilter) string {
	if strings.TrimSpace(filter.AggKey) != "" || strings.TrimSpace(filter.FlowKey) != "" || strings.TrimSpace(filter.Status) != "" {
		return table
	}

	switch strings.TrimSpace(filter.Sort) {
	case "agg_key":
		return table + " FORCE INDEX (`idx_w_agg_event_agg_key`)"
	case "flow_key":
		return table + " FORCE INDEX (`idx_w_agg_event_flow_key`)"
	default:
		return table + " FORCE INDEX (`idx_w_agg_event_started_time`)"
	}
}
