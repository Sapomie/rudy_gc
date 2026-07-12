package fetchsehuatang

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

type shPageJob struct {
	URL      string
	Page     int64
	ForumID  string
	PageText string
	Body     []byte
}

type shRunRef struct {
	ID     int64
	Source string
}

func BuildSHProgressStats(result *SHFetchResult) map[string]int64 {
	stats := map[string]int64{
		"total_page_count":       0,
		"completed_page_count":   0,
		"current_page":           0,
		"handled_page_count":     0,
		"seen_count":             0,
		"matched_count":          0,
		"skipped_forum_count":    0,
		"skipped_existing_count": 0,
		"fetched_count":          0,
		"failed_count":           0,
	}
	if result == nil {
		return stats
	}
	stats["total_page_count"] = result.TotalPageCount
	stats["completed_page_count"] = result.CompletedPageCount
	stats["current_page"] = result.CurrentPage
	stats["handled_page_count"] = result.HandledPageCount
	stats["seen_count"] = result.SeenCount
	stats["matched_count"] = result.MatchedCount
	stats["skipped_forum_count"] = result.SkippedForumCount
	stats["skipped_existing_count"] = result.SkippedExistingCount
	stats["fetched_count"] = result.FetchedCount
	stats["failed_count"] = result.FailedCount
	return stats
}

func BuildSHProgressStatsText(result *SHFetchResult) map[string]string {
	if result == nil {
		return nil
	}
	currentPageText := strings.TrimSpace(result.CurrentPageText)
	if currentPageText == "" {
		return nil
	}
	return map[string]string{
		"current_page_text": currentPageText,
	}
}

func reportSHProgress(ctx context.Context, result *SHFetchResult, progress taskctx.Progress) {
	progress.Stats = BuildSHProgressStats(result)
	progress.StatsText = BuildSHProgressStatsText(result)
	taskctx.ReportProgress(ctx, progress)
}

func (s *Service) FetchSHThreads(ctx context.Context, req SHFetchRequest) (*SHFetchResult, error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}
	if s.deps.ShSehuatangThreadModel == nil || s.deps.ShSehuatangFetchRunModel == nil || s.deps.ShSehuatangFetchItemModel == nil {
		return nil, fmt.Errorf("sh sehuatang models are nil")
	}

	req = normalizeSHFetchRequest(req)
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
		FetchMode:      req.FetchMode,
		Keyword:        req.Keyword,
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
		FetchMode:       req.FetchMode,
		Keyword:         req.Keyword,
		TargetForumIDs:  targetForumIDs,
		StartPage:       req.StartPage,
		EndPage:         req.EndPage,
		TotalPageCount:  int64(len(jobs)),
		CurrentPage:     jobs[0].Page,
		CurrentPageText: jobs[0].PageText,
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂新流程开始: run_id=%d | mode=%s | forums=%s | page=%d-%d", runID, req.FetchMode, targetForumIDs, req.StartPage, req.EndPage))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:           "fetch_sehuatang_sh_start",
		Message:         fmt.Sprintf("开始色花堂新流程，run_id=%d，模式=%s，版块=%s，页码=%d-%d", runID, req.FetchMode, targetForumIDs, req.StartPage, req.EndPage),
		CurrentPhaseKey: "list",
	})

	fetchErr := s.fetchSHPages(ctx, cfg, state, req, jobs, run, result)
	if aggErr := s.flushSHThreadAgg(context.Background(), "sh_fetch", result); aggErr != nil {
		fetchErr = errors.Join(fetchErr, aggErr)
	}
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

	reportInfoLog(ctx, fmt.Sprintf("色花堂新流程完成: run_id=%d | seen=%d | matched=%d | skipped_forum=%d | skipped_existing=%d | fetched=%d | failed=%d", runID, result.SeenCount, result.MatchedCount, result.SkippedForumCount, result.SkippedExistingCount, result.FetchedCount, result.FailedCount))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_done",
		Message:      fmt.Sprintf("色花堂新流程完成：发现=%d，命中=%d，跳过版块=%d，跳过已存在=%d，入库=%d，失败=%d", result.SeenCount, result.MatchedCount, result.SkippedForumCount, result.SkippedExistingCount, result.FetchedCount, result.FailedCount),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return result, nil
}

