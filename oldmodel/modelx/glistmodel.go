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

var _ GListModel = (*customGListModel)(nil)

type (
	// GListModel 是一个接口，用于自定义并增加更多方法的接口，
	// 并在 customGListModel 中实现新增的方法。
	GListModel interface {
		gListModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error
		FindAll(ctx context.Context) ([]*GList, error)

		FindByMovieJavId(ctx context.Context, javId string) ([]*GList, error)
		FindGList(ctx context.Context, scName string, isCome int64, page, pageSize int) ([]*GList, int64, error)
	}

	// customGListModel 提供基础模型实现及其扩展
	customGListModel struct {
		*defaultGListModel
	}
)

// NewGListModel 创建基础模型初始化，返回表的模型。
func NewGListModel(conn sqlx.SqlConn) GListModel {
	return &customGListModel{
		defaultGListModel: newGListModel(conn),
	}
}

func (f *GList) Description() interface{} {
	return f.Name
}

const (
	GListIsCome    int64 = 2
	GListIsNotCome int64 = 1
)

// FindDataByDescription 根据描述查找数据。
func (m *customGListModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	id, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, id) // 假设存在这个方法
}

// ExistByDescription 检查描述是否存在。
func (m *customGListModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sqlc.ErrNotFound) {
		return false, nil
	}
	return false, err
}

// InsertData 插入数据。
func (m *customGListModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*GList); ok { // 假设 GList 是正确的结构体
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData) // 假设存在 Insert 方法
		return err
	}
	return ErrInvalidData
}

// UpdateDataByDescription 更新数据。
func (m *customGListModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*GList); ok { // 假设 GList 是正确的结构体
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*GList); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData) // 假设存在 Update 方法
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}

func (m *customGListModel) FindAll(ctx context.Context) ([]*GList, error) {
	// 使用 squirrel 构建 SQL 查询
	resultQuery, resultArgs, err := squirrel.Select("*").
		From(m.tableName()).
		//Where("`owned_movie_number` > 0").
		ToSql()
	if err != nil {
		return nil, err
	}

	var lists []*GList
	if err := m.conn.QueryRowsCtx(ctx, &lists, resultQuery, resultArgs...); err != nil {
		return nil, err
	}

	return lists, nil
}

func (m *customGListModel) FindGList(ctx context.Context, scName string, isCome int64, page, pageSize int) ([]*GList, int64, error) {
	// 使用 squirrel 构建 SQL 查询
	db := squirrel.Select("*").From(m.tableName())
	if scName != "" {
		db = db.Where("sc_name = ?", scName)
	}
	if isCome != 0 {
		db = db.Where("is_come = ?", isCome)
	}
	if page == 0 {
		page = 1
	}

	var total int64
	countQuery, countArgs, err := db.ToSql()
	if err != nil {
		return nil, 0, err
	}
	finalCountQuery := "SELECT COUNT(*) FROM (" + countQuery + ") AS count_query"
	if err := m.conn.QueryRowCtx(ctx, &total, finalCountQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	query, args, err := db.Limit(uint64(pageSize)).Offset(uint64((page - 1) * pageSize)).OrderBy("id desc").ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build SQL query: %w", err)
	}

	var result []*GList
	// 执行查询
	if err := m.conn.QueryRowsCtx(ctx, &result, query, args...); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (m *customGListModel) FindByMovieJavId(ctx context.Context, javId string) ([]*GList, error) {
	query, values, err := squirrel.Select("*").
		From(m.tableName()).
		Where("movie_jav_id = ?", javId).
		OrderBy("`name` desc").
		ToSql()
	if err != nil {
		return nil, err
	}

	var result []*GList
	if err := m.conn.QueryRowsCtx(ctx, &result, query, values...); err != nil {
		return nil, err
	}

	return result, nil
}
