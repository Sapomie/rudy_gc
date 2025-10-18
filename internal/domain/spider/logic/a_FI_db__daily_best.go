// internal/spider/logic/a_FI_fetch_bestinv.go
package logic

import (
	"context"
	"rudy_gc/internal/consts"
)

func (l *CrawlLogic) FetchAndParseDailyBestinv(ctx context.Context) error {
	l.deps.Log.WithContext(ctx).Info("FetchAndParseDailyBestinv: begin")

	if err := l.FetchBestinv(ctx, consts.BestCategoryMonth, 25); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("FetchBestinv: %v", err)
		return err
	}

	if err := l.ParseBestinv(ctx); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseBestinv: %v", err)
		return err
	}

	l.deps.Log.WithContext(ctx).Info("FetchAndParseDailyBestinv: done")
	return nil
}
