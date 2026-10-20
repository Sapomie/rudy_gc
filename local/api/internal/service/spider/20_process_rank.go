package spider

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	consts "rudy_gc/local/api/internal/consts"
	dailybestsync "rudy_gc/local/api/internal/service/dailybest"
	"rudy_gc/local/api/internal/types"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	"rudy_gc/local/api/pkg/ptr"
	contracts "rudy_gc/shared/dailybest"
)

func (l *Service) ProcessDailyBestSyncRanks(
	ctx context.Context,
	batches []dailybestsync.Batch,
) ([]dailybesttypes.RankDayTiming, error) {
	rankBatches := make([]dailybesttypes.RankBatch, 0, len(batches))
	for _, batch := range batches {
		rankBatches = append(rankBatches, dailybesttypes.RankBatch{
			BusinessDate: batch.Manifest.BusinessDate,
			DayNumber:    batch.Manifest.DayNumber,
			Category:     batch.Manifest.Category,
			FetchedOn:    batch.Manifest.FinishedAt,
			Ranks:        batch.Ranks,
		})
	}
	return l.processDailyBestRankBatches(ctx, rankBatches)
}

func (l *Service) ProcessBestinvRank(ctx context.Context) ([]dailybesttypes.RankDayTiming, error) {
	log := l.deps.Log.WithContext(ctx)
	ids, err := l.deps.BestinvRepo.ListIDsByRankCheck(ctx, consts.BestinvNeedRankCheck, 0)
	if err != nil {
		return nil, fmt.Errorf("查询需要排名检查的 bestinv 失败: %w", err)
	}
	if len(ids) == 0 {
		log.Info("没有需要排名检查的 bestinv")
		return []dailybesttypes.RankDayTiming{}, nil
	}
	log.Infof("需要处理 %d 个 bestinv", len(ids))

	bestinvs := make([]*types.Bestinv, 0, len(ids))
	for _, id := range ids {
		bestinv, err := l.deps.BestinvRepo.FindOne(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("根据 id=%d 查询 bestinv 失败: %w", id, err)
		}
		bestinvs = append(bestinvs, bestinv)
	}
	pendingBatches, err := buildPendingBestinvRankBatches(bestinvs)
	if err != nil {
		return nil, err
	}
	rankBatches := make([]dailybesttypes.RankBatch, 0, len(pendingBatches))
	for _, pending := range pendingBatches {
		rankBatches = append(rankBatches, pending.Batch)
	}
	affectedDayTimings, err := l.processDailyBestRankBatches(ctx, rankBatches)
	if err != nil {
		return nil, err
	}

	checkedAt := time.Now().Unix()
	for count, id := range ids {
		if err := l.deps.BestinvRepo.MarkRankChecked(ctx, id, checkedAt); err != nil {
			return nil, fmt.Errorf("标记 bestinv 已完成排名检查失败(id=%d): %w", id, err)
		}
		if (count+1)%5 == 0 {
			log.Infof("已完成 %d 个 bestinv 的排名处理", count+1)
		}
	}

	log.Info("Bestinv 排名处理完成")
	return affectedDayTimings, nil
}

