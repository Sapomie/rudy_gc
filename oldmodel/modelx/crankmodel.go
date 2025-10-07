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

var _ CRankModel = (*customCRankModel)(nil)
var _ orm.DataModel = (CRankModel)(nil) // CRankModel 实现了 orm.DataModel
var _ orm.DataStruct = (*CRank)(nil)    // CRank 实现了 orm.DataStruct

type (
	// CRankModel 是一个接口，用于自定义并添加更多方法
	CRankModel interface {
		cRankModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FirstRankTimeByJavId(ctx context.Context, javId string) (int64, error)
		FindAllDistinctMovieJavIds(ctx context.Context) ([]string, error)
		FindHighestRank(ctx context.Context, movieJavId string, limit uint64) ([]*CRank, error)

		DistinctJavId(ctx context.Context) ([]string, error)
		FindLatest(ctx context.Context) (*CRank, error)
		FindJavIdsByDayNumber(ctx context.Context, dayNumber int64) ([]string, error)

		During(ctx context.Context, start, end int64) ([]*CRank, error)
	}

	// customCRankModel 包含默认模型的实现
	customCRankModel struct {
		*defaultCRankModel
	}
)

const (
	CategoryMonth = 1 + iota
	CategoryAll
)

func (f *CRank) Description() interface{} {
	return f.Name
}

// NewCRankModel 返回数据库表的模型
func NewCRankModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CRankModel {
	return &customCRankModel{
		defaultCRankModel: newCRankModel(conn, c, opts...),
	}
}

// FindDataByDescription 根据描述查找数据
func (m *defaultCRankModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	name, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByName(ctx, name)
}

// ExistByDescription 检查描述所指数据是否存在
func (m *defaultCRankModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

// InsertData 插入新数据
func (m *defaultCRankModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*CRank); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

// UpdateDataByDescription 根据描述更新数据
func (m *defaultCRankModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*CRank); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Name)
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*CRank); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultCRankModel) FirstRankTimeByJavId(ctx context.Context, javId string) (int64, error) {
	var dateNumber int64

	query, args, err := squirrel.Select("day_number").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_jav_id": javId}).
		OrderBy("day_number ASC").
		Limit(1).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("构建 SQL 查询失败: %w", err)
	}

	err = m.QueryRowNoCacheCtx(ctx, &dateNumber, query, args...)
	if err != nil {
		return 0, err
	}

	return dateNumber, nil
}

func (m *defaultCRankModel) FindAllDistinctMovieJavIds(ctx context.Context) ([]string, error) {
	var movieJavIds []string

	// 使用 SQL 构建器来创建查询语句
	query, args, err := squirrel.Select("DISTINCT movie_jav_id").
		From(m.tableName()).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("构建 SQL 查询失败: %w", err)
	}

	// 执行查询并填充 movieJavIds 切片
	err = m.QueryRowsNoCacheCtx(ctx, &movieJavIds, query, args...)
	if err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil // 对于没有找到记录的情况，返回空切片
		}
		return nil, err
	}

	return movieJavIds, nil
}

func (m *defaultCRankModel) FindHighestRank(ctx context.Context, movieJavId string, limit uint64) ([]*CRank, error) {
	var ranks []*CRank

	// 使用 SQL 构建器来创建查询语句
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_jav_id": movieJavId}).
		OrderBy("number ASC, day_number ASC").
		Limit(limit).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	// 执行查询并填充 ranks 切片
	if err := m.QueryRowsNoCacheCtx(ctx, &ranks, query, args...); err != nil {
		return nil, err
	}

	return ranks, nil
}

func (m *defaultCRankModel) DistinctJavId(ctx context.Context) ([]string, error) {
	var javIds []string

	// 使用 SQL 构建器来创建查询语句
	query, args, err := squirrel.Select("Distinct(movie_jav_id)").
		From(m.tableName()).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	// 执行查询并填充 ranks 切片
	if err := m.QueryRowsNoCacheCtx(ctx, &javIds, query, args...); err != nil {
		return nil, err
	}

	return javIds, nil
}

func (m *defaultCRankModel) FindLatest(ctx context.Context) (*CRank, error) {
	var latest CRank
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		OrderBy("day_number desc").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	err = m.QueryRowNoCacheCtx(ctx, &latest, query, args...)
	if err != nil {
		return nil, err
	}

	return &latest, nil
}

func (m *defaultCRankModel) FindJavIdsByDayNumber(ctx context.Context, dayNumber int64) ([]string, error) {
	var result []string
	// 使用 SQL 构建器来创建查询语句
	query, args, err := squirrel.Select("movie_jav_id").
		From(m.tableName()).
		Where(squirrel.Eq{"day_number": dayNumber}).
		OrderBy("number ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	// 执行查询并填充 ranks 切片
	if err := m.QueryRowsNoCacheCtx(ctx, &result, query, args...); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *defaultCRankModel) During(ctx context.Context, start, end int64) ([]*CRank, error) {
	var ranks []*CRank
	query, args, err := squirrel.Select("*").
		From(m.tableName()).
		Where("`day_number` >= ? and `day_number` <= ?", start, end).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	err = m.QueryRowsNoCacheCtx(ctx, &ranks, query, args...)
	if err != nil {
		return nil, err
	}

	return ranks, nil
}
