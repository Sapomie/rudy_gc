package fetchsehuatang

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

type SHDetailFetchRequest struct {
	TriggerSource     string            `json:"trigger_source"`
	SelectionCriteria string            `json:"selection_criteria"`
	TargetForumIDs    []string          `json:"target_forum_ids"`
	Keyword           string            `json:"keyword"`
	DetailStatus      string            `json:"detail_status"`
	ThreadTIDs        []string          `json:"thread_tids"`
	PostTimeFrom      int64             `json:"post_time_from"`
	HasPostTimeFrom   bool              `json:"has_post_time_from"`
	PostTimeTo        int64             `json:"post_time_to"`
	HasPostTimeTo     bool              `json:"has_post_time_to"`
	BaseStats         map[string]int64  `json:"-"`
	BaseStatsText     map[string]string `json:"-"`
}

type shListPageCoverage struct {
	Page                 int64  `json:"page"`
	URL                  string `json:"url"`
	Status               string `json:"status"`
	PostTimeFrom         int64  `json:"post_time_from"`
	PostTimeTo           int64  `json:"post_time_to"`
	FirstThreadTID       string `json:"first_thread_tid"`
	LastThreadTID        string `json:"last_thread_tid"`
	SeenCount            int64  `json:"seen_count"`
	InsertedCount        int64  `json:"inserted_count"`
	ExistingCount        int64  `json:"existing_count"`
	SkippedExistingCount int64  `json:"skipped_existing_count"`
	FailedCount          int64  `json:"failed_count"`
	StartedOn            int64  `json:"started_on"`
	FinishedOn           int64  `json:"finished_on"`
	Error                string `json:"error"`
}

type shListForumCoverage struct {
	PostTimeFrom    int64                `json:"post_time_from"`
	PostTimeTo      int64                `json:"post_time_to"`
	LastSuccessPage int64                `json:"last_success_page"`
	Pages           []shListPageCoverage `json:"pages"`
}

type shListCoverage map[string]*shListForumCoverage

func (s *Service) FetchSHThreadList(ctx context.Context, req SHFetchRequest) (result *SHFetchResult, err error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}
	if s.deps.ShSehuatangThreadModel == nil || s.deps.ShSehuatangThreadListFetchRunModel == nil || s.deps.ShSehuatangFetchItemModel == nil {
		return nil, fmt.Errorf("sh sehuatang models are nil")
	}
	jobID := taskctx.JobID(ctx)
	if jobID == "" {
		return nil, fmt.Errorf("job_id 不能为空")
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
	run := &moviex.ShSehuatangThreadListFetchRun{
		JobID:          jobID,
		TargetForumIDs: targetForumIDs,
		StartPage:      req.StartPage,
		EndPage:        req.EndPage,
		CurrentForumID: jobs[0].ForumID,
		CurrentPage:    jobs[0].Page,
		CurrentURL:     jobs[0].URL,
		TotalPageCount: int64(len(jobs)),
		PageCoverage:   "{}",
		Status:         SHRunStatusRunning,
		StartedOn:      now,
		CreatedOn:      now,
		UpdatedOn:      now,
	}
	insertResult, err := s.deps.ShSehuatangThreadListFetchRunModel.Insert(ctx, run)
	if err != nil {
		return nil, err
	}
	runID, err := insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	run.Id = runID
	defer func() {
		finishTime := time.Now().Unix()
		run.FinishedOn = finishTime
		run.UpdatedOn = finishTime
		if result != nil {
			syncSHListRunCounts(run, result)
		}
		if err != nil {
			run.Status = SHRunStatusFailed
			run.StopReason = "failed"
			run.LastError = err.Error()
			if ctx.Err() != nil {
				run.Status = SHRunStatusStopped
				run.StopReason = "context_canceled"
			}
		} else {
			run.Status = SHRunStatusSuccess
			run.StopReason = "range_completed"
		}
		run.CoverageComplete = run.CompletedPageCount == run.TotalPageCount && run.FailedPageCount == 0
		if updateErr := s.deps.ShSehuatangThreadListFetchRunModel.Update(context.Background(), run); updateErr != nil && err == nil {
			err = updateErr
		}
	}()

	result = &SHFetchResult{
		RunID:           runID,
		FetchMode:       SHFetchModeForumList,
		ExistingPolicy:  req.ExistingPolicy,
		TargetForumIDs:  targetForumIDs,
		StartPage:       req.StartPage,
		EndPage:         req.EndPage,
		TotalPageCount:  int64(len(jobs)),
		CurrentPage:     jobs[0].Page,
		CurrentPageText: jobs[0].PageText,
		CurrentForumID:  jobs[0].ForumID,
	}
	defer func() {
		if aggErr := s.flushSHThreadAgg(context.Background(), "sh_list", result); aggErr != nil && err == nil {
			err = aggErr
		}
	}()

	reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池开始: run_id=%d | forums=%s | page=%d-%d", runID, targetForumIDs, req.StartPage, req.EndPage))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:           "fetch_sehuatang_sh_list_start",
		Message:         fmt.Sprintf("开始色花堂列表建池，run_id=%d，版块=%s，页码=%d-%d", runID, targetForumIDs, req.StartPage, req.EndPage),
		CurrentPhaseKey: "list",
	})

	coverage := make(shListCoverage, len(req.TargetForumIDs))
	if err = s.fetchSHListPages(ctx, cfg, state, req, jobs, run, result, coverage); err != nil {
		result.LastError = err.Error()
		return result, err
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池完成: run_id=%d | pages=%d/%d | seen=%d | matched=%d | inserted=%d | existing=%d | skipped_existing=%d | persist_failed=%d | skipped_top=%d", runID, result.SuccessPageCount, result.FailedPageCount, result.SeenCount, result.MatchedCount, result.InsertedCount, result.ExistingCount, result.SkippedExistingCount, result.PersistFailedCount, result.SkippedTopCount))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_list_done",
		Message:      fmt.Sprintf("列表抓取完成：页面成功=%d，页面失败=%d，新增=%d，更新已有=%d，跳过已有=%d，入库失败=%d，跳过置顶=%d", result.SuccessPageCount, result.FailedPageCount, result.InsertedCount, result.ExistingCount, result.SkippedExistingCount, result.PersistFailedCount, result.SkippedTopCount),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return result, nil
}

