package types

// Item 表示解析出的影片条目
type Item struct {
	Id         int64
	Name       string
	JavId      string
	Prefix     string
	SearchType int64
	CoverUrl   string
	SearchBy   string

	HasDetail        int64
	HasDownloadCover int64
	HasChinese       int64
	DetailNeedScan   int64
	DetailBirthTime  int64
	DetailUpdateTime int64

	CreatedOn int64
	UpdatedOn int64
}

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
