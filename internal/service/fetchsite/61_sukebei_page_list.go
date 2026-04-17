package fetchsite

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
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

	filter := buildSukebeiPageFilter(query)

	total, err := s.deps.SukebeiTorrentFetchModel.CountPageRows(ctx, filter)
	if err != nil {
		return nil, err
	}
	statusStats, err := s.deps.SukebeiTorrentFetchModel.CountByFetchStatus(ctx, filter)
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

	if err := s.fillSukebeiInventory(ctx, items); err != nil {
		return nil, err
	}

	return &SukebeiPageResult{
		Items:        items,
		Page:         page,
		PageSize:     pageSize,
		Total:        total,
		SuccessCount: statusStats[FetchStatusSuccess],
		PendingCount: statusStats[FetchStatusPending],
		FailedCount:  statusStats[FetchStatusFailed],
	}, nil
}

func (s *Service) ListSukebeiFetchTasksByPageQuery(ctx context.Context, query SukebeiPageQuery) ([]*SukebeiFetchTask, error) {
	rows, err := s.deps.SukebeiTorrentFetchModel.ListPageRows(ctx, 0, 0, sukebeiPageOrderBy(query.Sort, query.Order), buildSukebeiPageFilter(query))
	if err != nil {
		return nil, err
	}
	return s.BuildSukebeiFetchTasksByRows(rows), nil
}

func buildSukebeiPageFilter(query SukebeiPageQuery) moviex.SukebeiFetchPageFilter {
	filter := moviex.SukebeiFetchPageFilter{
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
		MediaBirthFrom:     query.MediaBirthFrom,
		MediaBirthTo:       query.MediaBirthTo,
		HasMediaBirthFrom:  query.HasMediaBirthFrom,
		HasMediaBirthTo:    query.HasMediaBirthTo,
	}
	if query.Sort == "media_birth_time" {
		filter.RequireWMedia = true
	}
	return filter
}

func (s *Service) fillSukebeiInventory(ctx context.Context, items []*SukebeiPageItem) error {
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

	mediaRows, err := s.deps.WMediaModel.ListByMovieJavIds(ctx, movieJavIDs)
	if err != nil {
		return err
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
		mediaRow := mediaMap[item.MovieJavID]
		item.WMediaOwnedText, item.WMediaBirthText = buildWMediaInventoryText(mediaRow)
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

func buildWMediaInventoryText(row *moviex.WMedia) (string, string) {
	if row == nil {
		return strings.Join([]string{
			strconv.FormatInt(consts.MovieAll, 10),
			strconv.FormatInt(consts.OwnedNotOwned, 10),
		}, "/"), ""
	}
	return buildOwnedInventoryText(row.IsRemoved, row.HasSub, row.BirthTime)
}

func buildOwnedInventoryText(isRemoved int64, hasSub int64, birthTime int64) (string, string) {
	states := []string{strconv.FormatInt(consts.MovieAll, 10)}
	states = append(states, strconv.FormatInt(consts.OwnedAll, 10))
	if isRemoved == consts.FilmIsRemoved {
		states = append(states, strconv.FormatInt(consts.OwnedRemoved, 10))
	} else {
		states = append(states, strconv.FormatInt(consts.OwnedAllNotRemoved, 10))
		if hasSub == consts.FilmHasSub {
			states = append(states, strconv.FormatInt(consts.OwnedHasSubNotRemoved, 10))
		} else {
			states = append(states, strconv.FormatInt(consts.OwnedNoSubNotRemoved, 10))
		}
	}
	return strings.Join(states, "/"), pageFormatDate(birthTime)
}

func buildOwnedState(isRemoved int64, hasSub int64) int64 {
	if isRemoved == consts.FilmIsRemoved {
		return consts.OwnedRemoved
	}
	if hasSub == consts.FilmHasSub {
		return consts.OwnedHasSubNotRemoved
	}
	return consts.OwnedNoSubNotRemoved
}

func sukebeiPageOrderBy(sortField string, sortOrder string) string {
	column := mapSukebeiPageSortColumn(sortField)
	order := normalizePageSortOrder(sortOrder)
	return column + " " + order + ", sf.`id` DESC"
}

func normalizePageSortOrder(order string) string {
	if strings.EqualFold(strings.TrimSpace(order), "asc") {
		return "ASC"
	}
	return "DESC"
}

func mapSukebeiPageSortColumn(sortField string) string {
	switch strings.TrimSpace(sortField) {
	case "movie_name":
		return "sf.`movie_name`"
	case "release_date":
		return "sf.`release_date`"
	case "fetch_status":
		return "sf.`fetch_status`"
	case "last_fetch_time":
		return "sf.`last_fetch_time`"
	case "last_result_count":
		return "sf.`last_result_count`"
	case "torrent_hash_count":
		return "sf.`torrent_hash_count`"
	case "latest_publish_time":
		return "sf.`latest_publish_time`"
	case "media_birth_time":
		return "wm.`birth_time`"
	default:
		return "sf.`updated_on`"
	}
}
