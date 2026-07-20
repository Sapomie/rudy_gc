package sourceclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	UserAgent string
	Cookie    string
	Timeout   time.Duration
}

type Response struct {
	Body       []byte
	StatusCode int
}

type Client struct {
	http      *http.Client
	userAgent string
	cookie    string
}

func New(cfg Config) *Client {
	return &Client{
		http:      &http.Client{Timeout: cfg.Timeout},
		userAgent: strings.TrimSpace(cfg.UserAgent),
		cookie:    strings.TrimSpace(cfg.Cookie),
	}
}

func (c *Client) Get(ctx context.Context, target string) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return Response{}, err
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{Body: body, StatusCode: resp.StatusCode}, nil
}
