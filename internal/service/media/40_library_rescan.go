package media

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/wfoldertree"
)

type LibraryRescanResult struct {
	ConfiguredRoots int                  `json:"configured_roots"`
	SelectedRoots   []string             `json:"selected_roots"`
	ScannedRoots    []string             `json:"scanned_roots"`
	SkippedRoots    []string             `json:"skipped_roots"`
	SelectedTargets []string             `json:"selected_targets"`
	ScannedTargets  []string             `json:"scanned_targets"`
	SkippedTargets  []string             `json:"skipped_targets"`
	TotalFiles      int                  `json:"total_files"`
	Matched         int                  `json:"matched"`
	Unchanged       int                  `json:"unchanged"`
	Moved           int                  `json:"moved"`
	Restored        int                  `json:"restored"`
	MarkedRemoved   int                  `json:"marked_removed"`
	Unmatched       int                  `json:"unmatched"`
	Errors          int                  `json:"errors"`
	MovedItems      []*LibraryRescanItem `json:"moved_items"`
	RestoredItems   []*LibraryRescanItem `json:"restored_items"`
	RemovedItems    []*LibraryRescanItem `json:"removed_items"`
	UnmatchedItems  []*LibraryRescanItem `json:"unmatched_items"`
	ErrorItems      []*LibraryRescanItem `json:"error_items"`
}

type LibraryRescanItem struct {
	MovieName    string `json:"movie_name"`
	MovieJavId   string `json:"movie_jav_id"`
	FileName     string `json:"file_name"`
	MatchBy      string `json:"match_by"`
	Path         string `json:"path"`
	PreviousPath string `json:"previous_path"`
	RootDir      string `json:"root_dir"`
	PreviousRoot string `json:"previous_root"`
	Error        string `json:"error"`
}

type libraryRescanScope struct {
	Root   string
	Scope  string
	Branch string
	Path   string
	Label  string
}

const (
	scopeMedia   = "media"
	scopeWatched = "watched"
)

