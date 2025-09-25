package types

// NeedScan 常量：是否需要后续解析扫描
const (
	InventoryNoNeedScan int64 = 0
	InventoryNeedScan   int64 = 1
)

// Category 常量：与老项目一致
const (
	InventoryCategoryByPrefix int64 = 1
	InventoryCategoryByLabel  int64 = 2
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
