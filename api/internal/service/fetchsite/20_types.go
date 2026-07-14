package fetchsite

import "rudy_gc/internal/repository/moviex"

type JavbusFetchTask struct {
	MovieJavID string
	MovieName  string
	Row        *moviex.TJavbusMagnetFetch
}

type SukebeiFetchTask struct {
	MovieJavID string
	MovieName  string
	Row        *moviex.TSukebeiTorrentFetch
}

type RunFetchTasksResult struct {
	Queued  int
	Handled int
	Success int
	Failed  int
}
