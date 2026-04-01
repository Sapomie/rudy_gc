package loop

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/service/fetchsite"
)

func buildSukebeiPageQueryFromTask(req StartTaskRequest) (fetchsite.SukebeiPageQuery, error) {
	triggerSort := strings.TrimSpace(req.TriggerSort)
	triggerOrder := strings.TrimSpace(req.TriggerOrder)
	if triggerSort == "" {
		triggerSort = strings.TrimSpace(req.Sort)
	}
	if triggerOrder == "" {
		triggerOrder = strings.TrimSpace(req.Order)
	}

	out := fetchsite.SukebeiPageQuery{
		Sort:    normalizeSukebeiTaskSortField(triggerSort),
		Order:   normalizeSukebeiTaskSortOrder(triggerOrder),
		Keyword: strings.TrimSpace(req.Keyword),
	}

	if v, ok, err := parseSukebeiTaskOwned(req.Owned); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("VFilm 库存筛选错误: %w", err)
	} else if ok {
		out.Owned = v
	}
	if v, ok, err := parseSukebeiTaskOwned(req.MediaOwned); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("WMedia 库存筛选错误: %w", err)
	} else if ok {
		out.MediaOwned = v
	}
	if values, ok, err := parseSukebeiTaskStatuses(req.Statuses, req.Status); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 状态错误: %w", err)
	} else if ok {
		out.Statuses = values
		out.HasStatuses = true
	}

	if ts, ok, err := parseSukebeiTaskDateStart(req.LastFetchFrom); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 最后抓取开始日期错误: %w", err)
	} else if ok {
		out.LastFetchFrom = ts
		out.HasLastFetchFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.LastFetchTo); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 最后抓取结束日期错误: %w", err)
	} else if ok {
		out.LastFetchTo = ts
		out.HasLastFetchTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.ReleaseDateFrom); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 发行时间开始日期错误: %w", err)
	} else if ok {
		out.ReleaseDateFrom = ts
		out.HasReleaseDateFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.ReleaseDateTo); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 发行时间结束日期错误: %w", err)
	} else if ok {
		out.ReleaseDateTo = ts
		out.HasReleaseDateTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.FilmBirthFrom); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 下载时间开始日期错误: %w", err)
	} else if ok {
		out.FilmBirthFrom = ts
		out.HasFilmBirthFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.FilmBirthTo); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei 下载时间结束日期错误: %w", err)
	} else if ok {
		out.FilmBirthTo = ts
		out.HasFilmBirthTo = true
	}
	if ts, ok, err := parseSukebeiTaskDateStart(req.MediaBirthFrom); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei M下载时间开始日期错误: %w", err)
	} else if ok {
		out.MediaBirthFrom = ts
		out.HasMediaBirthFrom = true
	}
	if ts, ok, err := parseSukebeiTaskDateEnd(req.MediaBirthTo); err != nil {
		return fetchsite.SukebeiPageQuery{}, fmt.Errorf("Sukebei M下载时间结束日期错误: %w", err)
	} else if ok {
		out.MediaBirthTo = ts
		out.HasMediaBirthTo = true
	}

	return out, nil
}

func normalizeSukebeiTaskSortField(raw string) string {
	switch strings.TrimSpace(raw) {
	case "movie_name", "release_date", "fetch_status", "last_fetch_time", "last_result_count", "torrent_hash_count", "latest_publish_time", "film_birth_time", "media_birth_time":
		return strings.TrimSpace(raw)
	default:
		return "last_fetch_time"
	}
}

func normalizeSukebeiTaskSortOrder(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "asc") {
		return "asc"
	}
	return "desc"
}

func parseSukebeiTaskOwned(raw string) (int64, bool, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return 0, false, nil
	}
	value, err := strconv.ParseInt(current, 10, 64)
	if err != nil {
		return 0, false, err
	}
	if value < 1 || value > 7 {
		return 0, false, fmt.Errorf("must be between 1 and 7")
	}
	return value, true, nil
}

func parseSukebeiTaskStatuses(raws []string, legacy string) ([]int64, bool, error) {
	merged := make([]string, 0, len(raws)+1)
	for _, raw := range raws {
		current := strings.TrimSpace(raw)
		if current == "" {
			continue
		}
		merged = append(merged, current)
	}
	if len(merged) == 0 {
		current := strings.TrimSpace(legacy)
		if current != "" {
			merged = append(merged, current)
		}
	}
	if len(merged) == 0 {
		return nil, false, nil
	}

	values := make([]int64, 0, len(merged))
	seen := make(map[int64]struct{}, len(merged))
	for _, raw := range merged {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, false, err
		}
		if value < 1 || value > 4 {
			return nil, false, fmt.Errorf("must be between 1 and 4")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, len(values) > 0, nil
}

func parseSukebeiTaskDateStart(raw string) (int64, bool, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return 0, false, nil
	}
	t, err := time.ParseInLocation("2006-01-02", current, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Unix(), true, nil
}

func parseSukebeiTaskDateEnd(raw string) (int64, bool, error) {
	current := strings.TrimSpace(raw)
	if current == "" {
		return 0, false, nil
	}
	t, err := time.ParseInLocation("2006-01-02", current, time.Local)
	if err != nil {
		return 0, false, err
	}
	return t.Add(24*time.Hour - time.Second).Unix(), true, nil
}
