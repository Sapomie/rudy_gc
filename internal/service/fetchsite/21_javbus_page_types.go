package fetchsite

type JavbusPageQuery struct {
	Page               int64
	PageSize           int64
	Sort               string
	Order              string
	MediaOwned         int64
	Keyword            string
	Statuses           []int64
	HasStatuses        bool
	LastFetchFrom      int64
	LastFetchTo        int64
	HasLastFetchFrom   bool
	HasLastFetchTo     bool
	ReleaseDateFrom    int64
	ReleaseDateTo      int64
	HasReleaseDateFrom bool
	HasReleaseDateTo   bool
	MediaBirthFrom     int64
	MediaBirthTo       int64
	HasMediaBirthFrom  bool
	HasMediaBirthTo    bool
}

type JavbusPageItem struct {
	MovieJavID          string
	MovieName           string
	ReleaseDate         int64
	ReleaseDateText     string
	FetchStatus         int64
	FetchStatusText     string
	TryCount            int64
	LastFetchTime       int64
	LastFetchText       string
	LastResultCount     int64
	TorrentHashCount    int64
	LatestPublishTime   int64
	LatestPublishText   string
	LastError           string
	OwnedWMedia         int64
	VideoURLWMedia      string
	FilmBirthDateWMedia string
	WMediaOwnedText     string
	WMediaBirthText     string
}

type JavbusPageResult struct {
	Items        []*JavbusPageItem
	Page         int64
	PageSize     int64
	Total        int64
	SuccessCount int64
	PendingCount int64
	FailedCount  int64
}
