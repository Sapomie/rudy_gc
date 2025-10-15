// internal/infra/film_infra/film_repo_sqlx.go
package film_infra

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/film_repo"
	"rudy_gc/internal/types"
)

var _ film_repo.FilmRepo = (*FilmRepoSqlx)(nil)

type FilmRepoSqlx struct {
	m moviex.VFilmModel
}

func NewFilmRepoSqlx(m moviex.VFilmModel) *FilmRepoSqlx {
	return &FilmRepoSqlx{m: m}
}

func (r *FilmRepoSqlx) FindOne(ctx context.Context, id int64) (*types.Film, error) {
	row, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapModelxToTypes(row), nil
}

func (r *FilmRepoSqlx) FindOneByMovieJavId(ctx context.Context, javId string) (*types.Film, error) {
	row, err := r.m.FindOneByMovieJavId(ctx, javId)
	if err != nil {
		return nil, err
	}
	return mapModelxToTypes(row), nil
}

func (r *FilmRepoSqlx) FindOneByMovieName(ctx context.Context, name string) (*types.Film, error) {
	row, err := r.m.FindOneByMovieName(ctx, name)
	if err != nil {
		return nil, err
	}
	return mapModelxToTypes(row), nil
}

func (r *FilmRepoSqlx) UpsertFilm(ctx context.Context, in *types.Film) (*types.Film, consts.UpsertStatus, error) {
	if in == nil {
		return nil, 0, errors.New("nil input")
	}

	if row, err := r.m.FindOneByMovieJavId(ctx, in.MovieJavId); err == nil && row != nil {
		changed := false

		// 一次性对比 & 同步所有业务字段（除了 Id/CreatedOn）
		if row.MovieName != in.MovieName {
			row.MovieName = in.MovieName
			changed = true
		}
		if row.FileName != in.FileName {
			row.FileName = in.FileName
			changed = true
		}
		if row.DirectoryId != in.DirectoryId {
			row.DirectoryId = in.DirectoryId
			changed = true
		}
		if row.RootDir != in.RootDir {
			row.RootDir = in.RootDir
			changed = true
		}
		if row.FullDir != in.FullDir {
			row.FullDir = in.FullDir
			changed = true
		}

		if row.Dir1Id != in.Dir1Id {
			row.Dir1Id = in.Dir1Id
			changed = true
		}
		if row.Dir2Id != in.Dir2Id {
			row.Dir2Id = in.Dir2Id
			changed = true
		}
		if row.Dir3Id != in.Dir3Id {
			row.Dir3Id = in.Dir3Id
			changed = true
		}
		if row.Dir4Id != in.Dir4Id {
			row.Dir4Id = in.Dir4Id
			changed = true
		}

		if row.Alias != in.Alias {
			row.Alias = in.Alias
			changed = true
		}
		if row.Size != in.Size {
			row.Size = in.Size
			changed = true
		}
		if row.Width != in.Width {
			row.Width = in.Width
			changed = true
		}
		if row.Height != in.Height {
			row.Height = in.Height
			changed = true
		}
		if row.BitRate != in.BitRate {
			row.BitRate = in.BitRate
			changed = true
		}
		if row.Duration != in.Duration {
			row.Duration = in.Duration
			changed = true
		}
		if row.FrameAverage != in.FrameAverage {
			row.FrameAverage = in.FrameAverage
			changed = true
		}

		if row.HasSub != in.HasSub {
			row.HasSub = in.HasSub
			changed = true
		}
		if row.SelfMake != in.SelfMake {
			row.SelfMake = in.SelfMake
			changed = true
		}
		if row.HasMask != in.HasMask {
			row.HasMask = in.HasMask
			changed = true
		}

		if row.NeedScanMeta != in.NeedScanMeta {
			row.NeedScanMeta = in.NeedScanMeta
			changed = true
		}
		if row.IsRemoved != in.IsRemoved {
			row.IsRemoved = in.IsRemoved
			changed = true
		}
		if row.RemoveTime != in.RemoveTime {
			row.RemoveTime = in.RemoveTime
			changed = true
		}
		if row.ScTimes != in.ScTimes {
			row.ScTimes = in.ScTimes
			changed = true
		}
		if row.ComeTimes != in.ComeTimes {
			row.ComeTimes = in.ComeTimes
			changed = true
		}
		if row.LastScTime != in.LastScTime {
			row.LastScTime = in.LastScTime
			changed = true
		}
		if row.BirthTime != in.BirthTime {
			row.BirthTime = in.BirthTime
			changed = true
		}
		if row.ReleasingDate != in.ReleasingDate {
			row.ReleasingDate = in.ReleasingDate
			changed = true
		}

		if changed {
			row.UpdatedOn = time.Now().Unix()
			if err := r.m.Update(ctx, row); err != nil {
				return nil, 0, err
			}
			return mapModelxToTypes(row), consts.UpsertUpdated, nil
		}
		return mapModelxToTypes(row), consts.UpsertUnchanged, nil
	}

	// 不存在：插入
	mv := mapTypesToModelx(in)
	now := time.Now().Unix()
	mv.CreatedOn = now
	mv.UpdatedOn = now

	if _, err := r.m.Insert(ctx, mv); err != nil {
		if row2, err2 := r.m.FindOneByMovieJavId(ctx, in.MovieJavId); err2 == nil && row2 != nil {
			return mapModelxToTypes(row2), consts.UpsertUpdated, nil // 并发插入时算更新
		}
		return nil, 0, err
	}

	row3, err := r.m.FindOneByMovieJavId(ctx, in.MovieJavId)
	if err != nil || row3 == nil {
		return nil, 0, err
	}

	return mapModelxToTypes(row3), consts.UpsertInserted, nil
}

