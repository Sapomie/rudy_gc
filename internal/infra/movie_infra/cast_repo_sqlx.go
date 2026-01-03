package movie_infra

import (
	"context"
	"errors"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/movie_repo"
	"rudy_gc/internal/types"
)

var _ movie_repo.CastRepo = (*CastRepoSqlx)(nil)

type CastRepoSqlx struct {
	m moviex.AmCastModel
}

func NewCastRepoSqlx(m moviex.AmCastModel) movie_repo.CastRepo {
	return &CastRepoSqlx{m: m}
}

// ====== 已有：保持不变（返回 id）
func (r *CastRepoSqlx) GetOrCreateByName(ctx context.Context, name, javId string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}
	if row != nil {
		return row.Id, nil
	}
	now := time.Now().Unix()
	res, err := r.m.Insert(ctx, &moviex.AmCast{
		Name:      name,
		JavId:     javId,
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// ====== 新版：FindOne / FindOneByName 返回 types.Cast
func (r *CastRepoSqlx) FindOne(ctx context.Context, id int64) (*types.Cast, error) {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapAmCastToTypes(row), nil
}

func (r *CastRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Cast, error) {
	row, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return mapAmCastToTypes(row), nil
}

// ====== 新增：Upsert（以 name 作为幂等键）
func (r *CastRepoSqlx) Upsert(ctx context.Context, in *types.Cast) (*types.Cast, error) {
	if in == nil {
		return nil, errors.New("nil input")
	}
	now := time.Now().Unix()

	// 以 name 为幂等键
	old, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}

	if old != nil {
		// 覆盖式更新（按你的字段一一映射）
		old.JavId = in.JavId
		old.MovieNumber = in.MovieNumber
		old.OwnedMovieNumber = in.OwnedMovieNumber
		old.ScTimes = in.ScTimes
		old.ComeTimes = in.ComeTimes
		old.LastScTime = in.LastScTime
		old.Rank500MovieNumber = in.Rank500MovieNumber
		old.Rank20MovieNumber = in.Rank20MovieNumber
		old.Rank1MovieNumber = in.Rank1MovieNumber
		old.HighestRank = in.HighestRank
		old.RankTimes = in.RankTimes
		if in.CreatedOn > 0 {
			old.CreatedOn = in.CreatedOn
		}
		old.UpdatedOn = now

		if err := r.m.Update(ctx, old); err != nil {
			return nil, err
		}
		return mapAmCastToTypes(old), nil
	}

	// 插入
	row := &moviex.AmCast{
		Name:               in.Name,
		JavId:              in.JavId,
		MovieNumber:        in.MovieNumber,
		OwnedMovieNumber:   in.OwnedMovieNumber,
		ScTimes:            in.ScTimes,
		ComeTimes:          in.ComeTimes,
		LastScTime:         in.LastScTime,
		Rank500MovieNumber: in.Rank500MovieNumber,
		Rank20MovieNumber:  in.Rank20MovieNumber,
		Rank1MovieNumber:   in.Rank1MovieNumber,
		HighestRank:        in.HighestRank,
		RankTimes:          in.RankTimes,
		CreatedOn:          ifElseInt64(in.CreatedOn > 0, in.CreatedOn, now),
		UpdatedOn:          now,
	}
	if _, err := r.m.Insert(ctx, row); err != nil {
		// 并发兜底
		if again, e2 := r.m.FindOneByName(ctx, in.Name); e2 == nil && again != nil {
			return mapAmCastToTypes(again), nil
		}
		return nil, err
	}
	ins, err := r.m.FindOneByName(ctx, in.Name)
	if err != nil {
		return nil, err
	}
	return mapAmCastToTypes(ins), nil
}

/******** helpers ********/

func mapAmCastToTypes(v *moviex.AmCast) *types.Cast {
	if v == nil {
		return nil
	}
	return &types.Cast{
		Id:                 v.Id,
		Name:               v.Name,
		JavId:              v.JavId,
		MovieNumber:        v.MovieNumber,
		OwnedMovieNumber:   v.OwnedMovieNumber,
		ScTimes:            v.ScTimes,
		ComeTimes:          v.ComeTimes,
		LastScTime:         v.LastScTime,
		Rank500MovieNumber: v.Rank500MovieNumber,
		Rank20MovieNumber:  v.Rank20MovieNumber,
		Rank1MovieNumber:   v.Rank1MovieNumber,
		HighestRank:        v.HighestRank,
		RankTimes:          v.RankTimes,
		CreatedOn:          v.CreatedOn,
		UpdatedOn:          v.UpdatedOn,
	}
}

func ifElseInt64(cond bool, a, b int64) int64 {
	if cond {
		return a
	}
	return b
}

func (r *CastRepoSqlx) UpdateMovieNumbersByID(ctx context.Context, id int64, ownedRemovedStatus int64, now int64) error {
	movieNumber, ownedMovieNumber, err := r.m.GetMovieNumbersByID(ctx, id, ownedRemovedStatus)
	if err != nil {
		return err
	}

	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now

	return r.m.Update(ctx, row)
}