func (s *Service) FetchSHThreadDetails(ctx context.Context, req SHDetailFetchRequest) (result *SHFetchResult, err error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}
	if s.deps.ShSehuatangThreadModel == nil || s.deps.ShMagnetModel == nil || s.deps.ShSehuatangThreadDetailFetchRunModel == nil || s.deps.ShSehuatangFetchItemModel == nil {
		return nil, fmt.Errorf("sh sehuatang models are nil")
	}
	jobID := taskctx.JobID(ctx)
	if jobID == "" {
		return nil, fmt.Errorf("job_id 不能为空")
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

	targetForumText := strings.Join(uniqueSHStrings(targetForumIDs), ",")
	selectedThreadTIDsJSON, err := json.Marshal(req.ThreadTIDs)
	if err != nil {
		return nil, err
	}
	selectionCriteria := strings.TrimSpace(req.SelectionCriteria)
	if selectionCriteria == "" {
		selectionCriteria = "{}"
	}
	now := time.Now().Unix()
	run := &moviex.ShSehuatangThreadDetailFetchRun{
		JobID:              jobID,
		TriggerSource:      normalizeSHDetailTriggerSource(req.TriggerSource),
		TargetForumIDs:     targetForumText,
		SelectionCriteria:  selectionCriteria,
		SelectedThreadTIDs: string(selectedThreadTIDsJSON),
		SelectedCount:      int64(len(rows)),
		Status:             SHRunStatusRunning,
		StartedOn:          now,
		CreatedOn:          now,
		UpdatedOn:          now,
	}
	insertResult, err := s.deps.ShSehuatangThreadDetailFetchRunModel.Insert(ctx, run)
	if err != nil {
		return nil, err
	}
	runID, err := insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	run.Id = runID
	defer func() {
		finishTime := time.Now().Unix()
		run.FinishedOn = finishTime
		run.UpdatedOn = finishTime
		if result != nil {
			run.CompletedCount = result.CompletedPageCount
			run.FetchedCount = result.FetchedCount
			run.FailedCount = result.FailedCount
			if result.LastError != "" {
				run.LastError = result.LastError
			}
		}
		if err != nil {
			run.Status = SHRunStatusFailed
			run.StopReason = "failed"
			run.LastError = err.Error()
			if ctx.Err() != nil {
				run.Status = SHRunStatusStopped
				run.StopReason = "context_canceled"
			}
		} else {
			run.Status = SHRunStatusSuccess
			run.StopReason = "completed"
		}
		if updateErr := s.deps.ShSehuatangThreadDetailFetchRunModel.Update(context.Background(), run); updateErr != nil && err == nil {
			err = updateErr
		}
	}()

	result = &SHFetchResult{
		RunID:           runID,
		FetchMode:       SHFetchModeDetail,
		TargetForumIDs:  targetForumText,
		TotalPageCount:  int64(len(rows)),
		CurrentPageText: "detail=0",
	}
	defer func() {
		if aggErr := s.flushSHThreadAgg(context.Background(), "sh_detail", result); aggErr != nil && err == nil {
			err = aggErr
		}
	}()
	reportInfoLog(ctx, fmt.Sprintf("色花堂选中帖子详情抓取开始: selected=%d", len(rows)))
	reportSHDetailProgress(ctx, result, req, taskctx.Progress{
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
		run.CurrentIndex = int64(index + 1)
		run.CurrentThreadTID = row.ThreadTID
		run.CurrentThreadURL = row.ThreadURL
		run.CompletedCount = int64(index)
		run.UpdatedOn = time.Now().Unix()
		if err := s.deps.ShSehuatangThreadDetailFetchRunModel.Update(ctx, run); err != nil {
			return result, err
		}

		reportSHDetailProgress(ctx, result, req, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_detail_item_start",
			Message:           fmt.Sprintf("开始抓取帖子详情：%s", row.ThreadTitle),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: index,
			PhaseTotalCount:   len(rows),
		})

		topic := shThreadRowToTopic(row)
		if err := s.fetchSHTopicDetailWithOptions(ctx, cfg, state, SHFetchRequest{FetchMode: SHFetchModeDetail}, shRunRef{ID: run.Id, Source: SHRunSourceDetail}, result, topic, false); err != nil {
			return result, err
		}

		result.CompletedPageCount = int64(index + 1)
		run.CompletedCount = result.CompletedPageCount
		run.FetchedCount = result.FetchedCount
		run.FailedCount = result.FailedCount
		run.LastError = result.LastError
		run.UpdatedOn = time.Now().Unix()
		if err := s.deps.ShSehuatangThreadDetailFetchRunModel.Update(ctx, run); err != nil {
			return result, err
		}
		reportSHDetailProgress(ctx, result, req, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_detail_item_done",
			Message:           fmt.Sprintf("帖子详情处理完成：%s", row.ThreadTitle),
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: index + 1,
			PhaseTotalCount:   len(rows),
		})
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂选中帖子详情抓取完成: selected=%d | fetched=%d | failed=%d", result.SeenCount, result.FetchedCount, result.FailedCount))
	reportSHDetailProgress(ctx, result, req, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_detail_done",
		Message:      fmt.Sprintf("色花堂选中帖子详情抓取完成：选中=%d，成功=%d，失败=%d", len(rows), result.FetchedCount, result.FailedCount),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return result, nil
}

func reportSHDetailProgress(ctx context.Context, result *SHFetchResult, req SHDetailFetchRequest, progress taskctx.Progress) {
	if len(req.BaseStats) == 0 {
		reportSHProgress(ctx, result, progress)
		return
	}
	stats := make(map[string]int64, len(req.BaseStats)+20)
	for key, value := range req.BaseStats {
		stats[key] = value
	}
	for key, value := range BuildSHProgressStats(result) {
		stats["detail_"+key] = value
	}
	statsText := make(map[string]string, len(req.BaseStatsText)+3)
	for key, value := range req.BaseStatsText {
		statsText[key] = value
	}
	for key, value := range BuildSHProgressStatsText(result) {
		statsText["detail_"+key] = value
	}
	progress.Stats = stats
	progress.StatsText = statsText
	taskctx.ReportProgress(ctx, progress)
}

func (s *Service) fetchSHListPages(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest, jobs []shPageJob, run *moviex.ShSehuatangThreadListFetchRun, result *SHFetchResult, coverage shListCoverage) error {
	seenThreads := make(map[string]struct{}, 128)
	for index, job := range jobs {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return err
		}

		pageStartedOn := time.Now().Unix()
		result.CurrentPage = job.Page
		result.CurrentPageText = job.PageText
		result.CurrentForumID = job.ForumID
		result.CompletedPageCount = int64(index)
		run.CurrentForumID = job.ForumID
		run.CurrentURL = job.URL
		run.CurrentPage = job.Page
		run.CompletedPageCount = int64(index)
		run.UpdatedOn = time.Now().Unix()
		if err := s.updateSHListRun(ctx, run, result, coverage); err != nil {
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
			result.FailedPageCount++
			result.LastError = fetchErr.Error()
			run.FailedPageCount++
			run.CompletedPageCount = int64(index + 1)
			run.LastError = fetchErr.Error()
			result.CompletedPageCount = int64(index + 1)
			setSHListPageCoverage(coverage, job.ForumID, shListPageCoverage{
				Page: job.Page, URL: job.URL, Status: SHRunStatusFailed,
				StartedOn: pageStartedOn, FinishedOn: time.Now().Unix(), Error: fetchErr.Error(),
			})
			if updateErr := s.updateSHListRun(ctx, run, result, coverage); updateErr != nil {
				return updateErr
			}
			reportErrorLog(ctx, fmt.Sprintf("列表页抓取失败: %s | err=%v", job.PageText, fetchErr))
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_list_page_failed",
				Message:           fmt.Sprintf("列表页抓取失败：%s，错误=%v", job.PageText, fetchErr),
				CurrentPhaseKey:   "list",
				PhaseKey:          "list",
				PhaseHandledCount: index + 1,
				PhaseTotalCount:   len(jobs),
				PhaseSuccessCount: int(result.SuccessPageCount),
				PhaseFailedCount:  int(result.FailedPageCount),
			})
			continue
		}

		topics, skippedTopCount, err := parseSHPageTopics(SHFetchModeForum, body, job)
		if err != nil {
			result.FailedCount++
			result.FailedPageCount++
			result.LastError = err.Error()
			run.FailedPageCount++
			run.CompletedPageCount = int64(index + 1)
			run.LastError = err.Error()
			result.CompletedPageCount = int64(index + 1)
			setSHListPageCoverage(coverage, job.ForumID, shListPageCoverage{
				Page: job.Page, URL: job.URL, Status: SHRunStatusFailed,
				StartedOn: pageStartedOn, FinishedOn: time.Now().Unix(), Error: err.Error(),
			})
			if updateErr := s.updateSHListRun(ctx, run, result, coverage); updateErr != nil {
				return updateErr
			}
			reportErrorLog(ctx, fmt.Sprintf("列表页解析失败: %s | err=%v", job.PageText, err))
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_list_page_failed",
				Message:           fmt.Sprintf("列表页解析失败：%s，错误=%v", job.PageText, err),
				CurrentPhaseKey:   "list",
				PhaseKey:          "list",
				PhaseHandledCount: index + 1,
				PhaseTotalCount:   len(jobs),
				PhaseSuccessCount: int(result.SuccessPageCount),
				PhaseFailedCount:  int(result.FailedPageCount),
			})
			continue
		}

		page := shListPageCoverage{
			Page:      job.Page,
			URL:       job.URL,
			Status:    SHRunStatusSuccess,
			StartedOn: pageStartedOn,
		}
		threadTIDs := make([]string, 0, len(topics))
		pageNow := time.Now()
		for _, topic := range topics {
			if topic == nil || strings.TrimSpace(topic.ThreadTID) == "" {
				continue
			}
			threadTIDs = append(threadTIDs, topic.ThreadTID)
			if page.FirstThreadTID == "" {
				page.FirstThreadTID = topic.ThreadTID
			}
			page.LastThreadTID = topic.ThreadTID
			postTime := parseForumTimeText(topic.ListPostAt, pageNow)
			if postTime > 0 && (page.PostTimeFrom == 0 || postTime < page.PostTimeFrom) {
				page.PostTimeFrom = postTime
			}
			if postTime > page.PostTimeTo {
				page.PostTimeTo = postTime
			}
		}
		existingTargets, err := s.deps.ShSehuatangThreadModel.FindThreadAggTargetsByTIDs(ctx, threadTIDs)
		if err != nil {
			return err
		}

		result.HandledPageCount++
		run.SkippedTopCount += int64(skippedTopCount)
		result.SkippedTopCount += int64(skippedTopCount)
		for _, topic := range topics {
			if topic == nil || topic.ThreadTID == "" {
				continue
			}
			result.SeenCount++
			page.SeenCount++
			if _, ok := seenThreads[topic.ThreadTID]; ok {
				continue
			}
			seenThreads[topic.ThreadTID] = struct{}{}
			result.MatchedCount++
			oldTarget, existed := existingTargets[topic.ThreadTID]
			if shListTopicShouldSkip(req.ExistingPolicy, existed) {
				result.SkippedExistingCount++
				page.SkippedExistingCount++
				if err = s.recordSHSkippedExistingListTopic(ctx, req, shRunRef{ID: run.Id, Source: SHRunSourceList}, topic); err != nil {
					result.FailedCount++
					result.PersistFailedCount++
					page.FailedCount++
					result.LastError = err.Error()
					run.LastError = err.Error()
					reportErrorLog(ctx, fmt.Sprintf("记录跳过已有帖子失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
				} else {
					reportInfoLog(ctx, fmt.Sprintf("跳过已存在帖子: run_id=%d | tid=%s | %s", run.Id, topic.ThreadTID, topic.Title))
				}
				continue
			}
			needsDetail := shListTopicNeedsDetail(req.ExistingPolicy, existed)
			if existed {
				markSHThreadAggTarget(result, oldTarget)
			}
			if err = s.saveSHListTopic(ctx, req, shRunRef{ID: run.Id, Source: SHRunSourceList}, result, topic); err != nil {
				result.FailedCount++
				result.PersistFailedCount++
				page.FailedCount++
				result.LastError = err.Error()
				run.LastError = err.Error()
				reportErrorLog(ctx, fmt.Sprintf("列表帖子基础入库失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
			} else {
				result.FetchedCount++
				if needsDetail {
					result.DetailThreadTIDs = append(result.DetailThreadTIDs, topic.ThreadTID)
				}
				if existed {
					page.ExistingCount++
					run.ExistingCount++
					result.ExistingCount++
				} else {
					page.InsertedCount++
					run.InsertedCount++
					result.InsertedCount++
				}
			}
		}

		reportInfoLog(ctx, fmt.Sprintf("色花堂列表建池页完成: run_id=%d | %s | topics=%d | skipped_top=%d | seen=%d | inserted=%d | existing=%d | skipped_existing=%d | persist_failed=%d", run.Id, job.PageText, len(topics), skippedTopCount, result.SeenCount, result.InsertedCount, result.ExistingCount, result.SkippedExistingCount, result.PersistFailedCount))
		result.CompletedPageCount = int64(index + 1)
		run.CompletedPageCount = int64(index + 1)
		run.SuccessPageCount++
		result.SuccessPageCount++
		page.FinishedOn = time.Now().Unix()
		if page.FailedCount > 0 {
			page.Error = result.LastError
		}
		setSHListPageCoverage(coverage, job.ForumID, page)
		if err = s.updateSHListRun(ctx, run, result, coverage); err != nil {
			return err
		}
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_list_page_done",
			Message:           fmt.Sprintf("列表页完成：%s，帖子=%d，新增=%d，更新已有=%d，跳过已有=%d，入库失败=%d，跳过置顶=%d", job.PageText, len(topics), result.InsertedCount, result.ExistingCount, result.SkippedExistingCount, result.PersistFailedCount, skippedTopCount),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: index + 1,
			PhaseTotalCount:   len(jobs),
			PhaseSuccessCount: int(result.SuccessPageCount),
			PhaseFailedCount:  int(result.FailedPageCount),
		})
	}
	return nil
}

