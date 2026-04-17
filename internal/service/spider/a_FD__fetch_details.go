package spider

import (
	"context"
	"errors"
	"fmt"
	"rudy_gc/internal/types"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (l *CrawlLogic) RefreshOldestDetail(ctx context.Context, num int64) (int, error) {
	if num <= 0 {
		num = 1
	}

	items, err := l.deps.ItemRepo.FindOldestByLastQueryDetailTime(ctx, num)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.deps.Log.WithContext(ctx).Info("RefreshOldestDetail: 没有可更新的 Item")
			return 0, nil
		}
		return 0, err
	}
	if len(items) == 0 {
		l.deps.Log.WithContext(ctx).Info("RefreshOldestDetail: Item 列表为空，跳过")
		return 0, nil
	}

	first := items[0]
	l.deps.Log.WithContext(ctx).Infof(
		"RefreshOldestDetail: 准备更新 %d 条最久未查询的详情，first javId=%s, name=%s, lastQuery=%d",
		len(items), first.JavId, first.Name, first.LastQueryDetailTime,
	)

	// 直接复用批量逻辑：按顺序刷新 num 条
	total, changedMovieJavIDs, err := l.handleFetchAndParseDetails(ctx, items)
	if err != nil {
		return 0, err
	}
	l.movieSvc.EnqueueAggRebuildByMovieJavIDs("spider_refresh_oldest_detail", changedMovieJavIDs...)
	return total, nil
}

func (l *CrawlLogic) HandleFetchDetailsById(ctx context.Context, javIds []string) (int, []string, error) {
	if len(javIds) == 0 {
		l.deps.Log.WithContext(ctx).Info("handleFetchDetailsById: 空列表，跳过")
		return 0, nil, nil
	}

	items := make([]*types.Item, 0, len(javIds))
	for _, javId := range javIds {
		javId = strings.TrimSpace(javId)
		if javId == "" {
			continue
		}

		it, err := l.deps.ItemRepo.FindOneByJavId(ctx, javId)
		if err != nil {
			l.deps.Log.WithContext(ctx).Warnf("handleFetchDetailsById: 根据 javId=%s 查询 Item 失败: %v（将尝试抓取）", javId, err)
			// 查询失败时，宁可尝试抓取（以免漏处理）
			// 但此时没有 it 无法继续，跳过该 id
			continue
		}
		if it == nil {
			l.deps.Log.WithContext(ctx).Warnf("handleFetchDetailsById: 根据 javId=%s 未找到 Item（跳过）", javId)
			continue
		}

		// 读取 Movie 的上映日，失败/缺失时按“需要更新”处理
		var releasingDate int64
		if m, merr := l.deps.MovieRepo.FindOneByJavId(ctx, javId); merr != nil {
			l.deps.Log.WithContext(ctx).Warnf("handleFetchDetailsById: MovieRepo.FindOneByJavId(%s) 失败: %v（视为需要更新）", javId, merr)
			releasingDate = 0 // 让 shouldSkipUpdate 返回 false
		} else if m == nil {
			// Movie 还未建立，必然需要抓详情
			releasingDate = 0
		} else {
			releasingDate = m.ReleasingDate
		}

		// 只有“不应跳过”时才加入待处理列表
		if !l.shouldSkipUpdate(ctx, it.LastQueryDetailTime, releasingDate, it.Name) {
			items = append(items, it)
		}

	}

	total, changedMovieJavIDs, err := l.handleFetchAndParseDetails(ctx, items)
	if err != nil {
		return 0, nil, err
	}
	return total, changedMovieJavIDs, nil
}

