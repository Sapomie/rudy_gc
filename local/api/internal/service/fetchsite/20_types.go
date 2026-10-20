package fetchsite

import "rudy_gc/local/api/internal/model/modelx/moviex"

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
