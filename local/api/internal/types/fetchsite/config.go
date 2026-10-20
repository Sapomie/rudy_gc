package fetchsite

import "time"

const (
	CodeJavlib     = "javlib"
	CodeJavbus     = "javbus"
	CodeSukebei    = "sukebei"
	CodeSehuatang  = "sehuatang"
	CodeBaiduFanyi = "baidu_fanyi"
)

type Config struct {
	SiteCode string
	SiteName string
	BaseURL  string

	UserAgent string
	Cookie    string
	Proxy     string

	Timeout       time.Duration
	RequestSleep  time.Duration
	RetrySleep    time.Duration
	MaxRetryTimes int64
	Status        int64
}