func (s *Service) fetchSHPages(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest, jobs []shPageJob, run *moviex.ShSehuatangFetchRun, result *SHFetchResult) error {
	targetForums := targetForumSet(req.TargetForumIDs)
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

		reportInfoLog(ctx, fmt.Sprintf("色花堂新流程列表抓取: run_id=%d | %d/%d | %s", run.Id, index+1, len(jobs), job.PageText))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_page_start",
			Message:           fmt.Sprintf("开始抓取列表页：%s", job.PageText),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: index,
			PhaseTotalCount:   len(jobs),
		})

		body := job.Body
		var fetchErr error
		if len(body) > 0 {
			reportInfoLog(ctx, fmt.Sprintf("复用搜索初始化结果: run_id=%d | %s", run.Id, job.PageText))
		} else {
			body, fetchErr = s.fetchWithSafeWithOptions(ctx, cfg, state, job.URL, true)
		}
		if fetchErr != nil {
			result.FailedCount++
			result.LastError = fetchErr.Error()
			run.LastError = fetchErr.Error()
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if updateErr := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); updateErr != nil {
				return updateErr
			}
			reportErrorLog(ctx, fmt.Sprintf("列表页抓取失败: %s | err=%v", job.PageText, fetchErr))
			result.CompletedPageCount = int64(index + 1)
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_page_failed",
				Message:           fmt.Sprintf("列表页抓取失败：%s，错误=%v", job.PageText, fetchErr),
				CurrentPhaseKey:   "list",
				PhaseKey:          "list",
				PhaseHandledCount: index + 1,
				PhaseTotalCount:   len(jobs),
				PhaseFailedCount:  int(result.FailedCount),
			})
			continue
		}

		topics, skippedTopCount, err := parseSHPageTopics(req.FetchMode, body, job)
		if err != nil {
			result.FailedCount++
			result.LastError = err.Error()
			run.LastError = err.Error()
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if updateErr := s.deps.ShSehuatangFetchRunModel.Update(ctx, run); updateErr != nil {
				return updateErr
			}
			reportErrorLog(ctx, fmt.Sprintf("列表页解析失败: %s | err=%v", job.PageText, err))
			result.CompletedPageCount = int64(index + 1)
			reportSHProgress(ctx, result, taskctx.Progress{
				Stage:             "fetch_sehuatang_sh_page_failed",
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
			if req.FetchMode == SHFetchModeSearch {
				if _, ok := targetForums[topic.SourceForumID]; !ok {
					result.SkippedForumCount++
					continue
				}
			}
			if _, ok := seenThreads[topic.ThreadTID]; ok {
				continue
			}
			seenThreads[topic.ThreadTID] = struct{}{}
			result.MatchedCount++
			if err = s.fetchSHTopicDetail(ctx, cfg, state, req, run, result, topic); err != nil {
				return err
			}
			run.UpdatedOn = time.Now().Unix()
			syncSHRunCounts(run, result)
			if err = s.deps.ShSehuatangFetchRunModel.Update(ctx, run); err != nil {
				return err
			}
		}

		reportInfoLog(ctx, fmt.Sprintf("色花堂新流程列表完成: run_id=%d | %s | topics=%d | skipped_top=%d | seen=%d | matched=%d", run.Id, job.PageText, len(topics), skippedTopCount, result.SeenCount, result.MatchedCount))
		result.CompletedPageCount = int64(index + 1)
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:             "fetch_sehuatang_sh_page_done",
			Message:           fmt.Sprintf("列表页完成：%s，帖子=%d，跳过置顶=%d，累计命中=%d", job.PageText, len(topics), skippedTopCount, result.MatchedCount),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: index + 1,
			PhaseTotalCount:   len(jobs),
		})
	}
	return nil
}

