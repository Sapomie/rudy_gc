package types

type ListMovieLiteRequest struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	OrderBy  string `form:"od"`
}

type ListMovieResponse struct {
	List  []*MovieType
	Total int64
}

type ListMovieFullRequest struct {
	CastNames  string `form:"cn"`
	GenreNames string `form:"gn"`

	DirectorName string `form:"dn"`
	PrefixName   string `form:"pn"`
	MakerName    string `form:"mn"`
	LabelName    string `form:"ln"`

	ReleasingDateStart string  `form:"rs"`  //ok
	ReleasingDateEnd   string  `form:"re"`  //ok
	CastAgeMin         float64 `form:"cay"` //ok
	CastAgeMax         float64 `form:"cao"` //ok

	StartRankingDate string `form:"srd"`
	NeedDownload     int64  `form:"nd"`
	Word             string `form:"wd"`

	Owned              int64  `form:"owned"`
	ComeTimesMin       int64  `form:"comin"`
	LastScTimeMin      string `form:"lsct"`
	ScTimesMin         int64  `form:"scmin"`
	ScTimesMax         *int64 `form:"scmax"`
	FilmBirthTimeStart string `form:"bs"`
	FilmBirthTimeEnd   string `form:"be"`
	Dir1               string `form:"d1"`
	Dir2               string `form:"d2"`
	Dir3               string `form:"d3"`
	Dir4               string `form:"d4"`

	OrderBy  string `form:"od"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
}
