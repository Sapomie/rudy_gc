package config

import "github.com/zeromicro/go-zero/core/stores/cache"

type Config struct {
	DataSource string
	Port       string
	Cache      cache.CacheConf
}
