package moviex

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (m *customAMovieModel) ListPageWithTotal(ctx context.Context, offset, limit int64, orderKey string) ([]*AMovie, int64, error) {
	// 1) 先查总数（后续如加筛选条件，把 WHERE 同步到 COUNT 和 SELECT）
	var total int64
	{
		qc, args, err := squirrel.Select("COUNT(*)").From(m.table).ToSql()
		if err != nil {
			return nil, 0, err
		}
		if err := m.QueryRowNoCacheCtx(ctx, &total, qc, args...); err != nil {
			return nil, 0, err
		}
		if total == 0 {
			return []*AMovie{}, 0, nil
		}
	}

	// 2) 列映射（把上层 orderKey 转换为确定的 ORDER BY）
	orderMap := map[string]string{
		"rd":  "`releasing_date` DESC,`name` DESC",
		"du":  "`detail_update_time` DESC,`name` DESC",
		"cad": "`cast_average_age` DESC,`name` DESC",
		"caa": "`cast_average_age` ASC,`name` DESC",
		"vw":  "`viewers_number_watched` DESC,`name` DESC",
	}
	orderBy := orderMap[orderKey]
	if orderBy == "" {
		orderBy = "`releasing_date` DESC,`name` DESC"
	}

	// 3) 分页查询
	q, args, err := squirrel.
		Select(aMovieRows).
		From(m.table).
		OrderBy(orderBy).
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	var list []*AMovie
	if err := m.QueryRowsNoCacheCtx(ctx, &list, q, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*AMovie{}, total, nil
		}
		return nil, 0, err
	}
	return list, total, nil
}
