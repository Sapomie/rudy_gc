package sc

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *ScService) BuildActorScPage(ctx context.Context, actorName string, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	return s.buildActorScPageByName(ctx, actorName, recentLimit, movieLimit)
}

func (s *ScService) BuildActorScPageByPersonID(ctx context.Context, personID int64, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	return s.buildActorScPageByPersonID(ctx, personID, recentLimit, movieLimit)
}

func (s *ScService) buildActorScPageByName(ctx context.Context, actorName string, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	actorName = strings.TrimSpace(actorName)
	if actorName == "" {
		return nil, types.ErrNotFound
	}

	castRow, err := s.castFindOneByName(ctx, actorName)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if castRow == nil {
		return nil, types.ErrNotFound
	}
	if castRow.PersonId > 0 {
		return s.buildActorScPageByPersonID(ctx, castRow.PersonId, recentLimit, movieLimit)
	}

	actor := personFromCast(castRow)
	req := &types.ListMovieFullRequest{
		CastNames:  actorName,
		MediaOwned: consts.MovieAll,
		OrderBy:    consts.OrderByReleasingDate,
		Page:       1,
		PageSize:   999999,
	}
	return s.buildActorScPageWithMovies(ctx, actor, req, recentLimit, movieLimit, buildActorPageHrefByName(actorName), buildActorCardsHrefByName(actorName))
}

func (s *ScService) buildActorScPageByPersonID(ctx context.Context, personID int64, recentLimit, movieLimit int) (*types.ActorScPage, error) {
	if personID <= 0 {
		return nil, types.ErrNotFound
	}
	_ = movieLimit
	actor, err := s.personFindOne(ctx, personID)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if actor == nil {
		return nil, types.ErrNotFound
	}
	recentEvents, err := s.buildActorRecentEventsFromSnapshot(ctx, personID, recentLimit)
	if err != nil {
		return nil, err
	}

	actorAliases, err := s.buildActorAliases(ctx, actor)
	if err != nil {
		return nil, err
	}

	return &types.ActorScPage{
		Actor:         actor,
		ActorAliases:  actorAliases,
		RecentEvents:  recentEvents,
		ActorPageHref: buildActorPageHrefByID(personID),
		CardsHref:     buildActorCardsHrefByID(personID),
	}, nil
}

func (s *ScService) buildActorScPageWithMovies(ctx context.Context, actor *types.Person, req *types.ListMovieFullRequest, recentLimit, movieLimit int, actorHref, cardsHref string) (*types.ActorScPage, error) {
	resp, err := s.movieSvc.ListMovieFull(ctx, req)
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

	actorAliases, err := s.buildActorAliases(ctx, actor)
	if err != nil {
		return nil, err
	}

	if movieLimit > 0 && len(movies) > movieLimit {
		movies = movies[:movieLimit]
	}

	return &types.ActorScPage{
		Actor:         actor,
		ActorAliases:  actorAliases,
		RecentEvents:  recentEvents,
		Movies:        movies,
		TotalMovies:   len(resp.List),
		ActorPageHref: actorHref,
		CardsHref:     cardsHref,
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
		if row.IsSc != consts.GListIsSc {
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
	if movieType.OwnedWMedia > consts.MovieAll {
		return 1
	}
	return 0
}

func movieLastScTime(movieType *types.MovieType) int64 {
	if movieType == nil {
		return 0
	}
	return movieType.LastScTime
}

func buildActorPageHrefByName(name string) string {
	return "/cast?name=" + url.QueryEscape(name)
}

func buildActorCardsHrefByName(name string) string {
	return "/cards?cn=" + url.QueryEscape(name)
}

func buildActorPageHrefByID(id int64) string {
	return "/cast/" + strconv.FormatInt(id, 10)
}

func buildActorCardsHrefByID(id int64) string {
	return "/cards?pid=" + strconv.FormatInt(id, 10)
}

func personFromCast(cast *types.Cast) *types.Person {
	if cast == nil {
		return nil
	}
	return &types.Person{
		Id:                cast.PersonId,
		Name:              cast.Name,
		Chinese:           cast.Chinese,
		BirthDay:          cast.BirthDay,
		Height:            cast.Height,
		MovieNumber:       cast.MovieNumber,
		OwnedMovieNumber:  cast.OwnedWMediaNumber,
		OwnedWMediaNumber: cast.OwnedWMediaNumber,
		ScTimes:           cast.ScTimes,
		ComeTimes:         cast.ComeTimes,
		LastScTime:        cast.LastScTime,
		HighestRank:       cast.HighestRank,
		RankTimes:         cast.RankTimes,
		CreatedOn:         cast.CreatedOn,
		UpdatedOn:         cast.UpdatedOn,
	}
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

func buildScEventHref(name string) string {
	return "/sc-events/" + url.PathEscape(name)
}

func (s *ScService) buildActorRecentEventsFromSnapshot(ctx context.Context, personID int64, recentLimit int) ([]*types.ActorScEventItem, error) {
	if s == nil || s.deps == nil || s.deps.CPersonScModel == nil || personID <= 0 {
		return nil, nil
	}

	limit := int64(recentLimit)
	rows, err := s.deps.CPersonScModel.ListRecentByPersonID(ctx, personID, limit)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		if err := s.deps.CPersonScModel.RebuildByPersonIDs(ctx, []int64{personID}, 0); err != nil {
			return nil, err
		}
		rows, err = s.deps.CPersonScModel.ListRecentByPersonID(ctx, personID, limit)
		if err != nil {
			return nil, err
		}
	}

	out := make([]*types.ActorScEventItem, 0, len(rows))
	for _, row := range rows {
		item, err := mapCPersonScRowToActorScEvent(row)
		if err != nil {
			return nil, err
		}
		if item == nil {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func mapCPersonScRowToActorScEvent(row *moviex.CPersonSc) (*types.ActorScEventItem, error) {
	if row == nil || strings.TrimSpace(row.ScName) == "" {
		return nil, nil
	}

	var snapshots []struct {
		JavId  string `json:"javId"`
		Name   string `json:"name"`
		IsCome bool   `json:"isCome"`
	}
	if strings.TrimSpace(row.MoviesJson) != "" {
		if err := json.Unmarshal([]byte(row.MoviesJson), &snapshots); err != nil {
			return nil, err
		}
	}

	movies := make([]*types.ActorScEventMovie, 0, len(snapshots))
	for _, snapshot := range snapshots {
		name := strings.TrimSpace(snapshot.Name)
		if name == "" {
			name = strings.TrimSpace(snapshot.JavId)
		}
		movie := &types.ActorScEventMovie{
			Name:   name,
			IsCome: snapshot.IsCome,
		}
		if movie.Name != "" {
			movie.Href = buildMovieHref(movie.Name)
		}
		movies = append(movies, movie)
	}

	return &types.ActorScEventItem{
		ScName:     row.ScName,
		ScTime:     row.ScTime,
		Cooldown:   row.Cooldown,
		MovieCount: int(row.MovieCount),
		IsCome:     row.HasCome > 0,
		Href:       buildScEventHref(row.ScName),
		Movies:     movies,
	}, nil
}

func actorMovieIsCome(gl *types.GList) bool {
	return gl != nil && gl.IsCome == consts.GListIsCome
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
