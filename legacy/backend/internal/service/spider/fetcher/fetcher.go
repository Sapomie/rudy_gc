package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"time"
)

type Config struct {
	UserAgent string
	Cookie    string
	Proxy     string
	Timeout   time.Duration
}

type Response struct {
	Body   []byte
	Status int
	URL    string
}

type RequestOptions struct {
	UseProxy bool
	Referer  string
	Headers  map[string]string
}

type Fetcher struct {
	cfg      Config
	siteCfgs map[string]Config
}

func NewFetcher(cfg Config) *Fetcher {
	return &Fetcher{
		cfg:      cfg,
		siteCfgs: make(map[string]Config),
	}
}

func (f *Fetcher) SetSiteConfig(siteCode string, cfg Config) {
	if siteCode == "" {
		return
	}
	f.siteCfgs[siteCode] = cfg
}

func (f *Fetcher) Get(ctx context.Context, u string) (*Response, error) {
	return f.doRequestWithConfig(ctx, f.cfg, u, RequestOptions{})
}

func (f *Fetcher) GetWithProxy(ctx context.Context, u string) (*Response, error) {
	return f.doRequestWithConfig(ctx, f.cfg, u, RequestOptions{UseProxy: true})
}

func (f *Fetcher) GetBySite(ctx context.Context, siteCode, u string) (*Response, error) {
	return f.GetBySiteWithOptions(ctx, siteCode, u, RequestOptions{})
}

func (f *Fetcher) GetBySiteWithOptions(ctx context.Context, siteCode, u string, opts RequestOptions) (*Response, error) {
	cfg, ok := f.siteCfgs[siteCode]
	if !ok {
		cfg = f.cfg
	}
	if shouldUseCurlSiteFetch(siteCode) {
		return doCurlRequestWithConfig(ctx, cfg, u, opts)
	}
	if cfg.Proxy != "" {
		opts.UseProxy = true
	}
	return f.doRequestWithConfig(ctx, cfg, u, opts)
}

func shouldUseCurlSiteFetch(siteCode string) bool {
	switch siteCode {
	case "javbus", "sukebei":
		return true
	default:
		return false
	}
}

func doCurlRequestWithConfig(ctx context.Context, cfg Config, u string, opts RequestOptions) (*Response, error) {
	args := []string{
		"-sS",
		"--compressed",
	}
	if cfg.Timeout > 0 {
		args = append(args, "--max-time", strconv.FormatFloat(cfg.Timeout.Seconds(), 'f', -1, 64))
	}
	if cfg.Proxy != "" {
		args = append(args, "-x", cfg.Proxy)
	}
	if cfg.UserAgent != "" {
		args = append(args, "-A", cfg.UserAgent)
	}
	if cfg.Cookie != "" {
		args = append(args, "-H", "Cookie: "+cfg.Cookie)
	}
	if opts.Referer != "" {
		args = append(args, "-e", opts.Referer)
	}
	for key, value := range opts.Headers {
		if key == "" || value == "" {
			continue
		}
		args = append(args, "-H", key+": "+value)
	}
	args = append(args, "-w", "\n__CODEX_STATUS__:%{http_code}", u)

	cmd := exec.CommandContext(ctx, "curl", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	statusLine := []byte("\n__CODEX_STATUS__:")
	idx := bytes.LastIndex(output, statusLine)
	if idx < 0 {
		return nil, fmt.Errorf("curl response missing status marker")
	}

	body := append([]byte(nil), output[:idx]...)
	statusRaw := string(output[idx+len(statusLine):])
	status, err := strconv.Atoi(statusRaw)
	if err != nil {
		return nil, fmt.Errorf("parse curl status failed: %w", err)
	}

	return &Response{
		Body:   body,
		Status: status,
		URL:    u,
	}, nil
}

func (f *Fetcher) doRequestWithConfig(ctx context.Context, cfg Config, u string, opts RequestOptions) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	if cfg.UserAgent != "" {
		req.Header.Add("User-Agent", cfg.UserAgent)
	}
	if cfg.Cookie != "" {
		req.Header.Add("Cookie", cfg.Cookie)
	}
	if opts.Referer != "" {
		req.Header.Add("Referer", opts.Referer)
	}
	for key, value := range opts.Headers {
		if key == "" || value == "" {
			continue
		}
		req.Header.Add(key, value)
	}

	client := &http.Client{Timeout: cfg.Timeout}
	if opts.UseProxy && cfg.Proxy != "" {
		proxyURL, _ := url.Parse(cfg.Proxy)
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return &Response{Body: body, Status: resp.StatusCode, URL: u}, err
}
