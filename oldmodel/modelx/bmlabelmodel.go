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

var _ BmLabelModel = (*customBmLabelModel)(nil)

type (
	// BmLabelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmLabelModel.
	BmLabelModel interface {
		bmLabelModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
	}

	customBmLabelModel struct {
		*defaultBmLabelModel
	}
)

// NewBmLabelModel 创建一个新的 BmLabelModel 实例
func NewBmLabelModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmLabelModel {
	return &customBmLabelModel{
		defaultBmLabelModel: newBmLabelModel(conn, c, opts...),
	}
}

func (l *BmLabel) Description() interface{} {
	return l.Name
}

func (m *defaultBmLabelModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

func (m *defaultBmLabelModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmLabelModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmLabel); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmLabelModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmLabel); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmLabel); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}
