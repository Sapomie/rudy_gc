package types

type MovieDetail struct {
	MovieType *MovieType
	FilmInfo  *FilmInfo
	HasFilm   int64
	RankInfos []*RankInfo
	SC        []*MovieScEvent
}

type RankInfo struct {
	Date string
	Rank int64
}
