package fetchsite

type SukebeiPageQuery struct {
	Page           int64
	PageSize       int64
	Sort           string
	Order          string
	Keyword        string
	Status         int64
	StatusSet      bool
	HasErrorOnly   bool
	HasNoErrorOnly bool
}

type SukebeiPageItem struct {
	MovieJavID        string
	MovieCode         string
	ReleaseDate       int64
	ReleaseDateText   string
	FetchStatus       int64
	FetchStatusText   string
	TryCount          int64
	LastFetchTime     int64
	LastFetchText     string
	LastResultCount   int64
	TorrentHashCount  int64
	LatestPublishTime int64
	LatestPublishText string
	LastError         string
}

type SukebeiPageResult struct {
	Items    []*SukebeiPageItem
	Page     int64
	PageSize int64
	Total    int64
}
