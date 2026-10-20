package media

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/service/wfoldertree"
)

const (
	libraryMediaTypeJAV = "jav"
	libraryMediaTypeFC2 = "fc2"
	scopeMedia          = "media"
	scopeWatched        = "watched"
)

type LibraryRescanTypeStat struct {
	ActiveRecords   int `json:"active_records"`
	DiskFiles       int `json:"disk_files"`
	Matched         int `json:"matched"`
	Healthy         int `json:"healthy"`
	Missing         int `json:"missing"`
	Untracked       int `json:"untracked"`
	SizeMismatch    int `json:"size_mismatch"`
	DuplicateActive int `json:"duplicate_active"`
	Errors          int `json:"errors"`
}

type LibraryRescanResult struct {
	ConfiguredRoots int                    `json:"configured_roots"`
	SelectedRoots   []string               `json:"selected_roots"`
	ScannedRoots    []string               `json:"scanned_roots"`
	SkippedRoots    []string               `json:"skipped_roots"`
	SelectedTargets []string               `json:"selected_targets"`
	ScannedTargets  []string               `json:"scanned_targets"`
	SkippedTargets  []string               `json:"skipped_targets"`
	ActiveRecords   int                    `json:"active_records"`
	DiskFiles       int                    `json:"disk_files"`
	Matched         int                    `json:"matched"`
	Healthy         int                    `json:"healthy"`
	Missing         int                    `json:"missing"`
	Untracked       int                    `json:"untracked"`
	SizeMismatch    int                    `json:"size_mismatch"`
	DuplicateActive int                    `json:"duplicate_active"`
	Errors          int                    `json:"errors"`
	JAVStats        *LibraryRescanTypeStat `json:"jav_stats"`
	FC2Stats        *LibraryRescanTypeStat `json:"fc2_stats"`
	MissingItems    []*LibraryRescanItem   `json:"missing_items"`
	UntrackedItems  []*LibraryRescanItem   `json:"untracked_items"`
	SizeItems       []*LibraryRescanItem   `json:"size_mismatch_items"`
	DuplicateItems  []*LibraryRescanItem   `json:"duplicate_items"`
	ErrorItems      []*LibraryRescanItem   `json:"error_items"`
}

type LibraryRescanItem struct {
	MediaType    string `json:"media_type"`
	Identity     string `json:"identity"`
	RecordID     int64  `json:"record_id"`
	MovieName    string `json:"movie_name"`
	MovieJavId   string `json:"movie_jav_id"`
	FC2MovieID   int64  `json:"fc2_movie_id"`
	FC2Key       string `json:"fc2_key"`
	FC2Code      string `json:"fc2_code"`
	FileName     string `json:"file_name"`
	Path         string `json:"path"`
	RootDir      string `json:"root_dir"`
	DatabaseSize int64  `json:"database_size"`
	DiskSize     int64  `json:"disk_size"`
	DuplicateKey string `json:"duplicate_key"`
	Error        string `json:"error"`
}

type libraryRescanConfiguredRoot struct {
	MediaType string
	Root      string
}

type libraryRescanScope struct {
	MediaType string
	Root      string
	Scope     string
	Branch    string
	Path      string
	Label     string
}

type libraryDiskFile struct {
	MediaType string
	Path      string
	FileName  string
	RootDir   string
	Size      int64
}

type libraryActiveRecord struct {
	MediaType  string
	RecordID   int64
	MovieName  string
	MovieJavID string
	FC2MovieID int64
	FC2Key     string
	FC2Code    string
	FileName   string
	RootDir    string
	FullDir    string
	Size       int64
	IsRemoved  int64
}

