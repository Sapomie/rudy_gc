package media

import (
	"path/filepath"
	"strings"

	"rudy_gc/internal/dep"
)

type Service struct {
	deps *dep.Dep
}

func NewService(d *dep.Dep) *Service {
	return &Service{
		deps: d,
	}
}

func (s *Service) mediaRoots() []string {
	raw := s.deps.Config.Media.RootDirs
	if len(raw) == 0 {
		return nil
	}

	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, root := range raw {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		cleaned := filepath.Clean(root)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}
