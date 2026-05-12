package types

type ScEventListItem struct {
	Event           *GSc
	TotalMovieCount int64
	ScMovieCount    int64
}

type ScEventListPage struct {
	Items []*ScEventListItem
	Total int64
}

type ScEventCardItem struct {
	Event              *GSc
	DetailHref         string
	ComeMovieHref      string
	ComeMovieName      string
	ComeMovieJacketImg string
	HasComeMovieCover  bool
}

type ScEventCardPage struct {
	Items []*ScEventCardItem
	Total int64
}

type ScEventDetail struct {
	Event                     *GSc
	Items                     []*ScEventMovie
	FailedItems               []*ScEventMovie
	ComeCount                 int64
	EditImageName             string
	EditMovieCount            int64
	EditScMovieCount          int64
	EditComeMovieJavId        string
	EditComeMovieOptions      []*ScEventEditMovieOption
	EditCurrentMovieCastNames []string
}

type ScEventEditForm struct {
	Name            string
	ComeMovieJavId  string
	Kind            string
	DurationMinutes int64
	Fg              string
	Vessel          string
	MovieCast       string
	Remarks         string
}

type ScEventEditMovieOption struct {
	MovieJavId  string
	MovieName   string
	CastOptions []string
	IsCome      bool
}

type ScEventMovie struct {
	Entry      *GList
	MovieType  *MovieType
	MovieName  string
	IsCome     bool
	LoadError  string
	CastNames  []string
	GenreNames []string
}
