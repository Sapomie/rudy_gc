package spider

type CrawlAction string

const (
	// 按启用的 query 爬库存（非“全量全站”）
	ActionActiveQueries CrawlAction = "active_queries"

	// 每日 Best Inventory（你的“每日排行”概念，沿用 bestinv 命名）
	ActionDailyBestinv CrawlAction = "daily_bestinv"

	// 同步远程的每日 Best Inventory（用于补历史/本地非常驻时追补）
	ActionSyncDailyBestinv CrawlAction = "sync_daily_bestinv"
)

type Notification struct {
	Action CrawlAction
	Meta   map[string]string // 可选：site、date、batch_id 等
}
