package modelx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"

	"rudy_gc/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ VVideoModel = (*customVVideoModel)(nil)
var _ orm.DataModel = (VVideoModel)(nil)
var _ orm.DataStruct = (*VVideo)(nil)

type (
	// VVideoModel 是一个接口，用于自定义并增加更多方法的接口，
	// 并在 customVVideoModel 中实现新增的方法。
	VVideoModel interface {
		vVideoModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
		FindAll(ctx context.Context, existStatus int) ([]*VVideo, error)
	}

	// customVVideoModel 提供基础模型实现及其扩展
	customVVideoModel struct {
		*defaultVVideoModel
	}
)

// NewVVideoModel 创建基础模型初始化，返回表的模型
func NewVVideoModel(conn sqlx.SqlConn) VVideoModel {
	return &customVVideoModel{
		defaultVVideoModel: newVVideoModel(conn),
	}
}

func (f *VVideo) Description() interface{} {
	return f.BaseName
}

func (m *customVVideoModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	videoId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByBaseName(ctx, videoId) // 假设存在这个方法
}

func (m *customVVideoModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sqlc.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (m *customVVideoModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*VVideo); ok { // 假设 VVideo 是正确的结构体
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData) // 假设存在 Insert 方法
		return err
	}
	return ErrInvalidData
}

func (m *customVVideoModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*VVideo); ok { // 假设 VVideo 是正确的结构体
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*VVideo); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData) // 假设存在 Update 方法
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *customVVideoModel) FindAll(ctx context.Context, existStatus int) ([]*VVideo, error) {
	// 使用 squirrel 构建 SQL 查询
	s := squirrel.Select("*").From(m.tableName())
	if existStatus != 0 {
		s = s.Where("is_removed = ?", existStatus)
	}

	query, args, err := s.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result []*VVideo
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}

	return result, nil
}
