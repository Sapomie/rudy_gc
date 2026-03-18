package sc

import (
	"context"
	"net/url"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (l *ScService) ListEventPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) (*types.ScEventListPage, error) {
	rows, total, err := l.deps.ScRepo.ListPage(ctx, page, pageSize, sortField, sortOrder)
	if err != nil {
		return nil, err
	}

	items := make([]*types.ScEventListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &types.ScEventListItem{
			Event: row,
		})
	}

	return &types.ScEventListPage{
		Items: items,
		Total: total,
	}, nil
}

func (l *ScService) ListEventCardPage(ctx context.Context, page, pageSize int, sortField, sortOrder string) (*types.ScEventCardPage, error) {
	rows, total, err := l.deps.ScRepo.ListPage(ctx, page, pageSize, sortField, sortOrder)
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

func (l *ScService) GetEventDetail(ctx context.Context, name string) (*types.ScEventDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, types.ErrNotFound
	}

	event, err := l.deps.ScRepo.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}

	rows, err := l.deps.GListRepo.FindByScName(ctx, name)
	if err != nil {
		return nil, err
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
		Event:       event,
		Items:       items,
		FailedItems: failedItems,
		ComeCount:   comeCount,
	}, nil
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

	rows, err := l.deps.GListRepo.FindByScName(ctx, event.Name)
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
