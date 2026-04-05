package fetchsite

import (
	"context"
	"path/filepath"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BuildJavbusPage(ctx context.Context, query JavbusPageQuery) (*JavbusPageResult, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	filter := buildJavbusPageFilter(query)

	total, err := s.deps.JavbusMagnetFetchModel.CountPageRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	statusStats, err := s.deps.JavbusMagnetFetchModel.CountByFetchStatus(ctx, filter)
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	rows, err := s.deps.JavbusMagnetFetchModel.ListPageRows(ctx, offset, pageSize, javbusPageOrderBy(query.Sort, query.Order), filter)
	if err != nil {
		return nil, err
	}

	items := make([]*JavbusPageItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &JavbusPageItem{
			MovieJavID:        row.MovieJavId,
			MovieName:         row.MovieName,
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

	if err := s.fillJavbusInventory(ctx, items); err != nil {
		return nil, err
	}

	return &JavbusPageResult{
		Items:        items,
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
		SuccessCount: statusStats[FetchStatusSuccess],
		PendingCount: statusStats[FetchStatusPending],
		FailedCount:  statusStats[FetchStatusFailed],
	}, nil
}

func (s *Service) ListJavbusFetchTasksByPageQuery(ctx context.Context, query JavbusPageQuery) ([]*JavbusFetchTask, error) {
	rows, err := s.deps.JavbusMagnetFetchModel.ListPageRows(ctx, 0, 0, javbusPageOrderBy(query.Sort, query.Order), buildJavbusPageFilter(query))
	if err != nil {
		return nil, err
	}
	return s.BuildJavbusFetchTasksByRows(rows), nil
}

func buildJavbusPageFilter(query JavbusPageQuery) moviex.JavbusFetchPageFilter {
	filter := moviex.JavbusFetchPageFilter{
		Owned:              query.Owned,
		MediaOwned:         query.MediaOwned,
		Keyword:            query.Keyword,
		FetchStatuses:      query.Statuses,
		HasFetchStatuses:   query.HasStatuses,
		LastFetchFrom:      query.LastFetchFrom,
		LastFetchTo:        query.LastFetchTo,
		HasLastFetchFrom:   query.HasLastFetchFrom,
		HasLastFetchTo:     query.HasLastFetchTo,
		ReleaseDateFrom:    query.ReleaseDateFrom,
		ReleaseDateTo:      query.ReleaseDateTo,
		HasReleaseDateFrom: query.HasReleaseDateFrom,
		HasReleaseDateTo:   query.HasReleaseDateTo,
		FilmBirthFrom:      query.FilmBirthFrom,
		FilmBirthTo:        query.FilmBirthTo,
		HasFilmBirthFrom:   query.HasFilmBirthFrom,
		HasFilmBirthTo:     query.HasFilmBirthTo,
		MediaBirthFrom:     query.MediaBirthFrom,
		MediaBirthTo:       query.MediaBirthTo,
		HasMediaBirthFrom:  query.HasMediaBirthFrom,
		HasMediaBirthTo:    query.HasMediaBirthTo,
	}
	if query.Sort == "film_birth_time" {
		filter.RequireVFilm = true
	}
	if query.Sort == "media_birth_time" {
		filter.RequireWMedia = true
	}
	return filter
}

func (s *Service) fillJavbusInventory(ctx context.Context, items []*JavbusPageItem) error {
	if len(items) == 0 {
		return nil
	}

	movieJavIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.MovieJavID) == "" {
			continue
		}
		if _, ok := seen[item.MovieJavID]; ok {
			continue
		}
		seen[item.MovieJavID] = struct{}{}
		movieJavIDs = append(movieJavIDs, item.MovieJavID)
	}

	filmRows, err := s.deps.FilmModel.ListByMovieJavIds(ctx, movieJavIDs)
	if err != nil {
		return err
	}
	mediaRows, err := s.deps.WMediaModel.ListByMovieJavIds(ctx, movieJavIDs)
	if err != nil {
		return err
	}

	filmMap := make(map[string]*moviex.VFilm, len(filmRows))
	for _, row := range filmRows {
		if row == nil || strings.TrimSpace(row.MovieJavId) == "" {
			continue
		}
		filmMap[row.MovieJavId] = row
	}

	mediaMap := make(map[string]*moviex.WMedia, len(mediaRows))
	for _, row := range mediaRows {
		if row == nil || strings.TrimSpace(row.MovieJavId) == "" {
			continue
		}
		mediaMap[row.MovieJavId] = row
	}

	for _, item := range items {
		if item == nil || strings.TrimSpace(item.MovieJavID) == "" {
			continue
		}
		filmRow := filmMap[item.MovieJavID]
		mediaRow := mediaMap[item.MovieJavID]
		item.VFilmOwnedText, item.VFilmBirthText = buildVFilmInventoryText(filmRow)
		item.WMediaOwnedText, item.WMediaBirthText = buildWMediaInventoryText(mediaRow)
		if filmRow != nil {
			item.Owned = buildOwnedState(filmRow.IsRemoved, filmRow.HasSub)
			item.VideoURL = filmRow.FullDir + string(filepath.Separator) + filmRow.FileName
			item.FilmBirthDate = pageFormatDate(filmRow.BirthTime)
		} else {
			item.Owned = consts.OwnedNotOwned
		}
		if mediaRow != nil {
			item.OwnedWMedia = buildOwnedState(mediaRow.IsRemoved, mediaRow.HasSub)
			if mediaRow.IsRemoved != consts.FilmIsRemoved {
				item.VideoURLWMedia = mediaRow.FullDir + string(filepath.Separator) + mediaRow.FileName
				item.FilmBirthDateWMedia = pageFormatDate(mediaRow.BirthTime)
			}
		} else {
			item.OwnedWMedia = consts.OwnedNotOwned
		}
	}

	return nil
}

func javbusPageOrderBy(sortField string, sortOrder string) string {
	column := mapJavbusPageSortColumn(sortField)
	order := normalizePageSortOrder(sortOrder)
	return column + " " + order + ", jf.`id` DESC"
}

func mapJavbusPageSortColumn(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "movie_name":
		return "jf.`movie_name`"
	case "release_date":
		return "jf.`release_date`"
	case "fetch_status":
		return "jf.`fetch_status`"
	case "last_fetch_time":
		return "jf.`last_fetch_time`"
	case "last_result_count":
		return "jf.`last_result_count`"
	case "torrent_hash_count":
		return "jf.`torrent_hash_count`"
	case "latest_publish_time":
		return "jf.`latest_publish_time`"
	case "film_birth_time":
		return "vf.`birth_time`"
	case "media_birth_time":
		return "wm.`birth_time`"
	default:
		return "jf.`updated_on`"
	}
}
