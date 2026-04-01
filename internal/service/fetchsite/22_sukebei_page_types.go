package fetchsite

type SukebeiPageQuery struct {
	Page               int64
	PageSize           int64
	Sort               string
	Order              string
	Owned              int64
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
	FilmBirthFrom      int64
	FilmBirthTo        int64
	HasFilmBirthFrom   bool
	HasFilmBirthTo     bool
	MediaBirthFrom     int64
	MediaBirthTo       int64
	HasMediaBirthFrom  bool
	HasMediaBirthTo    bool
}

type SukebeiPageItem struct {
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
	Owned               int64
	OwnedWMedia         int64
	VideoURL            string
	VideoURLWMedia      string
	FilmBirthDate       string
	FilmBirthDateWMedia string
	VFilmOwnedText      string
	VFilmBirthText      string
	WMediaOwnedText     string
	WMediaBirthText     string
}

type SukebeiPageResult struct {
	Items        []*SukebeiPageItem
	Page         int64
	PageSize     int64
	Total        int64
	SuccessCount int64
	PendingCount int64
	FailedCount  int64
}
