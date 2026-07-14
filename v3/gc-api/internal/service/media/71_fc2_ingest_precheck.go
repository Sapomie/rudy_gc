package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	fc2domain "rudy_gc/internal/domain/fc2"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/taskctx"
)

func (s *Service) fc2IngestPrecheckRoot(ctx context.Context, root fc2MediaRoot, now time.Time) (*fc2IngestRootResult, error) {
	layout := buildRootLayout(root.rootDir)
	if err := ensureRootLayout(layout); err != nil {
		return nil, err
	}
	files, err := listFC2SourceFiles(root.sourceDir)
	if err != nil {
		return nil, err
	}
	result := &fc2IngestRootResult{Items: make([]*FC2IngestFileItem, 0, len(files))}
	preparedItems := make([]*fc2IngestPreparedItem, 0, len(files))
	seenMovieIDs := make(map[int64]string, len(files))
	if len(files) > 0 {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:       "fc2_media_precheck_start",
			Message:     fmt.Sprintf("开始预处理校验，共 %d 个文件", len(files)),
			QueuedCount: len(files),
		})
	}

	for index, sourcePath := range files {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		prepared, item := s.precheckOneFC2IngestFile(ctx, root, sourcePath, now)
		if prepared != nil && !reserveFC2BatchMovie(prepared, item, seenMovieIDs) {
			prepared = nil
		}
		result.Precheck.Total++
		if prepared == nil {
			result.Precheck.Failed++
			item.Status = ingestItemStatusFail
			result.Items = append(result.Items, item)
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:        "fc2_media_precheck_item_failed",
				Message:      fmt.Sprintf("预处理失败：%s，err=%s", filepath.Base(sourcePath), item.Error),
				HandledCount: index + 1,
				SuccessCount: result.Precheck.Passed,
				FailedCount:  result.Precheck.Failed,
				QueuedCount:  len(files) - (index + 1),
			})
			continue
		}
		result.Precheck.Passed++
		item.Status = ingestItemStatusPass
		result.Items = append(result.Items, item)
		preparedItems = append(preparedItems, prepared)
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "fc2_media_precheck_item_pass",
			Message:      fmt.Sprintf("预处理通过：%s", filepath.Base(sourcePath)),
			HandledCount: index + 1,
			SuccessCount: result.Precheck.Passed,
			FailedCount:  result.Precheck.Failed,
			QueuedCount:  len(files) - (index + 1),
		})
	}

	plan := buildFC2IngestPlan(root, preparedItems, result.Items)
	if err := saveFC2IngestPlan(layout, plan); err != nil {
		return result, err
	}
	result.GeneratedAt = plan.GeneratedAt
	return result, nil
}

func reserveFC2BatchMovie(prepared *fc2IngestPreparedItem, item *FC2IngestFileItem, seenMovieIDs map[int64]string) bool {
	if prepared == nil || item == nil || prepared.fc2MovieID <= 0 {
		return false
	}
	if firstSourcePath, exists := seenMovieIDs[prepared.fc2MovieID]; exists {
		item.Error = fmt.Sprintf("同批次 FC2影片重复: %s，首个文件=%s", prepared.meta.key, filepath.Base(firstSourcePath))
		return false
	}
	seenMovieIDs[prepared.fc2MovieID] = prepared.sourcePath
	return true
}

func listFC2SourceFiles(sourceDir string) ([]string, error) {
	return listProcessVideoFiles(sourceDir)
}

