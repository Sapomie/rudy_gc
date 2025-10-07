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

var _ BmGenreModel = (*customBmGenreModel)(nil)

type (
	// BmGenreModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmGenreModel.
	BmGenreModel interface {
		bmGenreModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
	}

	customBmGenreModel struct {
		*defaultBmGenreModel
	}
)

// NewBmGenreModel returns a model for the database table.
func NewBmGenreModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmGenreModel {
	return &customBmGenreModel{
		defaultBmGenreModel: newBmGenreModel(conn, c, opts...),
	}
}

func (g *BmGenre) Description() interface{} {
	return g.Name
}

func (m *defaultBmGenreModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultBmGenreModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmGenreModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmGenre); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmGenreModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmGenre); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmGenre); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}
