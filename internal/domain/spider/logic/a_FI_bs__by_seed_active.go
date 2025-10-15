// internal/spiderx/logic/a_FI_bs__by_seed_active.go
package logic

import (
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

// FetchInventoriesBySeedActive
// - 读取启用的 seed 列表（prefix/label）
// - 依据断点(pageNow/start-end/offset)分页请求
// - 保存至 raw_inventory（落库见 fetchAndSaveInventory）
// - 成功页推进断点；空页/异常的退避与记录
func (l *CrawlLogic) FetchInventoriesBySeedActive() error {
	l.deps.Log.Info("FetchInventoriesBySeedActive: begin")

	// 1) Prefix
	if err := l.fetchByNameType(nameTypePrefix); err != nil {
		return err
	}
	// 2) Label
	if err := l.fetchByNameType(nameTypeLabel); err != nil {
		return err
	}

	l.deps.Log.Info("FetchInventoriesBySeedActive: done")
	return nil
}

func (l *CrawlLogic) fetchByNameType(nameType int64) error {
	seeds, err := l.deps.SeedRepo.FindActiveByNameType(l.ctx, nameType)
	if err != nil {
		return fmt.Errorf("FindActiveSeeds(nameType=%d): %w", nameType, err)
	}

	l.deps.Log.WithFields(logrus.Fields{
		"nameType": nameType,
		"count":    len(seeds),
	}).Info("active seeds fetched")

	for i, s := range seeds {
		if err := l.handleSeed(s); err != nil {
			return err
		}
		// 轻微打点 + 随机 sleep，避免被限流
		l.deps.Log.WithFields(logrus.Fields{
			"idx":   i + 1,
			"total": len(seeds),
			"name":  s.Name,
		}).Info("seed done")

		time.Sleep(getRandomSleepDuration())
	}
	return nil
}

// 处理单个 seed：计算页区间 -> 逐页抓取并保存 -> 推进断点
func (l *CrawlLogic) handleSeed(s *types.Seed) error {
	pageStart, pageEnd := determinePageRange(s)
	if pageStart <= 0 {
		pageStart = 1
	}
	if pageEnd < pageStart {
		// 无需抓取
		return nil
	}

	l.deps.Log.WithFields(logrus.Fields{
		"name":       s.Name,
		"searchType": s.SearchType,
		"pageStart":  pageStart,
		"pageEnd":    pageEnd,
	}).Info("seed begin")

	newPageNow := s.PageNow
	queryBy := buildQueryPath(s.NameType, s.Name) // 与老项目一致

	for p := pageStart; p <= pageEnd; p++ {
		// 抓取并保存单页
		if err := l.fetchAndSaveInventory(s.NameType, s.Name, queryBy, p); err != nil {
			if errors.Is(err, ErrBlankPage) {
				newPageNow = p - 1
				l.deps.Log.WithFields(logrus.Fields{
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
		// 微小抖动
		time.Sleep(getRandomSleepDuration())
	}

	// 回写进度：ok/empty
	status := consts.SeedStatusOK
	errMsg := ""
	if newPageNow < s.PageNow {
		status = consts.SeedStatusEmpty
	}
	if err := l.deps.SeedRepo.UpdateProgress(
		l.ctx, s.Id, newPageNow, time.Now().Unix(), status, errMsg,
	); err != nil {
		l.deps.Log.Errorf("update seed progress failed: %v", err)
	}

	l.deps.Log.WithFields(logrus.Fields{
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
func (l *CrawlLogic) fetchAndSaveInventory(nameType int64, keyword, queryBy string, page int64) error {
	// 构造 URL（与老项目一致）：https://{JavAddress}/cn + /{queryBy}&page={page}
	queryWithPage := fmt.Sprintf("/%s&page=%d", queryBy, page)
	base := fmt.Sprintf("https://%s/cn", l.deps.Config.Fetcher.JavAddress)
	fullURL := base + queryWithPage

	// 抓取（带重试、空页判定；内部已走 l.deps.Fetcher.Get）
	content, err := l.fetchInventoryContentWithRetry(fullURL)
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
	if err := l.deps.InventoryRepo.Upsert(l.ctx, inv); err != nil {
		return fmt.Errorf("save inventory failed: %w", err)
	}

	l.deps.Log.WithFields(logrus.Fields{
		"url":      fullURL,
		"name":     name,
		"nameType": nameType,
		"keyword":  keyword,
		"page":     page,
		"bytes":    len(content),
	}).Info("fetched page")

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
