package logic

import (
	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) SyncDailyBestinv() error {
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: begin")
	// TODO: 同步远端/历史榜单；幂等合并；必要时触发详情补抓
	logx.WithContext(l.ctx).Info("SyncDailyBestinv: done")
	return nil
}
