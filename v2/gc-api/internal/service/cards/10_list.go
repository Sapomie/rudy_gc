package cards

import (
	"context"
	"fmt"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func (s *Service) List(ctx context.Context, req *types.CardsListRequest) (*types.CardsListResponse, error) {
	normalizeRequest(req)

	ids, total, err := s.repo.ListMovieIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	items, err := s.loadMovieCards(ctx, ids)
	if err != nil {
		return nil, err
	}

	response := &types.CardsListResponse{
		View:     req.View,
		Title:    viewTitle(req.View),
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		OrderBy:  req.OrderBy,
		Order:    req.Order,
		Views:    buildViewOptions(),
		FilterSnapshot: &types.CardsFilter{
			CastNames:             req.CastNames,
			PersonIds:             req.PersonIds,
			GenreNames:            req.GenreNames,
			DirectorName:          req.DirectorName,
			PrefixName:            req.PrefixName,
			MakerName:             req.MakerName,
			LabelName:             req.LabelName,
			LabelJavID:            req.LabelJavID,
			AlbumName:             req.AlbumName,
			ReleasingDateStart:    req.ReleasingDateStart,
			ReleasingDateEnd:      req.ReleasingDateEnd,
			MediaBirthTimeStart:   req.MediaBirthTimeStart,
			MediaBirthTimeEnd:     req.MediaBirthTimeEnd,
			CastAgeMin:            req.CastAgeMin,
			CastAgeMax:            req.CastAgeMax,
			StartRankingDateStart: req.StartRankingDateStart,
			StartRankingDateEnd:   req.StartRankingDateEnd,
			DaysInRankMin:         req.DaysInRankMin,
			NeedDownload:          req.NeedDownload,
			Word:                  req.Word,
			MediaOwned:            req.MediaOwned,
			ViewWatchedMin:        req.ViewWatchedMin,
			ViewWatchedMax:        req.ViewWatchedMax,
			ScoreMin:              req.ScoreMin,
			ScoreMax:              req.ScoreMax,
			LastScTimeMin:         req.LastScTimeMin,
			LastScTimeMax:         req.LastScTimeMax,
			ScTimesMin:            req.ScTimesMin,
			ScTimesMax:            req.ScTimesMax,
			ComeTimesMin:          req.ComeTimesMin,
			ComeTimesMax:          req.ComeTimesMax,
			MediaDir1:             req.MediaDir1,
			MediaDir2:             req.MediaDir2,
			MediaDir3:             req.MediaDir3,
			MediaDir4:             req.MediaDir4,
		},
	}
	if req.View == "cardsrandom" {
		response.RandomRequested = req.RandomN
	}
	_ = viewSummary(req)
	return response, nil
}

func (s *Service) loadMovieCards(ctx context.Context, ids []string) ([]*types.MovieCard, error) {
	if len(ids) == 0 {
		return []*types.MovieCard{}, nil
	}
	items := make([]*types.MovieCard, len(ids))
	missing := make([]string, 0, len(ids))
	missingPos := make(map[string][]int, len(ids))
	for i, id := range ids {
		if s.cache != nil {
			card, err := s.cache.GetMovieCard(ctx, id)
			if err != nil {
				return nil, err
			}
			if card != nil {
				items[i] = card
				continue
			}
		}
		missing = append(missing, id)
		missingPos[id] = append(missingPos[id], i)
	}
	if len(missing) > 0 {
		cards, err := s.repo.LoadMovieCardsByIDs(ctx, missing)
		if err != nil {
			return nil, err
		}
		for _, card := range cards {
			if card == nil {
				continue
			}
			if s.cache != nil {
				if err := s.cache.SetMovieCard(ctx, card.MovieJavID, card, movieCardCacheTTL); err != nil {
					return nil, err
				}
			}
			for _, pos := range missingPos[card.MovieJavID] {
				items[pos] = card
			}
		}
	}
	out := make([]*types.MovieCard, 0, len(items))
	for _, item := range items {
		if item != nil {
			out = append(out, item)
		}
	}
	return out, nil
}

func normalizeRequest(req *types.CardsListRequest) {
	if req.View == "" {
		req.View = "cards"
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 18
	}
	if req.RandomN <= 0 {
		req.RandomN = 6
	}
	if req.OrderBy == "" {
		switch req.View {
		case "cardshasrank":
			req.OrderBy = consts.OrderByRankDate
		case "cardsmediamowned":
			req.OrderBy = consts.OrderByMediaBirthTime
		default:
			req.OrderBy = consts.OrderByReleasingDate
		}
	}
	if req.Order == "" {
		req.Order = "desc"
	}
	if req.View == "cardshasrank" && req.DaysInRankMin == nil {
		value := int64(1)
		req.DaysInRankMin = &value
	}
	if req.View == "cardsmediamowned" && req.MediaOwned == 0 {
		req.MediaOwned = consts.OwnedAllNotRemoved
	}
	if req.View == "cardsneeddownload" && req.NeedDownload == 0 {
		req.NeedDownload = consts.MovieNeedDownloadOK
	}
}

func viewTitle(view string) string {
	switch view {
	case "cardstoday":
		return "MovieCard Today"
	case "cardshasrank":
		return "MovieCard Ranked"
	case "cardsmediamowned":
		return "MovieCard Media Owned"
	case "cardsneeddownload":
		return "MovieCard Need Download"
	case "cardsrandom":
		return "MovieCard Random"
	default:
		return "MovieCard"
	}
}

func buildViewOptions() []*types.ViewOption {
	keys := []struct {
		key   string
		label string
	}{
		{key: "cards", label: "Cards"},
		{key: "cardstoday", label: "Today"},
		{key: "cardshasrank", label: "Ranked"},
		{key: "cardsmediamowned", label: "Media"},
		{key: "cardsneeddownload", label: "Need Download"},
		{key: "cardsrandom", label: "Random"},
	}
	options := make([]*types.ViewOption, 0, len(keys))
	for _, item := range keys {
		href := "/" + item.key
		if item.key == "cards" {
			href = "/cards"
		}
		options = append(options, &types.ViewOption{
			Key:   item.key,
			Label: item.label,
			Href:  href,
		})
	}
	return options
}

func viewSummary(req *types.CardsListRequest) string {
	return fmt.Sprintf("%s:%d:%d", req.View, req.Page, req.PageSize)
}
