package logic

import (
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/types"
	"rudy_gc/pkg/mylog"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) FetchDetails() (int, error) {
	// 1) 找待抓详情的 item（HasDetail = None）
	items, err := l.deps.ItemRepo.FindByDetailStatus(l.ctx, types.ItemDetailNone)
	if err != nil {
		return 0, fmt.Errorf("获取待抓取详情的条目失败: %w", err)
	}
	total := len(items)
	if total == 0 {
		logx.WithContext(l.ctx).Info("没有需要抓取的详情")
		return 0, nil
	}

	start := time.Now()
	logx.WithContext(l.ctx).Infof("有 %d 个详情需要抓取", total)

	for i, it := range items {
		// 2) 拼 URL（与老项目一致）
		url := fmt.Sprintf("https://%s/cn/?v=%s", l.deps.Config.Spider.JavAddress, it.JavId)

		// 3) 用“详情专用”的重试策略抓取
		respBody, ferr := l.fetchDetailWithRetry(it.Name, url)
		if ferr != nil {
			return i, fmt.Errorf("抓取 %s(%s) 详情失败: %w", it.Name, it.JavId, ferr)
		}

		// 4) 校验详情页内容
		if !isValidDetail(respBody) {
			return i, fmt.Errorf("详情页内容无效: %s(%s)", it.Name, it.JavId)
		}

		now := time.Now().Unix()

		// 先查是否已有 detail，决定 birthTime
		var birthTime int64
		if old, _ := l.deps.DetailRepo.FindOneByJavId(l.ctx, it.JavId); old != nil && old.CreatedOn > 0 {
			birthTime = old.CreatedOn
		} else {
			birthTime = now
		}

		// 5) 保存 raw_detail（幂等：按 JavId Upsert）
		detail := &types.Detail{
			Name:          it.Name,
			JavId:         it.JavId,
			Prefix:        it.Prefix,
			QueryUrl:      url,
			Content:       respBody,
			LastQueryTime: now,
			CreatedOn:     birthTime, // 首次创建时间；若已存在保持旧值
			UpdatedOn:     now,
		}
		if err := l.deps.DetailRepo.Upsert(l.ctx, detail); err != nil {
			return i, fmt.Errorf("保存详情失败 %s(%s): %w", it.Name, it.JavId, err)
		}

		// 6) 更新 item 的“详情元信息”
		if err := l.deps.ItemRepo.UpdateDetailMeta(
			l.ctx,
			it.Id,
			types.ItemDetailStatusNeedScan, // DetailNeedScan
			birthTime,                      // DetailBirthTime
			now,                            // DetailUpdateTime
			types.ItemDetailOK,             // HasDetail
			now,                            // UpdatedOn
		); err != nil {
			return i, fmt.Errorf("更新条目详情元信息失败 %s: %w", it.Name, err)
		}

		// 7) 轻微休眠（沿用你的随机 sleep）
		time.Sleep(getRandomSleepDuration())

		// 8) 进度日志（中文）
		logItemProgress(i+1, total, it.Name, start)
	}

	return total, nil
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
		resp, ferr := l.deps.Fetcher.Get(l.ctx, url)
		if ferr == nil {
			body = resp.Body
			// 详情页只要拿到响应就尝试返回（是否有效由 isValidDetail 再判断）
			return string(body), nil
		}
		err = ferr

		logx.WithContext(l.ctx).Infof("%s 第%d次尝试: %s", name, attempts, url)

		if attempts%20 == 0 {
			mylog.Warn(l.ctx, "多次重试失败，10分钟后重试", "name", name, "attempts", attempts, "url", url)
			time.Sleep(10 * time.Minute)
		} else {
			mylog.Warn(l.ctx, "请求错误，3秒后重试", "name", name, "attempts", attempts, "url", url, "err", err.Error())
			time.Sleep(3 * time.Second)
		}
	}
	return "", fmt.Errorf("达到最大重试次数，无法获取 URL: %s", url)
}

func logItemProgress(done, total int, name string, start time.Time) {
	elapsed := time.Since(start).Minutes()
	etaTotal := (elapsed / float64(done)) * float64(total)
	remain := etaTotal - elapsed
	logx.Infof("已完成 %d/%d: %s，用时 %.1f 分钟，预计剩余 %.1f 分钟",
		done, total, name, elapsed, remain)
}