func (s *Service) saveSHListTopic(ctx context.Context, req SHFetchRequest, run shRunRef, result *SHFetchResult, topic *topicLink) error {
	if run.ID <= 0 || strings.TrimSpace(run.Source) == "" || topic == nil || topic.ThreadTID == "" {
		return fmt.Errorf("帖子基础信息为空")
	}

	now := time.Now().Unix()
	item := &moviex.ShSehuatangFetchItem{
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
		Status:        SHFetchItemStatusDiscovered,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if _, err := s.deps.ShSehuatangFetchItemModel.Upsert(ctx, item); err != nil {
		return err
	}

	threadRow, err := s.deps.ShSehuatangThreadModel.UpsertListBasic(ctx, &moviex.ShSehuatangThread{
		ThreadTID:      topic.ThreadTID,
		ThreadURL:      topic.DetailURL,
		ThreadTitle:    topic.Title,
		AuthorName:     topic.AuthorName,
		PostTime:       parseForumTimeText(topic.ListPostAt, time.Now()),
		SourceForumID:  topic.SourceForumID,
		SourcePage:     topic.SourcePage,
		SourcePageURL:  topic.SourcePageURL,
		DetailStatus:   SHDetailStatusPending,
		RawContentText: "",
		MagnetText:     "",
		InfoHashText:   "",
		ContentHash:    "",
		FirstSeenTime:  now,
		LastSeenTime:   now,
		CreatedOn:      now,
		UpdatedOn:      now,
	})
	if err != nil {
		if _, updateErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, 0, err.Error(), time.Now().Unix()); updateErr != nil {
			return updateErr
		}
		return err
	}

	threadRowID := int64(0)
	if threadRow != nil {
		threadRowID = threadRow.Id
	}
	markSHThreadAggRow(result, threadRow)
	if _, err = s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusDiscovered, threadRowID, "", time.Now().Unix()); err != nil {
		return err
	}
	return nil
}

func (s *Service) fetchSHTopicDetail(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest, run *moviex.ShSehuatangFetchRun, result *SHFetchResult, topic *topicLink) error {
	return s.fetchSHTopicDetailWithOptions(ctx, cfg, state, req, shRunRef{ID: run.Id, Source: SHRunSourceFetch}, result, topic, true)
}

