package dailybestsource

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

func NewClient(cfg Config) *Client {
	return &Client{
		http:      &http.Client{Timeout: cfg.Timeout},
		userAgent: strings.TrimSpace(cfg.UserAgent),
		cookie:    strings.TrimSpace(cfg.Cookie),
	}
}

func (c *Client) Get(ctx context.Context, target string) (Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return Response{}, err
	}
	if c.userAgent != "" {
		request.Header.Set("User-Agent", c.userAgent)
	}
	if c.cookie != "" {
		request.Header.Set("Cookie", c.cookie)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{Body: body, StatusCode: response.StatusCode}, nil
}