func (s *Service) RescanLibrary(ctx context.Context, selections []LibraryRescanSelection) (result *LibraryRescanResult, err error) {
	allRoots := s.mediaRoots()
	safeSelections := sanitizeRescanSelections(selections, allRoots)

	result = &LibraryRescanResult{
		ConfiguredRoots: len(allRoots),
		SelectedRoots:   collectSelectedRoots(safeSelections),
		ScannedRoots:    []string{},
		SkippedRoots:    []string{},
		SelectedTargets: []string{},
		ScannedTargets:  []string{},
		SkippedTargets:  []string{},
		MovedItems:      []*LibraryRescanItem{},
		RestoredItems:   []*LibraryRescanItem{},
		RemovedItems:    []*LibraryRescanItem{},
		UnmatchedItems:  []*LibraryRescanItem{},
		ErrorItems:      []*LibraryRescanItem{},
	}

	if len(safeSelections) == 0 {
		return result, fmt.Errorf("至少选择一个已配置的扫描目录")
	}

	scopes := make([]libraryRescanScope, 0, len(safeSelections))
	nowUnix := time.Now().Unix()
	scannedRootSet := make(map[string]struct{}, len(safeSelections))
	skippedRootSet := make(map[string]struct{}, len(safeSelections))
	for _, selection := range safeSelections {
		layout := buildRootLayout(selection.Root)
		baseDir := rescanBaseDir(layout, selection.Scope)
		targetLabels := buildSelectionTargetLabels(baseDir, selection.Branches)
		result.SelectedTargets = append(result.SelectedTargets, targetLabels...)

		if !pathExists(layout.rootDir) || !pathExists(baseDir) {
			result.SkippedTargets = append(result.SkippedTargets, targetLabels...)
			skippedRootSet[selection.Root] = struct{}{}
			continue
		}

		if len(selection.Branches) == 0 {
			scopes = append(scopes, libraryRescanScope{
				Root:  selection.Root,
				Scope: selection.Scope,
				Path:  baseDir,
				Label: baseDir,
			})
			scannedRootSet[selection.Root] = struct{}{}
			continue
		}

		for _, branch := range selection.Branches {
			targetPath := filepath.Join(baseDir, branch)
			if !pathExists(targetPath) {
				result.SkippedTargets = append(result.SkippedTargets, targetPath)
				continue
			}
			scopes = append(scopes, libraryRescanScope{
				Root:   selection.Root,
				Scope:  selection.Scope,
				Branch: branch,
				Path:   targetPath,
				Label:  targetPath,
			})
			scannedRootSet[selection.Root] = struct{}{}
		}
	}

	result.ScannedRoots = sortedKeys(scannedRootSet)
	result.SkippedRoots = sortedKeys(skippedRootSet)

	if len(scopes) == 0 {
		sort.Strings(result.SelectedTargets)
		sort.Strings(result.SkippedTargets)
		return result, nil
	}

	prefixes := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		prefixes = append(prefixes, scope.Path)
		result.ScannedTargets = append(result.ScannedTargets, scope.Label)
	}

	rowsByScope, err := s.deps.WMediaModel.ListByFullDirPrefixes(ctx, prefixes)
	if err != nil {
		return result, err
	}
	trackedRows := make(map[int64]*moviex.WMedia, len(rowsByScope))
	for _, row := range rowsByScope {
		if row == nil {
			continue
		}
		trackedRows[row.Id] = row
	}

	seenIDs := make(map[int64]struct{}, len(trackedRows))
	touchedMovieJavIDs := make(map[string]struct{}, len(trackedRows))
	for _, scope := range scopes {
		if err := s.rescanOneScope(ctx, scope, nowUnix, result, seenIDs, touchedMovieJavIDs); err != nil {
			return result, err
		}
	}

	for _, row := range trackedRows {
		if row == nil {
			continue
		}
		if _, ok := seenIDs[row.Id]; ok {
			continue
		}
		if row.IsRemoved == consts.FilmIsRemoved {
			continue
		}

		updated := cloneWMedia(row)
		updated.IsRemoved = consts.FilmIsRemoved
		updated.RemoveTime = nowUnix
		updated.UpdatedOn = nowUnix
		if err := s.deps.WMediaModel.Update(ctx, updated); err != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				MovieName:  updated.MovieName,
				MovieJavId: updated.MovieJavId,
				FileName:   updated.FileName,
				Path:       joinMediaPath(updated.FullDir, updated.FileName),
				RootDir:    updated.RootDir,
				Error:      err.Error(),
			})
			continue
		}
		if err := s.syncPersonStatsByMovieJavIDs(ctx, updated.UpdatedOn, updated.MovieJavId); err != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				MovieName:  updated.MovieName,
				MovieJavId: updated.MovieJavId,
				FileName:   updated.FileName,
				Path:       joinMediaPath(updated.FullDir, updated.FileName),
				RootDir:    updated.RootDir,
				Error:      err.Error(),
			})
			continue
		}
		s.invalidateMovieTypeCaches(ctx, updated.MovieJavId)
		if strings.TrimSpace(updated.MovieJavId) != "" {
			touchedMovieJavIDs[strings.TrimSpace(updated.MovieJavId)] = struct{}{}
		}

		result.MarkedRemoved++
		result.RemovedItems = append(result.RemovedItems, &LibraryRescanItem{
			MovieName:  updated.MovieName,
			MovieJavId: updated.MovieJavId,
			FileName:   updated.FileName,
			Path:       joinMediaPath(updated.FullDir, updated.FileName),
			RootDir:    updated.RootDir,
		})
	}
	if len(touchedMovieJavIDs) > 0 {
		javIDs := make([]string, 0, len(touchedMovieJavIDs))
		for javID := range touchedMovieJavIDs {
			javIDs = append(javIDs, javID)
		}
		sort.Strings(javIDs)
		if err := s.syncPersonStatsByMovieJavIDs(ctx, nowUnix, javIDs...); err != nil {
			return result, err
		}
		s.movieSvc.EnqueueAggRebuildByMovieJavIDs("media_rescan", javIDs...)
	}

	sort.Strings(result.SelectedTargets)
	sort.Strings(result.ScannedTargets)
	sort.Strings(result.SkippedTargets)
	return result, nil
}

func sanitizeRescanSelections(selections []LibraryRescanSelection, configuredRoots []string) []LibraryRescanSelection {
	allowedRoots := make(map[string]struct{}, len(configuredRoots))
	for _, root := range configuredRoots {
		root = filepath.Clean(strings.TrimSpace(root))
		if root == "" {
			continue
		}
		allowedRoots[root] = struct{}{}
	}

	rootOrder := make([]string, 0, len(selections))
	scopeOrder := make(map[string][]string, len(selections))
	branchMap := make(map[string]map[string]map[string]struct{}, len(selections))
	fullScopeMap := make(map[string]map[string]bool, len(selections))
	for _, selection := range selections {
		root := filepath.Clean(strings.TrimSpace(selection.Root))
		if root == "" {
			continue
		}
		if _, ok := allowedRoots[root]; !ok {
			continue
		}
		scope := normalizeRescanScope(selection.Scope)
		if scope == "" {
			continue
		}
		if _, ok := branchMap[root]; !ok {
			branchMap[root] = map[string]map[string]struct{}{}
			fullScopeMap[root] = map[string]bool{}
			rootOrder = append(rootOrder, root)
		}
		if _, ok := branchMap[root][scope]; !ok {
			branchMap[root][scope] = map[string]struct{}{}
			scopeOrder[root] = append(scopeOrder[root], scope)
		}

		safeBranches := sanitizeRescanBranches(selection.Branches)
		if len(safeBranches) == 0 {
			fullScopeMap[root][scope] = true
			branchMap[root][scope] = map[string]struct{}{}
			continue
		}
		if fullScopeMap[root][scope] {
			continue
		}
		for _, branch := range safeBranches {
			branchMap[root][scope][branch] = struct{}{}
		}
	}

	out := make([]LibraryRescanSelection, 0, len(selections))
	for _, root := range rootOrder {
		for _, scope := range orderedRescanScopes(scopeOrder[root]) {
			if fullScopeMap[root][scope] {
				out = append(out, LibraryRescanSelection{
					Root:  root,
					Scope: scope,
				})
				continue
			}
			branches := make([]string, 0, len(branchMap[root][scope]))
			for branch := range branchMap[root][scope] {
				branches = append(branches, branch)
			}
			sort.Strings(branches)
			out = append(out, LibraryRescanSelection{
				Root:     root,
				Scope:    scope,
				Branches: branches,
			})
		}
	}
	return out
}

