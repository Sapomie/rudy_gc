package fetchsehuatang

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

type SHDetailFetchRequest struct {
	TargetForumIDs  []string `json:"target_forum_ids"`
	Keyword         string   `json:"keyword"`
	DetailStatus    string   `json:"detail_status"`
	ThreadTIDs      []string `json:"thread_tids"`
	PostTimeFrom    int64    `json:"post_time_from"`
	HasPostTimeFrom bool     `json:"has_post_time_from"`
	PostTimeTo      int64    `json:"post_time_to"`
	HasPostTimeTo   bool     `json:"has_post_time_to"`
}

func (s *Service) FetchSHThreadList(ctx context.Context, req SHFetchRequest) (*SHFetchResult, error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}
	if s.deps.ShSehuatangThreadModel == nil || s.deps.ShSehuatangFetchRunModel == nil || s.deps.ShSehuatangFetchItemModel == nil {
		return nil, fmt.Errorf("sh sehuatang models are nil")
	}

	req = normalizeSHFetchRequest(req)
	req.FetchMode = SHFetchModeForum
	req.Keyword = ""
	if err := validateSHFetchRequest(req); err != nil {
		return nil, err
	}

	cfg, ok := s.deps.GetFetchSiteConfig(siteCode())
	if !ok {
		return nil, fmt.Errorf("fetch site config not found: %s", siteCode())
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("fetch site base_url is empty: %s", siteCode())
	}
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout for site %s", siteCode())
	}
	if cfg.MaxRetryTimes <= 0 {
		return nil, fmt.Errorf("invalid max_retry_times for site %s", siteCode())
	}

	state := &remoteState{}
	jobs, err := s.buildSHPageJobs(ctx, cfg, state, req)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("没有可抓取的页面")
	}

	targetForumIDs := strings.Join(req.TargetForumIDs, ",")
	now := time.Now().Unix()
	run := &moviex.ShSehuatangFetchRun{
		FetchMode:      SHFetchModeForumList,
		TargetForumIDs: targetForumIDs,
		StartURL:       jobs[0].URL,
		CurrentURL:     jobs[0].URL,
		CurrentPage:    req.StartPage,
		Status:         SHRunStatusRunning,
		StartedOn:      now,
		CreatedOn:      now,
		UpdatedOn:      now,
	}
	insertResult, err := s.deps.ShSehuatangFetchRunModel.Insert(ctx, run)
	if err != nil {
		return nil, err
	}
	runID, err := insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	run.Id = runID

	result := &SHFetchResult{
		RunID:           runID,
		FetchMode:       SHFetchModeForumList,
		TargetForumIDs:  targetForumIDs,
		StartPage:       req.StartPage,
		EndPage:         req.EndPage,
		TotalPageCount:  int64(len(jobs)),
		CurrentPage:     jobs[0].Page,
		CurrentPageText: jobs[0].PageText,
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池开始: run_id=%d | forums=%s | page=%d-%d", runID, targetForumIDs, req.StartPage, req.EndPage))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:           "fetch_sehuatang_sh_list_start",
		Message:         fmt.Sprintf("开始色花堂列表建池，run_id=%d，版块=%s，页码=%d-%d", runID, targetForumIDs, req.StartPage, req.EndPage),
		CurrentPhaseKey: "list",
	})

	fetchErr := s.fetchSHListPages(ctx, cfg, state, req, jobs, run, result)
	finishTime := time.Now().Unix()
	run.FinishedOn = finishTime
	run.UpdatedOn = finishTime
	syncSHRunCounts(run, result)
	if fetchErr != nil {
		run.Status = SHRunStatusFailed
		run.LastError = fetchErr.Error()
		result.LastError = fetchErr.Error()
	} else {
		run.Status = SHRunStatusSuccess
	}
	if updateErr := s.deps.ShSehuatangFetchRunModel.Update(context.Background(), run); updateErr != nil && fetchErr == nil {
		return result, updateErr
	}
	if fetchErr != nil {
		return result, fetchErr
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池完成: run_id=%d | seen=%d | matched=%d | list_saved=%d | failed=%d", runID, result.SeenCount, result.MatchedCount, result.FetchedCount, result.FailedCount))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_list_done",
		Message:      fmt.Sprintf("色花堂列表建池完成：发现=%d，基础入库=%d，失败=%d", result.SeenCount, result.FetchedCount, result.FailedCount),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return result, nil
}

