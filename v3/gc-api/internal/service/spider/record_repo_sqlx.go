package spider

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type RecordRepoSqlx struct {
	m moviex.ERecordModel
}

func (r *RecordRepoSqlx) TryInsert(ctx context.Context, rec *types.Record) (bool, error) {
	if rec == nil {
		return false, nil
	}
	// 以 Name 为幂等键
	if rec.Name != "" {
		if _, err := r.m.FindOneByName(ctx, rec.Name); err == nil {
			return false, nil // 已存在
		} else if err != sqlc.ErrNotFound {
			return false, err // 真实错误
		}
	}

	now := time.Now().Unix()
	row := &moviex.ERecord{
		Name:         rec.Name,
		StartTime:    rec.StartTime,
		EndTime:      rec.EndTime,
		Type:         rec.Type,
		DetailNumber: rec.DetailNumber,
		CreatedOn:    now,
		UpdatedOn:    now,
	}

	res, err := r.m.Insert(ctx, row)
	if err != nil {
		return false, err
	}
	id, _ := res.LastInsertId()
	rec.Id = id
	rec.CreatedOn = now
	rec.UpdatedOn = now
	return true, nil
}

func (r *RecordRepoSqlx) Find(ctx context.Context, startTimeFrom int64, typ string, limit int) ([]*types.Record, error) {
	list, err := r.m.FindByStartTimeAndType(ctx, startTimeFrom, typ, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Record, 0, len(list))
	for _, x := range list {
		out = append(out, &types.Record{
			Id:           x.Id,
			Name:         x.Name,
			StartTime:    x.StartTime,
			EndTime:      x.EndTime,
			Type:         x.Type,
			DetailNumber: x.DetailNumber,
			CreatedOn:    x.CreatedOn,
			UpdatedOn:    x.UpdatedOn,
		})
	}
	return out, nil
}
