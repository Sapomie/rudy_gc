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

	HasDetail           int64
	HasDownloadCover    int64
	HasChinese          int64
	DetailNeedScan      int64
	DetailBirthTime     int64
	LastQueryDetailTime int64

	CreatedOn int64
	UpdatedOn int64
}
