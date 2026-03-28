package fetchsite

import "rudy_gc/internal/model/modelx/moviex"

type JavbusFetchTask struct {
	MovieJavID string
	MovieCode  string
	Row        *moviex.TJavbusMagnetFetch
}

type SukebeiFetchTask struct {
	MovieJavID string
	MovieCode  string
	Row        *moviex.TSukebeiTorrentFetch
}

type RunFetchTasksResult struct {
	Queued  int
	Handled int
	Success int
	Failed  int
}
