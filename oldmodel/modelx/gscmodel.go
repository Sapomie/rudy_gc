package modelx

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GScModel = (*customGScModel)(nil)
var _ orm.DataModel = (GScModel)(nil)
var _ orm.DataStruct = (*GSc)(nil)

type (
	// GScModel 是一个接口，用于自定义并增加更多方法的接口，
	// 并在 customGScModel 中实现新增的方法。
	GScModel interface {
		gScModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
		FindAll(ctx context.Context) ([]*GSc, error)
		FindTopNRecentSc(ctx context.Context, n uint64) ([]*GSc, error)
	}

	// customGScModel 提供基础模型实现及其扩展
	customGScModel struct {
		*defaultGScModel
	}
)

// NewGScModel 创建基础模型初始化，返回表的模型
func NewGScModel(conn sqlx.SqlConn) GScModel {
	return &customGScModel{
		defaultGScModel: newGScModel(conn),
	}
}

func (f *GSc) Description() interface{} {
	return f.Name
}

func (m *customGScModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	javId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, javId) // 假设存在这个方法
}

func (m *customGScModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sqlc.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (m *customGScModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*GSc); ok { // 假设 GSc 是正确的结构体
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData) // 假设存在 Insert 方法
		return err
	}
	return ErrInvalidData
}

func (m *customGScModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*GSc); ok { // 假设 GSc 是正确的结构体
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*GSc); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData) // 假设存在 Update 方法
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *customGScModel) FindAll(ctx context.Context) ([]*GSc, error) {
	query, values, err := squirrel.Select("*").
		From(m.tableName()).
		//Where("`film_number` > 0").
		Limit(100000).
		ToSql()

	if err != nil {
		return nil, err
	}

	var result []*GSc
	if err := m.conn.QueryRowsCtx(ctx, &result, query, values...); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *customGScModel) FindTopNRecentSc(ctx context.Context, n uint64) ([]*GSc, error) {
	query, values, err := squirrel.Select("*").
		From(m.tableName()).
		OrderBy("name desc").
		Limit(n).
		ToSql()

	if err != nil {
		return nil, err
	}

	var result []*GSc
	if err := m.conn.QueryRowsCtx(ctx, &result, query, values...); err != nil {
		return nil, err
	}

	return result, nil
}
