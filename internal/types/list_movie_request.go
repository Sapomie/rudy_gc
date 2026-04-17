package types

type ListMovieResponse struct {
	List   []*MovieType
	Total  int64
	JavIds []string
}

type ListMovieFullRequest struct {
	CastNames    string `form:"cn"`
	PersonIds    string `form:"pid"`
	GenreNames   string `form:"gn"`
	DirectorName string `form:"dn"`
	PrefixName   string `form:"pn"`
	MakerName    string `form:"mn"`
	LabelName    string `form:"ln"`
	LabelJavID   string `form:"lj"`
	AlbumName    string `form:"an"`

	ReleasingDateStart  string `form:"rs"`
	ReleasingDateEnd    string `form:"re"`
	MediaBirthTimeStart string `form:"mbs"`
	MediaBirthTimeEnd   string `form:"mbe"`

	CastAgeMin float64 `form:"cay"`
	CastAgeMax float64 `form:"cao"`

	StartRankingDateStart string `form:"srds"`
	StartRankingDateEnd   string `form:"srde"`

	DaysInRankMin int64  `form:"drkmin"`
	NeedDownload  int64  `form:"nd"`
	Word          string `form:"wd"`
	MediaOwned    int64  `form:"mowned"`

	ViewWatchedMin int64   `form:"vwmin"`
	ViewWatchedMax int64   `form:"vwmax"`
	ScoreMin       float64 `form:"smin"`
	ScoreMax       float64 `form:"smax"`

	LastScTimeMin string `form:"lsctmin"`
	LastScTimeMax string `form:"lsctmax"`
	ScTimesMin    int64  `form:"scmin"`
	ScTimesMax    *int64 `form:"scmax"`
	ComeTimesMin  int64  `form:"comin"`
	ComeTimesMax  *int64 `form:"comax"`

	MediaDir1    string `form:"md1"`
	MediaDir2    string `form:"md2"`
	MediaDir3    string `form:"md3"`
	MediaDir4    string `form:"md4"`
	MediaDirFull string `form:"-"`
	MediaDirSub  bool   `form:"-"`

	OrderBy  string `form:"od"`
	Order    string `form:"order"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
}
