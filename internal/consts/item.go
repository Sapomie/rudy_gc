package consts

const (
	SeedStatusIdle  int64 = 0 // 空闲，尚未执行
	SeedStatusOK    int64 = 1 // 正常，抓取成功
	SeedStatusEmpty int64 = 2 // 空页，提前停止
	SeedStatusError int64 = 3 // 错误
)

// ItemSearchType
const (
	ItemSearchByPrefix int64 = 1 + iota
	ItemSearchByLabel
	ItemSearchByBestMonth
	ItemSearchByBestAllTime
	ItemSearchByOld
)

// Item 业务映射状态
const (
	ItemBusNone     int64 = 1 + iota // 无业务映射
	ItemBusOK                        // 已建立
	ItemBusNotExist                  // 映射不存在
)

// Item 中文字幕状态
const (
	ItemChineseNone      int64 = 1 + iota // 无中文字幕
	ItemChineseOK                         // 有中文字幕
	ItemChineseError                      // 翻译错误
	ItemChineseSensitive                  // 含敏感词
)

// Item 封面状态
const (
	ItemCoverNone  int64 = 1 + iota // 无本地封面
	ItemCoverOK                     // 已有本地封面
	ItemCoverWrong                  // 错误封面链接
)

// Item 详情状态
const (
	ItemDetailNone int64 = 1 + iota // 无详情
	ItemDetailOK                    // 已有详情
)

// Item 封面状态
const (
	ItemDetailStatusNeedScan = 1 + iota
	ItemDetailStatusNoNeedScan
	ItemDetailStatusWrongContent
)