func (r *FilmRepoSqlx) FindAll(ctx context.Context, removedStatus int64) ([]*types.Film, error) {
	rows, err := r.m.FindAll(ctx, removedStatus)
	if err != nil {
		return nil, err
	}

	list := make([]*types.Film, 0, len(rows))
	for _, mv := range rows {
		list = append(list, mapModelxToTypes(mv))
	}
	return list, nil
}

/* ---------------- helpers ---------------- */

func mapTypesToModelx(in *types.Film) *moviex.VFilm {
	return &moviex.VFilm{
		Id:          in.Id,
		MovieJavId:  in.MovieJavId,
		MovieName:   in.MovieName,
		FileName:    in.FileName,
		DirectoryId: in.DirectoryId,
		RootDir:     in.RootDir,
		FullDir:     in.FullDir,

		// 新增映射
		Dir1Id: in.Dir1Id,
		Dir2Id: in.Dir2Id,
		Dir3Id: in.Dir3Id,
		Dir4Id: in.Dir4Id,

		Alias:        in.Alias,
		Size:         in.Size,
		Width:        in.Width,
		Height:       in.Height,
		BitRate:      in.BitRate,
		Duration:     in.Duration,
		FrameAverage: in.FrameAverage,

		HasSub:   in.HasSub,
		SelfMake: in.SelfMake,
		HasMask:  in.HasMask,

		NeedScanMeta:  in.NeedScanMeta,
		IsRemoved:     in.IsRemoved,
		RemoveTime:    in.RemoveTime,
		ScTimes:       in.ScTimes,
		ComeTimes:     in.ComeTimes,
		LastScTime:    in.LastScTime,
		BirthTime:     in.BirthTime,
		ReleasingDate: in.ReleasingDate,

		CreatedOn: in.CreatedOn,
		UpdatedOn: in.UpdatedOn,
	}
}

func mapModelxToTypes(mv *moviex.VFilm) *types.Film {
	return &types.Film{
		Id:          mv.Id,
		MovieJavId:  mv.MovieJavId,
		MovieName:   mv.MovieName,
		FileName:    mv.FileName,
		DirectoryId: mv.DirectoryId,
		RootDir:     mv.RootDir,
		FullDir:     mv.FullDir,

		// 新增映射
		Dir1Id: mv.Dir1Id,
		Dir2Id: mv.Dir2Id,
		Dir3Id: mv.Dir3Id,
		Dir4Id: mv.Dir4Id,

		Alias:        mv.Alias,
		Size:         mv.Size,
		Width:        mv.Width,
		Height:       mv.Height,
		BitRate:      mv.BitRate,
		Duration:     mv.Duration,
		FrameAverage: mv.FrameAverage,

		HasSub:   mv.HasSub,
		SelfMake: mv.SelfMake,
		HasMask:  mv.HasMask,

		NeedScanMeta:  mv.NeedScanMeta,
		IsRemoved:     mv.IsRemoved,
		RemoveTime:    mv.RemoveTime,
		ScTimes:       mv.ScTimes,
		ComeTimes:     mv.ComeTimes,
		LastScTime:    mv.LastScTime,
		BirthTime:     mv.BirthTime,
		ReleasingDate: mv.ReleasingDate,

		CreatedOn: mv.CreatedOn,
		UpdatedOn: mv.UpdatedOn,
	}
}
