package loop

import (
	"context"
	"fmt"
	"strings"

	"rudy_gc/internal/service/fetchsite"
)

const fetchSitePreviewDefaultPageSize int64 = 100

type FetchSitePreviewItem struct {
	MovieJavID string `json:"movie_jav_id"`
	MovieName  string `json:"movie_name"`
	Source     string `json:"source"`
}

type FetchSitePreviewResult struct {
	TaskType   string                  `json:"task_type"`
	Total      int64                   `json:"total"`
	Page       int64                   `json:"page"`
	PageSize   int64                   `json:"page_size"`
	TotalPages int64                   `json:"total_pages"`
	Items      []*FetchSitePreviewItem `json:"items"`
}

func (l *FetchLoopService) PreviewFetchSiteTargets(ctx context.Context, req StartTaskRequest, page int64, pageSize int64) (*FetchSitePreviewResult, error) {
	taskType := strings.TrimSpace(req.TaskType)
	movieReq, err := buildFetchSiteMovieRequest(req)
	if err != nil {
		return nil, err
	}
	durationFilter, err := buildFetchSiteDurationFilter(req)
	if err != nil {
		return nil, err
	}
	targetCount := normalizeFetchSiteNumber(req.Number)
	page = normalizeFetchSitePreviewPage(page)
	pageSize = normalizeFetchSitePreviewPageSize(pageSize)

	var items []*FetchSitePreviewItem
	switch taskType {
	case TaskSpiderFetchJavbus:
		tasks, err := l.buildFilteredJavbusFetchTasksWithFilterTarget(ctx, movieReq, targetCount, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		items = buildFetchSitePreviewItemsFromJavbus(tasks)
	case TaskSpiderFetchSukebei:
		tasks, err := l.buildFilteredSukebeiFetchTasksWithFilterTarget(ctx, movieReq, targetCount, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		items = buildFetchSitePreviewItemsFromSukebei(tasks)
	case TaskSpiderFetchSiteBoth:
		javbusTasks, err := l.buildFilteredJavbusFetchTasksWithFilterTarget(ctx, movieReq, targetCount, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		sukebeiTasks, err := l.buildFilteredSukebeiFetchTasksWithFilterTarget(ctx, movieReq, targetCount, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		items = mergeFetchSitePreviewItems(
			buildFetchSitePreviewItemsFromJavbus(javbusTasks),
			buildFetchSitePreviewItemsFromSukebei(sukebeiTasks),
		)
	default:
		return nil, fmt.Errorf("unsupported task_type: %s", taskType)
	}

	total := int64(len(items))
	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	pageItems := []*FetchSitePreviewItem{}
	if total > 0 && start < end {
		pageItems = items[start:end]
	}

	return &FetchSitePreviewResult{
		TaskType:   taskType,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Items:      pageItems,
	}, nil
}

func buildFetchSitePreviewItemsFromJavbus(tasks []*fetchsite.JavbusFetchTask) []*FetchSitePreviewItem {
	out := make([]*FetchSitePreviewItem, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		out = append(out, &FetchSitePreviewItem{
			MovieJavID: strings.TrimSpace(task.MovieJavID),
			MovieName:  strings.TrimSpace(task.MovieName),
			Source:     "JavBus",
		})
	}
	return out
}

func buildFetchSitePreviewItemsFromSukebei(tasks []*fetchsite.SukebeiFetchTask) []*FetchSitePreviewItem {
	out := make([]*FetchSitePreviewItem, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		out = append(out, &FetchSitePreviewItem{
			MovieJavID: strings.TrimSpace(task.MovieJavID),
			MovieName:  strings.TrimSpace(task.MovieName),
			Source:     "Sukebei",
		})
	}
	return out
}

func mergeFetchSitePreviewItems(left []*FetchSitePreviewItem, right []*FetchSitePreviewItem) []*FetchSitePreviewItem {
	out := make([]*FetchSitePreviewItem, 0, len(left)+len(right))
	indexByMovie := make(map[string]int, len(left)+len(right))
	appendItem := func(item *FetchSitePreviewItem) {
		if item == nil {
			return
		}
		movieJavID := strings.TrimSpace(item.MovieJavID)
		movieName := strings.TrimSpace(item.MovieName)
		source := strings.TrimSpace(item.Source)
		if movieJavID == "" && movieName == "" {
			return
		}
		key := movieJavID
		if key == "" {
			key = movieName
		}
		if key == "" {
			return
		}
		if idx, ok := indexByMovie[key]; ok {
			current := out[idx]
			if current.MovieName == "" {
				current.MovieName = movieName
			}
			if current.Source == "" {
				current.Source = source
				return
			}
			if source != "" && !strings.Contains(current.Source, source) {
				current.Source = current.Source + " / " + source
			}
			return
		}
		indexByMovie[key] = len(out)
		out = append(out, &FetchSitePreviewItem{
			MovieJavID: movieJavID,
			MovieName:  movieName,
			Source:     source,
		})
	}
	for _, item := range left {
		appendItem(item)
	}
	for _, item := range right {
		appendItem(item)
	}
	return out
}

func normalizeFetchSitePreviewPage(page int64) int64 {
	if page > 0 {
		return page
	}
	return 1
}

func normalizeFetchSitePreviewPageSize(pageSize int64) int64 {
	if pageSize > 0 {
		return pageSize
	}
	return fetchSitePreviewDefaultPageSize
}
