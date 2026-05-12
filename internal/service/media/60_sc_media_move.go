package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

const (
	scMediaMoveStatusPass = "pass"
	scMediaMoveStatusSkip = "skip"
	scMediaMoveStatusFail = "fail"
)

type ScMediaMoveResult struct {
	ScName       string             `json:"sc_name"`
	GeneratedAt  int64              `json:"generated_at"`
	Total        int                `json:"total"`
	Movable      int                `json:"movable"`
	Skipped      int                `json:"skipped"`
	Failed       int                `json:"failed"`
	Success      int                `json:"success"`
	CommitFailed int                `json:"commit_failed"`
	Items        []*ScMediaMoveItem `json:"items"`
}

type ScMediaMoveItem struct {
	Status     string `json:"status"`
	MovieJavId string `json:"movie_jav_id"`
	MovieName  string `json:"movie_name"`
	RootDir    string `json:"root_dir"`
	SourcePath string `json:"source_path"`
	TargetDir  string `json:"target_dir"`
	TargetPath string `json:"target_path"`
	Error      string `json:"error"`
}

func (s *Service) ScMediaMovePrecheck(ctx context.Context, scName string) (*ScMediaMoveResult, error) {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return nil, fmt.Errorf("sc_name 为空")
	}
	if len(s.mediaRoots()) == 0 {
		return nil, fmt.Errorf("media.rootDirs 未配置")
	}

	now := time.Now()
	items, err := s.buildScMediaMovePrecheckItems(ctx, scName, now)
	if err != nil {
		return nil, err
	}

	result := buildScMediaMovePrecheckResult(scName, now.Unix(), items)
	if err = s.saveScMediaMovePlan(scName, buildScMediaMovePlan(scName, result.GeneratedAt, result.Items)); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ScMediaMoveCommit(ctx context.Context, scName string) (result *ScMediaMoveResult, err error) {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return nil, fmt.Errorf("sc_name 为空")
	}
	plan, err := s.loadScMediaMovePlan(scName)
	if err != nil {
		if errors.Is(err, ErrScMediaMovePlanNotFound) {
			return &ScMediaMoveResult{
				ScName: scName,
				Items:  []*ScMediaMoveItem{},
			}, nil
		}
		return nil, err
	}

	now := time.Now()
	result = &ScMediaMoveResult{
		ScName:      scName,
		GeneratedAt: now.Unix(),
		Total:       plan.Total,
		Movable:     plan.Movable,
		Skipped:     plan.Skipped,
		Failed:      plan.Failed,
		Items:       make([]*ScMediaMoveItem, 0, len(plan.Entries)),
	}
	changedMovieJavIDs := make([]string, 0, len(plan.Entries))

	for _, entry := range plan.Entries {
		item := s.commitOneScMediaMove(ctx, entry, now)
		result.Items = append(result.Items, item)
		if strings.TrimSpace(item.Error) == "" {
			result.Success++
			if javID := strings.TrimSpace(item.MovieJavId); javID != "" {
				changedMovieJavIDs = append(changedMovieJavIDs, javID)
			}
		} else {
			result.CommitFailed++
		}
	}
	if len(changedMovieJavIDs) > 0 {
		s.movieSvc.EnqueueAggRebuildByMovieJavIDs("sc_media_move", changedMovieJavIDs...)
	}

	if clearErr := s.clearScMediaMovePlan(scName); clearErr != nil {
		return result, clearErr
	}
	return result, nil
}

func buildScMediaMovePrecheckResult(scName string, generatedAt int64, items []*ScMediaMoveItem) *ScMediaMoveResult {
	result := &ScMediaMoveResult{
		ScName:      strings.TrimSpace(scName),
		GeneratedAt: generatedAt,
		Items:       make([]*ScMediaMoveItem, 0, len(items)),
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		result.Items = append(result.Items, item)
		result.Total++
		switch normalizeScMediaMoveStatus(item.Status) {
		case scMediaMoveStatusPass:
			result.Movable++
		case scMediaMoveStatusSkip:
			result.Skipped++
		default:
			result.Failed++
		}
	}
	return result
}

