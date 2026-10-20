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

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

const (
	scMediaMoveModeSingle = "single"
	scMediaMoveModeAll    = "all"

	scMediaMoveStatusPass = "pass"
	scMediaMoveStatusSkip = "skip"
	scMediaMoveStatusFail = "fail"
)

type ScMediaMoveResult struct {
	Mode         string             `json:"mode"`
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
	Status     string   `json:"status"`
	MovieType  string   `json:"movie_type"`
	MovieJavId string   `json:"movie_jav_id"`
	FC2Key     string   `json:"fc2_key"`
	MovieName  string   `json:"movie_name"`
	ScNames    []string `json:"sc_names"`
	RootDir    string   `json:"root_dir"`
	SourcePath string   `json:"source_path"`
	TargetDir  string   `json:"target_dir"`
	TargetPath string   `json:"target_path"`
	Error      string   `json:"error"`
}

func (s *Service) ScMediaMovePrecheck(ctx context.Context, mode, scName string) (*ScMediaMoveResult, error) {
	mode = normalizeScMediaMoveMode(mode)
	scName = strings.TrimSpace(scName)
	if mode == scMediaMoveModeSingle && scName == "" {
		return nil, fmt.Errorf("sc_name 为空")
	}
	if len(s.mediaRoots()) == 0 && len(s.fc2ScMediaRoots()) == 0 {
		return nil, fmt.Errorf("media.rootDirs 与 fc2Media.rootDirs 均未配置")
	}

	now := time.Now()
	rows, err := s.listScMediaMoveRowsForScope(ctx, mode, scName)
	if err != nil {
		return nil, err
	}
	items, err := s.buildScMediaMovePrecheckItems(ctx, mode, scName, rows, now)
	if err != nil {
		return nil, err
	}

	result := buildScMediaMovePrecheckResult(mode, scName, now.Unix(), items)
	if err = s.saveScMediaMovePlan(mode, scName, buildScMediaMovePlan(mode, scName, result.GeneratedAt, result.Items)); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ScMediaMoveCommit(ctx context.Context, mode, scName string) (result *ScMediaMoveResult, err error) {
	mode = normalizeScMediaMoveMode(mode)
	scName = strings.TrimSpace(scName)
	if mode == scMediaMoveModeSingle && scName == "" {
		return nil, fmt.Errorf("sc_name 为空")
	}
	plan, err := s.loadScMediaMovePlan(mode, scName)
	if err != nil {
		if errors.Is(err, ErrScMediaMovePlanNotFound) {
			return &ScMediaMoveResult{
				Mode:   mode,
				ScName: scName,
				Items:  []*ScMediaMoveItem{},
			}, nil
		}
		return nil, err
	}

	now := time.Now()
	result = &ScMediaMoveResult{
		Mode:        mode,
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
		if len(item.ScNames) == 0 {
			item.ScNames = normalizeScMediaMoveItemScNames(mode, plan.ScName, entryScNames(entry))
		}
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
		s.movieSvc.EnqueueAggRebuildByMovieJavIDs(scMediaMoveAggFlowKey(mode), changedMovieJavIDs...)
	}

	if clearErr := s.clearScMediaMovePlan(mode, scName); clearErr != nil {
		return result, clearErr
	}
	return result, nil
}

func buildScMediaMovePrecheckResult(mode, scName string, generatedAt int64, items []*ScMediaMoveItem) *ScMediaMoveResult {
	result := &ScMediaMoveResult{
		Mode:        normalizeScMediaMoveMode(mode),
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

func (s *Service) listScMediaMoveRowsForScope(ctx context.Context, mode, scName string) ([]*moviex.GList, error) {
	listByType := func(movieType string) ([]*moviex.GList, error) {
		if normalizeScMediaMoveMode(mode) == scMediaMoveModeAll {
			return s.deps.GListModel.ListAllScOnlyByMovieType(ctx, movieType)
		}
		return s.deps.GListModel.ListScOnlyByScNameAndMovieType(ctx, strings.TrimSpace(scName), movieType)
	}

	javRows, err := listByType(moviedomain.SehuatangThreadMovieTypeJAV)
	if err != nil {
		return nil, err
	}
	fc2Rows, err := listByType(moviedomain.SehuatangThreadMovieTypeFC2PPV)
	if err != nil {
		return nil, err
	}
	rows := make([]*moviex.GList, 0, len(javRows)+len(fc2Rows))
	rows = append(rows, javRows...)
	rows = append(rows, fc2Rows...)
	return rows, nil
}

func (s *Service) buildScMediaMovePrecheckItems(ctx context.Context, mode, scName string, rows []*moviex.GList, now time.Time) ([]*ScMediaMoveItem, error) {
	if len(rows) == 0 {
		return []*ScMediaMoveItem{}, nil
	}
	javRows := make([]*moviex.GList, 0, len(rows))
	fc2Rows := make([]*moviex.GList, 0, len(rows))
	items := make([]*ScMediaMoveItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch normalizeScMediaMoveMovieType(row.MovieType, row.MovieJavId, "") {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			javRows = append(javRows, row)
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			fc2Rows = append(fc2Rows, row)
		default:
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieType: strings.TrimSpace(row.MovieType),
				MovieName: pickScMediaMovieName(row),
				ScNames:   normalizeScMediaMoveItemScNames(mode, scName, []string{row.ScName}),
				Error:     "g_list.movie_type 不支持: " + strings.TrimSpace(row.MovieType),
			})
		}
	}

	javItems, err := s.buildJAVScMediaMovePrecheckItems(ctx, mode, scName, javRows, now)
	if err != nil {
		return nil, err
	}
	items = append(items, javItems...)
	fc2Items, err := s.buildFC2ScMediaMovePrecheckItems(ctx, mode, scName, fc2Rows, now)
	if err != nil {
		return nil, err
	}
	items = append(items, fc2Items...)
	sortScMediaMoveItems(items)
	return items, nil
}

func (s *Service) buildJAVScMediaMovePrecheckItems(ctx context.Context, mode, scName string, rows []*moviex.GList, now time.Time) ([]*ScMediaMoveItem, error) {
	movieNames := make(map[string]string, len(rows))
	movieScNames := make(map[string][]string, len(rows))
	javIDs := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	items := make([]*ScMediaMoveItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		javID := strings.TrimSpace(row.MovieJavId)
		movieName := pickScMediaMovieName(row)
		currentScName := strings.TrimSpace(row.ScName)
		if javID == "" {
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieType: moviedomain.SehuatangThreadMovieTypeJAV,
				MovieName: movieName,
				ScNames:   normalizeScMediaMoveItemScNames(mode, scName, []string{currentScName}),
				Error:     "g_list.movie_jav_id 为空",
			})
			continue
		}
		if _, ok := seen[javID]; ok {
			if strings.TrimSpace(movieNames[javID]) == "" && movieName != "" {
				movieNames[javID] = movieName
			}
			movieScNames[javID] = appendScMediaMoveName(movieScNames[javID], currentScName)
			continue
		}
		seen[javID] = struct{}{}
		movieNames[javID] = movieName
		movieScNames[javID] = appendScMediaMoveName(movieScNames[javID], currentScName)
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
		item.ScNames = normalizeScMediaMoveItemScNames(mode, scName, movieScNames[javID])
		items = append(items, item)
	}

	return items, nil
}