func (l *Service) processDailyBestRankBatches(
	ctx context.Context,
	batches []dailybesttypes.RankBatch,
) ([]dailybesttypes.RankDayTiming, error) {
	normalized, err := normalizeDailyBestRankBatches(batches)
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return []dailybesttypes.RankDayTiming{}, nil
	}

	javIDSet := make(map[string]struct{})
	affectedDayTimings := make(map[int64]dailybesttypes.RankDayTiming, len(normalized))
	syncedOn := time.Now().Unix()
	for _, batch := range normalized {
		ranks := make([]*types.Rank, 0, len(batch.Ranks))
		for _, item := range batch.Ranks {
			ranks = append(ranks, &types.Rank{
				RankKey:    item.RankKey,
				MovieJavId: item.MovieJavID,
				DayNumber:  item.DayNumber,
				RankPos:    item.RankPos,
				Category:   item.Category,
			})
		}
		inserted, err := l.deps.RankRepo.InsertByDayNumberCategory(
			ctx,
			batch.DayNumber,
			batch.Category,
			ranks,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"首次写入 DailyBest date=%s category=%d 排名失败: %w",
				batch.BusinessDate,
				batch.Category,
				err,
			)
		}
		if !inserted {
			continue
		}
		for _, item := range batch.Ranks {
			javIDSet[item.MovieJavID] = struct{}{}
		}
		timing := affectedDayTimings[batch.DayNumber]
		timing.DayNumber = batch.DayNumber
		if batch.FetchedOn > timing.FetchedOn {
			timing.FetchedOn = batch.FetchedOn
		}
		batchSyncedOn := batch.SyncedOn
		if batchSyncedOn <= 0 {
			batchSyncedOn = syncedOn
		}
		if batchSyncedOn > timing.SyncedOn {
			timing.SyncedOn = batchSyncedOn
		}
		affectedDayTimings[batch.DayNumber] = timing
	}

	if len(affectedDayTimings) == 0 {
		return []dailybesttypes.RankDayTiming{}, nil
	}
	affectedCastIDs := make(map[int64]struct{})
	if err := l.UpdateMovieRankInfo(ctx, javIDSet, affectedCastIDs); err != nil {
		return nil, fmt.Errorf("更新影片排名信息失败: %w", err)
	}
	if err := l.rebuildCastRankStatsByIDs(ctx, affectedCastIDs); err != nil {
		return nil, fmt.Errorf("更新演员排名信息失败: %w", err)
	}
	return sortedRankDayTimings(affectedDayTimings), nil
}

type pendingBestinvRankBatch struct {
	Batch dailybesttypes.RankBatch
}

type bestinvRankBatchKey struct {
	BusinessDate string
	DayNumber    int64
	Category     int64
}

func buildPendingBestinvRankBatches(bestinvs []*types.Bestinv) ([]pendingBestinvRankBatch, error) {
	groups := make(map[bestinvRankBatchKey][]*types.Bestinv)
	for _, bestinv := range bestinvs {
		if bestinv == nil {
			return nil, fmt.Errorf("待处理 Bestinv 存在空记录")
		}
		key := bestinvRankBatchKey{
			BusinessDate: strings.TrimSpace(bestinv.Date),
			DayNumber:    bestinv.DayNumber,
			Category:     bestinv.Category,
		}
		groups[key] = append(groups[key], bestinv)
	}

	keys := make([]bestinvRankBatchKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].DayNumber != keys[j].DayNumber {
			return keys[i].DayNumber < keys[j].DayNumber
		}
		if keys[i].Category != keys[j].Category {
			return keys[i].Category < keys[j].Category
		}
		return keys[i].BusinessDate < keys[j].BusinessDate
	})

	result := make([]pendingBestinvRankBatch, 0, len(keys))
	for _, key := range keys {
		pages := groups[key]
		sort.Slice(pages, func(i, j int) bool { return pages[i].Page < pages[j].Page })
		ranks := make([]contracts.DailyBestRank, 0)
		fetchedOn := int64(0)
		for index, page := range pages {
			expectedPage := int64(index + 1)
			if page.Page != expectedPage {
				return nil, fmt.Errorf(
					"DailyBest date=%s category=%d 页面不完整: 期望第 %d 页，实际第 %d 页",
					key.BusinessDate,
					key.Category,
					expectedPage,
					page.Page,
				)
			}
			pageRanks, err := parseBestinvPageRanks(page)
			if err != nil {
				return nil, fmt.Errorf("解析 Bestinv(id=%d) 排名失败: %w", page.Id, err)
			}
			ranks = append(ranks, pageRanks...)
			pageFetchedOn := page.LastQueryTime
			if pageFetchedOn <= 0 {
				pageFetchedOn = page.CreatedOn
			}
			if pageFetchedOn > fetchedOn {
				fetchedOn = pageFetchedOn
			}
		}
		result = append(result, pendingBestinvRankBatch{Batch: dailybesttypes.RankBatch{
			BusinessDate: key.BusinessDate,
			DayNumber:    key.DayNumber,
			Category:     key.Category,
			FetchedOn:    fetchedOn,
			Ranks:        ranks,
		}})
	}
	return result, nil
}

