package fetchsite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/service/spider/fetcher"
)

func (s *Service) GetConfig(siteCode string) (SiteConfig, error) {
	cfg, ok := s.deps.GetFetchSiteConfig(siteCode)
	if !ok {
		return SiteConfig{}, fmt.Errorf("fetch site config not found: %s", siteCode)
	}
	return cfg, nil
}

func (s *Service) SleepRequest(ctx context.Context, siteCode string) error {
	cfg, err := s.GetConfig(siteCode)
	if err != nil {
		return err
	}
	return sleepWithContext(ctx, cfg.RequestSleep)
}

func (s *Service) SleepRetry(ctx context.Context, siteCode string) error {
	cfg, err := s.GetConfig(siteCode)
	if err != nil {
		return err
	}
	return sleepWithContext(ctx, cfg.RetrySleep)
}

func (s *Service) Get(ctx context.Context, siteCode, rawURL string, opts RequestOptions) (*fetcher.Response, error) {
	return s.deps.Fetcher.GetBySiteWithOptions(ctx, siteCode, rawURL, opts)
}

func (s *Service) FetchWithRetry(
	ctx context.Context,
	siteCode string,
	rawURL string,
	opts RequestOptions,
	validate func(*fetcher.Response) error,
) (*fetcher.Response, int64, error) {
	cfg, err := s.GetConfig(siteCode)
	if err != nil {
		return nil, 0, err
	}
	if cfg.MaxRetryTimes < 1 {
		return nil, 0, fmt.Errorf("invalid max_retry_times for site %s: %d", siteCode, cfg.MaxRetryTimes)
	}

	var lastErr error
	for attempts := int64(1); attempts <= cfg.MaxRetryTimes; attempts++ {
		resp, reqErr := s.Get(ctx, siteCode, rawURL, opts)
		if reqErr == nil && validate != nil {
			reqErr = validate(resp)
		}
		if reqErr == nil {
			return resp, attempts, nil
		}
		lastErr = reqErr
		if attempts == cfg.MaxRetryTimes {
			break
		}
		if sleepErr := s.SleepRetry(ctx, siteCode); sleepErr != nil {
			return nil, attempts, sleepErr
		}
	}
	return nil, cfg.MaxRetryTimes, fmt.Errorf("fetch %s failed after %d attempts: %w", rawURL, cfg.MaxRetryTimes, lastErr)
}

func (s *Service) BuildURL(siteCode string, parts ...string) (string, error) {
	cfg, err := s.GetConfig(siteCode)
	if err != nil {
		return "", err
	}
	raw := strings.TrimRight(cfg.BaseURL, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		raw += "/" + strings.TrimLeft(part, "/")
	}
	return raw, nil
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
