package config

import (
	"fmt"
	"net/url"
	"strings"
)

type DailyBestRemoteClientConf struct {
	BaseURL string
	Proxy   string
}

func (c DailyBestRemoteClientConf) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("DailyBestRemote.BaseURL 不能为空")
	}
	if _, err := c.ProxyURL(); err != nil {
		return err
	}
	return nil
}

func (c DailyBestRemoteClientConf) ProxyURL() (*url.URL, error) {
	raw := strings.TrimSpace(c.Proxy)
	if raw == "" {
		return nil, nil
	}
	proxyURL, err := url.Parse(raw)
	if err != nil || proxyURL.Host == "" {
		return nil, fmt.Errorf("DailyBestRemote.Proxy 必须为完整 HTTP 地址")
	}
	switch strings.ToLower(proxyURL.Scheme) {
	case "http", "https":
		return proxyURL, nil
	default:
		return nil, fmt.Errorf("DailyBestRemote.Proxy 必须为完整 HTTP 地址")
	}
}
