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

var _ RawDetailModel = (*customRawDetailModel)(nil)
var _ orm.DataModel = (RawDetailModel)(nil)
var _ orm.DataStruct = (*RawDetail)(nil)

type (
	// RawDetailModel 是一个接口，用于自定义并增加更多方法的接口，
	// 并在 customRawDetailModel 中实现新增的方法。
	RawDetailModel interface {
		rawDetailModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		NamesByNeedScanStatus(ctx context.Context, status int64) ([]int64, error)

		AllIds(ctx context.Context) ([]int64, error)
	}

	// customRawDetailModel 提供基础模型实现及其扩展
	customRawDetailModel struct {
		*defaultRawDetailModel
	}
)

// NewRawDetailModel 创建基础模型初始化，返回表的模型
func NewRawDetailModel(conn sqlx.SqlConn) RawDetailModel {
	return &customRawDetailModel{
		defaultRawDetailModel: newRawDetailModel(conn),
	}
}

const (
	DetailStatusNeedScan = 1 + iota
	DetailStatusNoNeedScan
	DetailStatusWrongContent
)

func (f *RawDetail) Description() interface{} {
	return f.JavId
}

func (m *customRawDetailModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	javId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByJavId(ctx, javId)
}

func (m *customRawDetailModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *customRawDetailModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*RawDetail); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return ErrInvalidData
}

func (m *customRawDetailModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*RawDetail); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*RawDetail); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData) // 假设存在 Update 方法
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *customRawDetailModel) NamesByNeedScanStatus(ctx context.Context, status int64) ([]int64, error) {
	// 构建 SQL 查询语句
	query, args, err := squirrel.Select("`id`").
		From(m.tableName()).
		Where("`need_scan` = ?", status).
		Limit(1000000).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var ids []int64
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (m *customRawDetailModel) AllIds(ctx context.Context) ([]int64, error) {
	// 构建 SQL 查询语句
	query, args, err := squirrel.Select("`id`").
		From(m.tableName()).
		//Where("`need_scan` = ?", status).
		Limit(1000000).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var ids []int64
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		return nil, err
	}
	return ids, nil
}
