package types

type MovieDetail struct {
	MovieType       *MovieType
	FilmInfo        *FilmInfo
	HasFilm         int64
	RankInfos       []*RankInfo
	SC              []*MovieScEvent
	JavbusFetch     *MovieFetchSiteStatus
	JavbusMagnets   []*MovieJavbusMagnet
	SukebeiFetch    *MovieFetchSiteStatus
	SukebeiTorrents []*MovieSukebeiTorrent
}

type RankInfo struct {
	Date string
	Rank int64
}

type MovieFetchSiteStatus struct {
	MovieJavID        string
	MovieCode         string
	ReleaseDate       string
	FetchStatus       int64
	FetchStatusText   string
	TryCount          int64
	LastFetchTime     string
	LastError         string
	TorrentHashCount  int64
	LatestPublishTime string
	SourceURL         string
}

type MovieJavbusMagnet struct {
	MagnetName     string
	MagnetURL      string
	InfoHash       string
	SizeText       string
	ShareDate      string
	HasHD          bool
	HasSubtitle    bool
	PageURL        string
	HasMatchedHash bool
}

type MovieSukebeiTorrent struct {
	TorrentTitle   string
	ViewURL        string
	TorrentURL     string
	MagnetURL      string
	InfoHash       string
	SizeText       string
	PublishTime    string
	Seeders        int64
	Leechers       int64
	Completed      int64
	HasMatchedHash bool
}
