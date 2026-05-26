package types

type MovieDetailResponse struct {
	Movie         *MovieCard              `json:"movie"`
	Media         *MovieMedia             `json:"media"`
	RankInfos     []*MovieRankInfo        `json:"rank_infos"`
	ScEvents      []*MovieScEvent         `json:"sc_events"`
	JavbusFetch   *MovieFetchSiteStatus   `json:"javbus_fetch"`
	JavbusMagnets []*MovieJavbusMagnet    `json:"javbus_magnets"`
	SukebeiFetch  *MovieFetchSiteStatus   `json:"sukebei_fetch"`
	SukebeiRows   []*MovieSukebeiTorrent  `json:"sukebei_rows"`
	SehuatangRows []*MovieSehuatangMagnet `json:"sehuatang_rows"`
}

type MovieMedia struct {
	MovieJavID        string  `json:"movie_jav_id"`
	MovieName         string  `json:"movie_name"`
	FileName          string  `json:"file_name"`
	Directory         string  `json:"directory"`
	FilePath          string  `json:"file_path"`
	SourceTorrentHash string  `json:"source_torrent_hash"`
	BirthTime         string  `json:"birth_time"`
	Size              float64 `json:"size"`
	Height            int64   `json:"height"`
	BitRate           float64 `json:"bit_rate"`
	DurationMinutes   float64 `json:"duration_minutes"`
	Frame             int64   `json:"frame"`
	SelfMake          int64   `json:"self_make"`
	IsRemoved         int64   `json:"is_removed"`
}

type MovieRankInfo struct {
	Date string `json:"date"`
	Rank int64  `json:"rank"`
	Href string `json:"href"`
}

type MovieScEvent struct {
	ScName   string `json:"sc_name"`
	ScTime   int64  `json:"sc_time"`
	Cooldown int64  `json:"cooldown"`
	IsCome   bool   `json:"is_come"`
	Href     string `json:"href"`
}

type MovieFetchSiteStatus struct {
	MovieJavID        string `json:"movie_jav_id"`
	MovieName         string `json:"movie_name"`
	ReleaseDate       string `json:"release_date"`
	FetchStatus       int64  `json:"fetch_status"`
	FetchStatusText   string `json:"fetch_status_text"`
	TryCount          int64  `json:"try_count"`
	LastFetchTime     string `json:"last_fetch_time"`
	LastFetchAgo      string `json:"last_fetch_ago"`
	LastError         string `json:"last_error"`
	TorrentHashCount  int64  `json:"torrent_hash_count"`
	LatestPublishTime string `json:"latest_publish_time"`
	SourceURL         string `json:"source_url"`
}

type MovieJavbusMagnet struct {
	RowID          int64  `json:"row_id"`
	MagnetName     string `json:"magnet_name"`
	InfoHash       string `json:"info_hash"`
	SizeText       string `json:"size_text"`
	ShareDate      string `json:"share_date"`
	HasHD          bool   `json:"has_hd"`
	HasSubtitle    bool   `json:"has_subtitle"`
	HasMatchedHash bool   `json:"has_matched_hash"`
}

type MovieSukebeiTorrent struct {
	RowID          int64  `json:"row_id"`
	TorrentTitle   string `json:"torrent_title"`
	ViewURL        string `json:"view_url"`
	InfoHash       string `json:"info_hash"`
	SizeText       string `json:"size_text"`
	PublishTime    string `json:"publish_time"`
	Seeders        int64  `json:"seeders"`
	Leechers       int64  `json:"leechers"`
	Completed      int64  `json:"completed"`
	HasMatchedHash bool   `json:"has_matched_hash"`
}

type MovieSehuatangMagnet struct {
	RowID          int64  `json:"row_id"`
	ThreadTitle    string `json:"thread_title"`
	ThreadURL      string `json:"thread_url"`
	InfoHash       string `json:"info_hash"`
	PostTime       string `json:"post_time"`
	HasMatchedHash bool   `json:"has_matched_hash"`
}