func (s *Service) fetchSHTopicDetailWithOptions(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest, run shRunRef, result *SHFetchResult, topic *topicLink, skipExisting bool) error {
	now := time.Now().Unix()
	item := &moviex.ShSehuatangFetchItem{
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
		Status:        SHFetchItemStatusDiscovered,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if _, err := s.deps.ShSehuatangFetchItemModel.Upsert(ctx, item); err != nil {
		return err
	}

	oldThread, err := s.deps.ShSehuatangThreadModel.FindOneByThreadTID(ctx, topic.ThreadTID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		result.FailedCount++
		result.LastError = err.Error()
		if _, updateErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, 0, err.Error(), time.Now().Unix()); updateErr != nil {
			return updateErr
		}
		return nil
	}
	if skipExisting && oldThread != nil {
		updatedThread, updateErr := s.deps.ShSehuatangThreadModel.UpdateLastSeenByThreadTID(ctx, topic.ThreadTID, time.Now().Unix())
		if updateErr != nil {
			result.FailedCount++
			result.LastError = updateErr.Error()
			if _, itemErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, oldThread.Id, updateErr.Error(), time.Now().Unix()); itemErr != nil {
				return itemErr
			}
			return nil
		}
		threadRowID := oldThread.Id
		if updatedThread != nil {
			threadRowID = updatedThread.Id
		}
		if _, err = s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusSkippedExisting, threadRowID, "", time.Now().Unix()); err != nil {
			return err
		}
		result.SkippedExistingCount++
		reportInfoLog(ctx, fmt.Sprintf("跳过已存在帖子: run_id=%d | tid=%s | %s", run.ID, topic.ThreadTID, topic.Title))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_skip_existing",
			Message:      fmt.Sprintf("跳过已存在帖子：%s", topic.Title),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
		})
		return nil
	}
	markSHThreadAggRow(result, oldThread)

	reportInfoLog(ctx, fmt.Sprintf("色花堂新流程详情抓取: run_id=%d | tid=%s | %s", run.ID, topic.ThreadTID, topic.Title))
	body, err := s.fetchWithSafeWithOptions(ctx, cfg, state, topic.DetailURL, true)
	if err != nil {
		result.FailedCount++
		result.LastError = err.Error()
		if _, updateErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, 0, err.Error(), time.Now().Unix()); updateErr != nil {
			return updateErr
		}
		updatedThread, updateErr := s.deps.ShSehuatangThreadModel.UpdateDetailFailureByThreadTID(ctx, topic.ThreadTID, err.Error(), time.Now().Unix())
		if updateErr != nil && !errors.Is(updateErr, moviex.ErrNotFound) {
			return updateErr
		}
		markSHThreadAggRow(result, updatedThread)
		reportErrorLog(ctx, fmt.Sprintf("详情抓取失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_detail_failed",
			Message:      fmt.Sprintf("详情抓取失败：%s，错误=%v", topic.Title, err),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
		})
		return nil
	}

	detail, err := parseSHThreadDetail(body, topic, time.Now())
	if err != nil {
		result.FailedCount++
		result.LastError = err.Error()
		if _, updateErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, 0, err.Error(), time.Now().Unix()); updateErr != nil {
			return updateErr
		}
		updatedThread, updateErr := s.deps.ShSehuatangThreadModel.UpdateDetailFailureByThreadTID(ctx, topic.ThreadTID, err.Error(), time.Now().Unix())
		if updateErr != nil && !errors.Is(updateErr, moviex.ErrNotFound) {
			return updateErr
		}
		markSHThreadAggRow(result, updatedThread)
		reportErrorLog(ctx, fmt.Sprintf("详情解析失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_detail_failed",
			Message:      fmt.Sprintf("详情解析失败：%s，错误=%v", topic.Title, err),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
		})
		return nil
	}
	if detail.ThreadTID == "" {
		detail.ThreadTID = topic.ThreadTID
	}
	if detail.ThreadTitle == "" {
		detail.ThreadTitle = topic.Title
	}
	if detail.AuthorName == "" {
		detail.AuthorName = topic.AuthorName
	}
	if detail.PostTime <= 0 {
		detail.PostTime = parseForumTimeText(topic.ListPostAt, time.Now())
	}

	detailFetchedOn := time.Now().Unix()
	threadRow, err := s.deps.ShSehuatangThreadModel.Upsert(ctx, &moviex.ShSehuatangThread{
		ThreadTID:       detail.ThreadTID,
		ThreadURL:       topic.DetailURL,
		ThreadTitle:     detail.ThreadTitle,
		AuthorName:      detail.AuthorName,
		PostTime:        detail.PostTime,
		SourceForumID:   topic.SourceForumID,
		SourcePage:      topic.SourcePage,
		SourcePageURL:   topic.SourcePageURL,
		DetailStatus:    SHDetailStatusFetched,
		DetailFetchedOn: detailFetchedOn,
		DetailFailedOn:  0,
		DetailError:     "",
		RawContentText:  detail.RawContentText,
		MagnetText:      detail.MagnetText,
		InfoHashText:    detail.InfoHashText,
		ContentHash:     contentHash(detail.RawContentText),
		FirstSeenTime:   now,
		LastSeenTime:    detailFetchedOn,
		CreatedOn:       now,
		UpdatedOn:       detailFetchedOn,
	})
	if err != nil {
		result.FailedCount++
		result.LastError = err.Error()
		if _, updateErr := s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFailed, 0, err.Error(), time.Now().Unix()); updateErr != nil {
			return updateErr
		}
		updatedThread, updateErr := s.deps.ShSehuatangThreadModel.UpdateDetailFailureByThreadTID(ctx, topic.ThreadTID, err.Error(), time.Now().Unix())
		if updateErr != nil && !errors.Is(updateErr, moviex.ErrNotFound) {
			return updateErr
		}
		markSHThreadAggRow(result, updatedThread)
		reportErrorLog(ctx, fmt.Sprintf("帖子入库失败: tid=%s | %s | err=%v", topic.ThreadTID, topic.Title, err))
		reportSHProgress(ctx, result, taskctx.Progress{
			Stage:        "fetch_sehuatang_sh_detail_failed",
			Message:      fmt.Sprintf("帖子入库失败：%s，错误=%v", topic.Title, err),
			HandledCount: int(result.MatchedCount),
			SuccessCount: int(result.FetchedCount),
			FailedCount:  int(result.FailedCount),
		})
		return nil
	}

	threadRowID := int64(0)
	if threadRow != nil {
		threadRowID = threadRow.Id
	}
	markSHThreadAggRow(result, threadRow)
	if _, err = s.deps.ShSehuatangFetchItemModel.UpdateStatus(ctx, run.Source, run.ID, topic.ThreadTID, SHFetchItemStatusFetched, threadRowID, "", time.Now().Unix()); err != nil {
		return err
	}

	result.FetchedCount++
	reportInfoLog(ctx, fmt.Sprintf("帖子详情入库完成: run_id=%d | tid=%s | title=%s | post_time=%d | magnets=%d", run.ID, detail.ThreadTID, detail.ThreadTitle, detail.PostTime, countLines(detail.InfoHashText)))
	reportSHProgress(ctx, result, taskctx.Progress{
		Stage:        "fetch_sehuatang_sh_detail_done",
		Message:      fmt.Sprintf("帖子详情入库完成：%s", detail.ThreadTitle),
		HandledCount: int(result.MatchedCount),
		SuccessCount: int(result.FetchedCount),
		FailedCount:  int(result.FailedCount),
	})
	return nil
}

