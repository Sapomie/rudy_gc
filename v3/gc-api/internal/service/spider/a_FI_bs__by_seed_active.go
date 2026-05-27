// internal/spiderx/logic/a_FI_bs__by_seed_active.go
package spider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"

	"github.com/sirupsen/logrus"
)

// ---- 与老项目保持一致的常量 ----
const (
	// NameType
	nameTypePrefix int64 = 1 // QueryNamePrefix
	nameTypeLabel  int64 = 2 // QueryNameLabel

	// SearchType
	searchByOffset   int64 = 1 // QueryByOffset
	searchByStartEnd int64 = 2 // QueryByStartEnd

	findByPrefixLargestPage = 72
)

var ErrBlankPage = errors.New("blank page")

func (l *CrawlLogic) FetchInventoriesBySeedActive(ctx context.Context) error {
	log := l.deps.Log.WithContext(ctx)
	log.Info("FetchInventoriesBySeedActive: begin")

	prefixSeeds, err := l.deps.SeedRepo.FindActiveByNameType(ctx, nameTypePrefix)
	if err != nil {
		return fmt.Errorf("FindActiveSeeds(nameType=%d): %w", nameTypePrefix, err)
	}
	labelSeeds, err := l.deps.SeedRepo.FindActiveByNameType(ctx, nameTypeLabel)
	if err != nil {
		return fmt.Errorf("FindActiveSeeds(nameType=%d): %w", nameTypeLabel, err)
	}

	totalSeeds := len(prefixSeeds) + len(labelSeeds)
	l.reportPhaseProgress(ctx, "seed", "seed_queue_ready", fmt.Sprintf("待处理种子 %d 个", totalSeeds), 0, totalSeeds, 0, 0)

	processed := 0
	if err := l.fetchSeedBatch(ctx, nameTypePrefix, prefixSeeds, &processed, totalSeeds); err != nil {
		return err
	}
	if err := l.fetchSeedBatch(ctx, nameTypeLabel, labelSeeds, &processed, totalSeeds); err != nil {
		return err
	}

	log.Info("FetchInventoriesBySeedActive: done")
	return nil
}

func (l *CrawlLogic) FetchInventoriesBySeedName(ctx context.Context, seedName string) error {
	seed, err := l.deps.SeedRepo.FindOneByName(ctx, seedName)
	if err != nil {
		return errors.New("seed not found " + seedName)
	}

	l.reportPhaseProgress(ctx, "seed", "seed_queue_ready", "待处理种子 1 个", 0, 1, 0, 0)
	if err := l.handleSeed(ctx, seed); err != nil {
		return fmt.Errorf("handleSeed(name=%s): %w", seed.Name, err)
	}
	l.reportPhaseProgress(ctx, "seed", "seed_item_done", fmt.Sprintf("种子完成：%s", seed.Name), 1, 1, 1, 0)
	return nil
}

func (l *CrawlLogic) fetchSeedBatch(ctx context.Context, nameType int64, seeds []*types.Seed, processed *int, total int) error {
	log := l.deps.Log.WithContext(ctx)
	log.WithFields(logrus.Fields{
		"nameType": nameType,
		"count":    len(seeds),
	}).Info("active seeds fetched")

	for i, s := range seeds {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		if err := l.handleSeed(ctx, s); err != nil {
			return fmt.Errorf("handleSeed(name=%s): %w", s.Name, err)
		}
		// ✅ 进度日志
		log.Infof("fetchByNameType: processed %d/%d seeds (%s)", i+1, len(seeds), s.Name)
		if processed != nil {
			*processed = *processed + 1
			l.reportPhaseProgress(ctx, "seed", "seed_item_done", fmt.Sprintf("种子完成：%s", s.Name), *processed, total, *processed, 0)
		}
		if err := l.sleepWithContext(ctx, getRandomSleepDuration()); err != nil {
			return err
		}
	}
	return nil
}