func shListTopicNeedsDetail(existingPolicy string, existed bool) bool {
	return !existed || NormalizeSHExistingPolicy(existingPolicy) == SHExistingPolicyUpsert
}

func shListTopicShouldSkip(existingPolicy string, existed bool) bool {
	return existed && NormalizeSHExistingPolicy(existingPolicy) == SHExistingPolicySkip
}

func (s *Service) recordSHSkippedExistingListTopic(ctx context.Context, req SHFetchRequest, run shRunRef, topic *topicLink) error {
	if run.ID <= 0 || strings.TrimSpace(run.Source) == "" || topic == nil || strings.TrimSpace(topic.ThreadTID) == "" {
		return fmt.Errorf("跳过帖子基础信息为空")
	}
	now := time.Now().Unix()
	_, err := s.deps.ShSehuatangFetchItemModel.Upsert(ctx, &moviex.ShSehuatangFetchItem{
		RunSource:     run.Source,
		RunID:         run.ID,
		FetchMode:     req.FetchMode,
		SourceForumID: topic.SourceForumID,
		SourcePage:    topic.SourcePage,
		SourcePageURL: topic.SourcePageURL,
		ThreadTID:     topic.ThreadTID,
		ThreadURL:     topic.DetailURL,
		ThreadTitle:   topic.Title,
		AuthorName:    topic.AuthorName,
		Status:        SHFetchItemStatusSkippedExisting,
		CreatedOn:     now,
		UpdatedOn:     now,
	})
	return err
}

