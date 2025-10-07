package modelx

import (
	"context"
	"fmt"
	"time"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ BmPrefixModel = (*customBmPrefixModel)(nil)

type (
	// BmPrefixModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmPrefixModel.
	BmPrefixModel interface {
		bmPrefixModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
	}

	customBmPrefixModel struct {
		*defaultBmPrefixModel
	}
)

// NewBmPrefixModel returns a model for the database table.
func NewBmPrefixModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmPrefixModel {
	return &customBmPrefixModel{
		defaultBmPrefixModel: newBmPrefixModel(conn, c, opts...),
	}
}

func (p *BmPrefix) Description() interface{} {
	return p.Name
}

func (m *defaultBmPrefixModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultBmPrefixModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmPrefixModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmPrefix); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmPrefixModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmPrefix); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmPrefix); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}
