package modelx

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RawBestinvModel = (*customRawBestinvModel)(nil)
var _ orm.DataModel = (RawBestinvModel)(nil) // RawBestinvModel 实现了 orm.DataModel
var _ orm.DataStruct = (*RawBestinv)(nil)    // RawBestinv 实现了 orm.DataStruct

type (
	// RawBestinvModel 是一个接口，用于自定义并添加更多方法
	RawBestinvModel interface {
		rawBestinvModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		NamesNeedScan(ctx context.Context) ([]int64, error)
		NeedRankCheck(ctx context.Context) ([]int64, error)
		LatestDayNumber(ctx context.Context) (int64, error)
	}

	// customRawBestinvModel 包含默认模型的实现
	customRawBestinvModel struct {
		*defaultRawBestinvModel
	}
)

// NewRawBestinvModel 返回数据库表的模型
func NewRawBestinvModel(conn sqlx.SqlConn) RawBestinvModel {
	return &customRawBestinvModel{
		defaultRawBestinvModel: newRawBestinvModel(conn),
	}
}

const (
	BestCategoryMonth = 1 + iota
	BestCategoryAllTime
)

const (
	NeedRankCheck = 1 + iota
	NoNeedRankCheck
)

func (f *RawBestinv) Description() interface{} {
	return f.Name
}

// FindDataByDescription 根据描述查找数据
func (m *defaultRawBestinvModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name) // 假设有 FindOneByJavId 方法
}

// ExistByDescription 检查描述所指数据是否存在
func (m *defaultRawBestinvModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

// InsertData 插入新数据
func (m *defaultRawBestinvModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*RawBestinv); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData) // 假设有 Insert 方法
		return err
	}
	return ErrInvalidData // 假设定义了此错误
}

// UpdateDataByDescription 根据描述更新数据
func (m *defaultRawBestinvModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*RawBestinv); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*RawBestinv); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData) // 假设有 Update 方法
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *defaultRawBestinvModel) NamesNeedScan(ctx context.Context) ([]int64, error) {
	query, values, err := squirrel.Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", InventoryNeedScan).
		Limit(100000).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, values...); err != nil {
		return nil, err
	}

	return ids, nil
}

func (m *defaultRawBestinvModel) NeedRankCheck(ctx context.Context) ([]int64, error) {
	query, values, err := squirrel.Select("`id`").
		From(m.tableName()).
		Where("`need_rank_check` = ?", NeedRankCheck).
		Limit(100000).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, values...); err != nil {
		return nil, err
	}

	return ids, nil
}

func (m *defaultRawBestinvModel) LatestDayNumber(ctx context.Context) (int64, error) {
	query, values, err := squirrel.Select("`day_number`").
		From(m.tableName()).
		OrderBy("day_number desc").
		Limit(1).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result int64
	if err = m.conn.QueryRowCtx(ctx, &result, query, values...); err != nil {
		return 0, err
	}

	return result, nil
}