func parseBestinvPageRanks(best *types.Bestinv) ([]contracts.DailyBestRank, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(best.Content)))
	if err != nil {
		return nil, err
	}

	ranks := make([]contracts.DailyBestRank, 0)
	rankPos := (best.Page-1)*20 + 1
	accepted := int64(0)
	doc.Find("div.videos div.video").Each(func(_ int, selection *goquery.Selection) {
		title := strings.TrimSpace(selection.Find("div.title").Text())
		if isBluRayTitle(title) {
			return
		}
		javID, _ := selection.Find("div.toolbar a").Attr("id")
		javID = strings.TrimSpace(javID)
		if javID == "" {
			return
		}
		ranks = append(ranks, contracts.DailyBestRank{
			MovieJavID: javID,
			DayNumber:  best.DayNumber,
			RankPos:    rankPos + accepted,
			Category:   best.Category,
		})
		accepted++
	})
	return ranks, nil
}

func normalizeDailyBestRankBatches(batches []dailybesttypes.RankBatch) ([]dailybesttypes.RankBatch, error) {
	normalized := make([]dailybesttypes.RankBatch, 0, len(batches))
	seenBatches := make(map[dailybesttypes.RankCategoryKey]struct{}, len(batches))
	for _, batch := range batches {
		batch.BusinessDate = strings.TrimSpace(batch.BusinessDate)
		dayNumber, err := contracts.RankDayNumber(batch.BusinessDate)
		if err != nil || dayNumber != batch.DayNumber {
			return nil, fmt.Errorf("DailyBest date=%s 日期序号校验失败", batch.BusinessDate)
		}
		if batch.Category <= 0 {
			return nil, fmt.Errorf("DailyBest date=%s 分类必须为正整数", batch.BusinessDate)
		}
		key := dailybesttypes.RankCategoryKey{DayNumber: batch.DayNumber, Category: batch.Category}
		if _, exists := seenBatches[key]; exists {
			return nil, fmt.Errorf(
				"DailyBest date=%s category=%d 存在重复批次",
				batch.BusinessDate,
				batch.Category,
			)
		}
		seenBatches[key] = struct{}{}
		if len(batch.Ranks) == 0 {
			return nil, fmt.Errorf(
				"DailyBest date=%s category=%d 排名为空",
				batch.BusinessDate,
				batch.Category,
			)
		}

		seenJavIDs := make(map[string]struct{}, len(batch.Ranks))
		seenRankPositions := make(map[int64]struct{}, len(batch.Ranks))
		ranks := make([]contracts.DailyBestRank, 0, len(batch.Ranks))
		for _, rank := range batch.Ranks {
			rank.MovieJavID = strings.TrimSpace(rank.MovieJavID)
			if rank.MovieJavID == "" || rank.RankPos <= 0 {
				return nil, fmt.Errorf(
					"DailyBest date=%s category=%d 排名明细无效",
					batch.BusinessDate,
					batch.Category,
				)
			}
			if rank.DayNumber != 0 && rank.DayNumber != batch.DayNumber {
				return nil, fmt.Errorf("DailyBest jav_id=%s 日期序号不一致", rank.MovieJavID)
			}
			if rank.Category != 0 && rank.Category != batch.Category {
				return nil, fmt.Errorf("DailyBest jav_id=%s 分类不一致", rank.MovieJavID)
			}
			if _, exists := seenJavIDs[rank.MovieJavID]; exists {
				return nil, fmt.Errorf("DailyBest 排名存在重复 JavID: %s", rank.MovieJavID)
			}
			if _, exists := seenRankPositions[rank.RankPos]; exists {
				return nil, fmt.Errorf("DailyBest 排名存在重复名次: %d", rank.RankPos)
			}
			seenJavIDs[rank.MovieJavID] = struct{}{}
			seenRankPositions[rank.RankPos] = struct{}{}
			rank.RankKey = types.BuildRankKey(batch.BusinessDate, batch.Category, rank.RankPos)
			rank.DayNumber = batch.DayNumber
			rank.Category = batch.Category
			ranks = append(ranks, rank)
		}
		sort.Slice(ranks, func(i, j int) bool {
			if ranks[i].RankPos != ranks[j].RankPos {
				return ranks[i].RankPos < ranks[j].RankPos
			}
			return ranks[i].MovieJavID < ranks[j].MovieJavID
		})
		batch.Ranks = ranks
		normalized = append(normalized, batch)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].DayNumber != normalized[j].DayNumber {
			return normalized[i].DayNumber < normalized[j].DayNumber
		}
		return normalized[i].Category < normalized[j].Category
	})
	return normalized, nil
}

