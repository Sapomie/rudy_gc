package sc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

const (
	eventDetailMovieFilterAll  = "all"
	eventDetailMovieFilterSc   = "sc"
	eventDetailMovieFilterNoSc = "nosc"
	eventDetailMovieFilterJAV  = "jav"
	eventDetailMovieFilterFC2  = "fc2"
)

type scEventListRelationInfo struct {
	JavMovieCount int64
	FC2MovieCount int64
	ComeMovieType string
	ComeMovieHref string
}

type scEventMovieSummary struct {
	TotalMovieCount int64
	ScMovieCount    int64
	NoScMovieCount  int64
	JavMovieCount   int64
	FC2MovieCount   int64
	ComeMovieType   string
}

func (l *ScService) ListEventPage(ctx context.Context, page, pageSize int, sortField, sortOrder string, startTime, endTime int64) (*types.ScEventListPage, error) {
	rows, total, err := l.scListPage(ctx, page, pageSize, sortField, sortOrder, startTime, endTime)
	if err != nil {
		return nil, err
	}

	scNameStats, err := l.loadScNameStats(ctx, rows)
	if err != nil {
		return nil, err
	}

	items := make([]*types.ScEventListItem, 0, len(rows))
	movieCastHrefMap, err := l.buildScEventMovieCastHrefMap(ctx, rows)
	if err != nil {
		return nil, err
	}
	relationInfoMap, err := l.buildScEventRelationInfoMap(ctx, rows)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		stat := scNameStats[name]
		relationInfo := relationInfoMap[name]
		scMovieCount := row.MovieNumber
		if stat != nil {
			scMovieCount = stat.ScMovieCount
		}
		item := &types.ScEventListItem{
			Event:           row,
			TotalMovieCount: statTotalMovieCount(stat),
			ScMovieCount:    scMovieCount,
			MovieCastHref:   movieCastHrefMap[name],
		}
		if relationInfo != nil {
			item.JavMovieCount = relationInfo.JavMovieCount
			item.FC2MovieCount = relationInfo.FC2MovieCount
			item.ComeMovieType = relationInfo.ComeMovieType
			item.ComeMovieHref = relationInfo.ComeMovieHref
		}
		items = append(items, item)
	}

	return &types.ScEventListPage{
		Items: items,
		Total: total,
	}, nil
}

func (l *ScService) buildScEventRelationInfoMap(ctx context.Context, events []*types.GSc) (map[string]*scEventListRelationInfo, error) {
	names := make([]string, 0, len(events))
	eventByName := make(map[string]*types.GSc, len(events))
	for _, event := range events {
		if event == nil || strings.TrimSpace(event.Name) == "" {
			continue
		}
		name := strings.TrimSpace(event.Name)
		names = append(names, name)
		eventByName[name] = event
	}
	rows, err := l.deps.GListModel.ListByScNames(ctx, names)
	if err != nil {
		return nil, err
	}
	fc2IDs := make([]int64, 0)
	comeByEvent := make(map[string]*moviex.GList, len(events))
	out := make(map[string]*scEventListRelationInfo, len(events))
	for _, row := range rows {
		if row == nil {
			continue
		}
		name := strings.TrimSpace(row.ScName)
		if name == "" {
			continue
		}
		info := out[name]
		if info == nil {
			info = &scEventListRelationInfo{}
			out[name] = info
		}
		switch row.MovieType {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			info.JavMovieCount++
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			info.FC2MovieCount++
		}
		if row.IsCome != consts.GListIsCome {
			continue
		}
		comeByEvent[name] = row
		info.ComeMovieType = row.MovieType
		if row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV && row.FC2MovieId > 0 {
			fc2IDs = append(fc2IDs, row.FC2MovieId)
		}
	}
	fc2Rows, err := l.deps.FC2MovieModel.ListByIDs(ctx, fc2IDs)
	if err != nil {
		return nil, err
	}
	fc2ByID := make(map[int64]*moviex.FC2Movie, len(fc2Rows))
	for _, row := range fc2Rows {
		if row != nil {
			fc2ByID[row.ID] = row
		}
	}
	for name, row := range comeByEvent {
		info := out[name]
		if info == nil {
			continue
		}
		switch row.MovieType {
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			if movie := fc2ByID[row.FC2MovieId]; movie != nil && strings.TrimSpace(movie.FC2Key) != "" {
				info.ComeMovieHref = "/fcdetail/" + url.PathEscape(strings.TrimSpace(movie.FC2Key))
			}
		case moviedomain.SehuatangThreadMovieTypeJAV:
			if event := eventByName[name]; event != nil {
				info.ComeMovieHref = buildScEventMovieHref(event.ComeMovieName)
			}
		}
	}
	return out, nil
}

