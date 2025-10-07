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

var _ DRecordModel = (*customDRecordModel)(nil)
var _ orm.DataModel = (DRecordModel)(nil) // DRecordModel 实现了 orm.DataModel
var _ orm.DataStruct = (*DRecord)(nil)    // DRecord 实现了 orm.DataStruct

type (
	// DRecordModel 是一个接口，用于自定义并添加更多方法
	DRecordModel interface {
		dRecordModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindAll(ctx context.Context, limit uint64) ([]*DRecord, error)
	}

	customDRecordModel struct {
		*defaultDRecordModel
	}
)

// NewDRecordModel 返回数据库表的模型
func NewDRecordModel(conn sqlx.SqlConn) DRecordModel {
	return &customDRecordModel{
		defaultDRecordModel: newDRecordModel(conn),
	}
}

func (d *DRecord) Description() interface{} {
	return d.Name
}

func (m *defaultDRecordModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultDRecordModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultDRecordModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*DRecord); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return ErrInvalidData
}

func (m *defaultDRecordModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*DRecord); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*DRecord); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *defaultDRecordModel) FindAll(ctx context.Context, limit uint64) ([]*DRecord, error) {
	query, values, err := squirrel.Select("*").
		From(m.tableName()).OrderBy("start_time desc").
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result []*DRecord
	if err = m.conn.QueryRowsCtx(ctx, &result, query, values...); err != nil {
		return nil, err
	}

	return result, nil
}
