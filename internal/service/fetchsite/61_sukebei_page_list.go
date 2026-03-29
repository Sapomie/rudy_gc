package fetchsite

import (
	"context"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BuildSukebeiPage(ctx context.Context, query SukebeiPageQuery) (*SukebeiPageResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	filter := moviex.SukebeiFetchPageFilter{
		Keyword:            query.Keyword,
		FetchStatus:        query.Status,
		FetchStatusSet:     query.StatusSet,
		LastFetchFrom:      query.LastFetchFrom,
		LastFetchTo:        query.LastFetchTo,
		HasLastFetchFrom:   query.HasLastFetchFrom,
		HasLastFetchTo:     query.HasLastFetchTo,
		ReleaseDateFrom:    query.ReleaseDateFrom,
		ReleaseDateTo:      query.ReleaseDateTo,
		HasReleaseDateFrom: query.HasReleaseDateFrom,
		HasReleaseDateTo:   query.HasReleaseDateTo,
	}
	if query.HasErrorOnly {
		filter.HasErrorOnly = true
	}
	if query.HasNoErrorOnly {
		filter.HasNoErrorOnly = true
	}

	total, err := s.deps.SukebeiTorrentFetchModel.CountPageRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.deps.SukebeiTorrentFetchModel.ListPageRows(ctx, offset, pageSize, sukebeiPageOrderBy(query.Sort, query.Order), filter)
	if err != nil {
		return nil, err
	}

	items := make([]*SukebeiPageItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &SukebeiPageItem{
			MovieJavID:        row.MovieJavId,
			MovieCode:         row.MovieCode,
			ReleaseDate:       row.ReleaseDate,
			ReleaseDateText:   pageFormatDate(row.ReleaseDate),
			FetchStatus:       row.FetchStatus,
			FetchStatusText:   pageFetchStatusText(row.FetchStatus),
			TryCount:          row.TryCount,
			LastFetchTime:     row.LastFetchTime,
			LastFetchText:     pageFormatUnix(row.LastFetchTime),
			LastResultCount:   row.LastResultCount,
			TorrentHashCount:  row.TorrentHashCount,
			LatestPublishTime: row.LatestPublishTime,
			LatestPublishText: pageFormatDate(row.LatestPublishTime),
			LastError:         row.LastError,
		})
	}

	return &SukebeiPageResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func sukebeiPageOrderBy(sortField string, sortOrder string) string {
	column := mapSukebeiPageSortColumn(sortField)
	order := normalizePageSortOrder(sortOrder)
	return column + " " + order + ", `id` DESC"
}

func normalizePageSortOrder(order string) string {
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		return "ASC"
	}
	return "DESC"
}

func mapSukebeiPageSortColumn(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "movie_code":
		return "`movie_code`"
	case "fetch_status":
		return "`fetch_status`"
	case "last_fetch_time":
		return "`last_fetch_time`"
	case "last_result_count":
		return "`last_result_count`"
	case "torrent_hash_count":
		return "`torrent_hash_count`"
	case "latest_publish_time":
		return "`latest_publish_time`"
	default:
		return "`updated_on`"
	}
}
