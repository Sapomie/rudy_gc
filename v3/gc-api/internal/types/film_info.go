package types

type FilmInfo struct {
	Name              string
	BirthTime         string
	SourceTorrentHash string
	Size              float64
	FilePath          string
	FileName          string
	Directory         string
	Height            int64
	BitRate           float64
	DurationMinutes   float64 // 分钟
	Frame             float64
	SizeDurationRatio float64
	HasSub            string
	SelfMake          string
	Erased            string
}