func (s *Service) RescanLibrary(ctx context.Context, selections []LibraryRescanSelection) (*LibraryRescanResult, error) {
	configuredRoots := s.libraryRescanConfiguredRoots()
	safeSelections := sanitizeRescanSelections(selections, configuredRoots)
	result := newLibraryRescanResult(len(configuredRoots), safeSelections)
	if len(safeSelections) == 0 {
		return result, fmt.Errorf("至少选择一个已配置的扫描目录")
	}

	scopes := resolveLibraryRescanScopes(safeSelections, result)
	if len(scopes) == 0 {
		finalizeLibraryRescanResult(result)
		return result, nil
	}

	diskFiles, err := scanLibraryRescanScopes(ctx, scopes, result)
	if err != nil {
		return result, err
	}
	records, err := s.loadLibraryActiveRecords(ctx, scopes)
	if err != nil {
		return result, err
	}

	auditLibraryAssociations(records, scopes, diskFiles, result)
	finalizeLibraryRescanResult(result)
	return result, nil
}

func (s *Service) libraryRescanConfiguredRoots() []libraryRescanConfiguredRoot {
	out := make([]libraryRescanConfiguredRoot, 0, len(s.mediaRoots())+len(s.fc2MediaRoots()))
	seen := make(map[string]struct{}, cap(out))
	appendRoot := func(mediaType, root string) {
		mediaType = normalizeLibraryMediaType(mediaType)
		root = filepath.Clean(strings.TrimSpace(root))
		if mediaType == "" || root == "." || root == "" {
			return
		}
		key := libraryMediaPathKey(mediaType, root)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, libraryRescanConfiguredRoot{MediaType: mediaType, Root: root})
	}
	for _, root := range s.mediaRoots() {
		appendRoot(libraryMediaTypeJAV, root)
	}
	for _, root := range s.fc2MediaRoots() {
		appendRoot(libraryMediaTypeFC2, root.rootDir)
	}
	return out
}

func (s *Service) loadLibraryActiveRecords(ctx context.Context, scopes []libraryRescanScope) ([]libraryActiveRecord, error) {
	selectedTypes := make(map[string]struct{}, 2)
	for _, scope := range scopes {
		selectedTypes[scope.MediaType] = struct{}{}
	}

	records := make([]libraryActiveRecord, 0)
	if _, ok := selectedTypes[libraryMediaTypeJAV]; ok {
		rows, err := s.deps.WMediaModel.ListNativeMediaAll(ctx)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			records = append(records, libraryActiveRecord{
				MediaType: libraryMediaTypeJAV, RecordID: row.Id, MovieName: row.MovieName, MovieJavID: row.MovieJavId,
				FileName: row.FileName, RootDir: row.RootDir, FullDir: row.FullDir, Size: row.Size, IsRemoved: row.IsRemoved,
			})
		}
	}
	if _, ok := selectedTypes[libraryMediaTypeFC2]; ok {
		rows, err := s.deps.FC2MediaModel.ListAll(ctx)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			records = append(records, libraryActiveRecord{
				MediaType: libraryMediaTypeFC2, RecordID: row.ID, FC2MovieID: row.FC2MovieID, FC2Key: row.FC2Key, FC2Code: row.FC2Code,
				FileName: row.FileName, RootDir: row.RootDir, FullDir: row.FullDir, Size: row.Size, IsRemoved: row.IsRemoved,
			})
		}
	}
	return records, nil
}

func newLibraryRescanResult(configuredRoots int, selections []LibraryRescanSelection) *LibraryRescanResult {
	return &LibraryRescanResult{
		ConfiguredRoots: configuredRoots,
		SelectedRoots:   collectSelectedRoots(selections),
		ScannedRoots:    []string{},
		SkippedRoots:    []string{},
		SelectedTargets: []string{},
		ScannedTargets:  []string{},
		SkippedTargets:  []string{},
		JAVStats:        &LibraryRescanTypeStat{},
		FC2Stats:        &LibraryRescanTypeStat{},
		MissingItems:    []*LibraryRescanItem{},
		UntrackedItems:  []*LibraryRescanItem{},
		SizeItems:       []*LibraryRescanItem{},
		DuplicateItems:  []*LibraryRescanItem{},
		ErrorItems:      []*LibraryRescanItem{},
	}
}

