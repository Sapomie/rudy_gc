package fetchsite

import (
	"strings"
	"sync"

	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type Registry struct {
	mu      sync.RWMutex
	configs map[string]fetchsitetypes.Config
}

func NewRegistry(configs map[string]fetchsitetypes.Config) *Registry {
	registry := &Registry{configs: make(map[string]fetchsitetypes.Config, len(configs))}
	for siteCode, cfg := range configs {
		registry.configs[siteCode] = cfg
	}
	return registry
}

func (r *Registry) Get(siteCode string) (fetchsitetypes.Config, bool) {
	if r == nil {
		return fetchsitetypes.Config{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	cfg, ok := r.configs[strings.TrimSpace(siteCode)]
	return cfg, ok
}

func (r *Registry) List() []fetchsitetypes.Config {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]fetchsitetypes.Config, 0, len(r.configs))
	for _, cfg := range r.configs {
		out = append(out, cfg)
	}
	return out
}

func (r *Registry) Set(cfg fetchsitetypes.Config) bool {
	if r == nil {
		return false
	}
	cfg.SiteCode = strings.TrimSpace(cfg.SiteCode)
	if cfg.SiteCode == "" {
		return false
	}
	cfg.BaseURL = NormalizeBaseURL(cfg.BaseURL)
	r.mu.Lock()
	if r.configs == nil {
		r.configs = make(map[string]fetchsitetypes.Config)
	}
	r.configs[cfg.SiteCode] = cfg
	r.mu.Unlock()
	return true
}

func (r *Registry) BaseURL(siteCode string) string {
	cfg, ok := r.Get(siteCode)
	if !ok {
		return ""
	}
	return cfg.BaseURL
}
