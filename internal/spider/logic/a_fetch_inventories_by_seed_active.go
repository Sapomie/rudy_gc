package logic

func (l *CrawlLogic) FetchInventoriesBySeedActive() error {
	// TODO:
	// - 读取启用的 seed 列表
	// - 依据断点(pageNow/start-end/offset)分页请求
	// - 将整页 HTML/JSON 写入 raw_inventory（建议带 content_hash/状态码/抓取时间）
	// - 成功页推进断点；空页/异常的退避与记录
	return nil
}
