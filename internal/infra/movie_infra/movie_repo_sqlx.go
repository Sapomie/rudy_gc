package movie_infra

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"
)

var _ movie_repo.MovieRepo = (*MovieRepoSqlx)(nil)

type MovieRepoSqlx struct {
	m moviex.AMovieModel
}

func NewMovieRepoSqlx(m moviex.AMovieModel) movie_repo.MovieRepo {
	return &MovieRepoSqlx{m: m}
}

// FindByJavId 查找电影
func (r *MovieRepoSqlx) FindOneByJavId(ctx context.Context, javId string) (*types.Movie, error) {
	row, err := r.m.FindOneByJavId(ctx, javId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find movie by javId failed: %w", err)
	}
	return toTypesMovie(row), nil
}

// UpsertByJavId 按 JavId 保存（幂等）
func (r *MovieRepoSqlx) UpsertByJavId(ctx context.Context, mv *types.Movie) (*types.Movie, error) {
	// 先查
	exist, err := r.m.FindOneByJavId(ctx, mv.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	if exist != nil {
		// 更新：保留 CreatedOn
		exist.Name = mv.Name
		exist.Title = mv.Title
		exist.ReleasingDate = mv.ReleasingDate
		exist.Length = mv.Length
		exist.Score = mv.Score
		exist.ViewersNumberWant = mv.ViewersNumberWant
		exist.ViewersNumberOwned = mv.ViewersNumberOwned
		exist.ViewersNumberWatched = mv.ViewersNumberWatched
		exist.PrefixId = mv.PrefixId
		exist.MakerId = mv.MakerId
		exist.LabelId = mv.LabelId
		exist.DirectorId = mv.DirectorId
		exist.CastNumber = mv.CastNumber
		exist.CastAverageAge = mv.CastAverageAge
		exist.DetailUpdateTime = mv.DetailUpdateTime
		exist.UpdatedOn = mv.UpdatedOn

		if err := r.m.Update(ctx, exist); err != nil {
			return nil, fmt.Errorf("update movie(%s) failed: %w", mv.JavId, err)
		}
		return toTypesMovie(exist), nil
	}

	// 插入
	row := &moviex.AMovie{
		Name:                 mv.Name,
		JavId:                mv.JavId,
		Title:                mv.Title,
		ReleasingDate:        mv.ReleasingDate,
		Length:               mv.Length,
		Score:                mv.Score,
		ViewersNumberWant:    mv.ViewersNumberWant,
		ViewersNumberOwned:   mv.ViewersNumberOwned,
		ViewersNumberWatched: mv.ViewersNumberWatched,
		PrefixId:             mv.PrefixId,
		MakerId:              mv.MakerId,
		LabelId:              mv.LabelId,
		DirectorId:           mv.DirectorId,
		CastNumber:           mv.CastNumber,
		CastAverageAge:       mv.CastAverageAge,
		DetailUpdateTime:     mv.DetailUpdateTime,
		CreatedOn:            mv.CreatedOn,
		UpdatedOn:            mv.UpdatedOn,
	}
	ret, err := r.m.Insert(ctx, row)
	if err != nil {
		return nil, fmt.Errorf("insert movie(%s) failed: %w", mv.JavId, err)
	}
	if id, e := ret.LastInsertId(); e == nil {
		row.Id = id // 关键：填回自增ID
	} else {
		// 少数驱动拿不到 last insert id 时，兜底回查
		got, fe := r.m.FindOneByJavId(ctx, mv.JavId)
		if fe != nil {
			return nil, fmt.Errorf("insert ok but re-fetch movie(%s) failed: %w", mv.JavId, fe)
		}
		row = got
	}
	return toTypesMovie(row), nil
}

// 内部转换函数
func toTypesMovie(mv *moviex.AMovie) *types.Movie {
	if mv == nil {
		return nil
	}
	return &types.Movie{
		Id:                   mv.Id,
		Name:                 mv.Name,
		JavId:                mv.JavId,
		Title:                mv.Title,
		ReleasingDate:        mv.ReleasingDate,
		Length:               mv.Length,
		Score:                mv.Score,
		ViewersNumberWant:    mv.ViewersNumberWant,
		ViewersNumberOwned:   mv.ViewersNumberOwned,
		ViewersNumberWatched: mv.ViewersNumberWatched,
		PrefixId:             mv.PrefixId,
		MakerId:              mv.MakerId,
		LabelId:              mv.LabelId,
		DirectorId:           mv.DirectorId,
		CastNumber:           mv.CastNumber,
		CastAverageAge:       mv.CastAverageAge,
		DetailUpdateTime:     mv.DetailUpdateTime,
		CreatedOn:            mv.CreatedOn,
		UpdatedOn:            mv.UpdatedOn,
	}
}
func (r *MovieRepoSqlx) CountAll(ctx context.Context) (int64, error) {
	return r.m.CountAll(ctx)
}

func (r *MovieRepoSqlx) ListMovies(ctx context.Context, offset, limit int64) ([]*types.Movie, error) {
	rows, err := r.m.ListPage(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list movies failed: %w", err)
	}

	result := make([]*types.Movie, 0, len(rows))
	for _, row := range rows {
		result = append(result, toTypesMovie(row))
	}
	return result, nil
}
