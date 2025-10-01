package logic

import "time"

func (l *CrawlLogic) insertRaw(raw *RawJavMovie) (*insertRawResponse, error) {
	ctx := l.ctx
	now := time.Now()

	return l.deps.DB.WithTx(ctx, func(tx Tx) (*insertRawResponse, error) {
		// 1) 解析/兜底
		length := parseIntSafe(raw.Length)
		viewWatched := parseIntSafe(raw.Watched)
		viewOwned := parseIntSafe(raw.Owned)
		viewWanted := parseIntSafe(raw.Subscribed)
		releaseUnix, err := parseDateSafe(raw.Date) // 用 time.Parse
		if err != nil {
			return nil, err
		}
		score := parseScoreSafe(raw.Score) // 正则提取

		// 2) 批量 upsert 基础字典（导演/厂商/标签/前缀/题材/演员）
		directorID, makerID, labelID, prefixID, err := upsertBasics(tx, raw)
		if err != nil {
			return nil, err
		}

		genreIDs, err := upsertGenres(tx, raw.Genres, cons.GenreUnused)
		if err != nil {
			return nil, err
		}

		castIDs, avgAge, err := upsertCastsAndComputeAvgAge(tx, raw.Casts, releaseUnix)
		if err != nil {
			return nil, err
		}

		// 3) 电影 upsert（带 encodeName）
		encode := buildEncodeName(prefixID, raw.Number)
		movie, err := upsertMovie(tx, &MovieInput{
			JavId: raw.JavId, Title: raw.Title, Designation: raw.Designation,
			Release: releaseUnix, Length: int64(length), Score: score,
			DirectorID: directorID, MakerID: makerID, LabelID: labelID, PrefixID: prefixID,
			ViewWanted: int64(viewWanted), ViewOwned: int64(viewOwned), ViewWatched: int64(viewWatched),
			AvgAge: avgAge, Encode: encode, Birth: raw.BirthTime, LastQuery: raw.LastQueryTime,
		}, now)
		if err != nil {
			return nil, err
		}

		// 4) murl upsert（保留旧的本地路径）
		if err := upsertMurlPreserveLocal(tx, raw, encode); err != nil {
			return nil, err
		}

		// 5) 关系表批量 upsert（movie_cast, movie_genre）
		if err := upsertMovieCasts(tx, movie.Id, castIDs); err != nil {
			return nil, err
		}
		if err := upsertMovieGenres(tx, movie.Id, genreIDs); err != nil {
			return nil, err
		}

		// 6) 回写 detail 状态（NoNeedScan）
		if err := markDetailScanned(tx, raw.JavId, now.Unix()); err != nil {
			return nil, err
		}

		// 7) 返回
		return &insertRawResponse{movie: movie}, nil
	})

}