func resolveLibraryRescanScopes(selections []LibraryRescanSelection, result *LibraryRescanResult) []libraryRescanScope {
	scopes := make([]libraryRescanScope, 0, len(selections))
	scannedRootSet := make(map[string]struct{}, len(selections))
	skippedRootSet := make(map[string]struct{}, len(selections))
	for _, selection := range selections {
		layout := buildRootLayout(selection.Root)
		baseDir := rescanBaseDir(layout, selection.Scope)
		targetLabels := buildSelectionTargetLabels(selection.MediaType, baseDir, selection.Branches)
		result.SelectedTargets = append(result.SelectedTargets, targetLabels...)
		rootLabel := libraryTargetLabel(selection.MediaType, selection.Root)

		if !pathExists(layout.rootDir) || !pathExists(baseDir) {
			result.SkippedTargets = append(result.SkippedTargets, targetLabels...)
			skippedRootSet[rootLabel] = struct{}{}
			continue
		}
		if len(selection.Branches) == 0 {
			label := libraryTargetLabel(selection.MediaType, baseDir)
			scopes = append(scopes, libraryRescanScope{MediaType: selection.MediaType, Root: selection.Root, Scope: selection.Scope, Path: baseDir, Label: label})
			scannedRootSet[rootLabel] = struct{}{}
			continue
		}
		for _, branch := range selection.Branches {
			targetPath := filepath.Join(baseDir, branch)
			if !pathExists(targetPath) {
				result.SkippedTargets = append(result.SkippedTargets, libraryTargetLabel(selection.MediaType, targetPath))
				continue
			}
			label := libraryTargetLabel(selection.MediaType, targetPath)
			scopes = append(scopes, libraryRescanScope{MediaType: selection.MediaType, Root: selection.Root, Scope: selection.Scope, Branch: branch, Path: targetPath, Label: label})
			result.ScannedTargets = append(result.ScannedTargets, label)
			scannedRootSet[rootLabel] = struct{}{}
		}
	}
	for _, scope := range scopes {
		if scope.Branch == "" {
			result.ScannedTargets = append(result.ScannedTargets, scope.Label)
		}
	}
	result.ScannedRoots = sortedKeys(scannedRootSet)
	result.SkippedRoots = sortedKeys(skippedRootSet)
	return scopes
}