func (l *ScService) buildScEventMovieCastHrefMap(ctx context.Context, rows []*types.GSc) (map[string]string, error) {
	out := make(map[string]string, len(rows))
	resolved := make(map[string]string, len(rows))
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.Name) == "" {
			continue
		}
		rawName := strings.TrimSpace(row.MovieCast)
		if rawName == "" {
			continue
		}
		baseName := normalizeScEventMovieCastName(rawName)
		if baseName == "" {
			baseName = rawName
		}
		href, ok := resolved[baseName]
		if !ok {
			var err error
			href, err = l.resolveScEventMovieCastHref(ctx, baseName)
			if err != nil {
				return nil, err
			}
			resolved[baseName] = href
		}
		out[strings.TrimSpace(row.Name)] = href
	}
	return out, nil
}

func (l *ScService) resolveScEventMovieCastHref(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}

	castRow, err := l.castFindOneByName(ctx, name)
	if err == nil && castRow != nil {
		if castRow.PersonId > 0 {
			return buildActorPageHrefByID(castRow.PersonId), nil
		}
		return buildActorPageHrefByName(castRow.Name), nil
	}
	if err != nil && !errors.Is(err, types.ErrNotFound) {
		return "", err
	}

	personRow, err := l.personFindOneByDisplayName(ctx, name)
	if err == nil && personRow != nil && personRow.Id > 0 {
		return buildActorPageHrefByID(personRow.Id), nil
	}
	if err != nil && !errors.Is(err, types.ErrNotFound) {
		return "", err
	}

	return buildActorPageHrefByName(name), nil
}

func normalizeScEventMovieCastName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if trimmed, ok := trimScEventAgeSuffix(raw, '(', ')'); ok {
		return trimmed
	}
	if trimmed, ok := trimScEventAgeSuffix(raw, '（', '）'); ok {
		return trimmed
	}
	return raw
}

func trimScEventAgeSuffix(raw string, left, right rune) (string, bool) {
	if raw == "" || !strings.HasSuffix(raw, string(right)) {
		return "", false
	}
	start := strings.LastIndex(raw, string(left))
	if start <= 0 {
		return "", false
	}
	inner := strings.TrimSpace(raw[start+len(string(left)) : len(raw)-len(string(right))])
	if inner == "" {
		return "", false
	}
	for _, ch := range inner {
		if (ch < '0' || ch > '9') && ch != '.' {
			return "", false
		}
	}
	return strings.TrimSpace(raw[:start]), true
}

func (l *ScService) loadScNameStats(ctx context.Context, rows []*types.GSc) (map[string]*moviex.GListScNameStat, error) {
	scNames := make([]string, 0, len(rows))
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.Name) == "" {
			continue
		}
		scNames = append(scNames, strings.TrimSpace(row.Name))
	}
	stats, err := l.deps.GListModel.ListScNameStats(ctx, scNames)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*moviex.GListScNameStat, len(stats))
	for _, stat := range stats {
		if stat == nil || strings.TrimSpace(stat.ScName) == "" {
			continue
		}
		out[strings.TrimSpace(stat.ScName)] = stat
	}
	return out, nil
}

func statTotalMovieCount(stat *moviex.GListScNameStat) int64 {
	if stat == nil {
		return 0
	}
	return stat.TotalMovieCount
}

func (l *ScService) ListEventCardPage(ctx context.Context, page, pageSize int, sortField, sortOrder string, startTime, endTime int64) (*types.ScEventCardPage, error) {
	rows, total, err := l.scListPage(ctx, page, pageSize, sortField, sortOrder, startTime, endTime)
	if err != nil {
		return nil, err
	}

	items := make([]*types.ScEventCardItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, l.buildScEventCardItem(ctx, row))
	}

	return &types.ScEventCardPage{
		Items: items,
		Total: total,
	}, nil
}

