package fetchsehuatang

import (
	"context"
	"strings"

	"rudy_gc/local/api/internal/service/fetchsite"
)

func combineCookie(baseCookie, safeID string) string {
	baseCookie = strings.TrimSpace(baseCookie)
	safeID = strings.TrimSpace(safeID)
	if safeID == "" {
		return baseCookie
	}
	safeCookie := "_safe=" + safeID
	if baseCookie == "" {
		return safeCookie
	}
	return baseCookie + "; " + safeCookie
}

func (s *Service) fetchWithSafe(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string) ([]byte, error) {
	resp, err := s.fetchWithSafeResult(ctx, cfg, state, targetURL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *Service) fetchWithSafeResult(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string) (*fetchHTTPResult, error) {
	return s.fetchWithSafeResultWithOptions(ctx, cfg, state, targetURL, true)
}

func (s *Service) fetchWithSafeWithOptions(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, useProxy bool) ([]byte, error) {
	resp, err := s.fetchWithSafeResultWithOptions(ctx, cfg, state, targetURL, useProxy)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (s *Service) fetchWithSafeResultWithOptions(ctx context.Context, cfg fetchsite.SiteConfig, state *remoteState, targetURL string, useProxy bool) (*fetchHTTPResult, error) {
	resp, err := s.fetchWithRetryResultWithOptions(ctx, cfg, state, targetURL, cfg.Cookie, useProxy)
	if err != nil {
		return nil, err
	}

	safeID := parseSafeID(resp.Body)
	if safeID == "" {
		return resp, nil
	}

	safeCookie := combineCookie(cfg.Cookie, safeID)
	return s.fetchWithRetryResultWithOptions(ctx, cfg, state, targetURL, safeCookie, useProxy)
}