func (s *Service) FetchSHThreadDetails(ctx context.Context, req SHDetailFetchRequest) (*SHFetchResult, error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}
	if s.deps.ShSehuatangThreadModel == nil || s.deps.ShSehuatangFetchRunModel == nil || s.deps.ShSehuatangFetchItemModel == nil {
		return nil, fmt.Errorf("sh sehuatang models are nil")
	}

	req = normalizeSHDetailFetchRequest(req)
	if err := validateSHDetailFetchRequest(req); err != nil {
		return nil, err
	}

	cfg, ok := s.deps.GetFetchSiteConfig(siteCode())
	if !ok {
		return nil, fmt.Errorf("fetch site config not found: %s", siteCode())
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("fetch site base_url is empty: %s", siteCode())
	}
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout for site %s", siteCode())
	}
	if cfg.MaxRetryTimes <= 0 {
		return nil, fmt.Errorf("invalid max_retry_times for site %s", siteCode())
	}

	rows := make([]*moviex.ShSehuatangThread, 0, len(req.ThreadTIDs))
	seenThreadTIDs := make(map[string]struct{}, len(req.ThreadTIDs))
	targetForumIDs := make([]string, 0, len(req.ThreadTIDs))
	for _, threadTID := range req.ThreadTIDs {
		threadTID = strings.TrimSpace(threadTID)
		if threadTID == "" {
			continue
		}
		if _, ok := seenThreadTIDs[threadTID]; ok {
			continue
		}
		seenThreadTIDs[threadTID] = struct{}{}
		row, err := s.deps.ShSehuatangThreadModel.FindOneByThreadTID(ctx, threadTID)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
		if row != nil && strings.TrimSpace(row.SourceForumID) != "" {
			targetForumIDs = append(targetForumIDs, strings.TrimSpace(row.SourceForumID))
		}
	}

	startURL := ""
	if len(rows) > 0 && rows[0] != nil {
		startURL = rows[0].ThreadURL
	}
	targetForumText := strings.Join(uniqueSHStrings(targetForumIDs), ",")
	now := time.Now().Unix()
	run := &moviex.ShSehuatangFetchRun{
		FetchMode:      SHFetchModeDetail,
		TargetForumIDs: targetForumText,
		StartURL:       startURL,
		CurrentURL:     startURL,
		Status:         SHRunStatusRunning,
		StartedOn:      now,
		CreatedOn:      now,
		UpdatedOn:      now,
	}
	insertResult, err := s.deps.ShSehuatangFetchRunModel.Insert(ctx, run)
	if err != nil {
		return nil, err
	}
	runID, err := insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	run.Id = runID

	result := &SHFetchResult{
		RunID:           runID,
		FetchMode:       SHFetchModeDetail,
		TargetForumIDs:  targetForumText,
		TotalPageCount:  int64(len(rows)),
		CurrentPageText: "detail=0",
	}
	reportInfoLog(ctx, fmt.Sprintf("色花堂选中帖子详情抓取开始: selected=%d", len(rows)))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:           "fetch_sehuatang_sh_detail_start",
		Message:         fmt.Sprintf("开始抓取选中帖子详情，选中=%d", len(rows)),
		CurrentPhaseKey: "detail",
	})

	state := &remoteState{}
	for index, row := range rows {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		if row == nil || strings.TrimSpace(row.ThreadTID) == "" {
			continue
		}

		result.SeenCount++
		result.MatchedCount++
		result.CurrentPage = int64(index + 1)
		result.CurrentPageText = fmt.Sprintf("detail=%d/%d tid=%s", index+1, len(rows), row.ThreadTID)
		result.CompletedPageCount = int64(index)
		run.CurrentURL = row.ThreadURL
		run.CurrentPage = int64(index + 1)
		run.UpdatedOn = time.Now().Unix()
		syncSHRunCounts(run, result)
		if err := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); err != nil {
			return result, err
		}

		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_detail_item_start",
			Message:           fmt.Sprintf("开始抓取帖子详情：%s", row.ThreadTitle),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: index,
			PhaseTotalCount:   len(rows),
		})

		topic := shThreadRowToTopic(row)
		if err := s.fetchSHTopicDetailWithOptions(ctx, cfg, state, SHFetchRequest{FetchMode: SHFetchModeDetail}, run, result, topic, false); err != nil {
			return result, err
		}

		result.CompletedPageCount = int64(index + 1)
		run.UpdatedOn = time.Now().Unix()
		syncSHRunCounts(run, result)
		if err := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); err != nil {
			return result, err
		}
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_detail_item_done",
			Message:           fmt.Sprintf("帖子详情处理完成：%s", row.ThreadTitle),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: index + 1,
			PhaseTotalCount:   len(rows),
		})
	}

	finishTime := time.Now().Unix()
	run.FinishedOn = finishTime
	run.UpdatedOn = finishTime
	syncSHRunCounts(run, result)
	run.Status = SHRunStatusSuccess
	if result.FailedCount > 0 {
		run.LastError = result.LastError
	}
	if err := s.deps.ShSehuatangFetchRunModel.Update(context.Background(), run); err != nil {
		return result, err
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂选中帖子详情抓取完成: selected=%d | fetched=%d | failed=%d", result.SeenCount, result.FetchedCount, result.FailedCount))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_detail_done",
		Message:      fmt.Sprintf("色花堂选中帖子详情抓取完成：选中=%d，成功=%d，失败=%d", len(rows), result.FetchedCount, result.FailedCount),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return result, nil
}