func (l *ScService) GetEventDetail(ctx context.Context, name, movieFilter string) (*types.ScEventDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, types.ErrNotFound
	}
	movieFilter = normalizeEventDetailMovieFilter(movieFilter)

	event, err := l.scFindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}

	allRows, err := l.glFindAllByScName(ctx, name)
	if err != nil {
		return nil, err
	}
	rows := allRows
	switch movieFilter {
	case eventDetailMovieFilterNoSc:
		rows = filterEventDetailMoviesByIsSc(allRows, false)
	case eventDetailMovieFilterSc:
		rows = filterEventDetailMoviesByIsSc(allRows, true)
	case eventDetailMovieFilterJAV:
		rows = filterEventDetailMoviesByMovieType(allRows, moviedomain.SehuatangThreadMovieTypeJAV)
	case eventDetailMovieFilterFC2:
		rows = filterEventDetailMoviesByMovieType(allRows, moviedomain.SehuatangThreadMovieTypeFC2PPV)
	}
	summary := summarizeScEventMovies(allRows)

	fc2MovieMap, err := l.loadFC2ScEventMovieMap(ctx, allRows)
	if err != nil {
		return nil, err
	}
	editComeGListID, editComeMovieOptions, editCurrentMovieCastNames := l.buildScEventEditMovieOptions(ctx, event, allRows, fc2MovieMap)
	editImageName := ""
	if event != nil && strings.TrimSpace(event.ImagePath) != "" {
		editImageName = filepath.Base(strings.TrimSpace(event.ImagePath))
	}

	items := make([]*types.ScEventMovie, 0, len(rows))
	failedItems := make([]*types.ScEventMovie, 0)
	var comeCount int64
	for _, row := range rows {
		item := &types.ScEventMovie{
			Entry:     row,
			MovieName: parseGListMovieName(row),
			IsCome:    row != nil && row.IsCome == consts.GListIsCome,
		}
		if item.IsCome {
			comeCount++
		}

		if row != nil && row.MovieType == moviedomain.SehuatangThreadMovieTypeJAV && row.MovieJavId != "" {
			mt, err := l.movieSvc.GetMovieType(ctx, row.MovieJavId)
			if err != nil {
				item.LoadError = err.Error()
			} else {
				item.MovieType = mt
				if mt != nil {
					if mt.Name != "" {
						item.MovieName = mt.Name
					}
					item.CastNames = collectCastNames(mt)
					item.GenreNames = append(item.GenreNames, mt.Genre...)
				}
			}
		} else if row != nil && row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV && row.FC2MovieId > 0 {
			item.FC2Movie = fc2MovieMap[row.FC2MovieId]
			if item.FC2Movie == nil {
				item.LoadError = fmt.Sprintf("FC2影片不存在: %d", row.FC2MovieId)
			} else {
				item.MovieName = item.FC2Movie.FC2Key
			}
		}

		items = append(items, item)
		if item.LoadError != "" {
			failedItems = append(failedItems, item)
		}
	}

	comeMovieHref := ""
	for _, option := range editComeMovieOptions {
		if option != nil && option.GListId == editComeGListID {
			comeMovieHref = option.MovieHref
			break
		}
	}
	return &types.ScEventDetail{
		Event:                     event,
		Items:                     items,
		FailedItems:               failedItems,
		ComeCount:                 comeCount,
		ComeMovieHref:             comeMovieHref,
		ComeMovieType:             summary.ComeMovieType,
		TotalMovieCount:           summary.TotalMovieCount,
		ScMovieCount:              summary.ScMovieCount,
		NoScMovieCount:            summary.NoScMovieCount,
		JavMovieCount:             summary.JavMovieCount,
		FC2MovieCount:             summary.FC2MovieCount,
		EditImageName:             editImageName,
		EditMovieCount:            int64(len(allRows)),
		EditScMovieCount:          summary.ScMovieCount,
		EditComeGListId:           editComeGListID,
		EditComeMovieOptions:      editComeMovieOptions,
		EditCurrentMovieCastNames: editCurrentMovieCastNames,
	}, nil
}

