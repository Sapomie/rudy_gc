package types

type ListMovieLiteRequest struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	OrderBy  string `form:"od"`
}

type ListMovieLiteResponse struct {
	List  []*MovieType
	Total int64
}

// ListMovieComplexRequest 用于 HTML/API 查询电影列表时的复杂筛选条件
type ListMovieComplexRequest struct {
	CastNames          string  `form:"cn"`
	GenreNames         string  `form:"gn"`
	DirectorName       string  `form:"dn"`
	PrefixName         string  `form:"pn"`
	MakerName          string  `form:"mn"`
	LabelName          string  `form:"ln"`
	Word               string  `form:"wd"`
	OrderBy            string  `form:"od"`
	Page               int64   `form:"page"`
	PageSize           int64   `form:"ps"`
	ReleasingDateStart string  `form:"rs"`
	ReleasingDateEnd   string  `form:"re"`
	CastAgeMin         float64 `form:"cay"`
	CastAgeMax         float64 `form:"cao"`
	StartRankingDate   string  `form:"srd"`
	Owned              int64   `form:"owned"`
	HasSub             int64   `form:"hs"`
	ComeTimesMin       int64   `form:"comin"`
	LastScTimeMin      string  `form:"lsct"`
	ScTimesMin         int64   `form:"scmin"`
	ScTimesMax         int64   `form:"scmax,default=1000"`
	FilmBirthTimeStart string  `form:"bs"`
	FilmBirthTimeEnd   string  `form:"be"`
	Dir1               string  `form:"d1"`
	Dir2               string  `form:"d2"`
	Dir3               string  `form:"d3"`
	Dir4               string  `form:"d4"`
}
