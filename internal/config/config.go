// internal/config/config.go
package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type MovieTypeCacheConf struct {
	Prefix  string        `json:",default=rudy:mt"`
	Version string        `json:",default=v1"`
	TTL     time.Duration `json:",default=24h"`
}

type Config struct {
	DataSource string
	Port       string
	Cache      cache.CacheConf // go-zero 生成代码那套用这个
	BizRedis   redis.RedisConf
	Fetcher    FetcherConf
	Spider     SpiderConf

	MovieTypeCache MovieTypeCacheConf `json:",optional"`
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