func (l *ScService) buildScEventEditMovieOptions(ctx context.Context, event *types.GSc, rows []*types.GList, fc2Movies map[int64]*types.FC2ScEventMovie) (int64, []*types.ScEventEditMovieOption, []string) {
	if len(rows) == 0 {
		return 0, nil, nil
	}

	currentComeGListID := int64(0)
	currentComeMovieName := ""
	if event != nil {
		currentComeMovieName = strings.TrimSpace(event.ComeMovieName)
	}

	options := make([]*types.ScEventEditMovieOption, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.IsCome == consts.GListIsCome && currentComeGListID == 0 {
			currentComeGListID = row.Id
		}
		if row.IsSc != consts.GListIsSc {
			continue
		}

		movieName := strings.TrimSpace(parseGListMovieName(row))
		movieHref := ""
		movieKey := ""
		castOptions := []string(nil)
		switch row.MovieType {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			movieKey = strings.TrimSpace(row.MovieJavId)
			if movieRow, movieErr := l.movieFindOneByJavID(ctx, row.MovieJavId); movieErr == nil && movieRow != nil && strings.TrimSpace(movieRow.Name) != "" {
				movieName = strings.TrimSpace(movieRow.Name)
			}
			castOptions, _ = l.listMovieCastDisplayNames(ctx, row.MovieJavId)
			movieHref = buildScEventMovieHref(movieName)
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			if fc2Movie := fc2Movies[row.FC2MovieId]; fc2Movie != nil {
				movieKey = fc2Movie.FC2Key
				movieName = fc2Movie.Title
				if movieName == "" {
					movieName = fc2Movie.FC2Key
				}
				movieHref = "/fcdetail/" + url.PathEscape(fc2Movie.FC2Key)
			}
		}
		if movieName == "" {
			movieName = movieKey
		}

		options = append(options, &types.ScEventEditMovieOption{
			GListId:     row.Id,
			MovieType:   row.MovieType,
			MovieJavId:  strings.TrimSpace(row.MovieJavId),
			FC2MovieId:  row.FC2MovieId,
			MovieKey:    movieKey,
			MovieName:   movieName,
			MovieHref:   movieHref,
			CastOptions: castOptions,
		})
	}

	if currentComeGListID == 0 && currentComeMovieName != "" {
		for _, option := range options {
			if option == nil {
				continue
			}
			if strings.TrimSpace(option.MovieName) == currentComeMovieName {
				currentComeGListID = option.GListId
				break
			}
		}
	}

	currentMovieCastNames := []string(nil)
	for _, option := range options {
		if option == nil {
			continue
		}
		option.IsCome = option.GListId == currentComeGListID
		if option.IsCome {
			currentMovieCastNames = append(currentMovieCastNames, option.CastOptions...)
		}
	}

	return currentComeGListID, options, currentMovieCastNames
}

func (l *ScService) loadFC2ScEventMovieMap(ctx context.Context, rows []*types.GList) (map[int64]*types.FC2ScEventMovie, error) {
	movieIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV && row.FC2MovieId > 0 {
			movieIDs = append(movieIDs, row.FC2MovieId)
		}
	}
	movieIDs = uniqueFC2MovieIDs(movieIDs)
	if len(movieIDs) == 0 {
		return map[int64]*types.FC2ScEventMovie{}, nil
	}
	movies, err := l.deps.FC2MovieModel.ListByIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}
	medias, err := l.deps.FC2MediaModel.ListByMovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}
	stats, err := l.deps.FC2ScStatModel.ListByMovieIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}
	mediaByMovieID := make(map[int64]*moviex.FC2Media, len(medias))
	for _, row := range medias {
		if row != nil {
			mediaByMovieID[row.FC2MovieID] = row
		}
	}
	statByMovieID := make(map[int64]*moviex.FC2ScStat, len(stats))
	for _, row := range stats {
		if row != nil {
			statByMovieID[row.FC2MovieID] = row
		}
	}
	out := make(map[int64]*types.FC2ScEventMovie, len(movies))
	for _, movie := range movies {
		if movie == nil {
			continue
		}
		item := &types.FC2ScEventMovie{
			ID: movie.ID, FC2Key: movie.FC2Key, Title: movie.Title,
			PerformerText: movie.PerformerText, PostTime: movie.PostTime,
		}
		if media := mediaByMovieID[movie.ID]; media != nil {
			item.DownloadTime = media.BirthTime
			if strings.TrimSpace(media.FullDir) != "" && strings.TrimSpace(media.FileName) != "" {
				item.VideoURL = filepath.Join(media.FullDir, media.FileName)
				item.ShowPlayButton = true
				item.PlayButtonText = "Play"
				switch {
				case media.IsRemoved == int64(consts.FilmIsRemoved):
					item.PlayButtonClass = "btn-secondary"
					item.PlayButtonText = "Removed"
				case media.HasSub == int64(consts.FilmHasSub):
					item.PlayButtonClass = "btn-success"
				default:
					item.PlayButtonClass = "btn-info"
				}
			}
		}
		if stat := statByMovieID[movie.ID]; stat != nil {
			item.ScTimes = stat.ScTimes
			item.ComeTimes = stat.ComeTimes
			item.LastScTime = stat.LastScTime
		}
		out[movie.ID] = item
	}
	return out, nil
}

