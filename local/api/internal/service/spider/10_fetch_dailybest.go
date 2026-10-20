// internal/spider/logic/a_FI_fetch_bestinv.go
package spider

import (
	"context"
	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/service/dailybest"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func (l *Service) FetchAndParseDailyBestinv(
	ctx context.Context,
	isSync bool,
	targets []dailybesttypes.SyncTarget,
) (*dailybestsync.Result, error) {
	l.deps.Log.WithContext(ctx).Info("FetchAndParseDailyBestinv: begin")

	if isSync {
		result, err := l.SyncBestinv(ctx, targets)
		if err != nil {
			l.deps.Log.WithContext(ctx).Errorf("FetchBestinv: %v", err)
			return nil, err
		}
		l.deps.Log.WithContext(ctx).Info("FetchAndParseDailyBestinv: done")
		return result, nil
	} else {
		if err := l.FetchBestinv(ctx, consts.BestCategoryMonth, 25); err != nil {
			l.deps.Log.WithContext(ctx).Errorf("FetchBestinv: %v", err)
			return nil, err
		}
	}

	if err := l.ParseBestinv(ctx); err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseBestinv: %v", err)
		return nil, err
	}

	l.deps.Log.WithContext(ctx).Info("FetchAndParseDailyBestinv: done")
	return nil, nil
}
