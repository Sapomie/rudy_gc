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

var _ BmDirectorModel = (*customBmDirectorModel)(nil)

type (
	// BmDirectorModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmDirectorModel.
	BmDirectorModel interface {
		bmDirectorModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
	}

	customBmDirectorModel struct {
		*defaultBmDirectorModel
	}
)

// NewBmDirectorModel returns a model for the database table.
func NewBmDirectorModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmDirectorModel {
	return &customBmDirectorModel{
		defaultBmDirectorModel: newBmDirectorModel(conn, c, opts...),
	}
}

func (d *BmDirector) Description() interface{} {
	return d.Name
}
func (d *BmDirector) GetID() int64 {
	return d.Id
}

func (m *defaultBmDirectorModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultBmDirectorModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmDirectorModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmDirector); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmDirectorModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmDirector); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmDirector); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}
