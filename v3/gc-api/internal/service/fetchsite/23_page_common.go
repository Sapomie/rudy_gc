package fetchsite

import "time"

func pageFetchStatusText(status int64) string {
	switch status {
	case FetchStatusPending:
		return "待抓取"
	case FetchStatusRunning:
		return "抓取中"
	case FetchStatusSuccess:
		return "成功"
	case FetchStatusFailed:
		return "失败"
	default:
		return "未入队"
	}
}

func pageFormatUnix(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
}

func pageFormatDate(sec int64) string {
	if sec <= 0 {
		return "-"
	}
	return time.Unix(sec, 0).Format("2006-01-02")
}
