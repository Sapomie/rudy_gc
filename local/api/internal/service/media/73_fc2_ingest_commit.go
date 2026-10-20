package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/taskctx"
)

func (s *Service) fc2IngestCommitRoot(ctx context.Context, root fc2MediaRoot, now time.Time) (*fc2IngestRootResult, error) {
	layout := buildRootLayout(root.rootDir)
	if err := ensureRootLayout(layout); err != nil {
		return nil, err
	}
	plan, err := loadFC2IngestPlan(layout)
	if err != nil {
		if errors.Is(err, ErrFC2IngestPlanNotFound) {
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "fc2_media_commit_skip", Message: "没有预处理计划，跳过该 root"})
			return &fc2IngestRootResult{Items: []*FC2IngestFileItem{}}, nil
		}
		return nil, err
	}
	if filepath.Clean(plan.SourceDir) != filepath.Clean(root.sourceDir) || filepath.Clean(plan.RootDir) != filepath.Clean(root.rootDir) {
		return nil, fmt.Errorf("FC2预处理计划目录与当前配置不一致")
	}
	result := &fc2IngestRootResult{
		Precheck: IngestPrecheckStat{Total: plan.Total, Passed: plan.Passed, Failed: plan.Failed},
		Items:    make([]*FC2IngestFileItem, 0, len(plan.Entries)),
	}
	if len(plan.Entries) == 0 {
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "fc2_media_commit_skip", Message: "没有预处理通过的文件，请先执行第一段预处理"})
		return result, nil
	}
	for _, entry := range plan.Entries {
		if err = taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}
		prepared, err := buildFC2PreparedFromPlan(root, entry)
		if err != nil {
			item := fc2CommitFailedItem(root, entry, err)
			result.Items = append(result.Items, item)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "fc2_media_commit_item_failed", Message: fmt.Sprintf("执行失败：%s，err=%s", filepath.Base(item.SourcePath), item.Error)})
			continue
		}
		existing, err := s.validateFC2PreparedReferences(ctx, prepared)
		if err != nil {
			item := fc2CommitFailedItem(root, entry, err)
			result.Items = append(result.Items, item)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "fc2_media_commit_item_failed", Message: fmt.Sprintf("执行失败：%s，err=%s", filepath.Base(item.SourcePath), item.Error)})
			continue
		}
		item := s.processOneFC2IngestFile(ctx, layout, prepared, existing, now)
		result.Items = append(result.Items, item)
		stage := "fc2_media_commit_item_done"
		if item.Error != "" {
			stage = "fc2_media_commit_item_failed"
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: stage, Message: fmt.Sprintf("执行入库：%s", filepath.Base(prepared.sourcePath))})
	}
	if err := clearFC2IngestPlan(layout); err != nil {
		return result, err
	}
	if _, err := s.rebalanceFC2MediaLeafDirsForLayout(ctx, layout); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) validateFC2PreparedReferences(ctx context.Context, prepared *fc2IngestPreparedItem) (*moviex.FC2Media, error) {
	movie, err := s.deps.FC2MovieModel.FindOneByKey(ctx, prepared.meta.key)
	if err != nil {
		return nil, fmt.Errorf("FC2影片资料复核失败: %w", err)
	}
	if movie.ID != prepared.fc2MovieID || strings.TrimSpace(movie.DataStatus) != "ready" {
		return nil, fmt.Errorf("FC2影片资料在预处理后发生变化: %s", prepared.meta.key)
	}
	existing, err := s.validateFC2PreparedMediaIngestState(ctx, prepared)
	if err != nil {
		return nil, err
	}
	thread, err := s.deps.ShSehuatangThreadModel.FindOne(ctx, prepared.sourceThreadID)
	if err != nil {
		return nil, fmt.Errorf("FC2来源帖子复核失败: %w", err)
	}
	if thread.FC2MovieID != prepared.fc2MovieID {
		return nil, fmt.Errorf("FC2来源帖子影片关联在预处理后发生变化: %s", prepared.meta.key)
	}
	magnets, err := s.deps.ShMagnetModel.ListByMovieKey(ctx, "", prepared.meta.code)
	if err != nil {
		return nil, fmt.Errorf("FC2磁力来源复核失败: %w", err)
	}
	for _, magnet := range magnets {
		if magnet != nil && strings.TrimSpace(magnet.ThreadTID) == strings.TrimSpace(thread.ThreadTID) && strings.EqualFold(strings.TrimSpace(magnet.InfoHash), prepared.sourceTorrentHash) {
			return existing, nil
		}
	}
	return nil, fmt.Errorf("FC2来源 Info Hash 在预处理后发生变化: %s", prepared.meta.key)
}

