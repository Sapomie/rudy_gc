package movie

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/convert"
	"sort"
	"time"
)

const (
	detailFilmInfoNone = 1
	detailFilmInfoOK   = 2
)

func (s *MovieService) buildMovieDetail(ctx context.Context, m *types.Movie) (*types.MovieDetail, error) {
	movieType, err := s.GetMovieType(ctx, m.JavId)
	if err != nil {
		return nil, err
	}
	rankInfos, err := s.findRankInfo(ctx, m.JavId)
	if err != nil {
		return nil, err
	}
	filmInfo, err := s.findFilmInfo(ctx, m.JavId)
	if err != nil {
		return nil, err
	}
	scInfo, err := s.findScInfo(ctx, m.JavId)
	if err != nil {
		return nil, err
	}
	var hasFilm int64 = detailFilmInfoOK
	if filmInfo == nil {
		hasFilm = detailFilmInfoNone
	}

	md := &types.MovieDetail{
		MovieType: movieType,
		FilmInfo:  filmInfo,
		HasFilm:   hasFilm,
		RankInfos: rankInfos,
		SC:        scInfo,
	}
	return md, nil
}

func (s *MovieService) findRankInfo(ctx context.Context, movieJavId string) ([]*types.RankInfo, error) {
	rankMonths, err := s.deps.RankRepo.FindHighestRank(ctx, movieJavId, 1000)
	if err != nil {
		return nil, err
	}

	rankInfos := make([]*types.RankInfo, len(rankMonths))
	for i, rankMonth := range rankMonths {
		info := types.RankInfo{
			Date: consts.GetDateStringByRankDayNumber(rankMonth.DayNumber),
			Rank: rankMonth.RankPos,
		}
		rankInfos[i] = &info
	}

	return rankInfos, nil
}

func (s *MovieService) findFilmInfo(ctx context.Context, movieJavId string) (*types.FilmInfo, error) {
	vf, err := s.deps.FilmRepo.FindOneByMovieJavId(ctx, movieJavId)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	filmInfo := &types.FilmInfo{
		Name:      vf.MovieName,
		BirthTime: time.Unix(vf.BirthTime, 0).Format(time.DateTime),
		Size:      convert.FloatTo(float64(vf.Size) / 1e9).Decimal(2),
		FilePath:  vf.FullDir,
		FileName:  vf.FileName,
		Directory: vf.FullDir,
		Height:    vf.Height,
		BitRate:   convert.FloatTo(float64(vf.BitRate) / 1e3).Decimal(0),
		Duration:  convert.FloatTo(float64(vf.Duration) / 60).Decimal(0),
		Frame:     vf.FrameAverage,
	}
	return filmInfo, nil
}

func (s *MovieService) findScInfo(ctx context.Context, movieJavId string) ([]*types.MovieScEvent, error) {
	gLists, err := s.deps.GListRepo.FindGListByMovieJavId(ctx, movieJavId)
	if err != nil {
		return nil, err
	}
	if len(gLists) == 0 {
		return nil, nil
	}

	nameSet := make(map[string]struct{}, len(gLists))
	names := make([]string, 0, len(gLists))
	for _, gl := range gLists {
		if gl == nil || gl.ScName == "" {
			continue
		}
		if _, ok := nameSet[gl.ScName]; ok {
			continue
		}
		nameSet[gl.ScName] = struct{}{}
		names = append(names, gl.ScName)
	}

	scEvents, err := s.deps.ScRepo.FindByNames(ctx, names)
	if err != nil {
		return nil, err
	}

	scMap := make(map[string]*types.GSc, len(scEvents))
	for _, ev := range scEvents {
		if ev == nil {
			continue
		}
		scMap[ev.Name] = ev
	}

	items := make([]*types.MovieScEvent, 0, len(gLists))
	seen := make(map[string]struct{}, len(gLists))
	for _, gl := range gLists {
		if gl == nil || gl.ScName == "" {
			continue
		}
		if _, ok := seen[gl.ScName]; ok {
			continue
		}
		seen[gl.ScName] = struct{}{}

		item := &types.MovieScEvent{
			ScName: gl.ScName,
			IsCome: gl.IsCome == consts.GListIsCome,
			Href:   "/sc-events/" + gl.ScName,
		}
		if ev := scMap[gl.ScName]; ev != nil {
			item.ScTime = ev.ScTime
			item.Cooldown = ev.Cooldown
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ScTime == items[j].ScTime {
			return items[i].ScName > items[j].ScName
		}
		return items[i].ScTime > items[j].ScTime
	})

	return items, nil
}
