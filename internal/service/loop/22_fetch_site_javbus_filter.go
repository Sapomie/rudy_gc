package loop

import (
	"fmt"
	"strings"

	"rudy_gc/internal/service/fetchsite"
)

func buildJavbusPageQueryFromTask(req StartTaskRequest) (fetchsite.JavbusPageQuery, error) {
	triggerSort := strings.TrimSpace(req.TriggerSort)
	triggerOrder := strings.TrimSpace(req.TriggerOrder)
	if triggerSort == "" {
		triggerSort = strings.TrimSpace(req.Sort)
	}
	if triggerOrder == "" {
		triggerOrder = strings.TrimSpace(req.Order)
	}

	out := fetchsite.JavbusPageQuery{
		Sort:    normalizeJavbusTaskSortField(triggerSort),
		Order:   normalizeJavbusTaskSortOrder(triggerOrder),
		Keyword: strings.TrimSpace(req.Keyword),
	}

	if v, ok, err := parseSukebeiTaskOwned(req.Owned); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("VFilm 库存筛选错误: %w", err)
	} else if ok {
		out.Owned = v
	}
	if v, ok, err := parseSukebeiTaskOwned(req.MediaOwned); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("WMedia 库存筛选错误: %w", err)
	} else if ok {
		out.MediaOwned = v
	}
	if values, ok, err := parseSukebeiTaskStatuses(req.Statuses, req.Status); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 状态错误: %w", err)
	} else if ok {
		out.Statuses = values
		out.HasStatuses = true
	}

	if ts, ok, err := parseSukebeiTaskDateStart(req.LastFetchFrom); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 最后抓取开始日期错误: %w", err)
	} else if ok {
		out.LastFetchFrom = ts
		out.HasLastFetchFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.LastFetchTo); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 最后抓取结束日期错误: %w", err)
	} else if ok {
		out.LastFetchTo = ts
		out.HasLastFetchTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.ReleaseDateFrom); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 发行时间开始日期错误: %w", err)
	} else if ok {
		out.ReleaseDateFrom = ts
		out.HasReleaseDateFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.ReleaseDateTo); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 发行时间结束日期错误: %w", err)
	} else if ok {
		out.ReleaseDateTo = ts
		out.HasReleaseDateTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.FilmBirthFrom); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 下载时间开始日期错误: %w", err)
	} else if ok {
		out.FilmBirthFrom = ts
		out.HasFilmBirthFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.FilmBirthTo); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus 下载时间结束日期错误: %w", err)
	} else if ok {
		out.FilmBirthTo = ts
		out.HasFilmBirthTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.MediaBirthFrom); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus M下载时间开始日期错误: %w", err)
	} else if ok {
		out.MediaBirthFrom = ts
		out.HasMediaBirthFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.MediaBirthTo); err != nil {
		return fetchsite.JavbusPageQuery{}, fmt.Errorf("JavBus M下载时间结束日期错误: %w", err)
	} else if ok {
		out.MediaBirthTo = ts
		out.HasMediaBirthTo = true
	}

	return out, nil
}

func normalizeJavbusTaskSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "film_birth_time", "media_birth_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeJavbusTaskSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}