func (l *ScService) listMovieCastDisplayNames(ctx context.Context, movieJavID string) ([]string, error) {
	releasingTs := int64(0)
	movieRow, err := l.movieFindOneByJavID(ctx, movieJavID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}
	if movieRow != nil {
		releasingTs = movieRow.ReleasingDate
	}

	castIDs, err := l.movieCastListCastIDsByMovieJavID(ctx, movieJavID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(castIDs))
	out := make([]string, 0, len(castIDs))
	for _, castID := range castIDs {
		castRow, err := l.castFindOne(ctx, castID)
		if err != nil {
			if errors.Is(err, types.ErrNotFound) {
				continue
			}
			return nil, err
		}

		displayName := strings.TrimSpace(castRow.Name)
		birthDay := int64(0)
		if castRow.PersonId > 0 {
			personRow, err := l.personFindOne(ctx, castRow.PersonId)
			if err != nil && !errors.Is(err, types.ErrNotFound) {
				return nil, err
			}
			if personRow != nil {
				if strings.TrimSpace(personRow.Chinese) != "" {
					displayName = strings.TrimSpace(personRow.Chinese)
				} else if strings.TrimSpace(personRow.Name) != "" {
					displayName = strings.TrimSpace(personRow.Name)
				}
				birthDay = personRow.BirthDay
			}
		}
		if displayName == "" {
			continue
		}
		displayKey := buildScEventCastDisplayKey(castRow.PersonId, castRow.Name)
		if displayKey != "" {
			if _, ok := seen[displayKey]; ok {
				continue
			}
			seen[displayKey] = struct{}{}
		}

		nameShow := displayName
		if birthDay > 0 && releasingTs > 0 {
			age := roundScEventCastAge(float64(releasingTs-birthDay) / (3600.0 * 24.0 * 365.0))
			nameShow = fmt.Sprintf("%s(%v)", displayName, age)
		}
		out = append(out, nameShow)
	}
	return out, nil
}

func buildScEventCastDisplayKey(personID int64, name string) string {
	if personID > 0 {
		return "p:" + strconv.FormatInt(personID, 10)
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ""
	}
	return "n:" + name
}

func roundScEventCastAge(v float64) float64 {
	return math.Round(v*10) / 10
}

func normalizeEventDetailMovieFilter(v string) string {
	switch strings.TrimSpace(v) {
	case eventDetailMovieFilterAll:
		return eventDetailMovieFilterAll
	case eventDetailMovieFilterNoSc:
		return eventDetailMovieFilterNoSc
	case eventDetailMovieFilterJAV:
		return eventDetailMovieFilterJAV
	case eventDetailMovieFilterFC2:
		return eventDetailMovieFilterFC2
	default:
		return eventDetailMovieFilterSc
	}
}

func filterEventDetailMoviesByMovieType(rows []*types.GList, movieType string) []*types.GList {
	if len(rows) == 0 {
		return rows
	}
	movieType = strings.TrimSpace(movieType)
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.MovieType == movieType {
			out = append(out, row)
		}
	}
	return out
}

