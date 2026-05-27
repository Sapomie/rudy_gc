package types

type MovieDetail struct {
	MovieType        *MovieType
	MediaInfo        *FilmInfo
	RankInfos        []*RankInfo
	SC               []*MovieScEvent
	JavbusFetch      *MovieFetchSiteStatus
	JavbusMagnets    []*MovieJavbusMagnet
	SukebeiFetch     *MovieFetchSiteStatus
	SukebeiTorrents  []*MovieSukebeiTorrent
	SehuatangMagnets []*MovieSehuatangMagnet
}

type RankInfo struct {
	Date string
	Rank int64
}

type MovieScEvent struct {
	ScName   string
	ScTime   int64
	Cooldown int64
	IsCome   bool
	Href     string
}

type MovieFetchSiteStatus struct {
	MovieJavID        string
	MovieName         string
	ReleaseDate       string
	FetchStatus       int64
	FetchStatusText   string
	TryCount          int64
	LastFetchTime     string
	LastFetchAgo      string
	LastError         string
	TorrentHashCount  int64
	LatestPublishTime string
	SourceURL         string
}

type MovieJavbusMagnet struct {
	RowID          int64
	MagnetName     string
	InfoHash       string
	SizeText       string
	ShareDate      string
	HasHD          bool
	HasSubtitle    bool
	HasMatchedHash bool
	IsFavorited    bool
	IsInDownload   bool
	IsInPending    bool
}

type MovieSukebeiTorrent struct {
	RowID          int64
	TorrentTitle   string
	ViewURL        string
	InfoHash       string
	SizeText       string
	PublishTime    string
	Seeders        int64
	Leechers       int64
	Completed      int64
	HasMatchedHash bool
	IsFavorited    bool
	IsInDownload   bool
	IsInPending    bool
}

type MovieSehuatangMagnet struct {
	RowID          int64
	ThreadTitle    string
	ThreadURL      string
	InfoHash       string
	PostTime       string
	HasMatchedHash bool
	IsFavorited    bool
	IsInDownload   bool
	IsInPending    bool
}
