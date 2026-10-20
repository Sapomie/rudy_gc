// internal/spiderx/types/notification.go
package types

// CrawlAction 定义爬虫调度时的动作类型
type CrawlAction string

const (
	// 按启用的查询种子抓取库存（非全量全站）
	ActionActiveQueries CrawlAction = "active_queries"

	// 每日 Best Inventory（每日排行，沿用 bestinv 命名）
	ActionDailyBestinv CrawlAction = "daily_bestinv"

	// 同步远程的每日 Best Inventory（用于补历史或本地非常驻时追补）
	ActionSyncDailyBestinv CrawlAction = "sync_daily_bestinv"
)

// Notification 是 loop/logic/transport 等模块传递的消息
type Notification struct {
	Action CrawlAction
	Meta   map[string]string // 可选元信息：site、date、batch_id 等
}
