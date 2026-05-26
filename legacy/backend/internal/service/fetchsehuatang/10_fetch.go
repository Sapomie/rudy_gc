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

func (s *Service) FetchMagnetsFromList(ctx context.Context, req FetchRequest) (*FetchResult, error) {
	if s == nil || s.deps == nil {
		return nil, fmt.Errorf("fetchsehuatang service is nil")
	}

	req = defaultRequest(req)
	req.ListURL = strings.TrimSpace(req.ListURL)
	req.Keyword = strings.TrimSpace(req.Keyword)
	if req.ListURL == "" {
		return nil, fmt.Errorf("list_url 不能为空")
	}
	if req.StartPage <= 0 {
		return nil, fmt.Errorf("start_page 必须大于 0")
	}
	if req.EndPage <= 0 {
		return nil, fmt.Errorf("end_page 必须大于 0")
	}
	if req.EndPage < req.StartPage {
		return nil, fmt.Errorf("end_page 不能小于 start_page")
	}

	cfg, ok := s.deps.GetFetchSiteConfig(siteCode())
	if !ok {
		return nil, fmt.Errorf("fetch site config not found: %s", siteCode())
	}
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("invalid timeout for site %s", siteCode())
	}
	if cfg.MaxRetryTimes <= 0 {
		return nil, fmt.Errorf("invalid max_retry_times for site %s", siteCode())
	}

	state := &remoteState{}
	if buildListPageURL(req.ListURL, req.StartPage) == "" {
		return nil, fmt.Errorf("list_url 不支持翻页: %s", req.ListURL)
	}
	totalPageCount := req.EndPage - req.StartPage + 1

	result := &FetchResult{
		ListURL:          req.ListURL,
		Keyword:          req.Keyword,
		StartPage:        req.StartPage,
		EndPage:          req.EndPage,
		HandledPageCount: 0,
		SkippedTopCount:  0,
		MatchedCount:     0,
		SuccessCount:     0,
		FailedCount:      0,
		InsertedCount:    0,
		UpdatedCount:     0,
		PersistFailCount: 0,
		SkippedExisting:  0,
		Items:            make([]*FetchTopicItem, 0, 64),
	}
	topics := make([]*topicLink, 0, 64)
	seenTopics := make(map[string]struct{}, 128)

	for pageNo := req.StartPage; pageNo <= req.EndPage; pageNo++ {
		pageIndex := pageNo - req.StartPage + 1
		pageURL := buildListPageURL(req.ListURL, pageNo)
		if pageURL == "" {
			return nil, fmt.Errorf("构造列表页失败: page=%d", pageNo)
		}

		reportInfoLog(ctx, fmt.Sprintf("色花堂列表抓取开始: 第 %d/%d 页 | page=%d | %s", pageIndex, totalPageCount, pageNo, pageURL))
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "fetch_sehuatang_list_start",
			Message:           fmt.Sprintf("开始抓取列表页: 第 %d/%d 页 | page=%d", pageIndex, totalPageCount, pageNo),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: int(pageIndex - 1),
			PhaseTotalCount:   int(totalPageCount),
			PhaseSuccessCount: int(pageIndex - 1),
			PhaseFailedCount:  0,
		})

		listBody, fetchErr := s.fetchWithSafe(ctx, cfg, state, pageURL)
		if fetchErr != nil {
			return nil, fetchErr
		}

		parsedList, parseErr := parseListTopics(listBody, pageURL, req.Keyword)
		if parseErr != nil {
			return nil, parseErr
		}
		pageTopics := parsedList.Topics

		addedCount := 0
		for _, topic := range pageTopics {
			if topic == nil {
				continue
			}
			if _, ok := seenTopics[topic.DetailURL]; ok {
				continue
			}
			seenTopics[topic.DetailURL] = struct{}{}
			topics = append(topics, topic)
			addedCount++
		}

		result.HandledPageCount++
		result.SkippedTopCount += parsedList.SkippedTopCount
		result.MatchedCount = len(topics)
		reportInfoLog(ctx, fmt.Sprintf("色花堂列表解析完成: 第 %d/%d 页 | page=%d | keyword=%s | 跳过置顶=%d | page_matched=%d | total_matched=%d", pageIndex, totalPageCount, pageNo, req.Keyword, parsedList.SkippedTopCount, len(pageTopics), len(topics)))
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "fetch_sehuatang_list_done",
			Message:           fmt.Sprintf("列表解析完成: 第 %d/%d 页 | page=%d | 跳过置顶=%d | 新增帖子=%d | 累计=%d", pageIndex, totalPageCount, pageNo, parsedList.SkippedTopCount, addedCount, len(topics)),
			HandledCount:      0,
			SuccessCount:      0,
			FailedCount:       0,
			QueuedCount:       len(topics),
			CurrentPhaseKey:   "list",
			PhaseKey:          "list",
			PhaseHandledCount: int(result.HandledPageCount),
			PhaseTotalCount:   int(totalPageCount),
			PhaseSuccessCount: int(result.HandledPageCount),
			PhaseFailedCount:  0,
		})
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:             "fetch_sehuatang_list_done",
		Message:           fmt.Sprintf("列表抓取完成，页码=%d-%d，共 %d 页，模式=%s，跳过置顶=%d，命中 %d 条帖子", result.StartPage, result.EndPage, result.HandledPageCount, req.PersistMode, result.SkippedTopCount, len(topics)),
		HandledCount:      0,
		SuccessCount:      0,
		FailedCount:       0,
		QueuedCount:       len(topics),
		CurrentPhaseKey:   "detail",
		PhaseKey:          "detail",
		PhaseHandledCount: 0,
		PhaseTotalCount:   len(topics),
		PhaseSuccessCount: 0,
		PhaseFailedCount:  0,
	})

	for _, topic := range topics {
		if topic == nil {
			continue
		}
		if waitErr := taskctx.WaitIfPaused(ctx); waitErr != nil {
			return nil, waitErr
		}

		item := &FetchTopicItem{
			Title:      topic.Title,
			DetailURL:  topic.DetailURL,
			MovieJavID: "",
			MovieName:  "",
			PostTime:   0,
			PostDate:   0,
			Magnets:    make([]string, 0, 4),
			InfoHashes: make([]string, 0, 4),
		}
		if req.PersistMode == PersistModeSkipOld {
			exists, existsErr := s.deps.SehuatangMagnetModel.ExistsByThreadURL(ctx, topic.DetailURL)
			if existsErr != nil {
				item.Error = existsErr.Error()
				result.FailedCount++
				result.Items = append(result.Items, item)
				reportErrorLog(ctx, fmt.Sprintf("帖子存在性检查失败: %s | err=%v", topic.Title, existsErr))
				taskctx.ReportProgress(ctx, taskctx.Progress{
					Stage:             "fetch_sehuatang_detail_failed",
					Message:           fmt.Sprintf("帖子存在性检查失败: %s", topic.Title),
					HandledCount:      result.SuccessCount + result.FailedCount + result.SkippedExisting,
					SuccessCount:      result.SuccessCount,
					FailedCount:       result.FailedCount,
					QueuedCount:       result.MatchedCount - result.SuccessCount - result.FailedCount - result.SkippedExisting,
					CurrentPhaseKey:   "detail",
					PhaseKey:          "detail",
					PhaseHandledCount: result.SuccessCount + result.FailedCount + result.SkippedExisting,
					PhaseTotalCount:   result.MatchedCount,
					PhaseSuccessCount: result.SuccessCount,
					PhaseFailedCount:  result.FailedCount,
				})
				continue
			}
			if exists {
				result.SkippedExisting++
				result.Items = append(result.Items, item)
				reportInfoLog(ctx, fmt.Sprintf("跳过已存在帖子: %s | %s", topic.Title, topic.DetailURL))
				taskctx.ReportProgress(ctx, taskctx.Progress{
					Stage:             "fetch_sehuatang_detail_skip_existing",
					Message:           fmt.Sprintf("跳过已存在帖子: %s", topic.Title),
					HandledCount:      result.SuccessCount + result.FailedCount + result.SkippedExisting,
					SuccessCount:      result.SuccessCount,
					FailedCount:       result.FailedCount,
					QueuedCount:       result.MatchedCount - result.SuccessCount - result.FailedCount - result.SkippedExisting,
					CurrentPhaseKey:   "detail",
					PhaseKey:          "detail",
					PhaseHandledCount: result.SuccessCount + result.FailedCount + result.SkippedExisting,
					PhaseTotalCount:   result.MatchedCount,
					PhaseSuccessCount: result.SuccessCount,
					PhaseFailedCount:  result.FailedCount,
				})
				continue
			}
		}
		reportInfoLog(ctx, fmt.Sprintf("开始抓取详情: %s", topic.Title))
		detailBody, fetchErr := s.fetchWithSafe(ctx, cfg, state, topic.DetailURL)
		if fetchErr != nil {
			item.Error = fetchErr.Error()
			result.FailedCount++
			result.Items = append(result.Items, item)
			reportErrorLog(ctx, fmt.Sprintf("详情抓取失败: %s | err=%v", topic.Title, fetchErr))
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_sehuatang_detail_failed",
				Message:           fmt.Sprintf("详情抓取失败: %s", topic.Title),
				HandledCount:      result.SuccessCount + result.FailedCount + result.SkippedExisting,
				SuccessCount:      result.SuccessCount,
				FailedCount:       result.FailedCount,
				QueuedCount:       result.MatchedCount - result.SuccessCount - result.FailedCount - result.SkippedExisting,
				CurrentPhaseKey:   "detail",
				PhaseKey:          "detail",
				PhaseHandledCount: result.SuccessCount + result.FailedCount + result.SkippedExisting,
				PhaseTotalCount:   result.MatchedCount,
				PhaseSuccessCount: result.SuccessCount,
				PhaseFailedCount:  result.FailedCount,
			})
			continue
		}

		now := time.Now()
		threadTitle := parseThreadTitle(detailBody, topic.Title)
		movieName := parseMovieName(threadTitle)
		if movieName == "" {
			movieName = parseMovieName(topic.Title)
		}
		movieJavID, resolveErr := s.resolveMovieJavIDByMovieName(ctx, movieName)
		if resolveErr != nil {
			item.Error = resolveErr.Error()
			result.FailedCount++
			result.Items = append(result.Items, item)
			reportErrorLog(ctx, fmt.Sprintf("影片映射失败: %s | movie_name=%s | err=%v", threadTitle, movieName, resolveErr))
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:             "fetch_sehuatang_detail_failed",
				Message:           fmt.Sprintf("影片映射失败: %s", threadTitle),
				HandledCount:      result.SuccessCount + result.FailedCount + result.SkippedExisting,
				SuccessCount:      result.SuccessCount,
				FailedCount:       result.FailedCount,
				QueuedCount:       result.MatchedCount - result.SuccessCount - result.FailedCount - result.SkippedExisting,
				CurrentPhaseKey:   "detail",
				PhaseKey:          "detail",
				PhaseHandledCount: result.SuccessCount + result.FailedCount + result.SkippedExisting,
				PhaseTotalCount:   result.MatchedCount,
				PhaseSuccessCount: result.SuccessCount,
				PhaseFailedCount:  result.FailedCount,
			})
			continue
		}
		postTime := parsePostTime(detailBody, topic.ListPostAt, now)
		postDate := parsePostDate(postTime, now)
		tag := parseThreadTag(threadTitle)
		parsedMagnets := parseMagnets(detailBody)

		item.Title = threadTitle
		item.MovieJavID = movieJavID
		item.MovieName = movieName
		item.PostTime = postTime
		item.PostDate = postDate
		for _, parsed := range parsedMagnets {
			item.InfoHashes = append(item.InfoHashes, parsed.InfoHash)
		}

		reportInfoLog(ctx, fmt.Sprintf("详情抓取完成: %s | movie_name=%s | movie_jav_id=%s | post_time=%d | magnets=%d", item.Title, item.MovieName, item.MovieJavID, item.PostTime, len(item.InfoHashes)))
		if item.MovieName == "" {
			reportInfoLog(ctx, fmt.Sprintf("影片番号未识别，继续按 info_hash 入库: %s", item.Title))
		}
		if item.MovieJavID == "" && item.MovieName != "" {
			reportInfoLog(ctx, fmt.Sprintf("未在本地库匹配到 movie_jav_id: %s | movie_name=%s", item.Title, item.MovieName))
		}
		if len(item.InfoHashes) == 0 {
			reportInfoLog(ctx, fmt.Sprintf("解析磁力: %s | 无磁力链接", item.Title))
		}
		for idx, infoHash := range item.InfoHashes {
			reportInfoLog(ctx, fmt.Sprintf("解析磁力: %s | %d/%d | info_hash=%s", item.Title, idx+1, len(item.InfoHashes), infoHash))
		}

		persistFailed := false
		for _, parsed := range parsedMagnets {
			if parsed.InfoHash == "" {
				continue
			}

			action, persistErr := s.upsertMagnet(ctx, &moviex.TSehuatangMagnet{
				MovieJavId:   item.MovieJavID,
				MovieName:    item.MovieName,
				Tag:          tag,
				ThreadTitle:  item.Title,
				ThreadUrl:    item.DetailURL,
				PostTime:     item.PostTime,
				PostDate:     item.PostDate,
				InfoHash:     parsed.InfoHash,
				LastSeenTime: 0,
				CreatedOn:    0,
				UpdatedOn:    0,
			}, now.Unix())
			if persistErr != nil {
				persistFailed = true
				result.PersistFailCount++
				reportErrorLog(ctx, fmt.Sprintf("入库失败: %s | info_hash=%s | err=%v", item.Title, parsed.InfoHash, persistErr))
				continue
			}

			switch action {
			case persistActionInsert:
				result.InsertedCount++
			case persistActionUpdate:
				result.UpdatedCount++
			}
			reportInfoLog(ctx, fmt.Sprintf("写入磁力: %s | action=%s | info_hash=%s", item.Title, action, parsed.InfoHash))
		}

		if persistFailed {
			item.Error = "部分磁力入库失败"
			result.FailedCount++
		} else {
			result.SuccessCount++
		}
		result.Items = append(result.Items, item)
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:             "fetch_sehuatang_detail_done",
			Message:           fmt.Sprintf("详情抓取完成: %s | magnets=%d | insert=%d | update=%d | persist_failed=%d", item.Title, len(item.InfoHashes), result.InsertedCount, result.UpdatedCount, result.PersistFailCount),
			HandledCount:      result.SuccessCount + result.FailedCount + result.SkippedExisting,
			SuccessCount:      result.SuccessCount,
			FailedCount:       result.FailedCount,
			QueuedCount:       result.MatchedCount - result.SuccessCount - result.FailedCount - result.SkippedExisting,
			CurrentPhaseKey:   "detail",
			PhaseKey:          "detail",
			PhaseHandledCount: result.SuccessCount + result.FailedCount + result.SkippedExisting,
			PhaseTotalCount:   result.MatchedCount,
			PhaseSuccessCount: result.SuccessCount,
			PhaseFailedCount:  result.FailedCount,
		})
	}

	reportInfoLog(ctx, fmt.Sprintf("色花堂抓取结束: pages=%d-%d, handled_pages=%d, mode=%s, skipped_top=%d, skipped_existing=%d, matched=%d, success=%d, failed=%d, insert=%d, update=%d, persist_failed=%d", result.StartPage, result.EndPage, result.HandledPageCount, req.PersistMode, result.SkippedTopCount, result.SkippedExisting, result.MatchedCount, result.SuccessCount, result.FailedCount, result.InsertedCount, result.UpdatedCount, result.PersistFailCount))
	return result, nil
}

func combineCookie(baseCookie, safeID string) string {
	baseCookie = strings.TrimSpace(baseCookie)
	safeID = strings.TrimSpace(safeID)
	if safeID == "" {
		return baseCookie
	}
	safeCookie := "_safe=" + safeID
	if baseCookie == "" {
		return safeCookie
	}
	return baseCookie + "; " + safeCookie
}

func (s *Service) fetchWithSafe(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string) ([]byte, error) {
	body, err := s.fetchWithRetry(ctx, cfg, state, targetURL, cfg.Cookie)
	if err != nil {
		return nil, err
	}

	safeID := parseSafeID(body)
	if safeID == "" {
		return body, nil
	}

	safeCookie := combineCookie(cfg.Cookie, safeID)
	return s.fetchWithRetry(ctx, cfg, state, targetURL, safeCookie)
}