func (s *Service) buildScMediaMovePrecheckItems(ctx context.Context, scName string, now time.Time) ([]*ScMediaMoveItem, error) {
	rows, err := s.deps.GListModel.ListScOnlyByScName(ctx, scName)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*ScMediaMoveItem{}, nil
	}

	movieNames := make(map[string]string, len(rows))
	javIDs := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	items := make([]*ScMediaMoveItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.MovieJavId)
		movieName := pickScMediaMovieName(row)
		if javID == "" {
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieName: movieName,
				Error:     "g_list.movie_jav_id 为空",
			})
			continue
		}
		if _, ok := seen[javID]; ok {
			continue
		}
		seen[javID] = struct{}{}
		movieNames[javID] = movieName
		javIDs = append(javIDs, javID)
	}

	mediaRows, err := s.deps.WMediaModel.ListByMovieJavIds(ctx, javIDs)
	if err != nil {
		return nil, err
	}
	mediaMap := make(map[string]*moviex.WMedia, len(mediaRows))
	for _, row := range mediaRows {
		if row == nil {
			continue
		}
		mediaMap[strings.TrimSpace(row.MovieJavId)] = row
	}

	for _, javID := range javIDs {
		movieName := strings.TrimSpace(movieNames[javID])
		row := mediaMap[javID]
		item := s.buildOneScMediaMovePrecheckItem(row, javID, movieName, now)
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if rankScMediaMoveStatus(items[i].Status) != rankScMediaMoveStatus(items[j].Status) {
			return rankScMediaMoveStatus(items[i].Status) < rankScMediaMoveStatus(items[j].Status)
		}
		if items[i].MovieName != items[j].MovieName {
			return items[i].MovieName < items[j].MovieName
		}
		return items[i].SourcePath < items[j].SourcePath
	})
	return items, nil
}

func (s *Service) buildOneScMediaMovePrecheckItem(row *moviex.WMedia, javID, fallbackMovieName string, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:     scMediaMoveStatusFail,
		MovieJavId: strings.TrimSpace(javID),
		MovieName:  strings.TrimSpace(fallbackMovieName),
	}
	if row == nil {
		item.Status = scMediaMoveStatusSkip
		item.Error = "没有 w_media"
		return item
	}

	if strings.TrimSpace(item.MovieName) == "" {
		item.MovieName = strings.TrimSpace(row.MovieName)
	}
	item.RootDir = strings.TrimSpace(row.RootDir)
	item.SourcePath = joinMediaPath(row.FullDir, row.FileName)
	if strings.TrimSpace(item.SourcePath) == "" {
		item.Error = "w_media 源路径为空"
		return item
	}
	if row.IsRemoved == consts.FilmIsRemoved {
		item.Status = scMediaMoveStatusSkip
		item.Error = "w_media 已删除"
		return item
	}

	root, layout, err := s.resolveScMediaLayoutForRow(row)
	if err != nil {
		item.Error = err.Error()
		return item
	}
	item.RootDir = root

	sourceDir := filepath.Clean(strings.TrimSpace(row.FullDir))
	if isPathWithin(sourceDir, layout.watched) {
		item.Status = scMediaMoveStatusSkip
		item.Error = "已在 watched 目录"
		return item
	}
	if !isPathWithin(sourceDir, layout.media) {
		item.Error = "w_media 不在 media 目录"
		return item
	}
	if _, statErr := os.Stat(item.SourcePath); statErr != nil {
		item.Error = "源文件不存在: " + statErr.Error()
		return item
	}

	targetDir, err := previewWatchedTargetDirectory(layout, now)
	if err != nil {
		item.Error = err.Error()
		return item
	}
	targetPath := filepath.Join(targetDir, strings.TrimSpace(row.FileName))
	if _, statErr := os.Stat(targetPath); statErr == nil {
		item.Error = "目标目录已存在同名文件"
		return item
	} else if !os.IsNotExist(statErr) {
		item.Error = "检查目标路径失败: " + statErr.Error()
		return item
	}

	item.Status = scMediaMoveStatusPass
	item.TargetDir = targetDir
	item.TargetPath = targetPath
	return item
}

