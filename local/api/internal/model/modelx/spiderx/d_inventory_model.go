package spiderx

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DInventoryModel = (*customDInventoryModel)(nil)

type (
	// DInventoryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDInventoryModel.
	DInventoryModel interface {
		dInventoryModel
		withSession(session sqlx.Session) DInventoryModel
		ListNeedScanIDs(ctx context.Context, flag int64, limit int64) ([]int64, error)
	}

	customDInventoryModel struct {
		*defaultDInventoryModel
	}
)

// NeedScan：1=需要扫描, 2=不需要扫描
const (
	InventoryNeedScan   = 1 + iota // 1
	InventoryNoNeedScan            // 2
)

// Category：1=Prefix, 2=Label
const (
	InventoryCategoryByPrefix = 1 + iota // 1
	InventoryCategoryByLabel             // 2
)

// NewDInventoryModel returns a model for the database table.
func NewDInventoryModel(conn sqlx.SqlConn) DInventoryModel {
	return &customDInventoryModel{
		defaultDInventoryModel: newDInventoryModel(conn),
	}
}

func (m *customDInventoryModel) withSession(session sqlx.Session) DInventoryModel {
	return NewDInventoryModel(sqlx.NewSqlConnFromSession(session))
}

// ListNeedScanIDs 查询 need_scan=flag 的若干 id
func (m *customDInventoryModel) ListNeedScanIDs(ctx context.Context, flag int64, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 100000
	}

	query, args, err := squirrel.
		Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", flag).
		OrderBy("`id` ASC").
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}
