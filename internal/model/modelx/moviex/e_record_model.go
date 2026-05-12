package moviex

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ERecordModel = (*customERecordModel)(nil)

type (
	ERecordListFilter struct {
		Type     string
		Sort     string
		Order    string
		Page     int64
		PageSize int64
	}

	// ERecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customERecordModel.
	ERecordModel interface {
		eRecordModel
		FindByStartTimeAndType(ctx context.Context, startFrom int64, typ string, limit int) ([]*ERecord, error)
		ListPage(ctx context.Context, filter ERecordListFilter) ([]*ERecord, int64, error)
	}

	customERecordModel struct {
		*defaultERecordModel
	}
)

// NewERecordModel returns a model for the database table.
func NewERecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ERecordModel {
	return &customERecordModel{
		defaultERecordModel: newERecordModel(conn, c, opts...),
	}
}

func (m *customERecordModel) FindByStartTimeAndType(ctx context.Context, startFrom int64, typ string, limit int) ([]*ERecord, error) {
	b := squirrel.
		Select(eRecordRows).
		From(m.table).
		Where("`start_time` >= ?", startFrom).
		OrderBy("`start_time` DESC").
		PlaceholderFormat(squirrel.Question)

	if typ != "" {
		b = b.Where("`type` = ?", typ)
	}
	if limit > 0 {
		b = b.Limit(uint64(limit))
	}

	query, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql failed: %w", err)
	}

	// 不走缓存，直接查
	var list []*ERecord
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*ERecord{}, nil
		}
		return nil, err
	}
	return list, nil
}

func (m *customERecordModel) ListPage(ctx context.Context, filter ERecordListFilter) ([]*ERecord, int64, error) {
	builder := squirrel.
		Select(eRecordRows).
		From(m.table)
	countBuilder := squirrel.
		Select("COUNT(*)").
		From(m.table)

	if recordType := strings.TrimSpace(filter.Type); recordType != "" {
		builder = builder.Where(squirrel.Eq{"type": recordType})
		countBuilder = countBuilder.Where(squirrel.Eq{"type": recordType})
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
			return []*ERecord{}, 0, nil
		}
		return nil, 0, err
	}

	query, args, err := builder.OrderBy(buildERecordOrderBy(filter.Sort, filter.Order)...).ToSql()
	if err != nil {
		return nil, 0, err
	}

	var rows []*ERecord
	if err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*ERecord{}, total, nil
		}
		return nil, 0, err
	}
	return rows, total, nil
}

func buildERecordOrderBy(sortField string, sortOrder string) []string {
	dir := "DESC"
	idDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		dir = "ASC"
		idDir = "ASC"
	}

	switch strings.TrimSpace(sortField) {
	case "type":
		return []string{"`type` " + dir, "`id` " + idDir}
	case "detail_number":
		return []string{"`detail_number` " + dir, "`id` " + idDir}
	case "duration":
		return []string{"GREATEST(`end_time` - `start_time`, 0) " + dir, "`id` " + idDir}
	case "name":
		return []string{"`name` " + dir, "`id` " + idDir}
	default:
		return []string{"`start_time` " + dir, "`id` " + idDir}
	}
}
