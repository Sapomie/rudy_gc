package sc

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (s *ScService) BuildActorScSummary(ctx context.Context, actorName string, recentLimit int) (*types.ActorScSummary, error) {
	page, err := s.buildActorScPage(ctx, actorName, recentLimit, 0)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, nil
	}

	return &types.ActorScSummary{
		Actor:         page.Actor,
		RecentEvents:  page.RecentEvents,
		ActorPageHref: page.ActorPageHref,
		CardsHref:     page.CardsHref,
		ScAggHref:     page.ScAggHref,
	}, nil
}

func (s *ScService) BuildActorScPage(ctx context.Context, actorName string, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	return s.buildActorScPage(ctx, actorName, recentLimit, movieLimit)
}

func (s *ScService) buildActorScPage(ctx context.Context, actorName string, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	actorName = strings.TrimSpace(actorName)
	if actorName == "" {
		return nil, types.ErrNotFound
	}

	actor, err := s.castFindOneByName(ctx, actorName)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if actor == nil {
		return nil, types.ErrNotFound
	}

	resp, err := s.movieSvc.ListMovieFull(ctx, &types.ListMovieFullRequest{
		CastNames: actorName,
		Owned:     consts.MovieAll,
		OrderBy:   consts.OrderByReleasingDate,
		Page:      1,
		PageSize:  999999,
	})
	if err != nil {
		return nil, err
	}

	movies := make([]*types.MovieType, 0, len(resp.List))
	movieMap := make(map[string]*types.MovieType, len(resp.List))
	for _, movieType := range resp.List {
		if movieType == nil {
			continue
		}
		movies = append(movies, movieType)
		if movieType.JavId != "" {
			movieMap[movieType.JavId] = movieType
		}
	}
	sortMoviesForScDisplay(movies)

	glists, err := s.glFindByMovieJavIDs(ctx, resp.JavIds)
	if err != nil {
		return nil, err
	}

	recentEvents, err := s.buildActorScEvents(ctx, glists, movieMap, recentLimit)
	if err != nil {
		return nil, err
	}

	if movieLimit > 0 && len(movies) > movieLimit {
		movies = movies[:movieLimit]
	}

	return &types.ActorScPage{
		Actor:         actor,
		RecentEvents:  recentEvents,
		Movies:        movies,
		TotalMovies:   len(resp.List),
		ActorPageHref: buildActorPageHref(actorName),
		CardsHref:     buildActorCardsHref(actorName),
		ScAggHref:     "/sc-agg",
	}, nil
}

func (s *ScService) buildActorScEvents(ctx context.Context, glists []*types.GList, movieMap map[string]*types.MovieType, recentLimit int) ([]*types.ActorScEventItem, error) {
	if len(glists) == 0 {
		return nil, nil
	}

	type actorEventAgg struct {
		movies map[string]*types.ActorScEventMovie
		isCome bool
	}

	aggMap := make(map[string]*actorEventAgg)
	for _, row := range glists {
		if row == nil || row.ScName == "" || row.MovieJavId == "" {
			continue
		}
		agg := aggMap[row.ScName]
		if agg == nil {
			agg = &actorEventAgg{movies: make(map[string]*types.ActorScEventMovie)}
			aggMap[row.ScName] = agg
		}
		movieItem := agg.movies[row.MovieJavId]
		if movieItem == nil {
			movieItem = buildActorScEventMovie(row, movieMap[row.MovieJavId])
			agg.movies[row.MovieJavId] = movieItem
		}
		if actorMovieIsCome(row) {
			agg.isCome = true
			movieItem.IsCome = true
		}
	}
	if len(aggMap) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(aggMap))
	for name := range aggMap {
		names = append(names, name)
	}
	scEvents, err := s.scFindByNames(ctx, names)
	if err != nil {
		return nil, err
	}

	scMap := make(map[string]*types.GSc, len(scEvents))
	for _, event := range scEvents {
		if event == nil {
			continue
		}
		scMap[event.Name] = event
	}

	items := make([]*types.ActorScEventItem, 0, len(aggMap))
	for name, agg := range aggMap {
		movies := flattenActorScMovies(agg.movies)
		item := &types.ActorScEventItem{
			ScName:     name,
			MovieCount: len(movies),
			IsCome:     agg.isCome,
			Href:       buildScEventHref(name),
			Movies:     movies,
		}
		if event := scMap[name]; event != nil {
			item.ScTime = event.ScTime
			item.Cooldown = event.Cooldown
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].ScTime == items[j].ScTime {
			return items[i].ScName > items[j].ScName
		}
		return items[i].ScTime > items[j].ScTime
	})
	if recentLimit > 0 && len(items) > recentLimit {
		items = items[:recentLimit]
	}
	return items, nil
}

func sortMoviesForScDisplay(movies []*types.MovieType) {
	sort.Slice(movies, func(i, j int) bool {
		left := movies[i]
		right := movies[j]

		leftOwned := movieOwnedRank(left)
		rightOwned := movieOwnedRank(right)
		if leftOwned != rightOwned {
			return leftOwned > rightOwned
		}

		leftLastSc := movieLastScTime(left)
		rightLastSc := movieLastScTime(right)
		if leftLastSc != rightLastSc {
			return leftLastSc > rightLastSc
		}

		if left.ScTimes != right.ScTimes {
			return left.ScTimes > right.ScTimes
		}
		if left.ComeTimes != right.ComeTimes {
			return left.ComeTimes > right.ComeTimes
		}
		if left.HighestRank != right.HighestRank {
			if left.HighestRank == 0 {
				return false
			}
			if right.HighestRank == 0 {
				return true
			}
			return left.HighestRank < right.HighestRank
		}
		if left.ReleasingDate != right.ReleasingDate {
			return left.ReleasingDate > right.ReleasingDate
		}
		return left.Name < right.Name
	})
}

func movieOwnedRank(movieType *types.MovieType) int {
	if movieType == nil {
		return 0
	}
	if movieType.Owned > consts.MovieAll {
		return 1
	}
	return 0
}

func movieLastScTime(movieType *types.MovieType) int64 {
	if movieType == nil || movieType.VFilm == nil {
		return 0
	}
	return movieType.VFilm.LastScTime
}

func buildActorPageHref(name string) string {
	return "/cast?name=" + url.QueryEscape(name)
}

func buildActorCardsHref(name string) string {
	return "/cards?cn=" + url.QueryEscape(name)
}

func buildActorScEventMovie(gl *types.GList, movieType *types.MovieType) *types.ActorScEventMovie {
	item := &types.ActorScEventMovie{
		IsCome: actorMovieIsCome(gl),
	}
	if movieType != nil {
		item.Name = movieType.Name
	}
	if item.Name == "" {
		item.Name = parseActorScMovieName(gl)
	}
	if item.Name == "" && gl != nil {
		item.Name = gl.MovieJavId
	}
	if item.Name != "" {
		item.Href = buildMovieHref(item.Name)
	}
	return item
}

func flattenActorScMovies(movieMap map[string]*types.ActorScEventMovie) []*types.ActorScEventMovie {
	out := make([]*types.ActorScEventMovie, 0, len(movieMap))
	for _, movie := range movieMap {
		if movie == nil {
			continue
		}
		out = append(out, movie)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsCome != out[j].IsCome {
			return out[i].IsCome
		}
		return out[i].Name < out[j].Name
	})

	return out
}

func buildMovieHref(name string) string {
	return "/movie/" + url.PathEscape(name)
}

func parseActorScMovieName(gl *types.GList) string {
	if gl == nil || gl.Name == "" {
		return ""
	}
	parts := strings.SplitN(gl.Name, "__", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