func (s *Service) fetchSHListPages(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest, jobs []shPageJob, run *moviex.ShSehuatangFetchRun, result *SHFetchResult) error {
	seenThreads := make(map[string]struct{}, 128)
	for index, job := range jobs {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return err
		}

		result.CurrentPage = job.Page
		result.CurrentPageText = job.PageText
		result.CompletedPageCount = int64(index)
		run.CurrentURL = job.URL
		run.CurrentPage = job.Page
		run.UpdatedOn = time.Now().Unix()
		syncSHRunCounts(run, result)
		if err := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); err != nil {
			return err
		}

		reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池抓取: run_id=%d | %d/%d | %s", run.Id, index+1, len(jobs), job.PageText))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_list_page_start",
			Message:           fmt.Sprintf("开始抓取列表页：%s", job.PageText),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: index,
			PhaseTotalCount:   len(jobs),
		})

		body, fetchErr := s.fetchWithSafeWithOptions(ctx, cfg, state, job.URL, true)
		if fetchErr != nil {
			result.FailedCount++
			result.LastError = fetchErr.Error()
			run.LastError = fetchErr.Error()
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if updateErr := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); updateErr != nil {
				return updateErr
			}
			result.CompletedPageCount = int64(index + 1)
			reportErrorLog(ctx, fmt.Sprintf("列表页抓取失败: %s | err=%v", job.PageText, fetchErr))
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_list_page_failed",
				Message:           fmt.Sprintf("列表页抓取失败：%s，错误=%v", job.PageText, fetchErr),
				CurrentPhaseKey:   "list",
				PhaseKey:          "list",
				PhaseHandledCount: index + 1,
				PhaseTotalCount:   len(jobs),
				PhaseFailedCount:  int(result.FailedCount),
			})
			continue
		}

		topics, skippedTopCount, err := parseSHPageTopics(SHFetchModeForum, body, job)
		if err != nil {
			result.FailedCount++
			result.LastError = err.Error()
			run.LastError = err.Error()
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if updateErr := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); updateErr != nil {
				return updateErr
			}
			result.CompletedPageCount = int64(index + 1)
			reportErrorLog(ctx, fmt.Sprintf("列表页解析失败: %s | err=%v", job.PageText, err))
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_list_page_failed",
				Message:           fmt.Sprintf("列表页解析失败：%s，错误=%v", job.PageText, err),
				CurrentPhaseKey:   "list",
				PhaseKey:          "list",
				PhaseHandledCount: index + 1,
				PhaseTotalCount:   len(jobs),
				PhaseFailedCount:  int(result.FailedCount),
			})
			continue
		}

		result.HandledPageCount++
		for _, topic := range topics {
			if topic == nil || topic.ThreadTID == "" {
				continue
			}
			result.SeenCount++
			if _, ok := seenThreads[topic.ThreadTID]; ok {
				continue
			}
			seenThreads[topic.ThreadTID] = struct{}{}
			result.MatchedCount++
			if err = s.saveSHListTopic(ctx, req, run, topic); err != nil {
				result.FailedCount++
				result.LastError = err.Error()
				run.LastError = err.Error()
				reportErrorLog(ctx, fmt.Sprintf("列表帖子基础入库失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
			} else {
				result.FetchedCount++
			}
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if err = s.deps.ShSehuatangFetchRunModel.Update(ctx, run); err != nil {
				return err
			}
		}

		reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池页完成: run_id=%d | %s | topics=%d | skipped_top=%d | seen=%d | saved=%d", run.Id, job.PageText, len(topics), skippedTopCount, result.SeenCount, result.FetchedCount))
		result.CompletedPageCount = int64(index + 1)
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_list_page_done",
			Message:           fmt.Sprintf("列表页完成：%s，帖子=%d，跳过置顶=%d，累计入库=%d", job.PageText, len(topics), skippedTopCount, result.FetchedCount),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: index + 1,
			PhaseTotalCount:   len(jobs),
		})
	}
	return nil
}

