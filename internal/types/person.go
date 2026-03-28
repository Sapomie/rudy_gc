package types

type Person struct {
	Id               int64
	Name             string
	Alias            string
	Chinese          string
	BirthDay         int64
	Height           int64
	Cup              string
	Bwh              string
	Avatar           string
	MovieNumber      int64
	OwnedMovieNumber int64
	ScTimes          int64
	ComeTimes        int64
	LastScTime       int64
	HighestRank      int64
	RankTimes        int64
	CreatedOn        int64
	UpdatedOn        int64
}

type PersonListFilter struct {
	Keyword         string
	OwnedMin        int64
	HasOwnedMin     bool
	OwnedMax        int64
	HasOwnedMax     bool
	ScTimesMin      int64
	HasScTimesMin   bool
	ScTimesMax      int64
	HasScTimesMax   bool
	ComeTimesMin    int64
	HasComeTimesMin bool
	ComeTimesMax    int64
	HasComeTimesMax bool
	LastScFrom      int64
	HasLastScFrom   bool
	LastScTo        int64
	HasLastScTo     bool
}