func auditLibraryAssociations(records []libraryActiveRecord, scopes []libraryRescanScope, diskFiles map[string]libraryDiskFile, result *LibraryRescanResult) {
	activeRecords := make([]libraryActiveRecord, 0, len(records))
	activeByPath := make(map[string][]libraryActiveRecord, len(records))
	activeByMovie := make(map[string][]libraryActiveRecord, len(records))
	for _, record := range records {
		if record.IsRemoved == consts.FilmIsRemoved {
			continue
		}
		path := cleanLibraryMediaPath(record.FullDir, record.FileName)
		if !libraryPathInScopes(record.MediaType, path, scopes) {
			continue
		}
		activeRecords = append(activeRecords, record)
		activeByPath[libraryMediaPathKey(record.MediaType, path)] = append(activeByPath[libraryMediaPathKey(record.MediaType, path)], record)
		duplicateKey := libraryActiveMovieKey(record)
		activeByMovie[libraryMediaPathKey(record.MediaType, duplicateKey)] = append(activeByMovie[libraryMediaPathKey(record.MediaType, duplicateKey)], record)
		result.ActiveRecords++
		libraryTypeStats(result, record.MediaType).ActiveRecords++
	}

	for _, diskFile := range diskFiles {
		result.DiskFiles++
		libraryTypeStats(result, diskFile.MediaType).DiskFiles++
	}
	for _, record := range activeRecords {
		path := cleanLibraryMediaPath(record.FullDir, record.FileName)
		diskFile, ok := diskFiles[libraryMediaPathKey(record.MediaType, path)]
		if !ok {
			result.Missing++
			libraryTypeStats(result, record.MediaType).Missing++
			result.MissingItems = append(result.MissingItems, libraryRescanItemFromRecord(record, path))
			continue
		}
		result.Matched++
		libraryTypeStats(result, record.MediaType).Matched++
		if record.Size != diskFile.Size {
			item := libraryRescanItemFromRecord(record, path)
			item.DiskSize = diskFile.Size
			result.SizeMismatch++
			libraryTypeStats(result, record.MediaType).SizeMismatch++
			result.SizeItems = append(result.SizeItems, item)
			continue
		}
		result.Healthy++
		libraryTypeStats(result, record.MediaType).Healthy++
	}

	for key, diskFile := range diskFiles {
		if len(activeByPath[key]) > 0 {
			continue
		}
		result.Untracked++
		libraryTypeStats(result, diskFile.MediaType).Untracked++
		result.UntrackedItems = append(result.UntrackedItems, &LibraryRescanItem{
			MediaType: diskFile.MediaType, FileName: diskFile.FileName, Path: diskFile.Path,
			RootDir: diskFile.RootDir, DiskSize: diskFile.Size,
		})
	}

	for _, duplicateRecords := range activeByMovie {
		if len(duplicateRecords) < 2 {
			continue
		}
		mediaType := duplicateRecords[0].MediaType
		duplicateKey := libraryActiveMovieKey(duplicateRecords[0])
		result.DuplicateActive++
		libraryTypeStats(result, mediaType).DuplicateActive++
		for _, record := range duplicateRecords {
			item := libraryRescanItemFromRecord(record, cleanLibraryMediaPath(record.FullDir, record.FileName))
			item.DuplicateKey = duplicateKey
			result.DuplicateItems = append(result.DuplicateItems, item)
		}
	}
}

func libraryRescanItemFromRecord(record libraryActiveRecord, path string) *LibraryRescanItem {
	return &LibraryRescanItem{
		MediaType: record.MediaType, Identity: libraryActiveMovieIdentity(record), RecordID: record.RecordID,
		MovieName: strings.TrimSpace(record.MovieName), MovieJavId: strings.TrimSpace(record.MovieJavID),
		FC2MovieID: record.FC2MovieID, FC2Key: strings.TrimSpace(record.FC2Key), FC2Code: strings.TrimSpace(record.FC2Code),
		FileName: strings.TrimSpace(record.FileName), Path: path, RootDir: strings.TrimSpace(record.RootDir), DatabaseSize: record.Size,
	}
}

func libraryActiveMovieIdentity(record libraryActiveRecord) string {
	if record.MediaType == libraryMediaTypeFC2 {
		if key := strings.TrimSpace(record.FC2Key); key != "" {
			return key
		}
		if code := strings.TrimSpace(record.FC2Code); code != "" {
			return code
		}
		if record.FC2MovieID > 0 {
			return strconv.FormatInt(record.FC2MovieID, 10)
		}
	} else {
		if javID := strings.TrimSpace(record.MovieJavID); javID != "" {
			return javID
		}
		if movieName := strings.TrimSpace(record.MovieName); movieName != "" {
			return movieName
		}
	}
	return strconv.FormatInt(record.RecordID, 10)
}

func libraryActiveMovieKey(record libraryActiveRecord) string {
	if record.MediaType == libraryMediaTypeFC2 {
		if key := strings.TrimSpace(record.FC2Key); key != "" {
			return "fc2_key:" + key
		}
		if record.FC2MovieID > 0 {
			return "fc2_movie_id:" + strconv.FormatInt(record.FC2MovieID, 10)
		}
		if code := strings.TrimSpace(record.FC2Code); code != "" {
			return "fc2_code:" + code
		}
	} else {
		if javID := strings.TrimSpace(record.MovieJavID); javID != "" {
			return "movie_jav_id:" + javID
		}
		if movieName := strings.TrimSpace(record.MovieName); movieName != "" {
			return "movie_name:" + movieName
		}
	}
	return "record_id:" + strconv.FormatInt(record.RecordID, 10)
}