func (l *CrawlLogic) handleFetchAndParseDetails(ctx context.Context, items []*types.Item) (int, []string, error) {
	total := len(items)
	if total == 0 {
		l.deps.Log.WithContext(ctx).Info("没有需要抓取的详情")
		return 0, nil, nil
	}

	start := time.Now()
	l.deps.Log.WithContext(ctx).Infof("有 %d 个详情需要抓取", total)
	l.reportPhaseProgress(ctx, "detail", "detail_queue_ready", fmt.Sprintf("待抓取详情 %d 条", total), 0, total, 0, 0)
	var failed int
	affected := newAffectedMovieNumbers()
	for i, it := range items {
		if err := l.waitIfPaused(ctx); err != nil {
			return i, nil, err
		}

		// -------------------------------
		// 1) 抓详情（失败 → log + continue）
		// -------------------------------
		if err := l.fetchAndSaveDetail(ctx, it); err != nil {
			l.deps.Log.WithContext(ctx).Errorf(
				"[%d/%d] fetch 失败: %s (err=%v)",
				i+1, total, it.Name, err,
			)
			failed++
			l.reportPhaseProgress(ctx, "detail", "detail_fetch_failed", fmt.Sprintf("抓取失败：%s", it.Name), i+1, total, (i+1)-failed, failed)
			if sleepErr := l.sleepWithContext(ctx, getRandomSleepDuration()); sleepErr != nil {
				return i + 1, nil, sleepErr
			}
			continue
		}

		// -------------------------------
		// 2) 重新查 item
		// -------------------------------
		item, err := l.deps.ItemRepo.FindOneByJavId(ctx, it.JavId)
		if err != nil {
			l.deps.Log.WithContext(ctx).Errorf(
				"[%d/%d] FindOneByJavId 失败: %s (err=%v)",
				i+1, total, it.Name, err,
			)
			continue
		}

		// -------------------------------
		// 3) 解析详情（失败 → log + continue）
		// -------------------------------
		resp, err := l.handleDetailParse(ctx, item)
		if err != nil {
			l.deps.Log.WithContext(ctx).Errorf(
				"[%d/%d] parse 失败: %s (err=%v)",
				i+1, total, it.Name, err,
			)
			failed++
			l.reportPhaseProgress(ctx, "detail", "detail_parse_failed", fmt.Sprintf("解析失败：%s", it.Name), i+1, total, (i+1)-failed, failed)
			continue
		}
		affected.addFromResponse(resp)

		// -------------------------------
		// 4) 进度日志
		// -------------------------------
		l.logItemProgress(ctx, i+1, total, it.Name, it.LastQueryDetailTime, start)
		l.reportPhaseProgress(ctx, "detail", "detail_item_done", fmt.Sprintf("详情完成：%s", it.Name), i+1, total, (i+1)-failed, failed)
		if err := l.sleepWithContext(ctx, getRandomSleepDuration()); err != nil {
			return i + 1, nil, err
		}
	}

	return total, affectedMovieJavIDs(affected), nil
}

const (
	oneDay      = 24 * 60 * 60
	fifteenDays = 15 * oneDay
	thirtyDays  = 30 * oneDay
	ninetyDays  = 90 * oneDay
)

// 返回 true 表示“可以跳过本次更新”（即近期已抓过/离上映较久且更新频率较低）
func (l *CrawlLogic) shouldSkipUpdate(ctx context.Context, lastQueryTime, releasingDate int64, name string) bool {
	log := l.deps.Log.WithContext(ctx)
	now := time.Now().Unix()

	// 若缺数据，默认不跳过（需要更新）
	if lastQueryTime <= 0 || releasingDate <= 0 {
		log.Infof("shouldSkipUpdate: 缺少时间数据 → 需要更新")
		return false
	}

	releasingFromNow := now - releasingDate
	lastQueryFromNow := now - lastQueryTime
	//daysSinceLast := float64(lastQueryFromNow) / float64(oneDay)

	var skip bool
	switch {
	case releasingFromNow <= fifteenDays && releasingFromNow > -2*oneDay:
		// 上映前2天到上映后15天：每天都可更新 → 若24h内已更新则跳过
		skip = lastQueryFromNow <= oneDay*2
	case releasingFromNow <= thirtyDays && releasingFromNow > fifteenDays:
		// 上映15-30天：5天内更新过就跳过
		skip = lastQueryFromNow <= 10*oneDay
	case releasingFromNow <= ninetyDays && releasingFromNow > thirtyDays:
		// 上映30-90天：10天内更新过就跳过
		skip = lastQueryFromNow <= 30*oneDay
	default:
		// 其余（更久）：30天内更新过就跳过
		skip = lastQueryFromNow <= 100*oneDay
	}

	//if !skip {
	//	// ✅ 需要更新时打印距离上次更新的天数
	//	log.Infof("shouldSkipUpdate: 距上次更新 %.1f 天 → 需要更新 %s", daysSinceLast, name)
	//}

	return skip
}

func affectedMovieJavIDs(affected *affectedMovieNumbers) []string {
	if affected == nil || len(affected.movies) == 0 {
		return nil
	}

	out := make([]string, 0, len(affected.movies))
	for javID := range affected.movies {
		if strings.TrimSpace(javID) == "" {
			continue
		}
		out = append(out, strings.TrimSpace(javID))
	}
	return out
}
