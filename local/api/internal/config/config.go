// internal/config/config.go
package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Config struct {
	DataSource      string
	DailyBestRemote DailyBestRemoteClientConf
	Port            string
	Cache           cache.CacheConf // go-zero 生成代码那套用这个
	BizRedis        redis.RedisConf
	Fetcher         FetcherConf
	MovieTypeCache  MovieTypeCacheConf
	Film            Film
	Media           Media
	FC2Media        FC2MediaConfig
	SCEvent         SCEventConfig
	LogursLevel     string
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

type Media struct {
	RootDirs []string
}

type FC2MediaConfig struct {
	RootDirs      []string
	LocalImageDir string
}

type SCEventConfig struct {
	ActiveImageDir  string
	DeletedImageDir string
}

type FetcherConf struct {
	UserAgent     string
	Cookie        string
	Timeout       int    // 秒
	JavAddress    string `json:",optional"` // 仅用于首次初始化 t_fetch_site 的 javlib 配置
	BusAddress    string
	LocalImageDir string
}

type MovieTypeCacheConf struct {
	Prefix  string        `json:",default=rudy:mt"`
	Version string        `json:",default=v1"`
	TTL     time.Duration `json:",default=24h"`
}
