package fetcher

import (
	"context"
	"io"
	"net/http"
	"net/url"
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

type Fetcher struct {
	cfg Config
}

func NewFetcher(cfg Config) *Fetcher {
	return &Fetcher{cfg: cfg}
}

func (f *Fetcher) Get(ctx context.Context, u string) (*Response, error) {
	return f.doRequest(ctx, u, false)
}

func (f *Fetcher) GetWithProxy(ctx context.Context, u string) (*Response, error) {
	return f.doRequest(ctx, u, true)
}

func (f *Fetcher) doRequest(ctx context.Context, u string, useProxy bool) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	if f.cfg.UserAgent != "" {
		req.Header.Add("User-Agent", f.cfg.UserAgent)
	}
	if f.cfg.Cookie != "" {
		req.Header.Add("Cookie", f.cfg.Cookie)
	}

	client := &http.Client{Timeout: f.cfg.Timeout}
	if useProxy && f.cfg.Proxy != "" {
		proxyURL, _ := url.Parse(f.cfg.Proxy)
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
