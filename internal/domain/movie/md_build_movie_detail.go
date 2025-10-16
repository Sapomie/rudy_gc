package movie

import (
	"context"
	"errors"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/convert"
	"time"
)

func (s *Service) buildMovieDetail(ctx context.Context, m *types.Movie) (*types.MovieDetail, error) {
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

	md := &types.MovieDetail{
		MovieType: movieType,
		FilmInfo:  filmInfo,
		RankInfos: rankInfos,
		SC:        scInfo,
	}
	return md, nil
}

func (s *Service) findRankInfo(ctx context.Context, movieJavId string) ([]*types.RankInfo, error) {
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

func (s *Service) findFilmInfo(ctx context.Context, movieJavId string) (*types.FilmInfo, error) {
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

func (s *Service) findScInfo(ctx context.Context, movieJavId string) ([]string, error) {
	gLists, err := s.deps.GListRepo.FindGListByMovieJavId(ctx, movieJavId)
	if err != nil {
		return nil, err
	}

	var resp []string
	for _, gl := range gLists {
		resp = append(resp, gl.ScName)
	}

	return resp, nil
}