func (s *Service) buildFC2ScMediaMovePrecheckItems(ctx context.Context, mode, scName string, rows []*moviex.GList, now time.Time) ([]*ScMediaMoveItem, error) {
	movieNames := make(map[int64]string, len(rows))
	movieScNames := make(map[int64][]string, len(rows))
	movieIDs := make([]int64, 0, len(rows))
	seenMovieIDs := make(map[int64]struct{}, len(rows))
	items := make([]*ScMediaMoveItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		movieID := row.FC2MovieId
		movieName := pickScMediaMovieName(row)
		currentScName := strings.TrimSpace(row.ScName)
		if movieID <= 0 {
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
				MovieName: movieName,
				ScNames:   normalizeScMediaMoveItemScNames(mode, scName, []string{currentScName}),
				Error:     "g_list 无法解析 fc2_key",
			})
			continue
		}
		if _, ok := seenMovieIDs[movieID]; ok {
			if strings.TrimSpace(movieNames[movieID]) == "" && movieName != "" {
				movieNames[movieID] = movieName
			}
			movieScNames[movieID] = appendScMediaMoveName(movieScNames[movieID], currentScName)
			continue
		}
		seenMovieIDs[movieID] = struct{}{}
		movieNames[movieID] = movieName
		movieScNames[movieID] = appendScMediaMoveName(movieScNames[movieID], currentScName)
		movieIDs = append(movieIDs, movieID)
	}

	movieRows, err := s.deps.FC2MovieModel.ListByIDs(ctx, movieIDs)
	if err != nil {
		return nil, err
	}
	movieMap := make(map[int64]*moviex.FC2Movie, len(movieRows))
	for _, row := range movieRows {
		if row != nil {
			movieMap[row.ID] = row
		}
	}

	fc2Keys := make([]string, 0, len(movieIDs))
	keyNames := make(map[string]string, len(movieIDs))
	keyScNames := make(map[string][]string, len(movieIDs))
	seenKeys := make(map[string]struct{}, len(movieIDs))
	for _, movieID := range movieIDs {
		movieRow := movieMap[movieID]
		if movieRow == nil {
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
				MovieName: strings.TrimSpace(movieNames[movieID]),
				ScNames:   normalizeScMediaMoveItemScNames(mode, scName, movieScNames[movieID]),
				Error:     "fc2_movie 不存在",
			})
			continue
		}
		fc2Key := strings.TrimSpace(movieRow.FC2Key)
		movieName := strings.TrimSpace(movieNames[movieID])
		if movieName == "" {
			movieName = strings.TrimSpace(movieRow.Title)
		}
		if movieName == "" {
			movieName = fc2Key
		}
		if fc2Key == "" {
			items = append(items, &ScMediaMoveItem{
				Status:    scMediaMoveStatusFail,
				MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
				MovieName: movieName,
				ScNames:   normalizeScMediaMoveItemScNames(mode, scName, movieScNames[movieID]),
				Error:     "fc2_movie.fc2_key 为空",
			})
			continue
		}
		if _, ok := seenKeys[fc2Key]; ok {
			if strings.TrimSpace(keyNames[fc2Key]) == "" && movieName != "" {
				keyNames[fc2Key] = movieName
			}
			for _, currentScName := range movieScNames[movieID] {
				keyScNames[fc2Key] = appendScMediaMoveName(keyScNames[fc2Key], currentScName)
			}
			continue
		}
		seenKeys[fc2Key] = struct{}{}
		fc2Keys = append(fc2Keys, fc2Key)
		keyNames[fc2Key] = movieName
		keyScNames[fc2Key] = append([]string{}, movieScNames[movieID]...)
	}

	mediaRows, err := s.deps.FC2MediaModel.ListByKeys(ctx, fc2Keys)
	if err != nil {
		return nil, err
	}
	mediaMap := make(map[string]*moviex.FC2Media, len(mediaRows))
	for _, row := range mediaRows {
		if row != nil {
			mediaMap[strings.TrimSpace(row.FC2Key)] = row
		}
	}
	for _, fc2Key := range fc2Keys {
		item := s.buildOneFC2ScMediaMovePrecheckItem(mediaMap[fc2Key], fc2Key, keyNames[fc2Key], now)
		item.ScNames = normalizeScMediaMoveItemScNames(mode, scName, keyScNames[fc2Key])
		items = append(items, item)
	}
	return items, nil
}

