package types

type MovieDetail struct {
	MovieType *MovieType
	FilmInfo  *FilmInfo
	RankInfos []*RankInfo
	SC        []string
}

type RankInfo struct {
	Date string
	Rank int64
}
