package media

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) ListMediaPage(ctx context.Context, page, pageSize int64, orderBy string) ([]*types.MediaListItem, int64, error) {
	return s.ListMediaPageWithFilter(ctx, page, pageSize, orderBy, types.MediaListFilter{})
}

func (s *Service) ListMediaPageWithFilter(ctx context.Context, page, pageSize int64, orderBy string, filter types.MediaListFilter) ([]*types.MediaListItem, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 24
	}

	offset := (page - 1) * pageSize
	total, err := s.deps.WMediaModel.CountNativeMedia(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*types.MediaListItem{}, 0, nil
	}

	rows, err := s.deps.WMediaModel.ListNativeMediaPage(ctx, offset, pageSize, orderBy, filter)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*types.MediaListItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		item := s.buildMediaListItem(ctx, row)
		items = append(items, item)
	}
	return items, total, nil
}

func (s *Service) buildMediaListItem(ctx context.Context, row *moviex.NativeMediaListRow) *types.MediaListItem {
	item := &types.MediaListItem{
		Id:              row.Id,
		MovieName:       row.MovieName,
		MovieHref:       "/movie/" + url.PathEscape(row.MovieName),
		CastName:        "-",
		Casts:           []*types.MediaListCastItem{},
		SizeGB:          formatMediaSizeGB(row.Size),
		Height:          row.Height,
		DurationMinutes: formatMediaDurationMinutes(row.Duration),
		BitRate:         row.BitRate,
		FrameAverage:    formatMediaFrameAverage(row.FrameAverage),
		SelfMakeText:    formatMediaSelfMake(row.SelfMake),
		HasMaskText:     formatMediaHasMask(row.HasMask),
		ScTimes:         row.ScTimes,
		LastScTime:      row.LastScTime,
		BirthTime:       row.BirthTime,
		ReleasingDate:   row.ReleasingDate,
	}

	if strings.TrimSpace(row.MovieJavId) == "" {
		return item
	}

	mt, err := s.movieSvc.GetMovieType(ctx, row.MovieJavId)
	if err != nil {
		s.deps.Log.Errorf("build media list item get movie type failed, javId=%s, err=%v", row.MovieJavId, err)
		return item
	}
	if mt == nil {
		return item
	}

	if casts := buildMediaListCastItems(mt.Cast); len(casts) > 0 {
		item.Casts = casts
		item.CastName = buildMediaListCastName(casts)
	}
	item.ScTimes = mt.ScTimes
	item.LastScTime = mt.LastScTime
	item.VideoURL, item.PlayButtonClass, item.PlayButtonText, item.ShowPlayButton = buildMediaListPlayButton(mt)
	return item
}

func buildMediaListCastItems(casts []*types.CastInfo) []*types.MediaListCastItem {
	if len(casts) == 0 {
		return nil
	}
	items := make([]*types.MediaListCastItem, 0, len(casts))
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
		seenKey := name
		if cast.PersonId > 0 {
			seenKey = "id:" + strconv.FormatInt(cast.PersonId, 10)
		}
		if _, ok := seen[seenKey]; ok {
			continue
		}
		seen[seenKey] = struct{}{}

		item := &types.MediaListCastItem{
			Id:   cast.PersonId,
			Name: name,
		}
		if cast.PersonId > 0 {
			item.Href = "/cast?id=" + strconv.FormatInt(cast.PersonId, 10)
		}
		items = append(items, item)
	}
	return items
}

func buildMediaListCastName(casts []*types.MediaListCastItem) string {
	if len(casts) == 0 {
		return ""
	}
	names := make([]string, 0, len(casts))
	for _, cast := range casts {
		if cast == nil {
			continue
		}
		name := strings.TrimSpace(cast.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return strings.Join(names, " / ")
}

func buildMediaListPlayButton(mt *types.MovieType) (string, string, string, bool) {
	if mt == nil || strings.TrimSpace(mt.VideoUrlWMedia) == "" {
		return "", "", "", false
	}
	switch mt.OwnedWMedia {
	case consts.OwnedNoSubNotRemoved:
		return mt.VideoUrlWMedia, "btn-info", "Play", true
	case consts.OwnedHasSubNotRemoved:
		return mt.VideoUrlWMedia, "btn-success", "Play", true
	case consts.OwnedRemoved:
		return mt.VideoUrlWMedia, "btn-secondary", "Removed", true
	default:
		return "", "", "", false
	}
}

func formatMediaSizeGB(size int64) string {
	if size <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", float64(size)/(1024*1024*1024))
}

func formatMediaDurationMinutes(sec int64) int64 {
	if sec <= 0 {
		return 0
	}
	return int64(math.Round(float64(sec) / 60.0))
}

func formatMediaFrameAverage(v float64) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

func formatMediaSelfMake(v int64) string {
	if v == consts.FilmSelfMake {
		return "是"
	}
	return "否"
}

func formatMediaHasMask(v int64) string {
	switch v {
	case consts.FilmErased:
		return "去码"
	case consts.FilmNoMosaic:
		return "无码"
	default:
		return "有码"
	}
}
