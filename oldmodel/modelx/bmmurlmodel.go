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

var _ BmMurlModel = (*customBmMurlModel)(nil)

type (
	// BmMurlModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmMurlModel.
	BmMurlModel interface {
		bmMurlModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
	}

	customBmMurlModel struct {
		*defaultBmMurlModel
	}
)

// NewBmMurlModel returns a model for the database table.
func NewBmMurlModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmMurlModel {
	return &customBmMurlModel{
		defaultBmMurlModel: newBmMurlModel(conn, c, opts...),
	}
}

func (m *BmMurl) Description() interface{} {
	return m.JavId
}

func (m *defaultBmMurlModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	javId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByJavId(ctx, javId)
}

func (m *defaultBmMurlModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmMurlModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmMurl); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmMurlModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmMurl); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmMurl); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}
