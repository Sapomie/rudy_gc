package movie

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const movieTypeTTL = 7 * 24 * time.Hour

func (s *Service) GetMovieType(ctx context.Context, javId string) (*types.MovieType, error) {
	if s.deps.MovieTypeCache != nil {
		if v, err := s.deps.MovieTypeCache.GetMovieType(ctx, javId); err == nil && v != nil {
			return v, nil
		}
	}

	mt, err := s.buildMovieTypeFromModels(ctx, javId)
	if err != nil || mt == nil {
		return mt, err
	}

	if s.deps.MovieTypeCache != nil {
		_ = s.deps.MovieTypeCache.SetMovieType(ctx, javId, mt, movieTypeTTL)
	}
	return mt, nil
}

func (s *Service) InvalidateMovieType(ctx context.Context, javId string) {
	if s.deps.MovieTypeCache == nil || javId == "" {
		return
	}
	if err := s.deps.MovieTypeCache.DelMovieType(ctx, javId); err != nil {
		logx.WithContext(ctx).Errorf("del MovieType cache failed, javId=%s, err=%v", javId, err)
		return
	}
	logx.WithContext(ctx).Infof("del MovieType cache ok, javId=%s", javId)
}

func (s *Service) InvalidateMovieTypes(ctx context.Context, javIds ...string) {
	if s.deps.MovieTypeCache == nil || len(javIds) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(javIds))
	for _, id := range javIds {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		if err := s.deps.MovieTypeCache.DelMovieType(ctx, id); err != nil {
			logx.WithContext(ctx).Errorf("del MovieType cache failed, javId=%s, err=%v", id, err)
		} else {
			logx.WithContext(ctx).Infof("del MovieType cache ok, javId=%s", id)
		}
	}
}

func (s *Service) buildMovieTypeFromModels(ctx context.Context, javId string) (*types.MovieType, error) {
	mvRow, err := s.deps.MovieModel.FindOneByJavId(ctx, javId)
	if err != nil {
		return nil, fmt.Errorf("MovieModel.FindOneByJavId failed: %s,%w", javId, err)
	}
	mv := mapAMovieToTypes(mvRow)

	murlRow, err := s.deps.MurlModel.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("MurlModel.FindOneByJavId failed: %w", err)
	}
	murl := mapBmMurlToTypes(murlRow)

	castInfos, err := s.getCastInfos(ctx, mv.JavId, mv.ReleasingDate)
	if err != nil {
		return nil, fmt.Errorf("getCastInfos failed: %w", err)
	}
	genreNames, err := s.getGenreNames(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("getGenreNames failed: %w", err)
	}

	directorRow, err := s.deps.DirectorModel.FindOne(ctx, mv.DirectorId)
	if err != nil {
		return nil, fmt.Errorf("DirectorModel.FindOne(%d) failed: %w", mv.DirectorId, err)
	}
	prefixRow, err := s.deps.PrefixModel.FindOne(ctx, mv.PrefixId)
	if err != nil {
		return nil, fmt.Errorf("PrefixModel.FindOne(%d) failed: %w", mv.PrefixId, err)
	}
	labelRow, err := s.deps.LabelModel.FindOne(ctx, mv.LabelId)
	if err != nil {
		return nil, fmt.Errorf("LabelModel.FindOne(%d) failed: %w", mv.LabelId, err)
	}
	makerRow, err := s.deps.MakerModel.FindOne(ctx, mv.MakerId)
	if err != nil {
		return nil, fmt.Errorf("MakerModel.FindOne(%d) failed: %w", mv.MakerId, err)
	}

	minfoRow, err := s.deps.MinfoModel.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("MinfoModel.FindOneByJavId failed: %w", err)
	}
	minfo := mapBmMinfoToTypes(minfoRow)

	itemRow, err := s.deps.ItemModel.FindOneByJavId(ctx, mv.JavId)
	if err != nil {
		return nil, fmt.Errorf("ItemModel.FindOneByJavId failed: %w", err)
	}

	javURL := "https://" + s.deps.Config.Fetcher.JavAddress + "/cn/?v=" + mv.JavId
	updateDate := time.Unix(mv.DetailUpdateTime, 0).Format(time.DateOnly)

	out := &types.MovieType{
		DbId:                 mv.Id,
		Name:                 mv.Name,
		JavId:                mv.JavId,
		Title:                chooseTitle(mv.Title, minfo.Chinese, itemRow.HasChinese),
		ReleasingDate:        tsToDate(mv.ReleasingDate),
		UpdateDate:           updateDate,
		Length:               mv.Length,
		Score:                round1(float64(mv.Score) / 10.0),
		ViewersNumberOwned:   mv.ViewersNumberOwned,
		ViewersNumberWatched: mv.ViewersNumberWatched,
		ViewersNumberWant:    mv.ViewersNumberWant,
		Maker:                makerRow.Name,
		Director:             directorRow.Name,
		Label:                labelRow.Name,
		Genre:                genreNames,
		Cast:                 castInfos,
		CastNumber:           mv.CastNumber,
		JavUrl:               javURL,
		SearchUrl:            getMovieSearchUrl(mv.Name),
		BusUrl:               s.getBusUrl(mv.Name),
		JacketImg:            murl.JacketImg,
		Prefix:               prefixRow.Name,
		Owned:                consts.MovieAll,
		NeedDownload:         minfo.NeedDownload,
		EncodeName:           mv.EncodeName,
		AMovie:               mv,
		BmMinfo:              minfo,
		BmMurl:               murl,
	}

	filmRow, err := s.deps.FilmModel.FindOneByMovieJavId(ctx, mv.JavId)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("FilmModel.FindOneByMovieJavId failed: %w", err)
	}
	if filmRow != nil {
		film := mapVFilmToTypes(filmRow)
		out.VFilm = film
		out.FilmBirthDate = tsToDate(film.BirthTime)
		out.VideoUrl = film.FullDir + string(filepath.Separator) + film.FileName
		out.ScTimes = film.ScTimes
		out.ComeTimes = film.ComeTimes
		out.Owned = determineOwnership(film)
	}

	if minfo.HighestRank < 1000 {
		out.HighestRank = minfo.HighestRank
	}
	if minfo.DaysInRank > 0 {
		out.FirstRankingDay = fmt.Sprintf("%s(%v)", consts.GetDateStringByRankDayNumber(minfo.FirstRankDayNumber), minfo.DaysInRank)
	}

	if itemRow.HasDownloadCover == consts.ItemCoverOK && murl.JacketImgLocal != "" {
		out.JacketImg = s.deps.Config.Fetcher.LocalImageDir + murl.JacketImgLocal
	}
	return out, nil
}

