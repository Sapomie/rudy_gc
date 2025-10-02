// internal/spiderx/logic/a_FI_bs_process_inventory.go
package logic

import (
	"fmt"
	consts "rudy_gc/internal/cnonsts"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zeromicro/go-zero/core/logx"

	"rudy_gc/internal/types"
)

func (l *CrawlLogic) ProcessInventory() error {
	log := logx.WithContext(l.ctx)

	ids, err := l.deps.InventoryRepo.ListNeedScanIDs(l.ctx, 1000) // 先给一个上限，避免一次性全扫
	if err != nil {
		return fmt.Errorf("list need-scan inventory ids: %w", err)
	}
	if len(ids) == 0 {
		log.Info("ProcessInventory: nothing to scan")
		return nil
	}
	log.Infof("ProcessInventory: %d inventories to scan", len(ids))

	for _, id := range ids {
		inv, err := l.deps.InventoryRepo.FindOne(l.ctx, id)
		if err != nil {
			return fmt.Errorf("find inventory id=%d: %w", id, err)
		}
		if err := l.makeAndInsertItemsByInventory(inv); err != nil {
			return err
		}
		// 标记已扫描
		if err := l.deps.InventoryRepo.MarkScanned(l.ctx, id, time.Now().Unix()); err != nil {
			return fmt.Errorf("mark inventory scanned id=%d: %w", id, err)
		}
	}

	log.Info("ProcessInventory: done")
	return nil
}

// makeAndInsertItemsByInventory 处理单个 inventory 的 content
func (l *CrawlLogic) makeAndInsertItemsByInventory(inv *types.Inventory) error {
	if inv == nil {
		return fmt.Errorf("nil inventory")
	}
	// searchBy 使用 inventory.Name（与老代码保持：用页面唯一名做来源标记）
	if err := l.makeAndInsertItems(inv.Content, inv.Name, inv.Category); err != nil {
		return err
	}
	return nil
}

func (l *CrawlLogic) makeAndInsertItems(content, searchBy string, category int64) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}

	searchType := mapInventoryCategoryToItemSearchType(category)
	var toInsert []*types.Item

	// 兼容旧站点结构：div.videos div.video
	doc.Find("div.videos div.video").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("div.title").Text())
		if isBluRayTitle(title) {
			return
		}

		name := strings.TrimSpace(s.Find("div.id").Text())
		if name == "" {
			return
		}

		javID, _ := s.Find("div.toolbar a").Attr("id")
		javID = strings.TrimSpace(javID)
		if javID == "" {
			return
		}

		coverURL, _ := s.Find("img").Attr("src")
		coverURL = strings.TrimSpace(coverURL)

		prefix, _ := splitPrefixAndNumber(name)
		if isRedundantPrefix(prefix) {
			logx.WithContext(l.ctx).Infow("skip redundant prefix", logx.Field("name", name), logx.Field("prefix", prefix))
			return
		}

		now := time.Now().Unix()
		it := &types.Item{
			Name:             name,
			JavId:            javID,
			Prefix:           prefix,
			SearchType:       searchType,
			CoverUrl:         coverURL,
			SearchBy:         searchBy,
			CreatedOn:        now,
			UpdatedOn:        now,
			HasDetail:        types.ItemDetailNone,
			HasDownloadCover: types.ItemCoverNone,
			HasChinese:       types.ItemChineseNone,
			DetailBirthTime:  0,
			DetailUpdateTime: 0,
			DetailNeedScan:   types.ItemDetailStatusNeedScan,
		}
		toInsert = append(toInsert, it)
	})

	// 幂等插入
	for _, it := range toInsert {
		inserted, err := l.deps.ItemRepo.TryInsert(l.ctx, it)
		if err != nil {
			return fmt.Errorf("insert item name=%s javId=%s: %w", it.Name, it.JavId, err)
		}
		if inserted {
			logx.WithContext(l.ctx).Infow("item inserted", logx.Field("name", it.Name), logx.Field("javId", it.JavId))
		}
	}

	return nil
}

// ---------- 辅助函数 ----------

// mapInventoryCategoryToItemSearchType 将 inventory 的类别映射到 item 的 SearchType
// 与旧项目行为一致：Prefix→Prefix，Label→Label；未知保持原值
func mapInventoryCategoryToItemSearchType(cat int64) int64 {
	switch cat {
	case types.InventoryCategoryByPrefix:
		if v := types.ItemSearchByPrefix; v != 0 {
			return v
		}
		return cat
	case types.InventoryCategoryByLabel:
		if v := types.ItemSearchByLabel; v != 0 {
			return v
		}
		return cat
	default:
		return cat
	}
}

// splitPrefixAndNumber 从 "ABC-123" 拆出 "ABC" 和 "123"
func splitPrefixAndNumber(name string) (prefix string, number string) {
	if idx := strings.Index(name, "-"); idx != -1 {
		return name[:idx], name[idx+1:]
	}
	return name, ""
}

// isBluRayTitle 过滤蓝光等不需要的条目（尽量保守，不绑死具体常量）
func isBluRayTitle(title string) bool {
	t := strings.ToLower(title)
	return strings.Contains(t, "blu-ray") || strings.Contains(title, "蓝光")
}

// isRedundantPrefix 冗余前缀过滤（可接入你的 cons.RedundantPrefixSet）
// 这里先留个可替换点，默认不过滤
func isRedundantPrefix(prefix string) bool {
	_, ok := consts.RedundantPrefixSet[prefix]
	return ok
}
