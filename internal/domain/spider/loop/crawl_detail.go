// internal/spider/loop/crawl_detail.go
package loop

import (
	"context"
	"rudy_gc/internal/domain/spider/logic"
	spider "rudy_gc/internal/domain/spider/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// CrawlDetailLoop 消费 DetailCh，按老项目模式并发处理
func (m *LoopServer) CrawlDetailLoop() {
	for n := range m.DetailCh {
		m.refDetailSemaphore <- struct{}{}
		// 和老项目的 threading.GoSafe 类似，这里交给 goroutine 并确保释放信号量
		go func(notify *spider.Notification) {
			defer func() { <-m.refDetailSemaphore }()
			m.processDetailNotification(notify)
		}(n)
	}
}

// 处理单次 Detail 通知
func (m *LoopServer) processDetailNotification(n *spider.Notification) {
	ctx, cancel := context.WithTimeout(m.ctx, 45*time.Minute)
	defer cancel()

	m.deps.Log.Info("detail: begin")
	l := logic.NewCrawlLogic(ctx, m.deps)

	// 1) 抓取/刷新详情（按你的实现命名来，下面先占位 CrawDetailAll）
	num, err := l.FetchAndParseDetails()
	if err != nil {
		logx.WithContext(ctx).Errorf("detail: CrawlDetailAll error: %v", err)
	}
	// 记录数量（如果你的 Notification 没有字段承载，就仅日志记录）
	logx.WithContext(ctx).Infow("detail: done count", logx.Field("count", num))

	// 2) 可选：写入一次记录（占位）todo:record 细化
	//if recErr := l.AddRecord(n, int64(num)); recErr != nil {
	//	logx.WithContext(ctx).Errorf("detail: AddRecord error: %v", recErr)
	//}

	// 3) 通知后续环节（翻译 / 下载封面）
	m.notifyTranslationIfNecessary()
	m.notifyDownloadCoverIfNecessary()

	// 4) 如果是 Bestinv 场景，处理 Rank（保持老项目逻辑）
	if n.Action == spider.ActionDailyBestinv || n.Action == spider.ActionSyncDailyBestinv {
		logx.WithContext(ctx).Info("detail: ProcessBestinvRank begin")
		if rerr := l.ProcessBestinvRank(); rerr != nil {
			logx.WithContext(ctx).Errorf("detail: ProcessBestinvRank error: %v", rerr)
		}
	}

	// 5) 更新演员统计（按老项目）todo
	//if uerr := l.UpdateCastsMovieNumberInfo(); uerr != nil {
	//	logx.WithContext(ctx).Errorf("detail: UpdateCastsMovieNumberInfo error: %v", uerr)
	//}

	logx.WithContext(ctx).Info("detail: end")
}

// 发送翻译通知（nil 安全）
func (m *LoopServer) notifyTranslationIfNecessary() {
	if m.TranslationCh == nil {
		m.logWarn("translation notify skipped: channel nil")
		return
	}
	select {
	case m.TranslationCh <- struct{}{}:
		m.logInfo("translation notify sent")
	default:
		m.logWarn("translation notify dropped: channel full")
	}
}

// 发送下载封面通知（nil 安全）
func (m *LoopServer) notifyDownloadCoverIfNecessary() {
	if m.DownloadCoverCh == nil {
		m.logWarn("download cover notify skipped: channel nil")
		return
	}
	select {
	case m.DownloadCoverCh <- struct{}{}:
		m.logInfo("download cover notify sent")
	default:
		m.logWarn("download cover notify dropped: channel full")
	}
}
