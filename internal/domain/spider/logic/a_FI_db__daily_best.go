// internal/spider/logic/a_FI_fetch_bestinv.go
package logic

import (
	"rudy_gc/internal/consts"
)

func (l *CrawlLogic) FetchAndParseDailyBestinv() error {
	l.deps.Log.WithContext(l.ctx).Info("FetchAndParseDailyBestinv: begin")

	if err := l.FetchBestinv(consts.BestCategoryMonth, 25); err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("FetchBestinv: %v", err)
		return err
	}

	if err := l.ParseBestinv(); err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("ParseBestinv: %v", err)
		return err
	}

	l.deps.Log.WithContext(l.ctx).Info("FetchAndParseDailyBestinv: done")
	return nil
}