func normalizeSHFetchRequest(req SHFetchRequest) SHFetchRequest {
	out := req
	out.FetchMode = strings.TrimSpace(out.FetchMode)
	switch out.FetchMode {
	case SHFetchModeForum:
	default:
		out.FetchMode = SHFetchModeSearch
	}
	out.Keyword = strings.TrimSpace(out.Keyword)
	out.TargetForumIDs = normalizeSHTargetForumIDs(out.TargetForumIDs)
	if out.StartPage <= 0 {
		out.StartPage = DefaultStartPage
	}
	if out.EndPage <= 0 {
		out.EndPage = DefaultEndPage
	}
	return out
}

func normalizeSHTargetForumIDs(raw []string) []string {
	allowed := map[string]struct{}{
		SHForumAsiaUncensoredOriginal: {},
		SHForumHDChineseSubtitle:      {},
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, 2)
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if _, ok := allowed[item]; !ok {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	if len(out) == 0 {
		return []string{SHForumAsiaUncensoredOriginal, SHForumHDChineseSubtitle}
	}
	return out
}

func validateSHFetchRequest(req SHFetchRequest) error {
	if req.FetchMode == SHFetchModeSearch && req.Keyword == "" {
		return fmt.Errorf("keyword 不能为空")
	}
	if req.StartPage <= 0 {
		return fmt.Errorf("start_page 必须大于 0")
	}
	if req.EndPage <= 0 {
		return fmt.Errorf("end_page 必须大于 0")
	}
	if req.EndPage < req.StartPage {
		return fmt.Errorf("end_page 不能小于 start_page")
	}
	return nil
}

func (s *Service) buildSHPageJobs(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, req SHFetchRequest) ([]shPageJob, error) {
	baseURL := cfg.BaseURL
	jobs := make([]shPageJob, 0, len(req.TargetForumIDs)*int(req.EndPage-req.StartPage+1))
	if req.FetchMode == SHFetchModeSearch {
		initURL, err := buildSHSearchSubmitURL(baseURL, req.Keyword)
		if err != nil {
			return nil, err
		}
		initResp, err := s.fetchWithSafeResultWithOptions(ctx, cfg, state, initURL, true)
		if err != nil {
			return nil, err
		}
		resultURL := strings.TrimSpace(initResp.FinalURL)
		if resultURL == "" {
			resultURL = initURL
		}
		reportInfoLog(ctx, fmt.Sprintf("色花堂搜索初始化完成: keyword=%s | result_url=%s", req.Keyword, resultURL))
		for pageNo := req.StartPage; pageNo <= req.EndPage; pageNo++ {
			pageURL := buildListPageURL(resultURL, pageNo)
			if pageURL == "" {
				return nil, fmt.Errorf("构造搜索结果页失败: page=%d", pageNo)
			}
			job := shPageJob{
				URL:      pageURL,
				Page:     pageNo,
				PageText: fmt.Sprintf("search=%s page=%d", req.Keyword, pageNo),
			}
			if req.StartPage == 1 && pageNo == 1 && len(initResp.Body) > 0 {
				job.Body = initResp.Body
			}
			jobs = append(jobs, job)
		}
		return jobs, nil
	}

	for pageNo := req.StartPage; pageNo <= req.EndPage; pageNo++ {
		for _, forumID := range req.TargetForumIDs {
			pageURL, err := buildSHForumPageURL(baseURL, forumID, pageNo)
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, shPageJob{
				URL:      pageURL,
				Page:     pageNo,
				ForumID:  forumID,
				PageText: fmt.Sprintf("forum=%s page=%d", forumID, pageNo),
			})
		}
	}
	return jobs, nil
}

