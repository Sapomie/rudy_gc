package spider

import (
	"context"
	"fmt"
	consts "rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/oldmodel/modelg"
	"time"
)

func (l *CrawlLogic) SyncBestinv(ctx context.Context) error {
	dbRemote, err := modelg.NewGormDBEngineRemote(l.deps.Config.DataSourceRemote, "error")
	if err != nil {
		return err
	}
	remoteModel := modelg.NewBestInvModel(dbRemote)
	latest, err := l.deps.BestinvRepo.LatestDayNumber(ctx)
	if err != nil {
		return err
	}
	startDay := latest + 1
	startDayString := consts.GetDateStringByRankDayNumber(startDay)
	endDayString := time.Now().Format(time.DateOnly)
	endDay := consts.GetRankDayNumber(endDayString)

	var dayNumber, count int64
	l.deps.Log.Infof("开始同步从 %v 到 %v 的BestInv\n", startDayString, endDayString)
	for dayNumber = startDay; dayNumber <= endDay; dayNumber++ {
		bestInvs, err := remoteModel.FindByDayNumberRankMonth(ctx, dayNumber)
		if err != nil {
			return err
		}
		for _, b := range bestInvs {
			now := time.Now().Unix()
			best := &types.Bestinv{
				Name:          b.Name,
				NeedScan:      consts.BestinvNeedScan,
				NeedRankCheck: consts.BestinvNeedRankCheck,
				Category:      b.Category,
				Page:          b.Page,
				DayNumber:     b.DayNumber,
				Content:       b.Content,
				LastQueryTime: b.LastQueryTime,
				Date:          b.Date,
				CreatedOn:     now,
				UpdatedOn:     now,
			}
			if err := l.deps.BestinvRepo.Upsert(ctx, best); err != nil {
				return fmt.Errorf("写入 Bestinv 失败(page=%d): %w", b.Page, err)
			}
			count++
		}
		l.deps.Log.Infof("完成%v个BestInv\n", count)
	}

	return nil
}
