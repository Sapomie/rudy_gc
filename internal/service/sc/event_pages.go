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
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

const (
	eventDetailMovieFilterAll  = "all"
	eventDetailMovieFilterSc   = "sc"
	eventDetailMovieFilterNoSc = "nosc"
)

func (l *ScService) ListEventPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) (*types.ScEventListPage, error) {
	rows, total, err := l.scListPage(ctx, page, pageSize, sortField, sortOrder)
	if err != nil {
		return nil, err
	}

	scNameStats, err := l.loadScNameStats(ctx, rows)
	if err != nil {
		return nil, err
	}

	items := make([]*types.ScEventListItem, 0, len(rows))
	for _, row := range rows {
		stat := scNameStats[strings.TrimSpace(row.Name)]
		scMovieCount := row.MovieNumber
		if stat != nil {
			scMovieCount = stat.ScMovieCount
		}
		items = append(items, &types.ScEventListItem{
			Event:           row,
			TotalMovieCount: statTotalMovieCount(stat),
			ScMovieCount:    scMovieCount,
		})
	}

	return &types.ScEventListPage{
		Items: items,
		Total: total,
	}, nil
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

func (l *ScService) ListEventCardPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) (*types.ScEventCardPage, error) {
	rows, total, err := l.scListPage(ctx, page, pageSize, sortField, sortOrder)
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
	if movieFilter == eventDetailMovieFilterNoSc {
		rows = filterEventDetailMoviesByIsSc(allRows, false)
	} else if movieFilter == eventDetailMovieFilterSc {
		rows = filterEventDetailMoviesByIsSc(allRows, true)
	}

	editComeMovieJavID, editComeMovieOptions, editCurrentMovieCastNames := l.buildScEventEditMovieOptions(ctx, event, allRows)
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

		if row != nil && row.MovieJavId != "" {
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
		}

		items = append(items, item)
		if item.LoadError != "" {
			failedItems = append(failedItems, item)
		}
	}

	return &types.ScEventDetail{
		Event:                     event,
		Items:                     items,
		FailedItems:               failedItems,
		ComeCount:                 comeCount,
		EditImageName:             editImageName,
		EditMovieCount:            int64(len(allRows)),
		EditScMovieCount:          int64(len(editComeMovieOptions)),
		EditComeMovieJavId:        editComeMovieJavID,
		EditComeMovieOptions:      editComeMovieOptions,
		EditCurrentMovieCastNames: editCurrentMovieCastNames,
	}, nil
}

func (l *ScService) buildScEventEditMovieOptions(ctx context.Context, event *types.GSc, rows []*types.GList) (string, []*types.ScEventEditMovieOption, []string) {
	if len(rows) == 0 {
		return "", nil, nil
	}

	currentComeMovieJavID := ""
	currentComeMovieName := ""
	if event != nil {
		currentComeMovieName = strings.TrimSpace(event.ComeMovieName)
	}

	seen := make(map[string]struct{}, len(rows))
	options := make([]*types.ScEventEditMovieOption, 0, len(rows))
	for _, row := range rows {
		if row == nil || strings.TrimSpace(row.MovieJavId) == "" {
			continue
		}
		if row.IsCome == consts.GListIsCome && currentComeMovieJavID == "" {
			currentComeMovieJavID = strings.TrimSpace(row.MovieJavId)
		}
		if row.IsSc != consts.GListIsSc {
			continue
		}
		if _, ok := seen[row.MovieJavId]; ok {
			continue
		}
		seen[row.MovieJavId] = struct{}{}

		movieName := strings.TrimSpace(parseGListMovieName(row))
		if movieRow, movieErr := l.movieFindOneByJavID(ctx, row.MovieJavId); movieErr == nil && movieRow != nil && strings.TrimSpace(movieRow.Name) != "" {
			movieName = strings.TrimSpace(movieRow.Name)
		}
		castOptions, castErr := l.listMovieCastDisplayNames(ctx, row.MovieJavId)
		if castErr != nil {
			castOptions = nil
		}
		if movieName == "" {
			movieName = strings.TrimSpace(row.MovieJavId)
		}

		options = append(options, &types.ScEventEditMovieOption{
			MovieJavId:  strings.TrimSpace(row.MovieJavId),
			MovieName:   movieName,
			CastOptions: castOptions,
		})
	}

	if currentComeMovieJavID == "" && currentComeMovieName != "" {
		for _, option := range options {
			if option == nil {
				continue
			}
			if strings.TrimSpace(option.MovieName) == currentComeMovieName {
				currentComeMovieJavID = option.MovieJavId
				break
			}
		}
	}

	currentMovieCastNames := []string(nil)
	for _, option := range options {
		if option == nil {
			continue
		}
		option.IsCome = option.MovieJavId == currentComeMovieJavID
		if option.IsCome {
			currentMovieCastNames = append(currentMovieCastNames, option.CastOptions...)
		}
	}

	return currentComeMovieJavID, options, currentMovieCastNames
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
	default:
		return eventDetailMovieFilterSc
	}
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
		return "/sc-events"
	}
	return "/sc-events/" + url.PathEscape(strings.TrimSpace(event.Name))
}

func buildScEventMovieHref(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	return "/movie/" + url.PathEscape(name)
}