func (s *Service) commitOneScMediaMove(ctx context.Context, entry *scMediaMovePlanEntry, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:     scMediaMoveStatusPass,
		MovieJavId: strings.TrimSpace(entryMovieJavID(entry)),
		MovieName:  strings.TrimSpace(entryMovieName(entry)),
		RootDir:    strings.TrimSpace(entryRootDir(entry)),
		SourcePath: strings.TrimSpace(entrySourcePath(entry)),
	}

	if item.MovieJavId == "" {
		item.Status = scMediaMoveStatusFail
		item.Error = "计划中的 movie_jav_id 为空"
		return item
	}

	row, err := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, item.MovieJavId, consts.WMediaSourceNative)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			item.Status = scMediaMoveStatusFail
			item.Error = "w_media 不存在"
			return item
		}
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	if row == nil {
		item.Status = scMediaMoveStatusFail
		item.Error = "w_media 不存在"
		return item
	}

	currentPath := joinMediaPath(row.FullDir, row.FileName)
	item.SourcePath = currentPath
	if row.IsRemoved == consts.FilmIsRemoved {
		item.Status = scMediaMoveStatusFail
		item.Error = "w_media 已删除"
		return item
	}
	if expected := strings.TrimSpace(entrySourcePath(entry)); expected != "" && filepath.Clean(expected) != filepath.Clean(currentPath) {
		item.Status = scMediaMoveStatusFail
		item.Error = "w_media 路径已变化，请重跑预处理"
		return item
	}

	root, layout, err := s.resolveScMediaLayoutForRow(row)
	if err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	item.RootDir = root

	currentDir := filepath.Clean(strings.TrimSpace(row.FullDir))
	if isPathWithin(currentDir, layout.watched) {
		item.Status = scMediaMoveStatusFail
		item.Error = "w_media 已在 watched 目录"
		return item
	}
	if !isPathWithin(currentDir, layout.media) {
		item.Status = scMediaMoveStatusFail
		item.Error = "w_media 不在 media 目录"
		return item
	}

	targetDir, folderID, err := s.allocateWatchedTargetDirectory(ctx, layout, now)
	if err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	targetPath := filepath.Join(targetDir, strings.TrimSpace(row.FileName))
	if _, statErr := os.Stat(targetPath); statErr == nil {
		item.Status = scMediaMoveStatusFail
		item.Error = "目标目录已存在同名文件"
		return item
	} else if !os.IsNotExist(statErr) {
		item.Status = scMediaMoveStatusFail
		item.Error = "检查目标路径失败: " + statErr.Error()
		return item
	}

	if err = moveFileWithFallback(ctx, currentPath, targetPath); err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}

	updated := cloneWMedia(row)
	updated.FullDir = filepath.Clean(targetDir)
	updated.DirectoryId = folderID
	updated.UpdatedOn = time.Now().Unix()
	if err = s.deps.WMediaModel.Update(ctx, updated); err != nil {
		if rollbackErr := moveFileWithFallback(ctx, targetPath, currentPath); rollbackErr != nil {
			item.Status = scMediaMoveStatusFail
			item.Error = err.Error() + "; 回滚文件失败: " + rollbackErr.Error()
			return item
		}
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	if err = s.syncPersonStatsByMovieJavIDs(ctx, updated.UpdatedOn, updated.MovieJavId); err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	s.invalidateMovieTypeCaches(ctx, updated.MovieJavId)

	item.TargetDir = updated.FullDir
	item.TargetPath = targetPath
	return item
}

func pickScMediaMovieName(row *moviex.GList) string {
	if row == nil {
		return ""
	}
	name := strings.TrimSpace(row.Name)
	if name == "" {
		return strings.TrimSpace(row.MovieJavId)
	}
	parts := strings.SplitN(name, "__", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return name
}

func (s *Service) resolveScMediaLayoutForRow(row *moviex.WMedia) (string, rootLayout, error) {
	roots := s.mediaRoots()
	if row == nil {
		return "", rootLayout{}, fmt.Errorf("w_media 为空")
	}

	cleanRoot := filepath.Clean(strings.TrimSpace(row.RootDir))
	for _, root := range roots {
		if cleanRoot != "" && cleanRoot == filepath.Clean(root) {
			layout := buildRootLayout(root)
			return root, layout, nil
		}
	}

	fullDir := filepath.Clean(strings.TrimSpace(row.FullDir))
	for _, root := range roots {
		layout := buildRootLayout(root)
		if isPathWithin(fullDir, layout.media) || isPathWithin(fullDir, layout.watched) {
			return root, layout, nil
		}
	}

	if cleanRoot != "" {
		return "", rootLayout{}, fmt.Errorf("w_media root_dir 未配置: %s", cleanRoot)
	}
	return "", rootLayout{}, fmt.Errorf("w_media 路径不在已配置 root 中: %s", fullDir)
}

func isPathWithin(pathValue, baseDir string) bool {
	pathValue = filepath.Clean(strings.TrimSpace(pathValue))
	baseDir = filepath.Clean(strings.TrimSpace(baseDir))
	if pathValue == "" || baseDir == "" {
		return false
	}
	rel, err := filepath.Rel(baseDir, pathValue)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func normalizeScMediaMoveStatus(status string) string {
	switch strings.TrimSpace(status) {
	case scMediaMoveStatusPass:
		return scMediaMoveStatusPass
	case scMediaMoveStatusSkip:
		return scMediaMoveStatusSkip
	default:
		return scMediaMoveStatusFail
	}
}

func rankScMediaMoveStatus(status string) int {
	switch normalizeScMediaMoveStatus(status) {
	case scMediaMoveStatusPass:
		return 1
	case scMediaMoveStatusSkip:
		return 2
	default:
		return 3
	}
}

func entryMovieJavID(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.MovieJavId
}

func entryMovieName(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.MovieName
}

func entryRootDir(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.RootDir
}

func entrySourcePath(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.SourcePath
}