func (s *Service) precheckOneFC2IngestFile(ctx context.Context, root fc2MediaRoot, sourcePath string, now time.Time) (*fc2IngestPreparedItem, *FC2IngestFileItem) {
	item := &FC2IngestFileItem{
		SourceDir:  root.sourceDir,
		RootDir:    root.rootDir,
		SourcePath: sourcePath,
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		item.Error = fmt.Sprintf("源文件读取失败: %v", err)
		return nil, item
	}
	item.Size = info.Size()
	item.BirthTime = info.ModTime().Unix()

	meta, err := parseFC2RawMediaMeta(filepath.Base(sourcePath))
	if err != nil {
		item.Error = err.Error()
		return nil, item
	}
	item.FC2Code = meta.code
	item.FC2Key = meta.key

	resolved, err := s.resolveFC2DownloadSource(ctx, meta.code)
	if err != nil {
		item.Error = err.Error()
		return nil, item
	}
	movie := resolved.movie
	if movie == nil || movie.ID <= 0 {
		item.Error = "FC2影片资料为空"
		return nil, item
	}
	meta.key = strings.TrimSpace(movie.FC2Key)
	item.FC2Key = meta.key
	if strings.TrimSpace(movie.DataStatus) != "ready" {
		item.Error = fmt.Sprintf("FC2影片资料状态不可入库: %s", movie.DataStatus)
		return nil, item
	}
	if movie.MetadataThreadID <= 0 {
		item.Error = "FC2影片缺少资料来源帖子"
		return nil, item
	}
	item.FC2MovieID = movie.ID
	item.SourceThreadID = resolved.thread.Id
	if err := s.ensureFC2MediaInsertable(ctx, movie.ID, movie.FC2Key); err != nil {
		item.Error = err.Error()
		return nil, item
	}

	item.SourceTorrentHash = resolved.sourceTorrentHash

	hasMask, err := fc2MediaMaskValue(movie.MaskType)
	if err != nil {
		item.Error = err.Error()
		return nil, item
	}
	targetDir, err := previewTargetDirectory(buildRootLayout(root.rootDir), now)
	if err != nil {
		item.Error = fmt.Sprintf("FC2目标目录预览失败: %v", err)
		return nil, item
	}
	item.TargetDir = targetDir
	item.TargetFileName = buildFC2TargetFileName(meta, hasMask)
	item.TargetPath = filepath.Join(targetDir, item.TargetFileName)
	item.Alias = buildFC2MediaAlias(meta.key, item.BirthTime, item.Size)

	return &fc2IngestPreparedItem{
		root:              root,
		sourcePath:        sourcePath,
		meta:              meta,
		fc2MovieID:        movie.ID,
		sourceThreadID:    resolved.thread.Id,
		sourceTorrentHash: item.SourceTorrentHash,
		favoriteSource:    resolved.favoriteSource,
		hasMask:           hasMask,
	}, item
}

type fc2ResolvedDownloadSource struct {
	movie             *moviex.FC2Movie
	thread            *moviex.ShSehuatangThread
	favoriteSource    *favoriteAlbumSourceInfo
	sourceTorrentHash string
}

func (s *Service) resolveFC2DownloadSource(ctx context.Context, code string) (*fc2ResolvedDownloadSource, error) {
	album, err := s.deps.AlbumModel.FindOneByName(ctx, fc2SourceAlbumName)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, fmt.Errorf("FC2下载中相册不存在")
		}
		return nil, fmt.Errorf("FC2下载中相册查询失败: %w", err)
	}
	magnets, err := s.deps.ShMagnetModel.ListByMovieKey(ctx, "", code)
	if err != nil {
		return nil, fmt.Errorf("FC2磁力来源查询失败: %w", err)
	}
	magnetByID := make(map[int64]*moviex.ShMagnet, len(magnets))
	magnetIDs := make([]int64, 0, len(magnets))
	for _, magnet := range magnets {
		if magnet == nil || magnet.Id <= 0 {
			continue
		}
		magnetByID[magnet.Id] = magnet
		magnetIDs = append(magnetIDs, magnet.Id)
	}
	items, err := s.deps.AlbumItemModel.ListByAlbumIdSourceTypeSourceRowIDs(ctx, album.Id, fc2AlbumSourceType, magnetIDs)
	if err != nil {
		return nil, fmt.Errorf("FC2下载中来源查询失败: %w", err)
	}
	candidates := make([]*fc2ResolvedDownloadSource, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		magnet := magnetByID[item.SourceRowId]
		if magnet == nil {
			continue
		}
		hash := strings.ToUpper(strings.TrimSpace(magnet.InfoHash))
		if hash == "" || !strings.EqualFold(strings.TrimSpace(item.InfoHash), hash) {
			return nil, fmt.Errorf("FC2下载中相册 Info Hash 不一致: source_row_id=%d", item.SourceRowId)
		}
		thread, findErr := s.deps.ShSehuatangThreadModel.FindOneByThreadTID(ctx, magnet.ThreadTID)
		if findErr != nil {
			return nil, fmt.Errorf("FC2来源帖子查询失败: thread_tid=%s: %w", magnet.ThreadTID, findErr)
		}
		if thread == nil || thread.FC2MovieID <= 0 {
			return nil, fmt.Errorf("FC2来源帖子缺少影片关联: thread_tid=%s", magnet.ThreadTID)
		}
		movie, findErr := s.deps.FC2MovieModel.FindOne(ctx, thread.FC2MovieID)
		if findErr != nil {
			return nil, fmt.Errorf("FC2影片资料查询失败: movie_id=%d: %w", thread.FC2MovieID, findErr)
		}
		if movie == nil || !strings.EqualFold(strings.TrimSpace(movie.Code), strings.TrimSpace(code)) {
			return nil, fmt.Errorf("FC2来源影片番号不一致: thread_tid=%s", magnet.ThreadTID)
		}
		candidates = append(candidates, &fc2ResolvedDownloadSource{
			movie:             movie,
			thread:            thread,
			sourceTorrentHash: hash,
			favoriteSource: &favoriteAlbumSourceInfo{
				favoriteAlbumID: album.Id,
				item:            item,
				infoHash:        hash,
			},
		})
	}
	return requireUniqueFC2DownloadSource(code, candidates)
}

