package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Config struct {
	DataSource     string
	Port           string
	LogxMode       string
	Cache          cache.CacheConf
	BizRedis       redis.RedisConf
	MovieTypeCache MovieTypeCacheConf
	Fetcher        FetcherConf
	Film           Film
	Media          Media
}

type MovieTypeCacheConf struct {
	Prefix  string
	Version string
	TTL     time.Duration
}

type FetcherConf struct {
	UserAgent     string
	Cookie        string
	Proxy         string
	Timeout       int
	JavAddress    string
	BusAddress    string
	LocalImageDir string
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
