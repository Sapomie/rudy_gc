package types

type FilmListFilter struct {
	MovieNameKeyword string

	SizeMin    int64
	HasSizeMin bool
	SizeMax    int64
	HasSizeMax bool

	HeightMin    int64
	HasHeightMin bool
	HeightMax    int64
	HasHeightMax bool

	DurationMin    int64
	HasDurationMin bool
	DurationMax    int64
	HasDurationMax bool

	BitRateMin    int64
	HasBitRateMin bool
	BitRateMax    int64
	HasBitRateMax bool

	FrameAverageMin    float64
	HasFrameAverageMin bool
	FrameAverageMax    float64
	HasFrameAverageMax bool

	SelfMake    int64
	HasSelfMake bool

	HasMask    int64
	HasHasMask bool

	ScTimesMin    int64
	HasScTimesMin bool
	ScTimesMax    int64
	HasScTimesMax bool

	LastScFrom    int64
	HasLastScFrom bool
	LastScTo      int64
	HasLastScTo   bool

	BirthTimeFrom    int64
	HasBirthTimeFrom bool
	BirthTimeTo      int64
	HasBirthTimeTo   bool

	ReleasingDateFrom    int64
	HasReleasingDateFrom bool
	ReleasingDateTo      int64
	HasReleasingDateTo   bool
}

type FilmListItem struct {
	Id              int64
	MovieName       string
	MovieHref       string
	CastName        string
	Casts           []*FilmListCastItem
	SizeGB          string
	Height          int64
	DurationMinutes int64
	BitRate         int64
	FrameAverage    string
	VideoURL        string
	PlayButtonClass string
	PlayButtonText  string
	ShowPlayButton  bool
	SelfMakeText    string
	HasMaskText     string
	ScTimes         int64
	LastScTime      int64
	BirthTime       int64
	ReleasingDate   int64
}

type FilmListCastItem struct {
	Id   int64
	Name string
	Href string
}
