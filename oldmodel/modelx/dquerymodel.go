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

var _ DQueryModel = (*customDQueryModel)(nil)
var _ orm.DataModel = (DQueryModel)(nil) // DQueryModel 实现了 orm.DataModel
var _ orm.DataStruct = (*DQuery)(nil)    // DQuery 实现了 orm.DataStruct

type (
	// DQueryModel is an interface to be customized, add more methods here.
	DQueryModel interface {
		dQueryModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
		FindQueriesActive(ctx context.Context, nameType int64) ([]*DQuery, error)
		FindAll(ctx context.Context) ([]*DQuery, error)
	}

	customDQueryModel struct {
		*defaultDQueryModel
	}
)

const (
	QueryInactive = 1 + iota
	QueryActive
)

const (
	QueryNamePrefix = 1 + iota
	QueryNameLabel
)

const (
	QueryByOffset = 1 + iota
	QueryByStartEnd
)

// NewDQueryModel returns a model for the database table.
func NewDQueryModel(conn sqlx.SqlConn) DQueryModel {
	return &customDQueryModel{
		defaultDQueryModel: newDQueryModel(conn),
	}
}

// Description returns the description of DQuery.
func (d *DQuery) Description() interface{} {
	return d.Name
}

func (m *defaultDQueryModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound // 如果类型不匹配，返回 ErrNotFound
	}
	return m.FindOneByName(ctx, name) // 假设实现了根据ID查找的方法
}

func (m *defaultDQueryModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil // 存在
	} else if err == sqlc.ErrNotFound {
		return false, nil // 不存在
	}
	return false, err // 发生错误
}

func (m *defaultDQueryModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*DQuery); ok {
		insertData.CreatedOn = time.Now().Unix() // 设置创建时间
		_, err := m.Insert(ctx, insertData)      // 插入数据
		return err
	}
	return ErrInvalidData // 返回无效数据错误
}

func (m *defaultDQueryModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*DQuery); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err // 查找时发生错误
		}
		if existingData, ok := dataDb.(*DQuery); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix() // 设置更新时间
			return m.Update(ctx, newData)         // 更新数据
		}
		return ErrInvalidData // 返回无效数据错误
	}
	return ErrInvalidData // 返回无效数据错误
}

func (m *customDQueryModel) FindQueriesActive(ctx context.Context, nameType int64) ([]*DQuery, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`active` = ?", QueryActive).
		Where("`name_type` = ?", nameType).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var queries []*DQuery
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &queries, query, args...); err != nil {
		return nil, err
	}

	return queries, nil
}

func (m *customDQueryModel) FindAll(ctx context.Context) ([]*DQuery, error) {
	query, args, err := squirrel.
		Select("*").
		From(m.tableName()).
		OrderBy("id DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var list []*DQuery
	if err := m.conn.QueryRowsCtx(ctx, &list, query, args...); err != nil {
		return nil, err
	}
	return list, nil
}
