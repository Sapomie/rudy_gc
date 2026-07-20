package config

import (
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
)

type Config struct {
	DataSource     string
	ListenAddr     string
	ReadToken      string
	AdminToken     string
	DailyFetchTime string
	Timezone       string
	FetchOnStart   bool
	LogursLevel    string
	Source         SourceConfig
}

type SourceConfig struct {
	UserAgent  string
	Cookie     string
	Timeout    int
	JavAddress string
}

func MustLoad(path string) Config {
	var cfg Config
	conf.MustLoad(path, &cfg)
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.DataSource) == "" {
		return fmt.Errorf("远端服务 DataSource 不能为空")
	}
	if strings.TrimSpace(c.ListenAddr) == "" {
		return fmt.Errorf("远端服务 ListenAddr 不能为空")
	}
	if !strings.HasPrefix(strings.TrimSpace(c.ListenAddr), "0.0.0.0:") {
		return fmt.Errorf("远端服务 ListenAddr 必须监听 0.0.0.0")
	}
	if strings.TrimSpace(c.ReadToken) == "" || strings.TrimSpace(c.AdminToken) == "" {
		return fmt.Errorf("远端服务 ReadToken 和 AdminToken 不能为空")
	}
	if strings.TrimSpace(c.DailyFetchTime) == "" {
		return fmt.Errorf("远端服务 DailyFetchTime 不能为空")
	}
	if strings.TrimSpace(c.Timezone) == "" {
		return fmt.Errorf("远端服务 Timezone 不能为空")
	}
	if strings.TrimSpace(c.Source.JavAddress) == "" {
		return fmt.Errorf("远端服务 Source.JavAddress 不能为空")
	}
	return nil
}