func sortScMediaMoveItems(items []*ScMediaMoveItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if rankScMediaMoveStatus(items[i].Status) != rankScMediaMoveStatus(items[j].Status) {
			return rankScMediaMoveStatus(items[i].Status) < rankScMediaMoveStatus(items[j].Status)
		}
		if items[i].MovieName != items[j].MovieName {
			return items[i].MovieName < items[j].MovieName
		}
		return items[i].SourcePath < items[j].SourcePath
	})
}

func (s *Service) buildOneScMediaMovePrecheckItem(row *moviex.WMedia, javID, fallbackMovieName string, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:     scMediaMoveStatusFail,
		MovieType:  moviedomain.SehuatangThreadMovieTypeJAV,
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

func (s *Service) buildOneFC2ScMediaMovePrecheckItem(row *moviex.FC2Media, fc2Key, fallbackMovieName string, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:    scMediaMoveStatusFail,
		MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2Key:    strings.TrimSpace(fc2Key),
		MovieName: strings.TrimSpace(fallbackMovieName),
	}
	if row == nil {
		item.Status = scMediaMoveStatusSkip
		item.Error = "没有 fc2_media"
		return item
	}

	if item.MovieName == "" {
		item.MovieName = strings.TrimSpace(row.FC2Key)
	}
	item.RootDir = strings.TrimSpace(row.RootDir)
	item.SourcePath = joinMediaPath(row.FullDir, row.FileName)
	if strings.TrimSpace(item.SourcePath) == "" {
		item.Error = "fc2_media 源路径为空"
		return item
	}
	if row.IsRemoved == int64(consts.FilmIsRemoved) {
		item.Status = scMediaMoveStatusSkip
		item.Error = "fc2_media 已删除"
		return item
	}

	root, layout, err := s.resolveScMediaLayoutForFC2Row(row)
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
		item.Error = "fc2_media 不在 media 目录"
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
	if entryMovieType(entry) == moviedomain.SehuatangThreadMovieTypeFC2PPV {
		return s.commitOneFC2ScMediaMove(ctx, entry, now)
	}
	return s.commitOneJAVScMediaMove(ctx, entry, now)
}

