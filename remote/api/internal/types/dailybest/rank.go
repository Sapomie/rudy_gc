package dailybest

import contracts "rudy_gc/shared/dailybest"

type SourcePage struct {
	PageNumber int64
	SourceURL  string
	Content    string
	Checksum   string
}

type MovieDraft struct {
	MovieJavID string
	MovieName  string
	MovieCode  string
	Title      string
	CoverURL   string
}

type ParsedRank struct {
	Rank  contracts.DailyBestRank
	Movie MovieDraft
}

type RankDraft struct {
	BusinessDate string
	DayNumber    int64
	Category     int64
	PageCount    int64
	Ranks        []contracts.DailyBestRank
	Movies       []MovieDraft
	Checksum     string
}

type RankFilter struct {
	PageQuery
	DayNumber  int64
	Category   int64
	RankPos    int64
	MovieJavID string
}

type RankDayFilter struct {
	PageQuery
	DayNumber      int64
	StartDayNumber int64
	EndDayNumber   int64
}

type RankDayStatFilter struct {
	PageQuery
	DayNumber int64
	RankDate  string
}