func normalizeSHDetailFetchRequest(req SHDetailFetchRequest) SHDetailFetchRequest {
	out := req
	out.Keyword = strings.TrimSpace(out.Keyword)
	out.TargetForumIDs = normalizeSHTargetForumIDs(out.TargetForumIDs)
	out.DetailStatus = normalizeSHDetailStatus(out.DetailStatus)
	out.ThreadTIDs = uniqueSHStrings(out.ThreadTIDs)
	return out
}

func validateSHDetailFetchRequest(req SHDetailFetchRequest) error {
	if len(req.ThreadTIDs) == 0 {
		return fmt.Errorf("请选择要抓取详情的帖子")
	}
	return nil
}

func normalizeSHDetailStatus(raw string) string {
	switch strings.TrimSpace(raw) {
	case SHDetailStatusFetched:
		return SHDetailStatusFetched
	case SHDetailStatusFailed:
		return SHDetailStatusFailed
	default:
		return SHDetailStatusPending
	}
}

func uniqueSHStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func shThreadRowToTopic(row *moviex.ShSehuatangThread) *topicLink {
	if row == nil {
		return nil
	}
	listPostAt := ""
	if row.PostTime > 0 {
		listPostAt = time.Unix(row.PostTime, 0).Format("2006-01-02 15:04")
	}
	return &topicLink{
		Title:         row.ThreadTitle,
		DetailURL:     row.ThreadURL,
		ThreadTID:     row.ThreadTID,
		AuthorName:    row.AuthorName,
		ListPostAt:    listPostAt,
		SourceForumID: row.SourceForumID,
		SourcePage:    row.SourcePage,
		SourcePageURL: row.SourcePageURL,
	}
}
