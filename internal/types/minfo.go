package types

type Minfo struct {
	Id                 int64
	JavId              string
	Name               string
	Chinese            string
	FirstRankDayNumber int64
	HighestRank        int64
	DaysInRank         int64
	NeedDownload       int64
	CreatedOn          int64
	UpdatedOn          int64
	ReleasingDate      int64
}