// 处理单个 seed：计算页区间 -> 逐页抓取并保存 -> 推进断点
func (l *CrawlLogic) handleSeed(ctx context.Context, s *types.Seed) error {
	log := l.deps.Log.WithContext(ctx)
	pageStart, pageEnd := determinePageRange(s)
	if pageStart <= 0 {
		pageStart = 1
	}
	if pageEnd < pageStart {
		// 无需抓取
		return nil
	}

	log.WithFields(logrus.Fields{
		"name":       s.Name,
		"searchType": s.SearchType,
		"pageStart":  pageStart,
		"pageEnd":    pageEnd,
	}).Info("seed begin")

	newPageNow := s.PageNow
	queryBy := buildQueryPath(s.NameType, s.Name) // 与老项目一致

	for p := pageStart; p <= pageEnd; p++ {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		// 抓取并保存单页
		if err := l.fetchAndSaveInventory(ctx, s.NameType, s.Name, queryBy, p); err != nil {
			if errors.Is(err, ErrBlankPage) {
				newPageNow = p - 1
				log.WithFields(logrus.Fields{
					"name":       s.Name,
					"page":       p,
					"newPageNow": newPageNow,
				}).Info("blank page hit, stop range")
				break
			}
			// 其它错误直接返回，让上层感知（可按需改成“记录后继续”）
			return fmt.Errorf("fetchAndSaveInventory(name=%s,page=%d): %w", s.Name, p, err)
		}
		newPageNow = p
		l.reportProgress(ctx, "seed_page_done", fmt.Sprintf("已抓取 %s 第 %d 页", s.Name, p), int(p-pageStart+1), int(p-pageStart+1), 0, int(pageEnd-p))
		// 微小抖动
		if err := l.sleepWithContext(ctx, getRandomSleepDuration()); err != nil {
			return err
		}
	}

	// 回写进度：ok/empty
	status := consts.SeedStatusOK
	errMsg := ""
	if newPageNow < s.PageNow {
		status = consts.SeedStatusEmpty
	}
	stats, statsErr := l.deps.SeedRepo.CalcMovieStats(ctx, s)
	if statsErr != nil {
		log.Errorf("calc seed movie stats failed: %v", statsErr)
	}
	if err := l.deps.SeedRepo.UpdateProgressAndMovieStats(
		ctx, s.Id, newPageNow, time.Now().Unix(), status, errMsg, stats,
	); err != nil {
		log.Errorf("update seed progress failed: %v", err)
	}

	log.WithFields(logrus.Fields{
		"name":       s.Name,
		"pageNowOld": s.PageNow,
		"pageNowNew": newPageNow,
	}).Info("seed progress")

	return nil
}

// determinePageRange 依据 SearchType 计算起止页
func determinePageRange(s *types.Seed) (start int64, end int64) {
	switch s.SearchType {
	case searchByOffset:
		start = s.PageNow - s.Offset
		if start < 1 {
			start = 1
		}
		end = findByPrefixLargestPage
	case searchByStartEnd:
		start = s.StartPage
		end = s.EndPage
	default:
		// 容错：未知类型按 offset 处理
		start = s.PageNow - s.Offset
		if start < 1 {
			start = 1
		}
		end = findByPrefixLargestPage
	}
	return
}

// 抓取并保存到 raw_inventory
func (l *CrawlLogic) fetchAndSaveInventory(ctx context.Context, nameType int64, keyword, queryBy string, page int64) error {
	// 构造 URL（与老项目一致）：https://{JavAddress}/cn + /{queryBy}&page={page}
	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
	base := fmt.Sprintf("https://%s/cn", l.deps.Config.Fetcher.JavAddress)
	fullURL := base + queryWithPage

	// 抓取（带重试、空页判定；内部已走 l.deps.Fetcher.Get）
	content, err := l.fetchInventoryContentWithRetry(ctx, fullURL)
	if err != nil {
		return err
	}

	// 生成 inventory 名称（Label 类追加日期后缀）
	now := time.Now()
	name := buildInventoryName(queryWithPage, nameType, now)

	// 落库 raw_inventory（Upsert）
	inv := &types.Inventory{
		Name:          name,
		NeedScan:      consts.InventoryNeedScan,
		Keyword:       keyword,
		Parent:        queryBy,
		Page:          page,
		Content:       content,
		Category:      nameType,
		LastQueryTime: now.Unix(),
		CreatedOn:     now.Unix(),
		UpdatedOn:     now.Unix(),
	}
	if err := l.deps.InventoryRepo.Upsert(ctx, inv); err != nil {
		return fmt.Errorf("save inventory failed: %w", err)
	}

	//log

	return nil
}

// 与老项目保持一致的查询片段
func buildQueryPath(nameType int64, name string) string {
	switch nameType {
	case nameTypeLabel:
		return fmt.Sprintf("vl_label.php?&mode=2&l=%v", name)
	case nameTypePrefix:
		return fmt.Sprintf("vl_searchbyid.php?&keyword=%v", name)
	default:
		return fmt.Sprintf("vl_searchbyid.php?&keyword=%v", name)
	}
}

func getRandomSleepDuration() time.Duration {
	return time.Second * time.Duration(time.Now().UnixNano()%3+2)
}
