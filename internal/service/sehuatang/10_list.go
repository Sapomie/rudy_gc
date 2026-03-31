package sehuatang

import (
	"context"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

type ListRequest struct {
	Page       int64
	PageSize   int64
	Sort       string
	Order      string
	Keyword    string
	MovieJavID string
	InfoHash   string
}

type ListRow struct {
	Id            int64
	MovieJavID    string
	MovieName     string
	MovieRouteKey string
	ThreadTitle   string
	ThreadURL     string
	PostTime      int64
	PostDate      int64
	InfoHash      string
	LastSeenTime  int64
	CreatedOn     int64
	UpdatedOn     int64
	CanFavorite   bool
	IsInDownload  bool
	IsInPending   bool
}

type ListResult struct {
	Page     int64
	PageSize int64
	Total    int64
	Items    []*ListRow
}

func (s *Service) ListPage(ctx context.Context, req ListRequest) (*ListResult, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	filter := moviex.SehuatangMagnetListFilter{
		Keyword:    strings.TrimSpace(req.Keyword),
		MovieJavID: strings.TrimSpace(req.MovieJavID),
		InfoHash:   strings.TrimSpace(req.InfoHash),
	}

	total, err := s.deps.SehuatangMagnetModel.CountAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	offset := (req.Page - 1) * req.PageSize
	rows, err := s.deps.SehuatangMagnetModel.ListPage(ctx, offset, req.PageSize, buildOrderBy(req.Sort, req.Order), filter)
	if err != nil {
		return nil, err
	}

	items := make([]*ListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		routeKey := strings.TrimSpace(row.MovieJavId)
		if routeKey == "" {
			routeKey = strings.TrimSpace(row.MovieName)
		}
		if routeKey == "" {
			routeKey = strings.TrimSpace(row.InfoHash)
		}
		items = append(items, &ListRow{
			Id:            row.Id,
			MovieJavID:    row.MovieJavId,
			MovieName:     row.MovieName,
			MovieRouteKey: routeKey,
			ThreadTitle:   row.ThreadTitle,
			ThreadURL:     row.ThreadUrl,
			PostTime:      row.PostTime,
			PostDate:      row.PostDate,
			InfoHash:      row.InfoHash,
			LastSeenTime:  row.LastSeenTime,
			CreatedOn:     row.CreatedOn,
			UpdatedOn:     row.UpdatedOn,
		})
	}

	if err := s.fillAlbumStatus(ctx, items); err != nil {
		return nil, err
	}

	return &ListResult{
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
		Items:    items,
	}, nil
}

func buildOrderBy(sort, order string) string {
	field := normalizeSortField(sort)
	direction := normalizeSortOrder(order)
	return "`" + field + "` " + strings.ToUpper(direction) + ", `id` DESC"
}

func normalizeSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_jav_id", "movie_name", "post_time", "post_date", "last_seen_time", "created_on", "updated_on":
		return strings.TrimSpace(raw)
	default:
		return "post_time"
	}
}

func normalizeSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}
