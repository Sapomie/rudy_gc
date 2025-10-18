// internal/spiderx/logic/a_FD__fetch_and_parse_detail.go
package logic

import "context"

func (l *CrawlLogic) FetchAndParseDetails(ctx context.Context) (int, error) {
	detailNum, err := l.FetchDetailsByItemDetailStatus(ctx)
	if err != nil {
		l.deps.Log.WithContext(ctx).Errorf("FetchDetailsByItemDetailStatus: %v", err)
		return 0, err
	}

	err = l.ParseDetails(ctx)
	if err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseDetails: %v", err)
		return 0, err
	}
	return detailNum, nil
}