func (s *Service) commitOneJAVScMediaMove(ctx context.Context, entry *scMediaMovePlanEntry, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:     scMediaMoveStatusPass,
		MovieType:  moviedomain.SehuatangThreadMovieTypeJAV,
		MovieJavId: strings.TrimSpace(entryMovieJavID(entry)),
		MovieName:  strings.TrimSpace(entryMovieName(entry)),
		ScNames:    append([]string{}, entryScNames(entry)...),
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
	if err = s.syncJAVScMediaSnapshot(ctx, updated.MovieJavId); err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	s.invalidateMovieTypeCaches(ctx, updated.MovieJavId)

	item.TargetDir = updated.FullDir
	item.TargetPath = targetPath
	return item
}

func (s *Service) commitOneFC2ScMediaMove(ctx context.Context, entry *scMediaMovePlanEntry, now time.Time) *ScMediaMoveItem {
	item := &ScMediaMoveItem{
		Status:     scMediaMoveStatusPass,
		MovieType:  moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2Key:     strings.TrimSpace(entryFC2Key(entry)),
		MovieName:  strings.TrimSpace(entryMovieName(entry)),
		ScNames:    append([]string{}, entryScNames(entry)...),
		RootDir:    strings.TrimSpace(entryRootDir(entry)),
		SourcePath: strings.TrimSpace(entrySourcePath(entry)),
	}
	if item.FC2Key == "" {
		item.Status = scMediaMoveStatusFail
		item.Error = "计划中的 fc2_key 为空"
		return item
	}

	row, err := s.deps.FC2MediaModel.FindOneByKey(ctx, item.FC2Key)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			item.Status = scMediaMoveStatusFail
			item.Error = "fc2_media 不存在"
			return item
		}
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	if row == nil {
		item.Status = scMediaMoveStatusFail
		item.Error = "fc2_media 不存在"
		return item
	}

	currentPath := joinMediaPath(row.FullDir, row.FileName)
	item.SourcePath = currentPath
	if row.IsRemoved == int64(consts.FilmIsRemoved) {
		item.Status = scMediaMoveStatusFail
		item.Error = "fc2_media 已删除"
		return item
	}
	if expected := strings.TrimSpace(entrySourcePath(entry)); expected != "" && filepath.Clean(expected) != filepath.Clean(currentPath) {
		item.Status = scMediaMoveStatusFail
		item.Error = "fc2_media 路径已变化，请重跑预处理"
		return item
	}

	root, layout, err := s.resolveScMediaLayoutForFC2Row(row)
	if err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	item.RootDir = root

	currentDir := filepath.Clean(strings.TrimSpace(row.FullDir))
	if isPathWithin(currentDir, layout.watched) {
		item.Status = scMediaMoveStatusFail
		item.Error = "fc2_media 已在 watched 目录"
		return item
	}
	if !isPathWithin(currentDir, layout.media) {
		item.Status = scMediaMoveStatusFail
		item.Error = "fc2_media 不在 media 目录"
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

	updated := *row
	updated.RootDir = filepath.Clean(root)
	updated.FullDir = filepath.Clean(targetDir)
	updated.DirectoryID = folderID
	updated.UpdatedOn = now.Unix()
	if err = s.deps.FC2MediaModel.Update(ctx, &updated); err != nil {
		if rollbackErr := moveFileWithFallback(ctx, targetPath, currentPath); rollbackErr != nil {
			item.Status = scMediaMoveStatusFail
			item.Error = err.Error() + "; 回滚文件失败: " + rollbackErr.Error()
			return item
		}
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}
	if err = s.syncFC2MediaSnapshots(ctx, updated.FC2MovieID); err != nil {
		item.Status = scMediaMoveStatusFail
		item.Error = err.Error()
		return item
	}

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
	if row == nil {
		return "", rootLayout{}, fmt.Errorf("w_media 为空")
	}
	return s.resolveScMediaLayout(row.RootDir, row.FullDir, "w_media", s.mediaRoots())
}

