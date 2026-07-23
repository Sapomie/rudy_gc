package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	DataSource             string
	ListenAddr             string
	DailyFetchTime         string
	FailureRetryDelay      string
	Timezone               string
	LogursLevel            string
	BrowserMonitorInterval string
	Source                 SourceConfig
	TargetProbe            TargetProbeConfig
}

type SourceConfig struct {
	Timeout    int
	JavAddress string
	Browser    BrowserClientConfig
}

type BrowserClientConfig struct {
	BaseURL string
	Session string
}

type TargetProbeConfig struct {
	URL      string
	Interval string
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
	if _, err := positiveDuration(c.FailureRetryDelay, "远端服务 FailureRetryDelay"); err != nil {
		return err
	}
	if _, err := positiveDuration(c.BrowserMonitorInterval, "远端服务 BrowserMonitorInterval"); err != nil {
		return err
	}
	if strings.TrimSpace(c.Source.JavAddress) == "" {
		return fmt.Errorf("远端服务 Source.JavAddress 不能为空")
	}
	if c.Source.Timeout <= 0 {
		return fmt.Errorf("远端服务 Source.Timeout 必须为正整数")
	}
	if err := validateHTTPURL(c.Source.Browser.BaseURL, "远端服务 Source.Browser.BaseURL"); err != nil {
		return err
	}
	if strings.TrimSpace(c.Source.Browser.Session) == "" {
		return fmt.Errorf("远端服务 Source.Browser.Session 不能为空")
	}
	if err := validateHTTPURL(c.TargetProbe.URL, "远端服务 TargetProbe.URL"); err != nil {
		return err
	}
	if _, err := positiveDuration(c.TargetProbe.Interval, "远端服务 TargetProbe.Interval"); err != nil {
		return err
	}
	return nil
}

func (c Config) FailureRetryDuration() time.Duration {
	value, _ := time.ParseDuration(strings.TrimSpace(c.FailureRetryDelay))
	return value
}

func (c Config) TargetProbeDuration() time.Duration {
	value, _ := time.ParseDuration(strings.TrimSpace(c.TargetProbe.Interval))
	return value
}

func positiveDuration(raw, field string) (time.Duration, error) {
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s 必须为正数时间长度", field)
	}
	return value, nil
}

func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%s 必须为完整 HTTP 地址", field)
	}
	return nil
}