func (s *Service) validateFC2PreparedMediaIngestState(ctx context.Context, prepared *fc2IngestPreparedItem) (*moviex.FC2Media, error) {
	if prepared == nil {
		return nil, fmt.Errorf("FC2预处理条目为空")
	}
	current, err := s.findExistingFC2MediaForIngest(ctx, prepared.fc2MovieID)
	if err != nil {
		return nil, err
	}

	switch prepared.action {
	case ingestActionInsert:
		if current != nil {
			return nil, fmt.Errorf("FC2预处理后媒体状态已变化，请重新预处理: %s", prepared.meta.key)
		}
		return nil, nil
	case ingestActionRestore:
		if current == nil || current.IsRemoved != int64(consts.FilmIsRemoved) || !fc2MediaSnapshotMatches(prepared.existing, current) {
			return nil, fmt.Errorf("FC2预处理后已移除媒体记录已变化，请重新预处理: %s", prepared.meta.key)
		}
		return current, nil
	case ingestActionReplace:
		if current == nil || current.IsRemoved == int64(consts.FilmIsRemoved) || !fc2MediaSnapshotMatches(prepared.existing, current) {
			return nil, fmt.Errorf("FC2预处理后活动媒体记录已变化，请重新预处理: %s", prepared.meta.key)
		}
		path := joinMediaPath(current.FullDir, current.FileName)
		info, statErr := os.Stat(path)
		if statErr != nil {
			return nil, fmt.Errorf("活动 FC2影片旧文件不可用: %s: %w", path, statErr)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("活动 FC2影片旧路径不是文件: %s", path)
		}
		return current, nil
	default:
		return nil, fmt.Errorf("FC2入库动作无效: %q", prepared.action)
	}
}