func (s *Service) resolveScMediaLayoutForFC2Row(row *moviex.FC2Media) (string, rootLayout, error) {
	if row == nil {
		return "", rootLayout{}, fmt.Errorf("fc2_media 为空")
	}
	return s.resolveScMediaLayout(row.RootDir, row.FullDir, "fc2_media", s.fc2ScMediaRoots())
}

func (s *Service) resolveScMediaLayout(rootDir, fullDir, mediaLabel string, roots []string) (string, rootLayout, error) {
	mediaLabel = strings.TrimSpace(mediaLabel)
	if mediaLabel == "" {
		mediaLabel = "media"
	}

	cleanRoot := ""
	if rootDir = strings.TrimSpace(rootDir); rootDir != "" {
		cleanRoot = filepath.Clean(rootDir)
	}
	for _, root := range roots {
		if cleanRoot != "" && cleanRoot == filepath.Clean(root) {
			layout := buildRootLayout(root)
			return root, layout, nil
		}
	}

	cleanFullDir := ""
	if fullDir = strings.TrimSpace(fullDir); fullDir != "" {
		cleanFullDir = filepath.Clean(fullDir)
	}
	for _, root := range roots {
		layout := buildRootLayout(root)
		if isPathWithin(cleanFullDir, layout.media) || isPathWithin(cleanFullDir, layout.watched) {
			return root, layout, nil
		}
	}

	if cleanRoot != "" {
		return "", rootLayout{}, fmt.Errorf("%s root_dir 未配置: %s", mediaLabel, cleanRoot)
	}
	return "", rootLayout{}, fmt.Errorf("%s 路径不在已配置 root 中: %s", mediaLabel, cleanFullDir)
}

func (s *Service) fc2ScMediaRoots() []string {
	configured := s.fc2MediaRoots()
	roots := make([]string, 0, len(configured))
	for _, root := range configured {
		if strings.TrimSpace(root.rootDir) != "" {
			roots = append(roots, root.rootDir)
		}
	}
	return roots
}

func isPathWithin(pathValue, baseDir string) bool {
	pathValue = strings.TrimSpace(pathValue)
	baseDir = strings.TrimSpace(baseDir)
	if pathValue == "" || baseDir == "" {
		return false
	}
	pathValue = filepath.Clean(pathValue)
	baseDir = filepath.Clean(baseDir)
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

func entryMovieType(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return normalizeScMediaMoveMovieType(entry.MovieType, entry.MovieJavId, entry.FC2Key)
}

func entryFC2Key(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.FC2Key
}

func entryMovieName(entry *scMediaMovePlanEntry) string {
	if entry == nil {
		return ""
	}
	return entry.MovieName
}

func entryScNames(entry *scMediaMovePlanEntry) []string {
	if entry == nil {
		return nil
	}
	return entry.ScNames
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

func normalizeScMediaMoveMode(mode string) string {
	if strings.TrimSpace(mode) == scMediaMoveModeAll {
		return scMediaMoveModeAll
	}
	return scMediaMoveModeSingle
}

func normalizeScMediaMoveMovieType(movieType, movieJavID, fc2Key string) string {
	normalizedType := strings.TrimSpace(movieType)
	switch normalizedType {
	case moviedomain.SehuatangThreadMovieTypeJAV:
		return moviedomain.SehuatangThreadMovieTypeJAV
	case moviedomain.SehuatangThreadMovieTypeFC2PPV:
		return moviedomain.SehuatangThreadMovieTypeFC2PPV
	}
	if normalizedType != "" {
		return normalizedType
	}
	if strings.TrimSpace(fc2Key) != "" {
		return moviedomain.SehuatangThreadMovieTypeFC2PPV
	}
	if strings.TrimSpace(movieJavID) != "" {
		return moviedomain.SehuatangThreadMovieTypeJAV
	}
	return ""
}

func scMediaMoveScopeText(mode, scName string) string {
	if normalizeScMediaMoveMode(mode) == scMediaMoveModeAll {
		return "全部符合条件"
	}
	return strings.TrimSpace(scName)
}

func scMediaMoveAggFlowKey(mode string) string {
	if normalizeScMediaMoveMode(mode) == scMediaMoveModeAll {
		return "sc_media_move_all"
	}
	return "sc_media_move"
}

func appendScMediaMoveName(names []string, scName string) []string {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return names
	}
	for _, name := range names {
		if strings.TrimSpace(name) == scName {
			return names
		}
	}
	return append(names, scName)
}
