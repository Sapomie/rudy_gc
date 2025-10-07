package modelx

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CCafoModel = (*customCCafoModel)(nil)

type (
	// CCafoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCCafoModel.
	CCafoModel interface {
		cCafoModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindAll(ctx context.Context) ([]*CCafo, error)
	}

	customCCafoModel struct {
		*defaultCCafoModel
	}
)

// NewCCafoModel returns a model for the database table.
func NewCCafoModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CCafoModel {
	return &customCCafoModel{
		defaultCCafoModel: newCCafoModel(conn, c, opts...),
	}
}

func (c *CCafo) Description() interface{} {
	return c.Name
}

func (m *defaultCCafoModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultCCafoModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultCCafoModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*CCafo); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultCCafoModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*CCafo); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*CCafo); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultCCafoModel) FindAll(ctx context.Context) ([]*CCafo, error) {
	// 使用 squirrel 构建 SQL 查询
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var cafos []*CCafo
	// 执行查询
	if err := m.QueryRowsNoCacheCtx(ctx, &cafos, query, args...); err != nil {
		return nil, err
	}

	return cafos, nil
}
