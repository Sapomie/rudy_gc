package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strconv"
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
	Mail                   MailConfig
}

type SourceConfig struct {
	Timeout    int
	JavAddress string `json:",optional"` // 仅用于首次导入数据库中的站点配置。
	Browser    BrowserClientConfig
}

type BrowserClientConfig struct {
	BaseURL string
	Session string
}

type TargetProbeConfig struct {
	Interval string
}

type MailConfig struct {
	UserName      string
	AuthCode      string
	Host          string
	PortStr       string
	MailTo        string
	SenderName    string
	DetailBaseURL string
}

func (c MailConfig) Validate() error {
	if err := c.ValidateSMTP(); err != nil {
		return err
	}
	if err := validateHTTPURL(c.DetailBaseURL, "邮件配置 DetailBaseURL"); err != nil {
		return err
	}
	return nil
}

func (c MailConfig) ValidateSMTP() error {
	if strings.TrimSpace(c.UserName) == "" {
		return fmt.Errorf("邮件配置 UserName 不能为空")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(c.UserName)); err != nil {
		return fmt.Errorf("邮件配置 UserName 必须为有效邮箱地址")
	}
	if strings.TrimSpace(c.AuthCode) == "" {
		return fmt.Errorf("邮件配置 AuthCode 不能为空")
	}
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("邮件配置 Host 不能为空")
	}
	port, err := strconv.ParseUint(strings.TrimSpace(c.PortStr), 10, 16)
	if err != nil || port == 0 {
		return fmt.Errorf("邮件配置 PortStr 必须为有效端口")
	}
	if strings.TrimSpace(c.MailTo) == "" {
		return fmt.Errorf("邮件配置 MailTo 不能为空")
	}
	if _, err := mail.ParseAddressList(strings.TrimSpace(c.MailTo)); err != nil {
		return fmt.Errorf("邮件配置 MailTo 必须为有效邮箱地址")
	}
	if strings.TrimSpace(c.SenderName) == "" {
		return fmt.Errorf("邮件配置 SenderName 不能为空")
	}
	return nil
}

func (c MailConfig) Address() string {
	return net.JoinHostPort(strings.TrimSpace(c.Host), strings.TrimSpace(c.PortStr))
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
	if c.Source.Timeout <= 0 {
		return fmt.Errorf("远端服务 Source.Timeout 必须为正整数")
	}
	if err := validateHTTPURL(c.Source.Browser.BaseURL, "远端服务 Source.Browser.BaseURL"); err != nil {
		return err
	}
	if strings.TrimSpace(c.Source.Browser.Session) == "" {
		return fmt.Errorf("远端服务 Source.Browser.Session 不能为空")
	}
	if _, err := positiveDuration(c.TargetProbe.Interval, "远端服务 TargetProbe.Interval"); err != nil {
		return err
	}
	if err := c.Mail.Validate(); err != nil {
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
