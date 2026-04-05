package media

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LibraryRescanSelection struct {
	Scope    string   `json:"scope"`
	Root     string   `json:"root"`
	Branches []string `json:"branches"`
}

type LibraryRescanRootOption struct {
	Root   string                      `json:"root"`
	Scopes []*LibraryRescanScopeOption `json:"scopes"`
}

type LibraryRescanScopeOption struct {
	Scope    string                    `json:"scope"`
	Label    string                    `json:"label"`
	BaseDir  string                    `json:"base_dir"`
	Exists   bool                      `json:"exists"`
	Children []*LibraryRescanDirOption `json:"children"`
}

type LibraryRescanDirOption struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (s *Service) ListLibraryRescanRootOptions(ctx context.Context) ([]*LibraryRescanRootOption, error) {
	_ = ctx

	roots := s.mediaRoots()
	out := make([]*LibraryRescanRootOption, 0, len(roots))
	for _, root := range roots {
		layout := buildRootLayout(root)
		option := &LibraryRescanRootOption{
			Root: root,
			Scopes: []*LibraryRescanScopeOption{
				buildLibraryRescanScopeOption(scopeMedia, "media", layout.media),
				buildLibraryRescanScopeOption(scopeWatched, "watched", layout.watched),
			},
		}
		for _, scope := range option.Scopes {
			if scope == nil || !scope.Exists {
				continue
			}
			children, err := listImmediateChildDirs(scope.BaseDir)
			if err != nil {
				return nil, err
			}
			scope.Children = children
		}
		out = append(out, option)
	}
	return out, nil
}

func buildLibraryRescanScopeOption(scope, label, baseDir string) *LibraryRescanScopeOption {
	return &LibraryRescanScopeOption{
		Scope:    normalizeRescanScope(scope),
		Label:    strings.TrimSpace(label),
		BaseDir:  filepath.Clean(strings.TrimSpace(baseDir)),
		Exists:   pathExists(baseDir),
		Children: []*LibraryRescanDirOption{},
	}
}

func listImmediateChildDirs(dir string) ([]*LibraryRescanDirOption, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]*LibraryRescanDirOption, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		out = append(out, &LibraryRescanDirOption{
			Name: name,
			Path: filepath.Join(dir, name),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}
