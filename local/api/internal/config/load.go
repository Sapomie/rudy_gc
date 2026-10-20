package config

import "github.com/zeromicro/go-zero/core/conf"

func MustLoad(path string) Config {
	var c Config
	conf.MustLoad(path, &c)
	return c
}
