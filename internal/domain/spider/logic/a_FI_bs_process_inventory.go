// internal/spiderx/logic/a_FI_bs_process_inventory.go
package logic

import (
	"fmt"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func (l *CrawlLogic) ProcessInventory() error {
	log := logx.WithContext(l.ctx)

	ids, err := l.deps.InventoryRepo.ListNeedScanIDs(l.ctx, 10000) // 先给一个上限，避免一次性全扫
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

func (l *CrawlLogic) makeAndInsertItemsByInventory(inv *types.Inventory) error {
	if inv == nil {
		return fmt.Errorf("nil inventory")
	}

	if err := l.makeAndInsertItems(inv.Content, inv.Name, getInventorySearchType(inv.Category)); err != nil {
		return err
	}
	return nil
}

func getInventorySearchType(category int64) int64 {

	switch category {
	case consts.InventoryCategoryByPrefix:
		return consts.ItemSearchByPrefix
	case consts.InventoryCategoryByLabel:
		return consts.ItemSearchByLabel
	default:
		return 0
	}

}