func requireUniqueFC2DownloadSource(code string, candidates []*fc2ResolvedDownloadSource) (*fc2ResolvedDownloadSource, error) {
	valid := make([]*fc2ResolvedDownloadSource, 0, len(candidates))
	keys := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.movie == nil || candidate.thread == nil || candidate.favoriteSource == nil || candidate.favoriteSource.item == nil {
			continue
		}
		valid = append(valid, candidate)
		keys = append(keys, strings.TrimSpace(candidate.movie.FC2Key))
	}
	switch len(valid) {
	case 0:
		return nil, fmt.Errorf("FC2下载中相册条目缺失: %s", code)
	case 1:
		return valid[0], nil
	default:
		return nil, fmt.Errorf("FC2下载中来源不唯一: %s，当前=%d，影片=%s", code, len(valid), strings.Join(keys, ","))
	}
}

func parseFC2RawMediaMeta(fileName string) (fc2RawMediaMeta, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	if _, ok := videoExts[ext]; !ok {
		return fc2RawMediaMeta{}, fmt.Errorf("不支持的视频扩展名: %s", ext)
	}
	base := strings.ToUpper(strings.TrimSuffix(fileName, filepath.Ext(fileName)))
	trimmed := trimMovieSuffix(base)
	code, _, ok := fc2domain.NormalizeCode(trimmed)
	if !ok {
		return fc2RawMediaMeta{}, fmt.Errorf("FC2文件名格式无效: %s", fileName)
	}
	return fc2RawMediaMeta{
		code:     code,
		ext:      ext,
		hasSub:   parseHasSub(base),
		selfMake: parseSelfMake(base),
	}, nil
}

func fc2MediaMaskValue(maskType string) (int64, error) {
	switch strings.TrimSpace(maskType) {
	case "无码":
		return int64(consts.FilmNoMosaic), nil
	case "有码":
		return int64(consts.FilmNotErased), nil
	default:
		return 0, fmt.Errorf("FC2有码状态无法识别: %s", maskType)
	}
}

func buildFC2TargetFileName(meta fc2RawMediaMeta, hasMask int64) string {
	return buildTargetFileName(rawMovieMeta{
		movieName: buildEncodedFC2MediaKey(meta.key),
		ext:       meta.ext,
		hasSub:    meta.hasSub,
		selfMake:  meta.selfMake,
		hasMask:   hasMask,
	})
}

func buildFC2MediaAlias(key string, birthTime, fileSize int64) string {
	return buildMediaAlias(rawMovieMeta{movieName: buildEncodedFC2MediaKey(key)}, birthTime, fileSize)
}

func buildEncodedFC2MediaKey(key string) string {
	_, code, number, codeSeq, ok := fc2domain.NormalizeKey(key)
	if !ok {
		return strings.ToUpper(strings.TrimSpace(key))
	}
	serial := fmt.Sprintf("%d", number)
	prefix := strings.TrimSuffix(code, "-"+serial)
	return fmt.Sprintf("%s-%d", encodeMediaAliasCode(prefix, serial), codeSeq)
}
