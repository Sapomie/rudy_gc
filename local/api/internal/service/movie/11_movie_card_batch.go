package movie

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

func (s *Service) getMovieTypesByJavIDs(ctx context.Context, javIDs []string) ([]*types.MovieType, error) {
	if len(javIDs) == 0 {
		return []*types.MovieType{}, nil
	}
	if s.deps.MovieCardBatchRepo == nil {
		out := make([]*types.MovieType, 0, len(javIDs))
		for _, javID := range javIDs {
			item, err := s.GetMovieType(ctx, javID)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		}
		return out, nil
	}

	byJavID := make(map[string]*types.MovieType, len(javIDs))
	missing := make([]string, 0, len(javIDs))
	missingSeen := make(map[string]struct{}, len(javIDs))
	for _, javID := range javIDs {
		if s.deps.MovieTypeCache != nil {
			if item, err := s.deps.MovieTypeCache.GetMovieType(ctx, javID); err == nil && item != nil {
				byJavID[javID] = item
				continue
			}
		}
		if _, ok := missingSeen[javID]; ok {
			continue
		}
		missingSeen[javID] = struct{}{}
		missing = append(missing, javID)
	}

	if len(missing) > 0 {
		data, err := s.deps.MovieCardBatchRepo.Load(ctx, missing)
		if err != nil {
			return nil, err
		}
		for _, javID := range missing {
			item, err := s.buildMovieTypeFromBatch(javID, data)
			if err != nil {
				return nil, err
			}
			byJavID[javID] = item
			if s.deps.MovieTypeCache != nil {
				_ = s.deps.MovieTypeCache.SetMovieType(ctx, javID, item, movieTypeTTL)
			}
		}
	}

	out := make([]*types.MovieType, 0, len(javIDs))
	for _, javID := range javIDs {
		item := byJavID[javID]
		if item == nil {
			return nil, fmt.Errorf("MovieCardBatchRepo.Load missing movie: %s", javID)
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *Service) buildMovieTypeFromBatch(javID string, data *moviex.MovieCardBatchData) (*types.MovieType, error) {
	mvRow := data.Movies[javID]
	if mvRow == nil {
		return nil, fmt.Errorf("MovieModel.FindOneByJavId failed: %s,%w", javID, moviex.ErrNotFound)
	}
	murlRow := data.Murls[javID]
	if murlRow == nil {
		return nil, fmt.Errorf("MurlModel.FindOneByJavId failed: %w", moviex.ErrNotFound)
	}
	minfoRow := data.Minfos[javID]
	if minfoRow == nil {
		return nil, fmt.Errorf("MinfoModel.FindOneByJavId failed: %w", moviex.ErrNotFound)
	}
	itemRow := data.Items[javID]
	if itemRow == nil {
		return nil, fmt.Errorf("ItemModel.FindOneByJavId failed: %w", moviex.ErrNotFound)
	}
	directorRow := data.Directors[mvRow.DirectorId]
	if directorRow == nil {
		return nil, fmt.Errorf("DirectorModel.FindOne(%d) failed: %w", mvRow.DirectorId, moviex.ErrNotFound)
	}
	prefixRow := data.Prefixes[mvRow.PrefixId]
	if prefixRow == nil {
		return nil, fmt.Errorf("PrefixModel.FindOne(%d) failed: %w", mvRow.PrefixId, moviex.ErrNotFound)
	}
	labelRow := data.Labels[mvRow.LabelId]
	if labelRow == nil {
		return nil, fmt.Errorf("LabelModel.FindOne(%d) failed: %w", mvRow.LabelId, moviex.ErrNotFound)
	}
	makerRow := data.Makers[mvRow.MakerId]
	if makerRow == nil {
		return nil, fmt.Errorf("MakerModel.FindOne(%d) failed: %w", mvRow.MakerId, moviex.ErrNotFound)
	}

	mv := mapAMovieToTypes(mvRow)
	murl := mapBmMurlToTypes(murlRow)
	minfo := mapBmMinfoToTypes(minfoRow)
	castInfos := buildMovieCardBatchCastInfos(data.Casts[javID], mv.ReleasingDate)
	genreNames := make([]string, 0, len(data.Genres[javID]))
	for _, row := range data.Genres[javID] {
		if row != nil {
			genreNames = append(genreNames, row.GenreName)
		}
	}

	needDownload := int64(consts.MovieNeedDownLoadNone)
	if _, ok := data.NeedDownload[javID]; ok {
		needDownload = consts.MovieNeedDownLoadOK
	}
	javURL := "https://" + s.deps.Config.Fetcher.JavAddress + "/cn/?v=" + mv.JavId
	out := &types.MovieType{
		DbId:                 mv.Id,
		Name:                 mv.Name,
		JavId:                mv.JavId,
		Title:                chooseTitle(mv.Title, minfo.Chinese, itemRow.HasChinese),
		ReleasingDate:        tsToDate(mv.ReleasingDate),
		UpdateDate:           time.Unix(mv.DetailUpdateTime, 0).Format(time.DateOnly),
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
		SearchUrl:            s.getMovieSearchUrl(mv.Name),
		BusUrl:               s.getBusUrl(mv.Name),
		JacketImg:            murl.JacketImg,
		Prefix:               prefixRow.Name,
		Owned:                consts.MovieAll,
		NeedDownload:         needDownload,
		EncodeName:           mv.EncodeName,
		AMovie:               mv,
		BmMinfo:              minfo,
		BmMurl:               murl,
	}

	if scStatRow := data.ScStats[javID]; scStatRow != nil {
		out.ScTimes = scStatRow.ScTimes
		out.ComeTimes = scStatRow.ComeTimes
		out.LastSelectedScTime = scStatRow.LastSelectedScTime
	}
	if mediaRow := data.Media[javID]; mediaRow != nil {
		media := mapWMediaToTypes(mediaRow)
		out.WMedia = media
		out.OwnedWMedia = determineOwnershipByState(media.IsRemoved, media.HasSub)
		out.Owned = out.OwnedWMedia
		if media.IsRemoved != consts.FilmIsRemoved {
			out.FilmBirthDateWMedia = tsToDate(media.BirthTime)
			out.VideoUrlWMedia = media.FullDir + string(filepath.Separator) + media.FileName
			out.FilmBirthDate = out.FilmBirthDateWMedia
			out.VideoUrl = out.VideoUrlWMedia
		}
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

func buildMovieCardBatchCastInfos(rows []*moviex.MovieCardBatchCastRow, releasingTs int64) []*types.CastInfo {
	infos := make([]*types.CastInfo, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row == nil || len(infos) >= 9 {
			continue
		}
		displayName := row.CastName
		birthDay := int64(0)
		ci := &types.CastInfo{
			PersonId:               row.PersonID,
			LastSelectedScTime:     row.LastSelectedScTime,
			LastParticipatedScTime: row.LastParticipatedScTime,
			Name:                   row.CastName,
		}
		if row.PersonExists > 0 {
			if row.PersonChinese != "" {
				displayName = row.PersonChinese
			} else if row.PersonName != "" {
				displayName = row.PersonName
			}
			birthDay = row.PersonBirthDay
			ci.Url = fmt.Sprintf("/castdetail/%d", row.PersonID)
		}
		displayKey := buildMovieTypeCastDisplayKey(row.PersonID, row.CastName)
		if displayKey != "" {
			if _, ok := seen[displayKey]; ok {
				continue
			}
			seen[displayKey] = struct{}{}
		}
		if ci.Url == "" {
			ci.Url = fmt.Sprintf("cards?cn=%s", url.QueryEscape(row.CastName))
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
	return infos
}
