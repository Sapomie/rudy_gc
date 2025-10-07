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

var _ BmCastModel = (*customBmCastModel)(nil) // 确保 customBmCastModel 实现了 BmCastModel 接口

type (
	// BmCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmCastModel.
	BmCastModel interface {
		bmCastModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindCasts(ctx context.Context) ([]*BmCast, error)
		FindAll(ctx context.Context) ([]*BmCast, error)
	}

	customBmCastModel struct {
		*defaultBmCastModel
	}
)

// NewBmCastModel returns a model for the database table.
func NewBmCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmCastModel {
	return &customBmCastModel{
		defaultBmCastModel: newBmCastModel(conn, c, opts...),
	}
}

func (f *BmCast) Description() interface{} {
	return f.Name
}

func (m *defaultBmCastModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultBmCastModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmCastModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmCast); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmCastModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmCast); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmCast); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmCastModel) FindCasts(ctx context.Context) ([]*BmCast, error) {
	// 使用 squirrel 构建 SQL 查询
	resultQuery, resultArgs, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`owned_movie_number` > 0").ToSql()
	if err != nil {
		return nil, err
	}

	var casts []*BmCast
	if err := m.QueryRowsNoCacheCtx(ctx, &casts, resultQuery, resultArgs...); err != nil {
		return nil, err
	}

	return casts, nil
}

func (m *defaultBmCastModel) FindAll(ctx context.Context) ([]*BmCast, error) {
	// 使用 squirrel 构建 SQL 查询
	resultQuery, resultArgs, err := squirrel.Select("*").
		From(m.tableName()).
		ToSql()
	if err != nil {
		return nil, err
	}

	var casts []*BmCast
	if err := m.QueryRowsNoCacheCtx(ctx, &casts, resultQuery, resultArgs...); err != nil {
		return nil, err
	}

	return casts, nil
}