func libraryTypeStats(result *LibraryRescanResult, mediaType string) *LibraryRescanTypeStat {
	if mediaType == libraryMediaTypeFC2 {
		return result.FC2Stats
	}
	return result.JAVStats
}

func cleanLibraryMediaPath(fullDir, fileName string) string {
	path := joinMediaPath(fullDir, fileName)
	if strings.TrimSpace(path) == "" {
		return ""
	}
	return filepath.Clean(path)
}

func libraryPathInScopes(mediaType, path string, scopes []libraryRescanScope) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	for _, scope := range scopes {
		if scope.MediaType == mediaType && isPathWithin(path, scope.Path) {
			return true
		}
	}
	return false
}

func libraryMediaPathKey(mediaType, path string) string {
	return normalizeLibraryMediaType(mediaType) + "\x00" + filepath.Clean(strings.TrimSpace(path))
}

func libraryTargetLabel(mediaType, path string) string {
	return strings.ToUpper(normalizeLibraryMediaType(mediaType)) + ": " + filepath.Clean(strings.TrimSpace(path))
}

func normalizeLibraryMediaType(mediaType string) string {
	switch strings.ToLower(strings.TrimSpace(mediaType)) {
	case libraryMediaTypeJAV:
		return libraryMediaTypeJAV
	case libraryMediaTypeFC2:
		return libraryMediaTypeFC2
	default:
		return ""
	}
}

func finalizeLibraryRescanResult(result *LibraryRescanResult) {
	if result == nil {
		return
	}
	sort.Strings(result.SelectedTargets)
	sort.Strings(result.ScannedTargets)
	sort.Strings(result.SkippedTargets)
	sortLibraryRescanItems(result.MissingItems)
	sortLibraryRescanItems(result.UntrackedItems)
	sortLibraryRescanItems(result.SizeItems)
	sort.Slice(result.DuplicateItems, func(i, j int) bool {
		if result.DuplicateItems[i].MediaType == result.DuplicateItems[j].MediaType && result.DuplicateItems[i].DuplicateKey == result.DuplicateItems[j].DuplicateKey {
			return result.DuplicateItems[i].Path < result.DuplicateItems[j].Path
		}
		if result.DuplicateItems[i].MediaType == result.DuplicateItems[j].MediaType {
			return result.DuplicateItems[i].DuplicateKey < result.DuplicateItems[j].DuplicateKey
		}
		return result.DuplicateItems[i].MediaType < result.DuplicateItems[j].MediaType
	})
	sortLibraryRescanItems(result.ErrorItems)
}

func sortLibraryRescanItems(items []*LibraryRescanItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i] == nil {
			return false
		}
		if items[j] == nil {
			return true
		}
		if items[i].MediaType == items[j].MediaType {
			return items[i].Path < items[j].Path
		}
		return items[i].MediaType < items[j].MediaType
	})
}

