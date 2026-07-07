package fetchsehuatang

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

func (s *Service) fetchWithRetry(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, cookie string) ([]byte, error) {
	resp, err := s.fetchWithRetryResult(ctx, cfg, state, targetURL, cookie)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

type fetchHTTPResult struct {
	Body     []byte
	FinalURL string
}

func (s *Service) fetchWithRetryResult(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, cookie string) (*fetchHTTPResult, error) {
	return s.fetchWithRetryResultWithOptions(ctx, cfg, state, targetURL, cookie, false)
}

func (s *Service) fetchWithRetryResultWithOptions(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, cookie string, useProxy bool) (*fetchHTTPResult, error) {
	var lastErr error

	for attempts := int64(1); attempts <= cfg.MaxRetryTimes; attempts++ {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if err := sleepBeforeRemoteRequest(ctx, cfg, state); err != nil {
			return nil, err
		}

		reportInfoLog(ctx, fmt.Sprintf("请求 %s 第 %d/%d 次", targetURL, attempts, cfg.MaxRetryTimes))
		resp, status, err := doDirectRequest(ctx, cfg, targetURL, cookie, useProxy)
		if err == nil {
			switch status {
			case http.StatusOK:
				if resp == nil || len(resp.Body) == 0 {
					err = fmt.Errorf("empty response body")
				} else {
					return resp, nil
				}
			default:
				err = fmt.Errorf("unexpected status: %d", status)
			}
		}
		lastErr = err
		reportWarnLog(ctx, fmt.Sprintf("请求失败 %s 第 %d/%d 次: %v", targetURL, attempts, cfg.MaxRetryTimes, err))

		if attempts >= cfg.MaxRetryTimes {
			break
		}
		if err := sleepWithContext(ctx, cfg.RetrySleep); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("fetch %s failed after %d attempts: %w", targetURL, cfg.MaxRetryTimes, lastErr)
}

func sleepBeforeRemoteRequest(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState) error {
	if state == nil {
		return nil
	}
	if state.hasPreviousRemoteRequest {
		if err := sleepWithContext(ctx, cfg.RequestSleep); err != nil {
			return err
		}
	}
	state.hasPreviousRemoteRequest = true
	return nil
}

func doDirectRequest(ctx context.Context, cfg fetchsite.SiteConfig, targetURL string, cookie string, useProxy bool) (*fetchHTTPResult, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, 0, err
	}

	if cfg.UserAgent != "" {
		req.Header.Set("User-Agent", cfg.UserAgent)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	transport := &http.Transport{
		Proxy: nil,
	}
	if useProxy && strings.TrimSpace(cfg.Proxy) != "" {
		proxyURL, parseErr := url.Parse(strings.TrimSpace(cfg.Proxy))
		if parseErr != nil {
			return nil, 0, parseErr
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	finalURL := ""
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return &fetchHTTPResult{
		Body:     body,
		FinalURL: finalURL,
	}, resp.StatusCode, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