func (s *Service) getCastInfos(ctx context.Context, movieJavId string, releasingTs int64) ([]*types.CastInfo, error) {
	castIDs, err := s.deps.MovieCastModel.ListCastIDsByMovieJavId(ctx, movieJavId)
	if err != nil {
		return nil, err
	}

	infos := make([]*types.CastInfo, 0, 9)
	seen := make(map[string]struct{}, len(castIDs))
	for _, id := range castIDs {
		if len(infos) >= 9 {
			break
		}
		castRow, err := s.deps.CastModel.FindOne(ctx, id)
		if err != nil {
			return nil, err
		}
		ci := &types.CastInfo{
			PersonId:   castRow.PersonId,
			LastScTime: castRow.LastScTime,
			Name:       castRow.Name,
		}

		displayName := castRow.Name
		birthDay := int64(0)
		if castRow.PersonId > 0 {
			personRow, err := s.deps.PersonModel.FindOne(ctx, castRow.PersonId)
			if err != nil && !errors.Is(err, moviex.ErrNotFound) {
				return nil, err
			}
			if personRow != nil {
				if strings.TrimSpace(personRow.Chinese) != "" {
					displayName = personRow.Chinese
				} else if strings.TrimSpace(personRow.Name) != "" {
					displayName = personRow.Name
				}
				birthDay = personRow.BirthDay
				ci.Url = fmt.Sprintf("cast?id=%d", castRow.PersonId)
			}
		}
		displayKey := buildMovieTypeCastDisplayKey(castRow.PersonId, castRow.Name)
		if displayKey != "" {
			if _, ok := seen[displayKey]; ok {
				continue
			}
			seen[displayKey] = struct{}{}
		}
		if ci.Url == "" {
			ci.Url = fmt.Sprintf("cards?cn=%s", url.QueryEscape(castRow.Name))
		}
		ci.DisplayName = displayName
		if birthDay > 0 && releasingTs > 0 {
			age := round1(float64(releasingTs-birthDay) / (3600.0 * 24.0 * 365.0))
			ci.NameShow = fmt.Sprintf("%s(%v)", displayName, age)
		} else {
			ci.NameShow = displayName
		}
		infos = append(infos, ci)
	}
	return infos, nil
}

