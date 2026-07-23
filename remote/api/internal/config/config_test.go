package config

import (
	"strings"
	"testing"
)

func TestValidateAcceptsFlareSolverrClient(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
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
			URL: "https://www.c97k.com/cn/vl_bestrated.php?&mode=&page=1", Interval: "180m",
		},
	}
}