func sanitizeRescanBranches(branches []string) []string {
	out := make([]string, 0, len(branches))
	seen := make(map[string]struct{}, len(branches))
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || branch == "." || branch == ".." {
			continue
		}
		if strings.Contains(branch, "/") || strings.Contains(branch, string(filepath.Separator)) {
			continue
		}
		if filepath.Base(branch) != branch {
			continue
		}
		if _, ok := seen[branch]; ok {
			continue
		}
		seen[branch] = struct{}{}
		out = append(out, branch)
	}
	return out
}

func collectSelectedRoots(selections []LibraryRescanSelection) []string {
	out := make([]string, 0, len(selections))
	seen := make(map[string]struct{}, len(selections))
	for _, selection := range selections {
		if _, ok := seen[selection.Root]; ok {
			continue
		}
		seen[selection.Root] = struct{}{}
		out = append(out, selection.Root)
	}
	return out
}

func buildSelectionTargetLabels(mediaDir string, branches []string) []string {
	if len(branches) == 0 {
		return []string{mediaDir}
	}
	out := make([]string, 0, len(branches))
	for _, branch := range branches {
		out = append(out, filepath.Join(mediaDir, branch))
	}
	return out
}

func rescanBaseDir(layout rootLayout, scope string) string {
	switch normalizeRescanScope(scope) {
	case scopeWatched:
		return layout.watched
	default:
		return layout.media
	}
}

func normalizeRescanScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case scopeWatched:
		return scopeWatched
	default:
		return scopeMedia
	}
}

func orderedRescanScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, candidate := range []string{scopeMedia, scopeWatched} {
		for _, scope := range scopes {
			scope = normalizeRescanScope(scope)
			if scope != candidate {
				continue
			}
			if _, ok := seen[scope]; ok {
				continue
			}
			seen[scope] = struct{}{}
			out = append(out, scope)
		}
	}
	return out
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func cloneWMedia(row *moviex.WMedia) *moviex.WMedia {
	if row == nil {
		return nil
	}
	cp := *row
	return &cp
}

func joinMediaPath(dir, fileName string) string {
	dir = strings.TrimSpace(dir)
	fileName = strings.TrimSpace(fileName)
	switch {
	case dir == "" && fileName == "":
		return ""
	case dir == "":
		return fileName
	case fileName == "":
		return dir
	default:
		return filepath.Join(dir, fileName)
	}
}

func buildRescanItemFromMedia(row *moviex.WMedia) *LibraryRescanItem {
	if row == nil {
		return &LibraryRescanItem{}
	}
	return &LibraryRescanItem{
		MovieName:  row.MovieName,
		MovieJavId: row.MovieJavId,
		FileName:   row.FileName,
		Path:       joinMediaPath(row.FullDir, row.FileName),
		RootDir:    row.RootDir,
	}
}

func updateRowLocation(row *moviex.WMedia, rootDir, fullDir, fileName string, directoryID, nowUnix int64) bool {
	changed := false
	cleanRoot := filepath.Clean(strings.TrimSpace(rootDir))
	cleanDir := filepath.Clean(strings.TrimSpace(fullDir))
	fileName = strings.TrimSpace(fileName)

	if row.DirectoryId != directoryID {
		row.DirectoryId = directoryID
		changed = true
	}
	if filepath.Clean(strings.TrimSpace(row.RootDir)) != cleanRoot {
		row.RootDir = cleanRoot
		changed = true
	}
	if filepath.Clean(strings.TrimSpace(row.FullDir)) != cleanDir {
		row.FullDir = cleanDir
		changed = true
	}
	if strings.TrimSpace(row.FileName) != fileName {
		row.FileName = fileName
		changed = true
	}
	if row.IsRemoved != consts.FilmIsNotRemoved {
		row.IsRemoved = consts.FilmIsNotRemoved
		changed = true
	}
	if row.RemoveTime != 0 {
		row.RemoveTime = 0
		changed = true
	}
	if changed {
		row.UpdatedOn = nowUnix
	}
	return changed
}

func ensureFolderChainForDir(ctx context.Context, svc *Service, fullDir string, nowUnix int64) (*moviex.WFolder, error) {
	return wfoldertree.EnsurePathChain(ctx, svc.deps.WFolderModel, consts.WFolderSourceNative, fullDir, nowUnix)
}
