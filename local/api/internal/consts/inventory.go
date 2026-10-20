package consts

// Bestinv 分类：月榜 / 总榜
const (
	BestCategoryMonth   = 1 + iota // 1
	BestCategoryAllTime            // 2
)

// 是否需要排名检查
const (
	BestinvNeedRankCheck   = 1 + iota // 1
	BestinvNoNeedRankCheck            // 2
)

// 是否需要扫描（旧逻辑 Inventory 用过，可以沿用）
const (
	BestinvNeedScan   = 1 + iota // 1
	BestinvNoNeedScan            // 2
)

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
