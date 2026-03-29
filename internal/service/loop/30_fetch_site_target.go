package loop

import (
	"context"
	"strings"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/types"
)

func (l *FetchLoopService) runJavbusFetchTasksWithFilterTarget(
	ctx context.Context,
	baseReq *types.ListMovieFullRequest,
	targetCount int64,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) (*fetchsite.RunFetchTasksResult, error) {
	basePageSize := int64(0)
	if baseReq != nil {
		basePageSize = baseReq.PageSize
	}
	pageSize := normalizeFetchSiteNumber(basePageSize)
	page := int64(1)
	selected := make([]*fetchsite.JavbusFetchTask, 0)
	seenMovie := make(map[string]struct{})

	for int64(len(selected)) < targetCount {
		req := cloneFetchSiteMovieRequest(baseReq, page, pageSize)
		resp, err := l.movieSvc.ListMovieFull(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(resp.List) <= 0 {
			break
		}

		pageTasks, _, _, err := l.fetchQueue.BuildFilteredJavbusFetchTasksByMovies(ctx, resp.List, lastFetchDurationDays, lastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		for _, task := range pageTasks {
			if task == nil {
				continue
			}
			movieJavID := strings.TrimSpace(task.MovieJavID)
			if movieJavID == "" {
				continue
			}
			if _, ok := seenMovie[movieJavID]; ok {
				continue
			}
			seenMovie[movieJavID] = struct{}{}
			selected = append(selected, task)
			if int64(len(selected)) >= targetCount {
				break
			}
		}
		if page*pageSize >= resp.Total {
			break
		}
		page++
	}

	return l.fetchQueue.RunPreparedJavbusFetchTasks(ctx, selected, "JavBus 筛选任务已加载")
}

func (l *FetchLoopService) runSukebeiFetchTasksWithFilterTarget(
	ctx context.Context,
	baseReq *types.ListMovieFullRequest,
	targetCount int64,
	lastFetchDurationDays int64,
	lastSuccessDurationDays int64,
) (*fetchsite.RunFetchTasksResult, error) {
	basePageSize := int64(0)
	if baseReq != nil {
		basePageSize = baseReq.PageSize
	}
	pageSize := normalizeFetchSiteNumber(basePageSize)
	page := int64(1)
	selected := make([]*fetchsite.SukebeiFetchTask, 0)
	seenMovie := make(map[string]struct{})

	for int64(len(selected)) < targetCount {
		req := cloneFetchSiteMovieRequest(baseReq, page, pageSize)
		resp, err := l.movieSvc.ListMovieFull(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(resp.List) <= 0 {
			break
		}

		pageTasks, _, _, err := l.fetchQueue.BuildFilteredSukebeiFetchTasksByMovies(ctx, resp.List, lastFetchDurationDays, lastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		for _, task := range pageTasks {
			if task == nil {
				continue
			}
			movieJavID := strings.TrimSpace(task.MovieJavID)
			if movieJavID == "" {
				continue
			}
			if _, ok := seenMovie[movieJavID]; ok {
				continue
			}
			seenMovie[movieJavID] = struct{}{}
			selected = append(selected, task)
			if int64(len(selected)) >= targetCount {
				break
			}
		}
		if page*pageSize >= resp.Total {
			break
		}
		page++
	}

	return l.fetchQueue.RunPreparedSukebeiFetchTasks(ctx, selected, "Sukebei 筛选任务已加载")
}

func cloneFetchSiteMovieRequest(baseReq *types.ListMovieFullRequest, page int64, pageSize int64) *types.ListMovieFullRequest {
	if baseReq == nil {
		return &types.ListMovieFullRequest{
			Page:     page,
			PageSize: pageSize,
		}
	}
	out := *baseReq
	out.Page = page
	out.PageSize = pageSize
	return &out
}
