package loop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"rudy_gc/internal/service/fetchsite"
)

const (
	fetchSiteSourceJavbus  = "javbus"
	fetchSiteSourceSukebei = "sukebei"
)

type FetchSitePreviewAggregate struct {
	MatchedMovies        int64 `json:"matched_movies"`
	SelectedItems        int64 `json:"selected_items"`
	JavbusCount          int64 `json:"javbus_count"`
	SukebeiCount         int64 `json:"sukebei_count"`
	BothCount            int64 `json:"both_count"`
	SkippedRecentFetch   int64 `json:"skipped_recent_fetch"`
	SkippedRecentSuccess int64 `json:"skipped_recent_success"`
}

type FetchSitePreviewResult struct {
	TaskType  string                    `json:"task_type"`
	Signature string                    `json:"signature"`
	Aggregate FetchSitePreviewAggregate `json:"aggregate"`
	Targets   []FetchSiteTarget         `json:"targets"`
}

func (l *FetchLoopService) PreviewFetchSiteTargets(ctx context.Context, req StartTaskRequest) (*FetchSitePreviewResult, error) {
	taskType := strings.TrimSpace(req.TaskType)
	if !isFetchSitePreviewTaskType(taskType) {
		return nil, fmt.Errorf("unsupported task_type: %s", taskType)
	}
	movieReq, err := buildFetchSiteMovieRequest(req)
	if err != nil {
		return nil, err
	}
	durationFilter, err := buildFetchSiteDurationFilter(req)
	if err != nil {
		return nil, err
	}
	movies, err := l.movieSvc.ListMovieFullAllOrdered(ctx, movieReq)
	if err != nil {
		return nil, err
	}

	targets := make([]FetchSiteTarget, 0)
	javbusMovies := make(map[string]struct{})
	sukebeiMovies := make(map[string]struct{})
	sequence := int64(0)
	skippedRecentFetch := 0
	skippedRecentSuccess := 0

	if taskType == TaskSpiderFetchJavbus || taskType == TaskSpiderFetchSiteBoth {
		tasks, skippedFetch, skippedSuccess, err := l.fetchQueue.BuildFilteredJavbusFetchTasksByMovies(ctx, movies.List, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		skippedRecentFetch += skippedFetch
		skippedRecentSuccess += skippedSuccess
		for _, task := range tasks {
			if task == nil || strings.TrimSpace(task.MovieJavID) == "" {
				continue
			}
			sequence++
			movieJavID := strings.TrimSpace(task.MovieJavID)
			javbusMovies[movieJavID] = struct{}{}
			targets = append(targets, fetchSiteTargetFromJavbusTask(task, sequence))
		}
	}
	if taskType == TaskSpiderFetchSukebei || taskType == TaskSpiderFetchSiteBoth {
		tasks, skippedFetch, skippedSuccess, err := l.fetchQueue.BuildFilteredSukebeiFetchTasksByMovies(ctx, movies.List, durationFilter.LastFetchDurationDays, durationFilter.LastSuccessDurationDays)
		if err != nil {
			return nil, err
		}
		skippedRecentFetch += skippedFetch
		skippedRecentSuccess += skippedSuccess
		for _, task := range tasks {
			if task == nil || strings.TrimSpace(task.MovieJavID) == "" {
				continue
			}
			sequence++
			movieJavID := strings.TrimSpace(task.MovieJavID)
			sukebeiMovies[movieJavID] = struct{}{}
			targets = append(targets, fetchSiteTargetFromSukebeiTask(task, sequence))
		}
	}

	bothCount := int64(0)
	for movieJavID := range javbusMovies {
		if _, ok := sukebeiMovies[movieJavID]; ok {
			bothCount++
		}
	}
	signature, err := fetchSitePreviewSignature(req)
	if err != nil {
		return nil, err
	}
	aggregate := FetchSitePreviewAggregate{
		MatchedMovies:        movies.Total,
		SelectedItems:        int64(len(targets)),
		JavbusCount:          int64(len(javbusMovies)),
		SukebeiCount:         int64(len(sukebeiMovies)),
		BothCount:            bothCount,
		SkippedRecentFetch:   int64(skippedRecentFetch),
		SkippedRecentSuccess: int64(skippedRecentSuccess),
	}
	return &FetchSitePreviewResult{
		TaskType:  taskType,
		Signature: signature,
		Aggregate: aggregate,
		Targets:   targets,
	}, nil
}

func fetchSiteTargetFromJavbusTask(task *fetchsite.JavbusFetchTask, sequence int64) FetchSiteTarget {
	return FetchSiteTarget{
		Sequence:   sequence,
		Source:     fetchSiteSourceJavbus,
		MovieJavID: strings.TrimSpace(task.MovieJavID),
		MovieName:  strings.TrimSpace(task.MovieName),
	}
}

func fetchSiteTargetFromSukebeiTask(task *fetchsite.SukebeiFetchTask, sequence int64) FetchSiteTarget {
	return FetchSiteTarget{
		Sequence:   sequence,
		Source:     fetchSiteSourceSukebei,
		MovieJavID: strings.TrimSpace(task.MovieJavID),
		MovieName:  strings.TrimSpace(task.MovieName),
	}
}

func fetchSitePreviewSignature(req StartTaskRequest) (string, error) {
	req.FetchSiteTargets = nil
	req.Number = 0
	payload, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func isFetchSitePreviewTaskType(taskType string) bool {
	switch strings.TrimSpace(taskType) {
	case TaskSpiderFetchJavbus, TaskSpiderFetchSukebei, TaskSpiderFetchSiteBoth:
		return true
	default:
		return false
	}
}
