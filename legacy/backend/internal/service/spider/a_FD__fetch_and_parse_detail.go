// internal/spiderx/logic/a_FD__fetch_and_parse_detail.go
package spider

import (
	"context"
	"fmt"
	"rudy_gc/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/threading"
)

func (l *CrawlLogic) FetchAndParseDetails(ctx context.Context) (int64, *affectedMovieNumbers, error) {
	detailNum, err := l.FetchDetailsByItemDetailStatus(ctx)
	if err != nil {
		l.deps.Log.WithContext(ctx).Errorf("FetchDetailsByItemDetailStatus: %v", err)
		return 0, nil, err
	}

	affected, err := l.ParseDetails(ctx)
	if err != nil {
		l.deps.Log.WithContext(ctx).Errorf("ParseDetails: %v", err)
		return 0, nil, err
	}

	threading.GoSafe(func() {
		if err := l.updateMovieNumbers(ctx, affected); err != nil {
			l.deps.Log.WithContext(ctx).Errorf("updateMovieNumbers: %v", err)
		}
	})

	return detailNum, affected, nil
}

func (l *CrawlLogic) saveRecord(ctx context.Context, typ string, start, end time.Time, detailNum int64) {
	rec := &types.Record{
		Name:         fmt.Sprintf("%s-%s", typ, start.Format("20060102-150405")),
		StartTime:    start.Unix(),
		EndTime:      end.Unix(),
		Type:         typ,
		DetailNumber: detailNum,
	}

	if _, err := l.deps.RecordRepo.TryInsert(ctx, rec); err != nil {
		l.deps.Log.WithContext(ctx).Warnf("saveRecord: insert %s failed: %v", typ, err)
	}
}
