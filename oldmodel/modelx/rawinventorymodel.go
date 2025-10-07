package modelx

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RawInventoryModel = (*customRawInventoryModel)(nil)
var _ orm.DataModel = (RawInventoryModel)(nil)
var _ orm.DataStruct = (*RawInventory)(nil)

type (
	RawInventoryModel interface {
		rawInventoryModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		ListNeedScanIDs(ctx context.Context) ([]int64, error)
	}

	customRawInventoryModel struct {
		*defaultRawInventoryModel
	}
)

// NewRawInventoryModel 返回数据库表的模型.
func NewRawInventoryModel(conn sqlx.SqlConn) RawInventoryModel {
	return &customRawInventoryModel{
		defaultRawInventoryModel: newRawInventoryModel(conn),
	}
}

const (
	InventoryNeedScan = 1 + iota
	InventoryNoNeedScan
)

const (
	InventoryCategoryByPrefix = 1 + iota
	InventoryCategoryByLabel
)

func (f *RawInventory) Description() interface{} {
	return f.Name
}

// FindDataByDescription 根据描述查找数据
func (m *defaultRawInventoryModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	inventoryKey, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, inventoryKey)
}

// ExistByDescription 检查数据是否存在
func (m *defaultRawInventoryModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

// InsertData 插入新数据
func (m *defaultRawInventoryModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*RawInventory); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return ErrInvalidData
}

// UpdateDataByDescription 根据描述更新数据
func (m *defaultRawInventoryModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*RawInventory); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*RawInventory); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *defaultRawInventoryModel) ListNeedScanIDs(ctx context.Context) ([]int64, error) {
	query, values, err := squirrel.Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", InventoryNeedScan).
		Limit(100000).
		ToSql()

	if err != nil {
		return nil, err
	}

	var ids []int64
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, values...); err != nil {
		return nil, err
	}

	return ids, nil
}
