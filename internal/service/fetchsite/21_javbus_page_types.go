package fetchsite

type JavbusPageQuery struct {
	Page               int64
	PageSize           int64
	Sort               string
	Order              string
	Keyword            string
	Status             int64
	StatusSet          bool
	HasErrorOnly       bool
	HasNoErrorOnly     bool
	LastFetchFrom      int64
	LastFetchTo        int64
	HasLastFetchFrom   bool
	HasLastFetchTo     bool
	ReleaseDateFrom    int64
	ReleaseDateTo      int64
	HasReleaseDateFrom bool
	HasReleaseDateTo   bool
}

type JavbusPageItem struct {
	MovieJavID        string
	MovieName         string
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

type JavbusPageResult struct {
	Items    []*JavbusPageItem
	Page     int64
	PageSize int64
	Total    int64
}
