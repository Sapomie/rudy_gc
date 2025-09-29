package types

// NeedScan：1=需要扫描, 2=不需要扫描
const (
	InventoryNeedScan   int64 = 1 + iota // 1
	InventoryNoNeedScan                  // 2
)

// Category：1=Prefix, 2=Label
const (
	InventoryCategoryByPrefix int64 = 1 + iota // 1
	InventoryCategoryByLabel                   // 2
)

// Inventory 供业务层/仓储层使用的领域模型（与 modelx.DInventory 一一对应）
type Inventory struct {
	Id            int64
	Name          string // 唯一名(由query+page组合)
	NeedScan      int64  // 0/1
	Keyword       string
	Parent        string
	Page          int64
	Content       string
	Category      int64 // 1=Prefix,2=Label
	LastQueryTime int64 // Unix秒

	CreatedOn int64
	UpdatedOn int64
}