func syncSHListRunCounts(run *moviex.ShSehuatangThreadListFetchRun, result *SHFetchResult) {
	if run == nil || result == nil {
		return
	}
	run.SeenCount = result.SeenCount
	run.FailedCount = result.FailedCount
	if result.LastError != "" {
		run.LastError = result.LastError
	}
}

func (s *Service) updateSHListRun(ctx context.Context, run *moviex.ShSehuatangThreadListFetchRun, result *SHFetchResult, coverage shListCoverage) error {
	syncSHListRunCounts(run, result)
	run.CoveragePostTimeFrom = 0
	run.CoveragePostTimeTo = 0
	for _, forum := range coverage {
		if forum == nil {
			continue
		}
		if forum.PostTimeFrom > 0 && (run.CoveragePostTimeFrom == 0 || forum.PostTimeFrom < run.CoveragePostTimeFrom) {
			run.CoveragePostTimeFrom = forum.PostTimeFrom
		}
		if forum.PostTimeTo > run.CoveragePostTimeTo {
			run.CoveragePostTimeTo = forum.PostTimeTo
		}
	}
	pageCoverage, err := json.Marshal(coverage)
	if err != nil {
		return err
	}
	run.PageCoverage = string(pageCoverage)
	run.UpdatedOn = time.Now().Unix()
	return s.deps.ShSehuatangThreadListFetchRunModel.Update(ctx, run)
}

