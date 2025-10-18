package logic

import (
	"bytes"
	"fmt"
	"rudy_gc/internal/consts"
	"strings"
	"time"

	"rudy_gc/internal/types"
	"rudy_gc/pkg/mylog"

	"github.com/PuerkitoBio/goquery"
)

func (l *CrawlLogic) FetchDetailsByItemDetailStatus() (int, error) {
	// 1) 找待抓详情的 item（HasDetail = None）
	items, err := l.deps.ItemRepo.FindByDetailStatus(l.ctx, consts.ItemDetailNone)
	if err != nil {
		return 0, fmt.Errorf("获取待抓取详情的条目失败: %w", err)
	}

	return l.handleFetchDetails(items)
}

// 抽出的函数：统一处理详情抓取逻辑
func (l *CrawlLogic) handleFetchDetails(items []*types.Item) (int, error) {
	total := len(items)
	if total == 0 {
		l.deps.Log.WithContext(l.ctx).Info("没有需要抓取的详情")
		return 0, nil
	}

	start := time.Now()
	l.deps.Log.WithContext(l.ctx).Infof("有 %d 个详情需要抓取", total)

	for i, it := range items {
		if err := l.fetchAndSaveDetail(it); err != nil {
			return i, err
		}
		// 7) 进度日志（中文）
		l.logItemProgress(i+1, total, it.Name, start)
		time.Sleep(getRandomSleepDuration())
	}

	return total, nil
}

func (l *CrawlLogic) fetchAndSaveDetail(it *types.Item) error {
	// 支持取消
	select {
	case <-l.ctx.Done():
		return l.ctx.Err()
	default:
	}

	// 1) 拼 URL（与老项目一致）
	url := fmt.Sprintf("https://%s/cn/?v=%s", l.deps.Config.Fetcher.JavAddress, it.JavId)

	// 2) 用“详情专用”的重试策略抓取
	respBody, ferr := l.fetchDetailWithRetry(it.Name, url)
	if ferr != nil {
		return fmt.Errorf("抓取 %s(%s) 详情失败: %w", it.Name, it.JavId, ferr)
	}

	now := time.Now().Unix()

	// 3) 先查是否已有 detail，决定 birthTime
	var birthTime int64
	if old, _ := l.deps.DetailRepo.FindOneByJavId(l.ctx, it.JavId); old != nil && old.CreatedOn > 0 {
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
	if err := l.deps.DetailRepo.Upsert(l.ctx, detail); err != nil {
		return fmt.Errorf("保存详情失败 %s(%s): %w", it.Name, it.JavId, err)
	}

	// 5) 更新 item 的“详情元信息”
	if err := l.deps.ItemRepo.UpdateDetailMeta(
		l.ctx,
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
func (l *CrawlLogic) fetchDetailWithRetry(name, url string) (string, error) {
	const maxRetries = 45
	var body []byte
	var err error

	for attempts := 1; attempts <= maxRetries; attempts++ {

		l.deps.Log.WithContext(l.ctx).Infof("第 %d 次尝试: %s", attempts, url)

		resp, ferr := l.deps.Fetcher.Get(l.ctx, url)
		if ferr == nil {
			body = resp.Body

			// 调用过滤函数，裁剪掉无用的部分
			content, ferr := filterDetailContent(body)
			if ferr == nil {
				if isValidDetail(string(body)) {
					return content, nil
				}
				// 请求成功但页面无效/空白
				return "", ErrBlankPage
			}
			// 如果过滤失败，仍然返回原始 HTML 作为兜底
			return string(body), nil
		}
		err = ferr

		l.deps.Log.WithContext(l.ctx).Infof("%s 第%d次尝试: %s", name, attempts, url)

		if attempts%20 == 0 {
			mylog.Warn(l.ctx, "多次重试失败，10分钟后重试",
				"name", name, "attempts", attempts, "url", url)
			time.Sleep(10 * time.Minute)
		} else {
			mylog.Warn(l.ctx, "请求错误，3秒后重试",
				"name", name, "attempts", attempts, "url", url, "err", err.Error())
			time.Sleep(3 * time.Second)
		}
	}
	return "", fmt.Errorf("达到最大重试次数，无法获取 URL: %s", url)
}

func (l *CrawlLogic) logItemProgress(done, total int, name string, start time.Time) {
	elapsed := time.Since(start).Minutes()
	etaTotal := (elapsed / float64(done)) * float64(total)
	remain := etaTotal - elapsed
	l.deps.Log.Infof("已完成 %d/%d: %s，用时 %.1f 分钟，预计剩余 %.1f 分钟",
		done, total, name, elapsed, remain)
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
