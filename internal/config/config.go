// internal/config/config.go
package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Config struct {
	DataSource       string
	DataSourceRemote string
	Port             string
	Cache            cache.CacheConf // go-zero 生成代码那套用这个
	BizRedis         redis.RedisConf
	Fetcher          FetcherConf
	MovieTypeCache   MovieTypeCacheConf
	Film             Film
	LogursLevel      string
}

type Film struct {
	Pairs               []FilmDirPair
	CopyDestinationPath string
	ScRootDir           string
	RenamePaths         []string
}

type FilmDirPair struct {
	RootDir             string
	MoveFilmDestination string
}

type FetcherConf struct {
	UserAgent     string
	Cookie        string
	Proxy         string
	Timeout       int // 秒
	JavAddress    string
	BusAddress    string
	LocalImageDir string
}

type MovieTypeCacheConf struct {
	Prefix  string        `json:",default=rudy:mt"`
	Version string        `json:",default=v1"`
	TTL     time.Duration `json:",default=24h"`
}
