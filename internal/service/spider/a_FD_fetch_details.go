package spider

import (
	"bytes"
	"context"
	"fmt"
	"rudy_gc/internal/consts"
	"strings"
	"time"

	"rudy_gc/internal/types"

	"github.com/PuerkitoBio/goquery"
)

func (l *CrawlLogic) FetchDetailsByItemDetailStatus(ctx context.Context) (int64, error) {
	// 1) 找待抓详情的 item（HasDetail = None）
	items, err := l.deps.ItemRepo.FindByDetailStatus(ctx, consts.ItemDetailNone)
	if err != nil {
		return 0, fmt.Errorf("获取待抓取详情的条目失败: %w", err)
	}

	return l.handleFetchDetails(ctx, items)
}

// 抽出的函数：统一处理详情抓取逻辑
func (l *CrawlLogic) handleFetchDetails(ctx context.Context, items []*types.Item) (int64, error) {
	total := len(items)
	if total == 0 {
		l.deps.Log.WithContext(ctx).Info("没有需要抓取的详情")
		return 0, nil
	}

	start := time.Now()
	l.deps.Log.WithContext(ctx).Infof("有 %d 个详情需要抓取", total)
	l.reportProgress(ctx, "detail_queue_ready", fmt.Sprintf("待抓取详情 %d 条", total), 0, 0, 0, total)

	for i, it := range items {
		if err := l.waitIfPaused(ctx); err != nil {
			return int64(i), err
		}
		if err := l.fetchAndSaveDetail(ctx, it); err != nil {
			return 0, err
		}
		// 7) 进度日志（中文）
		l.logItemProgress(i+1, total, it.Name, it.LastQueryDetailTime, start)
		l.reportProgress(ctx, "detail_item_done", fmt.Sprintf("详情完成：%s", it.Name), i+1, i+1, 0, total-(i+1))
		if err := l.sleepWithContext(ctx, getRandomSleepDuration()); err != nil {
			return int64(i + 1), err
		}
	}

	return int64(total), nil
}

func (l *CrawlLogic) fetchAndSaveDetail(ctx context.Context, it *types.Item) error {
	// 支持取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 1) 拼 URL（与老项目一致）
	url := fmt.Sprintf("https://%s/cn/?v=%s", l.deps.Config.Fetcher.JavAddress, it.JavId)

	// 2) 用“详情专用”的重试策略抓取
	respBody, ferr := l.fetchDetailWithRetry(ctx, it.Name, url)
	if ferr != nil {
		return fmt.Errorf("抓取 %s(%s) 详情失败: %w", it.Name, it.JavId, ferr)
	}

	now := time.Now().Unix()

	// 3) 先查是否已有 detail，决定 birthTime
	var birthTime int64
	if old, _ := l.deps.DetailRepo.FindOneByJavId(ctx, it.JavId); old != nil && old.CreatedOn > 0 {
		birthTime = old.CreatedOn
	} else {
		birthTime = now
	}

	// 4) 保存 raw_detail（幂等：按 JavId Upsert）
	detail := &types.Detail{
		Name:      it.Name,
		JavId:     it.JavId,
		Prefix:    it.Prefix,
		QueryUrl:  url,
		Content:   respBody,
		CreatedOn: birthTime, // 首次创建时间；若已存在保持旧值
		UpdatedOn: now,
	}
	if err := l.deps.DetailRepo.Upsert(ctx, detail); err != nil {
		return fmt.Errorf("保存详情失败 %s(%s): %w", it.Name, it.JavId, err)
	}

	// 5) 更新 item 的“详情元信息”
	if err := l.deps.ItemRepo.UpdateDetailMeta(
		ctx,
		it.Id,
		consts.ItemDetailStatusNeedScan, // needScan
		birthTime,                       // birthTime（仅首次写入）
		now,                             // updateTime（本次抓/解详情时间）
		now,                             // updatedOn（记录更新时间）
		consts.ItemDetailOK,             // hasDetail（已具备详情）
	); err != nil {
		return fmt.Errorf("更新条目详情元信息失败 %s: %w", it.Name, err)
	}

	return nil
}

// 详情页有效性：包含 "video_title" 且长度 >= 5000（与老项目一致）
func isValidDetail(content string) bool {
	return len(content) >= 5000 && strings.Contains(content, "video_title")
}

