package config

import "github.com/zeromicro/go-zero/core/conf"

func MustLoad(path string) Config {
	var cfg Config
	conf.MustLoad(path, &cfg)
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}
