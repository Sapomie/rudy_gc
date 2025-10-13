// internal/spider/logic/a_FI_fetch_bestinv.go
package logic

import (
	"rudy_gc/internal/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) CrawlDailyBestinv() error {
	logx.WithContext(l.ctx).Info("CrawlDailyBestinv: begin")
	if err := l.FetchBestinv(consts.BestCategoryMonth, 2); err != nil {
		logx.WithContext(l.ctx).Errorf("FetchBestinv: %v", err)
		return err
	}

	if err := l.ProcessBestinv(); err != nil {
		logx.WithContext(l.ctx).Errorf("ProcessBestinv: %v", err)
		return err
	}

	logx.WithContext(l.ctx).Info("CrawlDailyBestinv: done")
	return nil
}
