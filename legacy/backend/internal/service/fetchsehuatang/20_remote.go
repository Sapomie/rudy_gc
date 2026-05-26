package fetchsehuatang

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"rudy_gc/internal/service/fetchsite"
	"rudy_gc/internal/taskctx"
)

func (s *Service) fetchWithRetry(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, cookie string) ([]byte, error) {
	var lastErr error

	for attempts := int64(1); attempts <= cfg.MaxRetryTimes; attempts++ {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
		if err := sleepBeforeRemoteRequest(ctx, cfg, state); err != nil {
			return nil, err
		}

		reportInfoLog(ctx, fmt.Sprintf("请求 %s 第 %d/%d 次", targetURL, attempts, cfg.MaxRetryTimes))
		body, status, err := doDirectRequest(ctx, cfg, targetURL, cookie)
		if err == nil {
			switch status {
			case http.StatusOK:
				if len(body) == 0 {
					err = fmt.Errorf("empty response body")
				} else {
					return body, nil
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

func doDirectRequest(ctx context.Context, cfg fetchsite.SiteConfig, targetURL string, cookie string) ([]byte, int, error) {
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

	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			Proxy: nil, // 明确禁用代理
		},
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
	return body, resp.StatusCode, nil
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