// 专用详情重试：每 20 次休眠 10 分钟，其余 3 秒；打印中文日志
func (l *CrawlLogic) fetchDetailWithRetry(ctx context.Context, name, url string) (string, error) {
	const maxRetries = 45
	var body []byte
	var err error

	for attempts := 1; attempts <= maxRetries; attempts++ {
		if err := l.waitIfPaused(ctx); err != nil {
			return "", err
		}
		l.deps.Log.WithContext(ctx).Infof("第 %d 次尝试: %s", attempts, url)

		resp, ferr := l.deps.Fetcher.Get(ctx, url)
		if ferr != nil {
			// 记录详细错误信息
			err = ferr
			l.deps.Log.WithContext(ctx).Warnf(
				"请求失败 - %s 第%d次尝试: %s, 错误: %v",
				name, attempts, url, err,
			)

			if attempts%20 == 0 {
				l.deps.Log.WithContext(ctx).Warnf(
					"多次重试失败，10分钟后重试 - name: %s, attempts: %d, url: %s, last_error: %v",
					name, attempts, url, err,
				)
				if sleepErr := l.sleepWithContext(ctx, 10*time.Minute); sleepErr != nil {
					return "", sleepErr
				}
			} else {
				l.deps.Log.WithContext(ctx).Warnf(
					"请求错误，3秒后重试 - name: %s, attempts: %d, url: %s, error: %v",
					name, attempts, url, err,
				)
				if sleepErr := l.sleepWithContext(ctx, 3*time.Second); sleepErr != nil {
					return "", sleepErr
				}
			}
			continue
		}

		// 获取成功，处理响应
		body = resp.Body

		// 调用过滤函数，裁剪掉无用的部分
		content, ferr := filterDetailContent(body)
		if ferr != nil {
			// 记录过滤错误
			l.deps.Log.WithContext(ctx).Warnf(
				"过滤内容失败 - name: %s, url: %s, error: %v",
				name, url, ferr,
			)
			// 如果过滤失败，仍然返回原始 HTML 作为兜底
			return string(body), nil
		}

		if isValidDetail(string(body)) {
			return content, nil
		}

		// 请求成功但页面无效/空白
		l.deps.Log.WithContext(ctx).Warnf(
			"页面内容无效/空白 - name: %s, url: %s, attempts: %d",
			name, url, attempts,
		)
		return "", ErrBlankPage
	}

	// 达到最大重试次数
	l.deps.Log.WithContext(ctx).Errorf(
		"达到最大重试次数，无法获取 URL - name: %s, url: %s, max_retries: %d, last_error: %v",
		name, url, maxRetries, err,
	)
	return "", fmt.Errorf("达到最大重试次数 %d，无法获取 URL %s: %v", maxRetries, url, err)
}

func (l *CrawlLogic) logItemProgress(done, total int, name string, lastQuery int64, start time.Time) {
	now := time.Now().Unix()
	days := float64(now-lastQuery) / 86400.0

	elapsed := time.Since(start).Minutes()
	etaTotal := (elapsed / float64(done)) * float64(total)
	remain := etaTotal - elapsed

	l.deps.Log.Infof(
		"已完成 %d/%d: %s（距上次更新 %.1f 天），用时 %.1f 分钟，预计剩余 %.1f 分钟",
		done, total, name, days, elapsed, remain,
	)
}

// filterDetailContent 过滤掉 HTML 中无关部分，只保留电影详情相关节点
func filterDetailContent(body []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("解析 HTML 出错: %w", err)
	}

	// 新建一个干净的 DOM 作为输出
	out := &strings.Builder{}
	out.WriteString("<html><body>\n")

	// 只挑选需要的节点复制出来
	keepIDs := []string{
		"video_title",
		"video_id",
		"video_date",
		"video_length",
		"video_director",
		"video_maker",
		"video_label",
		"video_review",
		"video_genres",
		"video_cast",
		"video_favorite_edit",
		"video_jacket",
	}

	for _, id := range keepIDs {
		if sel := doc.Find(fmt.Sprintf("#%s", id)); sel.Length() > 0 {
			html, _ := sel.Html()
			out.WriteString(fmt.Sprintf("<div id=\"%s\">%s</div>\n", id, html))
		}
	}

	out.WriteString("</body></html>")
	return out.String(), nil
}
