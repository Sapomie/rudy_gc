package types

// Cast 对应 data/modelx/moviex.AmCast，用于对外透出（domain/service 层使用）
type Cast struct {
	Id                 int64
	Name               string
	JavId              string
	MovieNumber        int64
	OwnedMovieNumber   int64
	ScTimes            int64
	ComeTimes          int64
	LastScTime         int64
	Rank500MovieNumber int64
	Rank20MovieNumber  int64
	Rank1MovieNumber   int64
	HighestRank        int64
	RankTimes          int64
	CreatedOn          int64
	UpdatedOn          int64
}