func buildMovieTypeCastDisplayKey(personID int64, name string) string {
	if personID > 0 {
		return "p:" + strconv.FormatInt(personID, 10)
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}
	return "n:" + name
}

func (s *Service) getGenreNames(ctx context.Context, movieJavId string) ([]string, error) {
	genreIDs, err := s.deps.MovieGenreModel.ListGenreIDsByMovieJavId(ctx, movieJavId)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(genreIDs))
	for _, gid := range genreIDs {
		row, err := s.deps.GenreModel.FindOne(ctx, gid)
		if err != nil {
			return nil, err
		}
		names = append(names, row.Name)
	}
	return names, nil
}

func determineOwnership(film *types.Film) int64 {
	if film.IsRemoved == consts.FilmIsRemoved {
		return consts.OwnedRemoved
	}
	if film.HasSub == consts.FilmHasSub {
		return consts.OwnedHasSubNotRemoved
	}
	return consts.OwnedNoSubNotRemoved
}

func (s *Service) getBusUrl(movieName string) string {
	return fmt.Sprintf("https://%v/%v", s.deps.Config.Fetcher.BusAddress, movieName)
}

func getMovieSearchUrl(movieName string) string {
	parts := strings.Split(movieName, "-")
	if len(parts) != 2 {
		return ""
	}
	return fmt.Sprintf("https://sukebei.nyaa.si/?f=0&c=0_0&q=%v+%v", parts[0], parts[1])
}

func tsToDate(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(time.DateOnly)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func chooseTitle(title, chinese string, hasChinese int64) string {
	if hasChinese == 1 && chinese != "" {
		return chinese
	}
	if chinese != "" {
		return chinese
	}
	return title
}

func (s *Service) findRankInfo(ctx context.Context, movieJavId string) ([]*types.RankInfo, error) {
	rankRows, err := s.deps.RankModel.FindHighestRank(ctx, movieJavId, 1000)
	if err != nil {
		return nil, err
	}
	rankInfos := make([]*types.RankInfo, 0, len(rankRows))
	for _, rankRow := range rankRows {
		if rankRow == nil {
			continue
		}
		rankInfos = append(rankInfos, &types.RankInfo{
			Date: consts.GetDateStringByRankDayNumber(rankRow.DayNumber),
			Rank: rankRow.RankPos,
		})
	}
	return rankInfos, nil
}

func (s *Service) findFilmInfo(ctx context.Context, movieJavId string) (*types.FilmInfo, error) {
	vf, err := s.deps.FilmModel.FindOneByMovieJavId(ctx, movieJavId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	filmInfo := &types.FilmInfo{
		Name:      vf.MovieName,
		BirthTime: time.Unix(vf.BirthTime, 0).Format(time.DateTime),
		Size:      float64(vf.Size) / 1e9,
		FilePath:  vf.FullDir,
		FileName:  vf.FileName,
		Directory: vf.FullDir,
		Height:    vf.Height,
		BitRate:   float64(vf.BitRate) / 1e3,
		Duration:  float64(vf.Duration) / 60,
		Frame:     vf.FrameAverage,
	}
	return filmInfo, nil
}

func (s *Service) findScInfo(ctx context.Context, movieJavId string) ([]*types.MovieScEvent, error) {
	gLists, err := s.deps.GListModel.ListByMovieJavId(ctx, movieJavId)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, nil
		}
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

	scEvents, err := s.deps.ScModel.ListByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	scMap := make(map[string]*moviex.GSc, len(scEvents))
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
	hasFilm := int64(2)
	if filmInfo == nil {
		hasFilm = 1
	}
	return &types.MovieDetail{
		MovieType: movieType,
		FilmInfo:  filmInfo,
		HasFilm:   hasFilm,
		RankInfos: rankInfos,
		SC:        scInfo,
	}, nil
}
