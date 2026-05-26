package types

type FilmProbeMetaResult struct {
	Id              int64  `json:"id"`
	Height          int64  `json:"height"`
	DurationMinutes int64  `json:"duration_minutes"`
	BitRate         int64  `json:"bit_rate"`
	FrameAverage    string `json:"frame_average"`
}