func (s *Service) processOneFC2IngestFile(ctx context.Context, layout rootLayout, prepared *fc2IngestPreparedItem, existing *moviex.FC2Media, now time.Time) (item *FC2IngestFileItem) {
	if prepared == nil {
		return &FC2IngestFileItem{Status: ingestItemStatusFail, RootDir: layout.rootDir, Error: "nil ingest prepared item"}
	}
	item = &FC2IngestFileItem{
		Status:            ingestItemStatusPass,
		Action:            prepared.action,
		Notice:            prepared.notice,
		SourceDir:         prepared.root.sourceDir,
		RootDir:           prepared.root.rootDir,
		SourcePath:        prepared.sourcePath,
		ExistingPath:      fc2MediaSnapshotPath(prepared.existing),
		FC2Code:           prepared.meta.code,
		FC2Key:            prepared.meta.key,
		FC2MovieID:        prepared.fc2MovieID,
		SourceThreadID:    prepared.sourceThreadID,
		SourceTorrentHash: prepared.sourceTorrentHash,
	}
	var (
		processErr error
		holding    string
	)
	defer func() {
		if processErr == nil {
			return
		}
		item.Status = ingestItemStatusFail
		item.Error = processErr.Error()
		if holding == "" {
			return
		}
		failedPath, moveErr := moveIntoDirUnique(holding, layout.failed)
		if moveErr != nil {
			item.Error += "; move_to_failed_err=" + moveErr.Error()
			return
		}
		item.FailedPath = failedPath
	}()

	movedToTmp, err := moveIntoDirUnique(prepared.sourcePath, layout.tmp)
	if err != nil {
		processErr = err
		return item
	}
	holding = movedToTmp

	videoMeta, err := probeVideoMeta(movedToTmp)
	if err != nil {
		processErr = err
		return item
	}
	targetDir, folderID, err := ensureFC2TargetDirectory(ctx, s.deps.WFolderModel, prepared.root.rootDir, now)
	if err != nil {
		processErr = err
		return item
	}
	fileName, err := ensureUniqueFileName(targetDir, buildFC2TargetFileName(prepared.meta, prepared.hasMask))
	if err != nil {
		processErr = err
		return item
	}
	targetPath := filepath.Join(targetDir, fileName)
	if err := os.Rename(movedToTmp, targetPath); err != nil {
		processErr = err
		return item
	}
	holding = targetPath

	stat, err := os.Stat(targetPath)
	if err != nil {
		processErr = err
		return item
	}
	item.TargetFileName = filepath.Base(targetPath)
	item.TargetDir = filepath.Dir(targetPath)
	item.TargetPath = targetPath
	item.Size = stat.Size()
	item.BirthTime = stat.ModTime().Unix()
	item.Alias = buildFC2MediaAlias(prepared.meta.key, item.BirthTime, item.Size)
	nowUnix := time.Now().Unix()

	row := &moviex.FC2Media{
		FC2MovieID:        prepared.fc2MovieID,
		FC2Key:            prepared.meta.key,
		FC2Code:           prepared.meta.code,
		SourceThreadID:    prepared.sourceThreadID,
		SourceTorrentHash: prepared.sourceTorrentHash,
		FileName:          item.TargetFileName,
		DirectoryID:       folderID,
		RootDir:           filepath.Clean(prepared.root.rootDir),
		FullDir:           filepath.Clean(item.TargetDir),
		Alias:             item.Alias,
		Size:              item.Size,
		Width:             videoMeta.width,
		Height:            videoMeta.height,
		BitRate:           videoMeta.bitRate,
		Duration:          videoMeta.duration,
		FrameAverage:      videoMeta.frameAverage,
		HasSub:            prepared.meta.hasSub,
		SelfMake:          prepared.meta.selfMake,
		HasMask:           prepared.hasMask,
		NeedScanMeta:      int64(consts.FilmMetaDataNoNeedScan),
		IsRemoved:         int64(consts.FilmIsNotRemoved),
		RemoveTime:        0,
		BirthTime:         item.BirthTime,
		CreatedOn:         nowUnix,
		UpdatedOn:         nowUnix,
	}
	previous, removedPath, err := s.persistPreparedFC2MediaIngest(ctx, prepared, existing, row)
	if err != nil {
		processErr = err
		return item
	}
	item.FC2MediaID = row.ID
	item.RemovedPath = removedPath
	holding = ""
	if finalizeErr := s.finalizeFC2MediaIngest(ctx, previous, row); finalizeErr != nil {
		item.Notice = appendIngestNotice(item.Notice, "FC2媒体已入库，后置状态刷新失败: "+finalizeErr.Error())
	}
	if albumErr := s.moveFC2SourceItemToMediaAlbum(ctx, prepared.favoriteSource); albumErr != nil {
		item.Notice = appendIngestNotice(item.Notice, "FC2媒体已入库，来源相册迁移失败: "+albumErr.Error())
	}

	return item
}

func moveActiveFC2MediaFileToRemoved(ctx context.Context, existing *moviex.FC2Media) (string, string, error) {
	if existing == nil {
		return "", "", fmt.Errorf("活动 FC2媒体记录为空")
	}
	originalPath := joinMediaPath(existing.FullDir, existing.FileName)
	if originalPath == "" {
		return "", "", fmt.Errorf("活动 FC2媒体旧文件路径为空: %s", existing.FC2Key)
	}
	info, err := os.Stat(originalPath)
	if err != nil {
		return "", "", fmt.Errorf("活动 FC2媒体旧文件不可用: %s: %w", originalPath, err)
	}
	if !info.Mode().IsRegular() {
		return "", "", fmt.Errorf("活动 FC2媒体旧路径不是文件: %s", originalPath)
	}

	removedDir := buildRemovedDir(existing.RootDir)
	if removedDir == "" {
		return "", "", fmt.Errorf("活动 FC2媒体 root_dir 为空: %s", existing.FC2Key)
	}
	if err = os.MkdirAll(removedDir, defaultFilePerm); err != nil {
		return "", "", fmt.Errorf("创建 FC2 removed 目录失败: %w", err)
	}
	removedPath, err := nextAvailableTargetPath(removedDir, filepath.Base(originalPath))
	if err != nil {
		return "", "", err
	}
	if err = moveFileWithFallback(ctx, originalPath, removedPath); err != nil {
		return "", "", fmt.Errorf("移动活动 FC2旧文件失败: %w", err)
	}
	return originalPath, removedPath, nil
}

