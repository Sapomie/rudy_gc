package logic

func (l *CrawlLogic) SyncDailyBestinv() error {
	l.deps.Log.WithContext(l.ctx).Info("SyncDailyBestinv: begin")
	// TODO: 同步远端/历史榜单；幂等合并；必要时触发详情补抓
	l.deps.Log.WithContext(l.ctx).Info("SyncDailyBestinv: done")
	return nil
}
