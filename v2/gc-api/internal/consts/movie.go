package consts

const (
	OrderByDetailUpdateTime = "du"
	OrderByCastAgeDesc      = "cad"
	OrderByCastAgeAsc       = "caa"
	OrderByViewerWatched    = "vw"
	OrderByReleasingDate    = "rd"
	OrderByRankDate         = "rk"
	OrderByHighestRank      = "hrk"
	OrderByDaysInRank       = "drk"
	OrderByBirthTime        = "bt"
	OrderByMediaBirthTime   = "mbt"
	OrderByScTimes          = "sc"
	OrderByComeTimes        = "co"
	OrderByLastScTime       = "lsct"
)

const (
	OwnedUnknown          int64 = 0
	OwnedAll              int64 = 2
	OwnedAllNotRemoved    int64 = 3
	OwnedHasSubNotRemoved int64 = 4
	OwnedNoSubNotRemoved  int64 = 5
	OwnedRemoved          int64 = 6
	OwnedNotOwned         int64 = 7
)

const (
	MovieNeedDownloadNone int64 = 1
	MovieNeedDownloadOK   int64 = 2
)

const MovieNeedDownloadAlbumName = "电影稍后下载"

const (
	WMediaSourceNative int64 = 2
)

const (
	FilmIsNotRemoved int64 = 1
	FilmIsRemoved    int64 = 2
)

const (
	FilmNoSub  int64 = 1
	FilmHasSub int64 = 2
)