func sortedRankDayTimings(timingByDay map[int64]dailybesttypes.RankDayTiming) []dailybesttypes.RankDayTiming {
	dayNumbers := make([]int64, 0, len(timingByDay))
	for dayNumber := range timingByDay {
		if dayNumber > 0 {
			dayNumbers = append(dayNumbers, dayNumber)
		}
	}
	sort.Slice(dayNumbers, func(i, j int) bool { return dayNumbers[i] < dayNumbers[j] })
	timings := make([]dailybesttypes.RankDayTiming, 0, len(dayNumbers))
	for _, dayNumber := range dayNumbers {
		timings = append(timings, timingByDay[dayNumber])
	}
	return timings
}

// UpdateMovieRankInfo 批量按 javId 聚合更新排行信息
func (l *Service) UpdateMovieRankInfo(ctx context.Context, javIdSet map[string]struct{}, affectedCastIDs map[int64]struct{}) error {
	log := l.deps.Log.WithContext(ctx)
	log.Infof("开始汇总并更新排行信息，待处理 %d 个 javId", len(javIdSet))

	updated := 0
	for javId := range javIdSet {
		if err := l.AddRankInfo(ctx, javId, affectedCastIDs); err != nil {
			return fmt.Errorf("更新 javId=%s 的排行信息失败: %w", javId, err)
		}
		updated++
	}
	log.Infof("排行信息更新完成，共处理 %d 个", updated)
	return nil
}

// AddRankInfo 计算单个影片的：首次上榜日、最佳名次、上榜天数，并写回 bm_minfo
func (l *Service) AddRankInfo(ctx context.Context, javId string, affectedCastIDs map[int64]struct{}) error {
	// 1) 统计聚合（SQL 聚合更快；没有就先在 Repo 做一层）
	firstDay, bestRank, daysInRank, err := l.deps.RankRepo.AggregateByJavId(ctx, javId)
	if err != nil {
		return fmt.Errorf("统计排行聚合失败(javId=%s): %w", javId, err)
	}

	// 2) 取得 movie.id（用于回刷关联演员缓存 / 侧写）
	mv, err := l.deps.MovieRepo.FindOneByJavId(ctx, javId)
	if err != nil {
		return fmt.Errorf("查询电影失败(javId=%s): %w", javId, err)
	}
	// 3) 取演员 ids（保持与老项目等价的副作用：填充内存 Map）
	castIDs, err := l.deps.MovieCastRepo.ListCastIDsByMovieJavId(ctx, mv.JavId)
	if err != nil {
		return fmt.Errorf("查询电影演员关系失败(movieId=%d): %w", mv.Id, err)
	}
	for _, cid := range castIDs {
		affectedCastIDs[cid] = struct{}{}
	}

	// 4) 回写 bm_minfo 的排行三件套
	now := time.Now().Unix()

	patch := types.MinfoPatch{
		Chinese:            nil,
		FirstRankDayNumber: ptr.Int64(firstDay),
		HighestRank:        ptr.Int64(bestRank),
		DaysInRank:         ptr.Int64(daysInRank),
		UpdatedOn:          ptr.Int64(now),
	}
	err = l.deps.MinfoRepo.UpdatePartialByJavId(ctx, javId, patch)
	if err != nil {
		return fmt.Errorf("写回 bm_minfo 排行信息失败(javId=%s): %w", javId, err)
	}

	l.movieSvc.InvalidateMovieType(ctx, javId)

	return nil
}