func (s *Service) persistPreparedFC2MediaIngest(
	ctx context.Context,
	prepared *fc2IngestPreparedItem,
	existing *moviex.FC2Media,
	row *moviex.FC2Media,
) (*moviex.FC2Media, string, error) {
	if prepared == nil {
		return nil, "", fmt.Errorf("FC2预处理条目为空")
	}
	if prepared.action != ingestActionReplace {
		previous, err := s.persistFC2MediaIngest(ctx, row, prepared.existing)
		return previous, "", err
	}

	originalPath, removedPath, err := moveActiveFC2MediaFileToRemoved(ctx, existing)
	if err != nil {
		return nil, "", err
	}
	previous, err := s.persistFC2MediaIngest(ctx, row, prepared.existing)
	if err == nil {
		return previous, removedPath, nil
	}
	if rollbackErr := rollbackReplacedMediaFile(ctx, removedPath, originalPath); rollbackErr != nil {
		return nil, removedPath, fmt.Errorf("%w; rollback_old_fc2_media_err=%v", err, rollbackErr)
	}
	return nil, "", err
}

func (s *Service) persistFC2MediaIngest(ctx context.Context, row *moviex.FC2Media, expected *fc2IngestExistingMediaSnapshot) (*moviex.FC2Media, error) {
	if row == nil {
		return nil, fmt.Errorf("FC2媒体入库记录为空")
	}
	existing, err := s.findExistingFC2MediaForIngest(ctx, row.FC2MovieID)
	if err != nil {
		return nil, err
	}
	if expected == nil {
		if existing != nil {
			return nil, fmt.Errorf("FC2预处理后媒体状态已变化，请重新预处理: %s", row.FC2Key)
		}
		inserted, err := s.insertFC2Media(ctx, row)
		if err != nil {
			return nil, err
		}
		*row = *inserted
		return nil, nil
	}
	if existing == nil || !fc2MediaSnapshotMatches(expected, existing) {
		return nil, fmt.Errorf("FC2预处理后媒体记录已变化，请重新预处理: %s", row.FC2Key)
	}

	row.ID = existing.ID
	row.CreatedOn = existing.CreatedOn
	if err := s.deps.FC2MediaModel.Update(ctx, row); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) insertFC2Media(ctx context.Context, row *moviex.FC2Media) (*moviex.FC2Media, error) {
	if row == nil {
		return nil, fmt.Errorf("nil fc2_media row")
	}
	if insertedRow, err := s.deps.FC2MediaModel.Insert(ctx, row); err == nil {
		return insertedRow, nil
	} else if isDuplicateEntryErr(err) {
		return nil, fmt.Errorf("FC2影片媒体已存在，拒绝覆盖: movie_id=%d: %w", row.FC2MovieID, err)
	} else {
		return nil, err
	}
}

func (s *Service) finalizeFC2MediaIngest(ctx context.Context, previous, row *moviex.FC2Media) error {
	if row == nil {
		return fmt.Errorf("FC2媒体入库记录为空")
	}
	hashes := []string{row.SourceTorrentHash}
	if previous != nil {
		hashes = append([]string{previous.SourceTorrentHash}, hashes...)
	}
	var errs []error
	if err := s.refreshSHMagnetDownloadStates(ctx, hashes...); err != nil {
		errs = append(errs, fmt.Errorf("刷新 FC2磁力下载状态: %w", err))
	}
	if err := s.syncFC2MediaSnapshots(ctx, row.FC2MovieID); err != nil {
		errs = append(errs, fmt.Errorf("同步 FC2媒体快照: %w", err))
	}
	return errors.Join(errs...)
}

func fc2CommitFailedItem(root fc2MediaRoot, entry *fc2IngestPlanEntry, err error) *FC2IngestFileItem {
	item := &FC2IngestFileItem{Status: ingestItemStatusFail, SourceDir: root.sourceDir, RootDir: root.rootDir, Error: err.Error()}
	if entry != nil {
		item.Action = entry.Action
		item.Notice = entry.Notice
		item.SourcePath = entry.SourcePath
		item.ExistingPath = fc2MediaSnapshotPath(entry.Existing)
		item.FC2Code = entry.FC2Code
		item.FC2Key = entry.FC2Key
		item.FC2MovieID = entry.FC2MovieID
		item.SourceThreadID = entry.SourceThreadID
		item.SourceTorrentHash = entry.SourceTorrentHash
	}
	return item
}
