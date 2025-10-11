package sc

import (
	"context"
	"fmt"
	"rudy_gc/oldmodel/modelx"
)

type movieScInfo struct {
	ScTimes    int64
	ComeTimes  int64
	LastScTime int64
}

func (l *ScService) AddMovieAndCastScInfo(ctx context.Context, movieJavIdMap map[string]struct{}) error {
	// 1) 拉取原始列表:

	movieJavIdList := make([]string, len(movieJavIdMap))
	for javid := range movieJavIdMap {
		movieJavIdList = append(movieJavIdList, javid)
	}

	gls, err := l.deps.GListRepo.FindGListByMovieJavIds(ctx, movieJavIdList)
	if err != nil {
		return fmt.Errorf("find gList: %w", err)
	}
	if len(gls) == 0 {
		return nil
	}

	// 2) 第一阶段：按 movie 聚合 + 缓存 scTime；同时收集唯一 movieIds
	movieAgg := make(map[string]movieScInfo, len(gls)/2+1)
	uniqueMovieIDs := make(map[string]struct{}, len(gls)/2+1)
	scTimeCache := make(map[string]int64, 256) // ScName -> scTime

	for _, gl := range gls {
		mid := gl.MovieJavId
		uniqueMovieIDs[mid] = struct{}{}

		// 2.1 取 scTime（带缓存）
		scTime, ok := scTimeCache[gl.ScName]
		if !ok {
			var err2 error
			scTime, err2 = getScTime(gl.ScName)
			if err2 != nil {
				return fmt.Errorf("getScTime %q: %w", gl.ScName, err2)
			}
			scTimeCache[gl.ScName] = scTime
		}

		// 2.2 聚合到 movie
		agg := movieAgg[mid] // 值类型，取出来修改再写回
		agg.ScTimes++
		if gl.IsCome == modelx.GListIsCome {
			agg.ComeTimes++
		}
		if scTime > agg.LastScTime {
			agg.LastScTime = scTime
		}
		movieAgg[mid] = agg
	}

	// 3) 第二阶段：把 movie 的聚合结果摊到 cast
	//    只对每个唯一电影查一次完整信息，避免循环内反复查
	castAgg := make(map[string]movieScInfo, 1024)
	for movieJavId := range uniqueMovieIDs {
		// 3.1 查一次完整电影（只按唯一电影）
		movieType, err := l.deps.MovieTypeCache.GetMovieType(ctx, movieJavId)
		if err != nil {
			return fmt.Errorf("FindMovieCompleteByJavId %s: %w", movieJavId, err)
		}

		// 3.2 取该电影的聚合结果
		mInfo, ok := movieAgg[movieJavId]
		if !ok {
			// 理论上不可能，如果出现说明上游数据不一致
			continue
		}

		// 3.3 把该电影的次数与时间摊到所有演职员
		for _, c := range movieType.Cast {
			ci := castAgg[c.Name]
			ci.ScTimes += mInfo.ScTimes
			ci.ComeTimes += mInfo.ComeTimes
			if mInfo.LastScTime > ci.LastScTime {
				ci.LastScTime = mInfo.LastScTime
			}
			castAgg[c.Name] = ci
		}
	}

	// 4) 写回 cast
	for castName, info := range castAgg {
		bmCast, err := l.deps.CastRepo.FindOneByName(ctx, castName)
		if err != nil {
			return fmt.Errorf("cast FindOneByName %s: %w", castName, err)
		}
		bmCast.ScTimes = info.ScTimes
		bmCast.ComeTimes = info.ComeTimes
		bmCast.LastScTime = info.LastScTime

		_, err = l.deps.CastRepo.Upsert(ctx, bmCast)
		if err != nil {
			return fmt.Errorf("upsert:%w", err)
		}
	}

	// 5) 写回 movie
	for movieJavId, info := range movieAgg {
		vFilm, err := l.deps.FilmRepo.FindOneByMovieJavId(ctx, movieJavId)
		if err != nil {
			return fmt.Errorf("MinfoRepo.FindOneByJavId %s: %w", movieJavId, err)
		}
		vFilm.ScTimes = info.ScTimes
		vFilm.ComeTimes = info.ComeTimes
		vFilm.LastScTime = info.LastScTime
		_, _, err = l.deps.FilmRepo.UpsertFilm(ctx, vFilm)
		if err != nil {
			return fmt.Errorf("FilmRepo.UpsertFilm %s: %w", movieJavId, err)
		}
		l.movieSvc.InvalidateMovieType(ctx, movieJavId)
	}

	return nil
}
