// data/modelx/spiderx/d_bestinv_model.go
package spiderx

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DBestinvModel = (*customDBestinvModel)(nil)

type (
	// DBestinvModel 是可扩展接口，新增方法写这里，并在 customDBestinvModel 中实现
	DBestinvModel interface {
		dBestinvModel
		withSession(session sqlx.Session) DBestinvModel
		// ListNeedScanIDs 查询 need_scan=1 的若干 id（按 id 升序，limit 上限可控）
		ListNeedScanIDs(ctx context.Context, limit int64) ([]int64, error)
	}

	customDBestinvModel struct {
		*defaultDBestinvModel
	}
)

// NewDBestinvModel 返回定制后的 model
func NewDBestinvModel(conn sqlx.SqlConn) DBestinvModel {
	return &customDBestinvModel{
		defaultDBestinvModel: newDBestinvModel(conn),
	}
}

func (m *customDBestinvModel) withSession(session sqlx.Session) DBestinvModel {
	return NewDBestinvModel(sqlx.NewSqlConnFromSession(session))
}

const bestinvNeedScan int64 = 1

func (m *customDBestinvModel) ListNeedScanIDs(ctx context.Context, limit int64) ([]int64, error) {
	if limit <= 0 {
		limit = 100000
	}

	query, args, err := squirrel.
		Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", bestinvNeedScan).
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
