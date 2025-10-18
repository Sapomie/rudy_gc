// internal/spider/logic/a_save_parsed_movie.go
package logic

import (
	"context"
	"fmt"
	"math"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strconv"
	"time"
)

// saveParsedMovieResponse 返回保存后的电影记录与演员 javId 集合（供后续统计/链路使用）
type saveParsedMovieResponse struct {
	movie        *types.Movie
	castJavIdMap map[string]struct{}
}

// 弱信号：可筛掉的分类（与老项目保持一致）
var genreUnused = map[string]struct{}{
	"单体作品":  {},
	"薄马赛克":  {},
	"数位马赛克": {},
}

func (l *CrawlLogic) saveParsedMovie(ctx context.Context, raw *RawJavMovie) (*saveParsedMovieResponse, error) {
	// ===== 1) 解析原始数值字段 =====
	length, err := strconv.Atoi(raw.Length)
	if err != nil {
		return nil, fmt.Errorf("片长解析失败: %w", err)
	}
	viewWatched, err := strconv.Atoi(raw.Watched)
	if err != nil {
		return nil, fmt.Errorf("已看人数解析失败: %w", err)
	}
	viewOwned, err := strconv.Atoi(raw.Owned)
	if err != nil {
		return nil, fmt.Errorf("拥有人数解析失败: %w", err)
	}
	viewWanted, err := strconv.Atoi(raw.Subscribed)
	if err != nil {
		return nil, fmt.Errorf("想看人数解析失败: %w", err)
	}
	releasingDate, err := raw.getReleasingDate()
	if err != nil {
		return nil, fmt.Errorf("发行日解析失败: %w", err)
	}
	score, err := raw.getScore()
	if err != nil {
		return nil, fmt.Errorf("评分解析失败: %w", err)
	}

	// ===== 2) 字典实体 Upsert =====
	dirID, err := l.deps.DirectorRepo.GetOrCreateByName(ctx, raw.Director.Name, raw.Director.JavId)
	if err != nil {
		return nil, fmt.Errorf("导演 Upsert 失败: %w", err)
	}
	mkrID, err := l.deps.MakerRepo.GetOrCreateByName(ctx, raw.Maker.Name, raw.Maker.JavId)
	if err != nil {
		return nil, fmt.Errorf("厂牌 Upsert 失败: %w", err)
	}
	labID, err := l.deps.LabelRepo.GetOrCreateByName(ctx, raw.Label.Name, raw.Label.JavId)
	if err != nil {
		return nil, fmt.Errorf("标签 Upsert 失败: %w", err)
	}
	pfxID, err := l.deps.PrefixRepo.GetOrCreateByName(ctx, raw.Prefix)
	if err != nil {
		return nil, fmt.Errorf("前缀 Upsert 失败: %w", err)
	}

	genreIDs := make([]int64, 0, len(raw.Genres))
	for _, g := range raw.Genres {
		if _, drop := genreUnused[g.Name]; drop {
			continue
		}
		gid, gerr := l.deps.GenreRepo.GetOrCreateByName(ctx, g.Name, g.JavId)
		if gerr != nil {
			return nil, fmt.Errorf("类型 Upsert 失败(%s): %w", g.Name, gerr)
		}
		genreIDs = append(genreIDs, gid)
	}

	castIDs := make([]int64, 0, len(raw.Casts))
	castJavIdMap := make(map[string]struct{}, len(raw.Casts))
	for _, c := range raw.Casts {
		cid, cerr := l.deps.CastRepo.GetOrCreateByName(ctx, c.Name, c.JavId)
		if cerr != nil {
			return nil, fmt.Errorf("演员 Upsert 失败(%s): %w", c.Name, cerr)
		}
		castIDs = append(castIDs, cid)
		castJavIdMap[c.JavId] = struct{}{}
	}

	//cast_age
	castAvgAgeTenth := int64(0)
	{
		var (
			sumYears float64
			cnt      int
		)
		for _, c := range raw.Casts {
			// 使用 Repo 查询生日
			birth, found, err := l.deps.CafoRepo.FindBirthByName(ctx, c.Name)
			if err != nil {
				return nil, fmt.Errorf("查询 Cafo 失败(%s): %w", c.Name, err)
			}
			if found && birth > 0 {
				years := float64(releasingDate-birth) / (3600.0 * 24.0 * 365.0)
				sumYears += years
				cnt++
			}
		}
		if cnt > 0 {
			avg := sumYears / float64(cnt)  // 平均岁数
			avg10 := math.Round(avg * 10.0) // 保留 1 位小数 → ×10 四舍五入
			castAvgAgeTenth = int64(avg10)  // 23.7 → 237
		}
	}

	// ===== 3) Upsert 电影主体（a_movie）=====
	now := time.Now().Unix()
	mv := &types.Movie{
		Name:                 raw.Designation,
		JavId:                raw.JavId,
		Title:                raw.Title,
		EncodeName:           fmt.Sprintf("%04d-%s", pfxID, raw.Number),
		ReleasingDate:        releasingDate,
		Length:               int64(length),
		Score:                score,
		ViewersNumberWant:    int64(viewWanted),
		ViewersNumberOwned:   int64(viewOwned),
		ViewersNumberWatched: int64(viewWatched),
		PrefixId:             pfxID,
		MakerId:              mkrID,
		LabelId:              labID,
		DirectorId:           dirID,
		CastNumber:           int64(len(castIDs)),
		//按“*10 的整数”策略存演员平均年龄（若将来有生日数据再计算）
		CastAverageAge:   castAvgAgeTenth,
		DetailUpdateTime: raw.LastQueryTime,
		CreatedOn:        now,
		UpdatedOn:        now,
	}

	//todo:1.事物        2.BatchTryLink(movieId, ids []int64)
	mvSaved, err := l.deps.MovieRepo.UpsertByJavId(ctx, mv)
	if err != nil {
		return nil, fmt.Errorf("保存电影失败: %w", err)
	}

	// ===== 4) Upsert bm_murl（海报/小图）=====
	murl := &types.Murl{
		Name:           raw.Designation,
		JavId:          raw.JavId,
		JacketImg:      raw.ImgUrl,
		JacketImgLocal: "", // 由下载流程后续写入
		CreatedOn:      now,
		UpdatedOn:      now,
	}
	if err := l.deps.MurlRepo.UpsertByJavIdPreserveLocal(ctx, murl); err != nil {
		return nil, fmt.Errorf("保存 MURL 失败: %w", err)
	}

	// ===== 5) Upsert bm_minfo（编码名/中文/下载需求/排行信息）=====
	minfo := &types.Minfo{
		Name:          raw.Designation,
		JavId:         raw.JavId,
		ReleasingDate: releasingDate,
		NeedDownload:  consts.MovieNeedDownLoadNone,
		// Chinese / FirstRankDayNumber / HighestRank / DaysInRank / NeedDownload
		// 这些由 Repo 内部“保留旧值”策略处理
		CreatedOn: now,
		UpdatedOn: now,
	}
	if err := l.deps.MinfoRepo.UpsertPreserve(ctx, minfo); err != nil {
		return nil, fmt.Errorf("保存 MINFO 失败: %w", err)
	}

	// ===== 6) 关系表（amr_movie_cast / amr_movie_genre）=====
	for _, cid := range castIDs {
		if err := l.deps.MovieCastRepo.TryLink(ctx, mvSaved.JavId, cid, now); err != nil {
			return nil, fmt.Errorf("建立关系 movie_cast 失败: %w", err)
		}
	}
	for _, gid := range genreIDs {
		if err := l.deps.MovieGenreRepo.TryLink(ctx, mvSaved.JavId, gid, now); err != nil {
			return nil, fmt.Errorf("建立关系 movie_genre 失败: %w", err)
		}
	}
	l.movieSvc.InvalidateMovieType(ctx, raw.JavId)

	return &saveParsedMovieResponse{
		movie:        mvSaved,
		castJavIdMap: castJavIdMap,
	}, nil
}

/************** RawJavMovie 的解析工具（延用旧项目逻辑） **************/

func (r *RawJavMovie) getScore() (score int64, err error) {
	if len(r.Score) == 0 {
		return 0, nil
	}
	scoreStr := r.Score[1 : len(r.Score)-1]
	scoreFloat, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return -1, err
	}
	return int64(scoreFloat * 10), nil
}

func (r *RawJavMovie) getReleasingDate() (dateUnix int64, err error) {
	yearStr := r.Date[:4]
	monthStr := r.Date[5:7]
	dayStr := r.Date[8:10]

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return -1, err
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return -1, err
	}
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return -1, err
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Unix(), nil
}
