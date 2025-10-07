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

var _ BmrMovieCastModel = (*customBmrMovieCastModel)(nil)

type (
	// BmrMovieCastModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBmrMovieCastModel.
	BmrMovieCastModel interface {
		bmrMovieCastModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindMovieIdsByCastId(ctx context.Context, castId int64) ([]int64, error)
		FindCastIdsByMovieId(ctx context.Context, movieId int64) ([]int64, error)
	}

	customBmrMovieCastModel struct {
		*defaultBmrMovieCastModel
	}
)

// NewBmrMovieCastModel returns a model for the database table.
func NewBmrMovieCastModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) BmrMovieCastModel {
	return &customBmrMovieCastModel{
		defaultBmrMovieCastModel: newBmrMovieCastModel(conn, c, opts...),
	}
}

func (g *BmrMovieCast) Description() interface{} {
	return makeFakeNameMovieColumn(g.MovieId, g.CastId)
}

func (m *defaultBmrMovieCastModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	desc, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	movieId, genreId, err := decodeFakeNameMovieColumn(desc)
	if err != nil {
		return nil, err
	}

	return m.FindOneByMovieIdCastId(ctx, movieId, genreId)
}

func (m *defaultBmrMovieCastModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if err == sqlc.ErrNotFound {
		return false, nil
	}
	return false, err
}

func (m *defaultBmrMovieCastModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*BmrMovieCast); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmrMovieCastModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*BmrMovieCast); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*BmrMovieCast); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return fmt.Errorf("invalid data")
	}
	return fmt.Errorf("invalid data type")
}

func (m *defaultBmrMovieCastModel) FindMovieIdsByCastId(ctx context.Context, castId int64) ([]int64, error) {
	query, args, err := squirrel.Select("movie_id").
		From(m.tableName()).
		Where(squirrel.Eq{"cast_id": castId}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("构建 SQL 查询失败: %w", err)
	}

	var movieIds []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &movieIds, query, args...); err != nil {
		return nil, err
	}

	return movieIds, nil
}

func (m *defaultBmrMovieCastModel) FindCastIdsByMovieId(ctx context.Context, movieId int64) ([]int64, error) {
	query, args, err := squirrel.Select("cast_id").
		From(m.tableName()).
		Where(squirrel.Eq{"movie_id": movieId}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("构建 SQL 查询失败: %w", err)
	}

	var movieIds []int64
	if err := m.QueryRowsNoCacheCtx(ctx, &movieIds, query, args...); err != nil {
		return nil, err
	}

	return movieIds, nil
}