func buildSHForumPageURL(baseURL string, forumID string, pageNo int64) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	forumID = strings.TrimSpace(forumID)
	if baseURL == "" || forumID == "" || pageNo <= 0 {
		return "", fmt.Errorf("构造版块页失败")
	}
	return fmt.Sprintf("%s/forum-%s-%d.html", baseURL, forumID, pageNo), nil
}

func buildSHSearchSubmitURL(baseURL string, keyword string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	keyword = strings.TrimSpace(keyword)
	if baseURL == "" || keyword == "" {
		return "", fmt.Errorf("构造搜索初始化页失败")
	}
	parsed, err := url.Parse(baseURL + "/search.php")
	if err != nil {
		return "", err
	}
	values := parsed.Query()
	values.Set("mod", "forum")
	values.Set("srchtxt", keyword)
	values.Set("searchsubmit", "yes")
	values.Set("orderby", "lastpost")
	values.Set("ascdesc", "desc")
	values.Set("srchtype", "title")
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func buildSHSearchPageURL(baseURL string, keyword string, pageNo int64) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	keyword = strings.TrimSpace(keyword)
	if baseURL == "" || keyword == "" || pageNo <= 0 {
		return "", fmt.Errorf("构造搜索页失败")
	}
	parsed, err := url.Parse(baseURL + "/search.php")
	if err != nil {
		return "", err
	}
	values := parsed.Query()
	values.Set("mod", "forum")
	values.Set("orderby", "lastpost")
	values.Set("ascdesc", "desc")
	values.Set("searchsubmit", "yes")
	values.Set("kw", keyword)
	values.Set("page", strconv.FormatInt(pageNo, 10))
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func parseSHPageTopics(fetchMode string, body []byte, job shPageJob) ([]*topicLink, int, error) {
	switch fetchMode {
	case SHFetchModeForum:
		parsed, err := parseSHForumTopics(body, job.URL, job.ForumID, job.Page)
		if err != nil {
			return nil, 0, err
		}
		return parsed.Topics, parsed.SkippedTopCount, nil
	default:
		topics, err := parseSHSearchTopics(body, job.URL, job.Page)
		return topics, 0, err
	}
}

func targetForumSet(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		out[id] = struct{}{}
	}
	return out
}

func syncSHRunCounts(run *moviex.ShSehuatangFetchRun, result *SHFetchResult) {
	if run == nil || result == nil {
		return
	}
	run.SeenCount = result.SeenCount
	run.MatchedCount = result.MatchedCount
	run.SkippedExistingCount = result.SkippedExistingCount
	run.SkippedForumCount = result.SkippedForumCount
	run.FetchedCount = result.FetchedCount
	run.FailedCount = result.FailedCount
	if result.LastError != "" {
		run.LastError = result.LastError
	}
}

func contentHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func countLines(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return len(strings.Split(text, "\n"))
}
