package types

// Cast 对应 internal/model/modelx/moviex.AmCast，用于对外透出（domain/service 层使用）
type Cast struct {
	Id                 int64
	PersonId           int64
	Name               string
	JavId              string
	Chinese            string
	BirthDay           int64
	Height             int64
	MovieNumber        int64
	OwnedMovieNumber   int64
	OwnedWMediaNumber  int64
	ScTimes            int64
	ComeTimes          int64
	LastScTime         int64
	LastScEventTime    int64
	Rank500MovieNumber int64
	Rank20MovieNumber  int64
	Rank1MovieNumber   int64
	HighestRank        int64
	RankTimes          int64
	CreatedOn          int64
	UpdatedOn          int64
}

type CastListFilter struct {
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
