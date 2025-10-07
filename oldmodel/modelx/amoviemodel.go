package modelx

import (
	"context"
	"errors"
	"rudy_gc/pkg/orm"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AMovieModel = (*customAMovieModel)(nil)

type (
	// AMovieModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAMovieModel.
	AMovieModel interface {
		aMovieModel
		FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error)
		ExistByDescription(ctx context.Context, description any) (bool, error)
		InsertData(ctx context.Context, data orm.DataStruct) error
		UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error

		FindMoviesComplex(ctx context.Context, r *FindMovieV1Request) ([]*AMovie, int64, error)
		FindMoviesNeedDownload(ctx context.Context, orderBy string, owned, page, pageSize int64) ([]*AMovie, int64, error)

		FindRandomMovies(ctx context.Context, count int) ([]*AMovie, error)
		FindRandomMoviesOwn(ctx context.Context, count int) ([]*AMovie, error)
		FindMoviesAll(ctx context.Context, limit uint64) ([]*AMovie, error)

		FindMoviesNeedRefresh(ctx context.Context) ([]*AMovie, error)
		FindMoviesHasNoChinese(ctx context.Context) ([]*AMovie, error)
		FindMoviesByName(ctx context.Context, name string) ([]*AMovie, error)
		FindMoviesByEncodeName(ctx context.Context, encodeName string) ([]*AMovie, error)
		FindMoviesHasNoCover(ctx context.Context) ([]*AMovie, error)

		FindMovieHasNoChinese(ctx context.Context) (*AMovie, int64, error)
		FindMovieHasNoCover(ctx context.Context) (*AMovie, int64, error)
	}

	customAMovieModel struct {
		*defaultAMovieModel
	}
)

// NewAMovieModel returns a model for the database table.
func NewAMovieModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AMovieModel {
	return &customAMovieModel{
		defaultAMovieModel: newAMovieModel(conn, c, opts...),
	}
}

func (f *AMovie) Description() interface{} {
	return f.JavId
}

func (m *defaultAMovieModel) FindDataByDescription(ctx context.Context, description any) (orm.DataStruct, error) {
	movieId, ok := description.(string)
	if !ok {
		return nil, sqlc.ErrNotFound
	}
	return m.FindOneByJavId(ctx, movieId)
}

func (m *defaultAMovieModel) ExistByDescription(ctx context.Context, description any) (bool, error) {
	_, err := m.FindDataByDescription(ctx, description)
	if err == nil {
		return true, nil
	} else if errors.Is(err, sqlc.ErrNotFound) {
		return false, nil
	}
	return false, err
}

func (m *defaultAMovieModel) InsertData(ctx context.Context, data orm.DataStruct) error {
	if insertData, ok := data.(*AMovie); ok {
		insertData.CreatedOn = time.Now().Unix()
		_, err := m.Insert(ctx, insertData)
		return err
	}
	return ErrInvalidData
}

func (m *defaultAMovieModel) UpdateDataByDescription(ctx context.Context, data orm.DataStruct) error {
	if newData, ok := data.(*AMovie); ok {
		dataDb, err := m.FindDataByDescription(ctx, newData.Description())
		if err != nil {
			return err
		}
		if existingData, ok := dataDb.(*AMovie); ok {
			newData.Id = existingData.Id
			newData.CreatedOn = existingData.CreatedOn
			newData.UpdatedOn = time.Now().Unix()
			return m.Update(ctx, newData)
		}
		return ErrInvalidData
	}
	return ErrInvalidData
}
