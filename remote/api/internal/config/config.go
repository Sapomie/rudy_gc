package config

import (
	"fmt"
	"strings"
)

type Config struct {
	DataSource     string
	ListenAddr     string
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
	if strings.TrimSpace(c.DailyFetchTime) == "" {
		return fmt.Errorf("远端服务 DailyFetchTime 不能为空")
	}
	if strings.TrimSpace(c.Timezone) == "" {
		return fmt.Errorf("远端服务 Timezone 不能为空")
	}
	if strings.TrimSpace(c.Source.JavAddress) == "" {
		return fmt.Errorf("远端服务 Source.JavAddress 不能为空")
	}
	if strings.TrimSpace(c.Source.UserAgent) == "" {
		return fmt.Errorf("远端服务 Source.UserAgent 不能为空")
	}
	if strings.TrimSpace(c.Source.Cookie) == "" {
		return fmt.Errorf("远端服务 Source.Cookie 不能为空")
	}
	if c.Source.Timeout <= 0 {
		return fmt.Errorf("远端服务 Source.Timeout 必须为正整数")
	}
	return nil
}