func setSHListPageCoverage(coverage shListCoverage, forumID string, page shListPageCoverage) {
	forumID = strings.TrimSpace(forumID)
	forum := coverage[forumID]
	if forum == nil {
		forum = &shListForumCoverage{Pages: make([]shListPageCoverage, 0)}
		coverage[forumID] = forum
	}
	replaced := false
	for index := range forum.Pages {
		if forum.Pages[index].Page == page.Page {
			forum.Pages[index] = page
			replaced = true
			break
		}
	}
	if !replaced {
		forum.Pages = append(forum.Pages, page)
	}
	if page.Status != SHRunStatusSuccess {
		return
	}
	forum.LastSuccessPage = page.Page
	if page.PostTimeFrom > 0 && (forum.PostTimeFrom == 0 || page.PostTimeFrom < forum.PostTimeFrom) {
		forum.PostTimeFrom = page.PostTimeFrom
	}
	if page.PostTimeTo > forum.PostTimeTo {
		forum.PostTimeTo = page.PostTimeTo
	}
}

func normalizeSHDetailFetchRequest(req SHDetailFetchRequest) SHDetailFetchRequest {
	out := req
	out.TriggerSource = normalizeSHDetailTriggerSource(out.TriggerSource)
	out.SelectionCriteria = strings.TrimSpace(out.SelectionCriteria)
	out.Keyword = strings.TrimSpace(out.Keyword)
	out.TargetForumIDs = normalizeSHTargetForumIDs(out.TargetForumIDs)
	out.DetailStatus = normalizeSHDetailStatus(out.DetailStatus)
	out.ThreadTIDs = uniqueSHStrings(out.ThreadTIDs)
	return out
}

func normalizeSHDetailTriggerSource(raw string) string {
	switch strings.TrimSpace(raw) {
	case "shpool":
		return "shpool"
	case "shfetchall":
		return "shfetchall"
	}
	return "shfill"
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
