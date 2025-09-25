// internal/config/config.go
package config

import "github.com/zeromicro/go-zero/core/stores/cache"

type Config struct {
	DataSource string
	Port       string
	Cache      cache.CacheConf
	Fetcher    FetcherConf
	Spider     SpiderConf
}

type FetcherConf struct {
	UserAgent string
	Cookie    string
	Proxy     string
	Timeout   int // 秒
}

type SpiderConf struct {
	JavAddress string
	BusAddress string
}
