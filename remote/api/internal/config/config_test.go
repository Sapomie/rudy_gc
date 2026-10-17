package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestValidateAcceptsFlareSolverrClient(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestServerConfigIsComplete(t *testing.T) {
	var cfg Config
	if err := conf.Load("../../config/server.yaml", &cfg); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigLoadsWithoutLegacyDomainAndIgnoresOldProbeURL(t *testing.T) {
	content, err := os.ReadFile("../../config/server.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "JavAddress:") {
			lines = append(lines, line)
		}
	}
	yaml := strings.Replace(strings.Join(lines, "\n"), "TargetProbe:\n", "TargetProbe:\n  URL: https://legacy.example.com/unused\n", 1)
	path := filepath.Join(t.TempDir(), "server.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := conf.Load(path, &cfg); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Source.JavAddress != "" || cfg.TargetProbe.Interval != "180m" {
		t.Fatalf("可选迁移配置加载异常")
	}
}

func TestValidateRejectsMissingFlareSolverrSession(t *testing.T) {
	cfg := validConfig()
	cfg.Source.Browser.Session = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "Source.Browser.Session") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateRejectsMissingBrowserMonitorInterval(t *testing.T) {
	cfg := validConfig()
	cfg.BrowserMonitorInterval = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "BrowserMonitorInterval") {
		t.Fatalf("err=%v", err)
	}
}

func TestValidateRejectsMissingProbeInterval(t *testing.T) {
	cfg := validConfig()
	cfg.TargetProbe.Interval = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "TargetProbe.Interval") {
		t.Fatalf("err=%v", err)
	}
}

func TestMailConfigValidate(t *testing.T) {
	cfg := validMailConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Address() != "smtp.qq.com:587" {
		t.Fatalf("address=%q", cfg.Address())
	}
}

func TestMailConfigRejectsMissingAuthCode(t *testing.T) {
	cfg := validMailConfig()
	cfg.AuthCode = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "AuthCode") {
		t.Fatalf("err=%v", err)
	}
}

func TestMailConfigRejectsInvalidRecipient(t *testing.T) {
	cfg := validMailConfig()
	cfg.MailTo = "invalid"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "MailTo") {
		t.Fatalf("err=%v", err)
	}
}

func TestMailConfigRejectsMissingDetailBaseURL(t *testing.T) {
	cfg := validMailConfig()
	cfg.DetailBaseURL = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "DetailBaseURL") {
		t.Fatalf("err=%v", err)
	}
}

func validMailConfig() MailConfig {
	return MailConfig{
		UserName: "sender@example.com", AuthCode: "secret", Host: "smtp.qq.com",
		PortStr: "587", MailTo: "receiver@example.com", SenderName: "Zero-GC",
		DetailBaseURL: "https://remote.example.com",
	}
}

func validConfig() Config {
	return Config{
		DataSource:             "user:password@tcp(127.0.0.1:3306)/database",
		ListenAddr:             "0.0.0.0:5031",
		DailyFetchTime:         "08:30",
		FailureRetryDelay:      "1h",
		Timezone:               "Asia/Shanghai",
		BrowserMonitorInterval: "10s",
		Source: SourceConfig{
			Timeout:    60,
			JavAddress: "www.c97k.com",
			Browser: BrowserClientConfig{
				BaseURL: "http://172.30.0.1:8191",
				Session: "rudy_gc_dailybest",
			},
		},
		TargetProbe: TargetProbeConfig{
			Interval: "180m",
		},
		Mail: validMailConfig(),
	}
}
