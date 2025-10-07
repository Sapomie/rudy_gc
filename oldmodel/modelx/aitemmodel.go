package modelx

import (
	"context"
	"errors"
	"rudy_gc/pkg/orm"
	"time"

	"github.com/Masterminds/squirrel"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AItemModel = (*customAItemModel)(nil)
var _ orm.DataModel = (AItemModel)(nil)
var _ orm.DataStruct = (*AItem)(nil)

type (
	// AItemModel 是一个接口，用于自定义并添加更多方法
	AItemModel interface {
		aItemModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindItemsByHasDetailStatus(ctx context.Context, hasDetail int64) ([]*AItem, error)
	}

	customAItemModel struct {
		*defaultAItemModel
	}
)

// NewAItemModel 返回数据库表的模型。
func NewAItemModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AItemModel {
	return &customAItemModel{
		defaultAItemModel: newAItemModel(conn, c, opts...),
	}
}

const (
	ItemSearchByPrefix = 1 + iota
	ItemSearchByLabel
	ItemSearchByBestMonth
	ItemSearchByBestAllTime
)

const (
	ItemStatusHasNoDetail = 1 + iota
	ItemStatusHasDetail
)

const (
	ItemHasNoBus = 1 + iota
	ItemHasBus
	ItemBusNotExist
)

// Description 返回 AItem 的描述。
func (f *AItem) Description() interface{} {
	return f.JavId
}

func (m *defaultAItemModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	javId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound // 如果类型不匹配，返回 ErrNotFound
	}
	return m.FindOneByJavId(ctx, javId) // 查找项
}

func (m *defaultAItemModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil // 存在
	} else if err == sqlc.ErrNotFound {
		return false, nil // 不存在
	}
	return false, err // 发生错误
}

func (m *defaultAItemModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*AItem); ok {
		insertData.CreatedOn = time.Now().Unix() // 设置创建时间
		_, err := m.Insert(ctx, insertData)      // 插入数据
		return err
	}
	return ErrInvalidData // 返回无效数据错误
}

func (m *defaultAItemModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*AItem); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err // 查找时发生错误
		}
		if existingData, ok := dataDb.(*AItem); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix() // 设置更新时间
			return m.Update(ctx, newData)         // 更新数据
		}
		return ErrInvalidData // 返回无效数据错误
	}
	return ErrInvalidData // 返回无效数据错误
}

func (m *defaultAItemModel) FindItemsByHasDetailStatus(ctx context.Context, hasDetail int64) ([]*AItem, error) {
	// 使用 aItemRows 来优化查询时的字段选择
	query, args, err := squirrel.Select(aItemRows).
		From(m.tableName()).
		Where(squirrel.Eq{"has_detail": hasDetail}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var items []*AItem
	err = m.QueryRowsNoCacheCtx(ctx, &items, query, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return items, nil
}
