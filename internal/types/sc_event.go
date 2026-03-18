package types

type ScEventListItem struct {
	Event *GSc
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
	Event       *GSc
	Items       []*ScEventMovie
	FailedItems []*ScEventMovie
	ComeCount   int64
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