func sanitizeRescanSelections(selections []LibraryRescanSelection, configuredRoots []libraryRescanConfiguredRoot) []LibraryRescanSelection {
	allowedRoots := make(map[string]struct{}, len(configuredRoots))
	for _, configured := range configuredRoots {
		allowedRoots[libraryMediaPathKey(configured.MediaType, configured.Root)] = struct{}{}
	}

	selectionOrder := make([]string, 0, len(selections))
	scopeOrder := make(map[string][]string, len(selections))
	branchMap := make(map[string]map[string]map[string]struct{}, len(selections))
	fullScopeMap := make(map[string]map[string]bool, len(selections))
	selectionByKey := make(map[string]LibraryRescanSelection, len(selections))
	for _, selection := range selections {
		mediaType := normalizeLibraryMediaType(selection.MediaType)
		root := filepath.Clean(strings.TrimSpace(selection.Root))
		selectionKey := libraryMediaPathKey(mediaType, root)
		if _, ok := allowedRoots[selectionKey]; !ok {
			continue
		}
		scope := normalizeRescanScope(selection.Scope)
		if scope == "" {
			continue
		}
		if _, ok := branchMap[selectionKey]; !ok {
			branchMap[selectionKey] = map[string]map[string]struct{}{}
			fullScopeMap[selectionKey] = map[string]bool{}
			selectionOrder = append(selectionOrder, selectionKey)
			selectionByKey[selectionKey] = LibraryRescanSelection{MediaType: mediaType, Root: root}
		}
		if _, ok := branchMap[selectionKey][scope]; !ok {
			branchMap[selectionKey][scope] = map[string]struct{}{}
			scopeOrder[selectionKey] = append(scopeOrder[selectionKey], scope)
		}

		safeBranches := sanitizeRescanBranches(selection.Branches)
		if len(safeBranches) == 0 {
			fullScopeMap[selectionKey][scope] = true
			branchMap[selectionKey][scope] = map[string]struct{}{}
			continue
		}
		if fullScopeMap[selectionKey][scope] {
			continue
		}
		for _, branch := range safeBranches {
			branchMap[selectionKey][scope][branch] = struct{}{}
		}
	}

	out := make([]LibraryRescanSelection, 0, len(selections))
	for _, selectionKey := range selectionOrder {
		base := selectionByKey[selectionKey]
		for _, scope := range orderedRescanScopes(scopeOrder[selectionKey]) {
			if fullScopeMap[selectionKey][scope] {
				out = append(out, LibraryRescanSelection{MediaType: base.MediaType, Root: base.Root, Scope: scope})
				continue
			}
			branches := make([]string, 0, len(branchMap[selectionKey][scope]))
			for branch := range branchMap[selectionKey][scope] {
				branches = append(branches, branch)
			}
			sort.Strings(branches)
			out = append(out, LibraryRescanSelection{MediaType: base.MediaType, Root: base.Root, Scope: scope, Branches: branches})
		}
	}
	return out
}

func sanitizeRescanBranches(branches []string) []string {
	out := make([]string, 0, len(branches))
	seen := make(map[string]struct{}, len(branches))
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || branch == "." || branch == ".." || strings.Contains(branch, "/") || strings.Contains(branch, string(filepath.Separator)) || filepath.Base(branch) != branch {
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
		label := libraryTargetLabel(selection.MediaType, selection.Root)
		if _, ok := seen[label]; ok {
			continue
		}
		seen[label] = struct{}{}
		out = append(out, label)
	}
	return out
}

func buildSelectionTargetLabels(mediaType, mediaDir string, branches []string) []string {
	if len(branches) == 0 {
		return []string{libraryTargetLabel(mediaType, mediaDir)}
	}
	out := make([]string, 0, len(branches))
	for _, branch := range branches {
		out = append(out, libraryTargetLabel(mediaType, filepath.Join(mediaDir, branch)))
	}
	return out
}

func rescanBaseDir(layout rootLayout, scope string) string {
	if normalizeRescanScope(scope) == scopeWatched {
		return layout.watched
	}
	return layout.media
}

func normalizeRescanScope(scope string) string {
	switch strings.TrimSpace(scope) {
	case scopeMedia:
		return scopeMedia
	case scopeWatched:
		return scopeWatched
	default:
		return ""
	}
}

func orderedRescanScopes(scopes []string) []string {
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, candidate := range []string{scopeMedia, scopeWatched} {
		for _, scope := range scopes {
			if normalizeRescanScope(scope) != candidate {
				continue
			}
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			out = append(out, candidate)
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