func summarizeScEventMovies(rows []*types.GList) scEventMovieSummary {
	var summary scEventMovieSummary
	for _, row := range rows {
		if row == nil {
			continue
		}
		summary.TotalMovieCount++
		switch row.IsSc {
		case consts.GListIsSc:
			summary.ScMovieCount++
		case consts.GListIsNotSc:
			summary.NoScMovieCount++
		}
		switch row.MovieType {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			summary.JavMovieCount++
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			summary.FC2MovieCount++
		}
		if row.IsCome == consts.GListIsCome {
			summary.ComeMovieType = row.MovieType
		}
	}
	return summary
}

func filterEventDetailMoviesByIsSc(rows []*types.GList, wantSc bool) []*types.GList {
	if len(rows) == 0 {
		return rows
	}
	out := make([]*types.GList, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if wantSc && row.IsSc == consts.GListIsSc {
			out = append(out, row)
			continue
		}
		if !wantSc && row.IsSc == consts.GListIsNotSc {
			out = append(out, row)
		}
	}
	return out
}

func collectCastNames(mt *types.MovieType) []string {
	if mt == nil || len(mt.Cast) == 0 {
		return nil
	}

	out := make([]string, 0, len(mt.Cast))
	for _, c := range mt.Cast {
		if c == nil {
			continue
		}
		name := strings.TrimSpace(c.NameShow)
		if name == "" {
			name = strings.TrimSpace(c.Name)
		}
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	return out
}

func parseGListMovieName(gl *types.GList) string {
	if gl == nil {
		return ""
	}
	parts := strings.SplitN(gl.Name, "__", 2)
	if len(parts) != 2 {
		return gl.MovieJavId
	}
	return parts[1]
}

func (l *ScService) buildScEventCardItem(ctx context.Context, event *types.GSc) *types.ScEventCardItem {
	item := &types.ScEventCardItem{
		Event:      event,
		DetailHref: buildScEventDetailHref(event),
	}
	if event == nil {
		return item
	}

	item.ComeMovieName = strings.TrimSpace(event.ComeMovieName)
	if item.ComeMovieName != "" {
		item.ComeMovieHref = buildScEventMovieHref(item.ComeMovieName)
	}

	rows, err := l.glFindByScName(ctx, event.Name)
	if err != nil {
		return item
	}

	for _, row := range rows {
		if row == nil || row.IsCome != consts.GListIsCome {
			continue
		}

		if row.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV && row.FC2MovieId > 0 {
			movieRow, findErr := l.deps.FC2MovieModel.FindOne(ctx, row.FC2MovieId)
			if findErr == nil && movieRow != nil {
				item.ComeMovieName = strings.TrimSpace(movieRow.Title)
				if item.ComeMovieName == "" {
					item.ComeMovieName = strings.TrimSpace(movieRow.FC2Key)
				}
				item.ComeMovieHref = "/fcdetail/" + url.PathEscape(strings.TrimSpace(movieRow.FC2Key))
			}
			break
		}

		if item.ComeMovieName == "" {
			item.ComeMovieName = parseGListMovieName(row)
			if item.ComeMovieName != "" {
				item.ComeMovieHref = buildScEventMovieHref(item.ComeMovieName)
			}
		}

		if row.MovieJavId == "" {
			break
		}

		mt, err := l.movieSvc.GetMovieType(ctx, row.MovieJavId)
		if err != nil || mt == nil {
			break
		}

		if item.ComeMovieName == "" && strings.TrimSpace(mt.Name) != "" {
			item.ComeMovieName = mt.Name
			item.ComeMovieHref = buildScEventMovieHref(item.ComeMovieName)
		}
		item.ComeMovieJacketImg = strings.TrimSpace(mt.JacketImg)
		item.HasComeMovieCover = item.ComeMovieJacketImg != ""
		break
	}

	return item
}

func buildScEventDetailHref(event *types.GSc) string {
	if event == nil || strings.TrimSpace(event.Name) == "" {
		return "/scevents"
	}
	return "/scevent/" + url.PathEscape(strings.TrimSpace(event.Name))
}

func buildScEventMovieHref(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "/moviedetail/" + url.PathEscape(name)
}
