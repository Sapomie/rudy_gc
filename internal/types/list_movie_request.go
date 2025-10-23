package types

type ListMovieResponse struct {
	List   []*MovieType
	Total  int64
	JavIds []string
}

type ListMovieFullRequest struct {
	CastNames  string `form:"cn"`
	GenreNames string `form:"gn"`

	DirectorName string `form:"dn"`
	PrefixName   string `form:"pn"`
	MakerName    string `form:"mn"`
	LabelName    string `form:"ln"`

	ReleasingDateStart string  `form:"rs"`
	ReleasingDateEnd   string  `form:"re"`
	CastAgeMin         float64 `form:"cay"`
	CastAgeMax         float64 `form:"cao"`

	StartRankingDateStart string `form:"srds"`
	StartRankingDateEnd   string `form:"srde"`

	DaysInRankMin int64  `form:"drkmin"`
	NeedDownload  int64  `form:"nd"`
	Word          string `form:"wd"`
	Owned         int64  `form:"owned"`

	ViewWatchedMin int64   `form:"vwmin"`
	ViewWatchedMax int64   `form:"vwmax"`
	ScoreMin       float64 `form:"smin"`
	ScoreMax       float64 `form:"smax"`

	LastScTimeMin      string `form:"lsctmin"`
	LastScTimeMax      string `form:"lsctmax"`
	ScTimesMin         int64  `form:"scmin"`
	ScTimesMax         *int64 `form:"scmax"`
	ComeTimesMin       int64  `form:"comin"`
	ComeTimesMax       *int64 `form:"comax"`
	FilmBirthTimeStart string `form:"bs"`
	FilmBirthTimeEnd   string `form:"be"`

	Dir1 string `form:"d1"`
	Dir2 string `form:"d2"`
	Dir3 string `form:"d3"`
	Dir4 string `form:"d4"`

	OrderBy  string `form:"od"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
}
