package vfilm

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

func (s *Service) ListFilmPage(ctx context.Context, page, pageSize int64, orderBy string) ([]*types.FilmListItem, int64, error) {
	return s.ListFilmPageWithFilter(ctx, page, pageSize, orderBy, types.FilmListFilter{})
}

func (s *Service) ListFilmPageWithFilter(ctx context.Context, page, pageSize int64, orderBy string, filter types.FilmListFilter) ([]*types.FilmListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 24
	}

	offset := (page - 1) * pageSize
	total, err := s.deps.FilmModel.CountAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.FilmListItem{}, 0, nil
	}

	rows, err := s.deps.FilmModel.ListPage(ctx, offset, pageSize, orderBy, filter)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*types.FilmListItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		film := mapFilmModelToTypes(row)
		item := s.buildFilmListItem(ctx, film)
		items = append(items, item)
	}
	return items, total, nil
}

func (s *Service) buildFilmListItem(ctx context.Context, film *types.Film) *types.FilmListItem {
	item := &types.FilmListItem{
		Id:              film.Id,
		MovieName:       film.MovieName,
		MovieHref:       "/movie/" + url.PathEscape(film.MovieName),
		CastName:        "-",
		SizeGB:          formatFilmSizeGB(film.Size),
		Height:          film.Height,
		DurationMinutes: formatFilmDurationMinutes(film.Duration),
		BitRate:         film.BitRate,
		FrameAverage:    formatFilmFrameAverage(film.FrameAverage),
		SelfMakeText:    formatFilmSelfMake(film.SelfMake),
		HasMaskText:     formatFilmHasMask(film.HasMask),
		ScTimes:         film.ScTimes,
		LastScTime:      film.LastScTime,
		BirthTime:       film.BirthTime,
		ReleasingDate:   film.ReleasingDate,
	}

	if film.MovieJavId == "" {
		return item
	}

	mt, err := s.movieSvc.GetMovieType(ctx, film.MovieJavId)
	if err != nil {
		s.deps.Log.Errorf("build film list item get movie type failed, javId=%s, err=%v", film.MovieJavId, err)
		return item
	}
	if mt == nil {
		return item
	}

	if castName := buildFilmListCastName(mt.Cast); castName != "" {
		item.CastName = castName
	}
	item.VideoURL, item.PlayButtonClass, item.PlayButtonText, item.ShowPlayButton = buildFilmListPlayButton(mt)
	return item
}

func buildFilmListCastName(casts []*types.CastInfo) string {
	if len(casts) == 0 {
		return ""
	}
	names := make([]string, 0, len(casts))
	seen := make(map[string]struct{}, len(casts))
	for _, cast := range casts {
		if cast == nil {
			continue
		}
		name := strings.TrimSpace(cast.DisplayName)
		if name == "" {
			name = strings.TrimSpace(cast.NameShow)
		}
		if name == "" {
			name = strings.TrimSpace(cast.Name)
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return strings.Join(names, " / ")
}

func buildFilmListPlayButton(mt *types.MovieType) (string, string, string, bool) {
	if mt == nil || strings.TrimSpace(mt.VideoUrl) == "" {
		return "", "", "", false
	}
	switch mt.Owned {
	case consts.OwnedNoSubNotRemoved:
		return mt.VideoUrl, "btn-info", "Play", true
	case consts.OwnedHasSubNotRemoved:
		return mt.VideoUrl, "btn-success", "Play", true
	case consts.OwnedRemoved:
		return mt.VideoUrl, "btn-secondary", "Removed", true
	default:
		return "", "", "", false
	}
}

func formatFilmSizeGB(size int64) string {
	if size <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", float64(size)/(1024*1024*1024))
}

func formatFilmDurationMinutes(sec int64) int64 {
	if sec <= 0 {
		return 0
	}
	return int64(math.Round(float64(sec) / 60.0))
}

func formatFilmFrameAverage(v float64) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

func formatFilmSelfMake(v int64) string {
	if v == consts.FilmSelfMake {
		return "是"
	}
	return "否"
}

func formatFilmHasMask(v int64) string {
	switch v {
	case consts.FilmErased:
		return "去码"
	case consts.FilmNoMosaic:
		return "无码"
	default:
		return "有码"
	}
}
