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
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/taskctx"
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
		if err := s.validateFC2PreparedReferences(ctx, prepared); err != nil {
			item := fc2CommitFailedItem(root, entry, err)
			result.Items = append(result.Items, item)
			taskctx.ReportProgress(ctx, taskctx.Progress{Stage: "fc2_media_commit_item_failed", Message: fmt.Sprintf("执行失败：%s，err=%s", filepath.Base(item.SourcePath), item.Error)})
			continue
		}
		item := s.processOneFC2IngestFile(ctx, layout, prepared, now)
		result.Items = append(result.Items, item)
		stage := "fc2_media_commit_item_done"
		if item.Error != "" {
			stage = "fc2_media_commit_item_failed"
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{Stage: stage, Message: fmt.Sprintf("执行插入：%s", filepath.Base(prepared.sourcePath))})
	}
	if err := clearFC2IngestPlan(layout); err != nil {
		return result, err
	}
	if _, err := s.rebalanceFC2MediaLeafDirsForLayout(ctx, layout); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) validateFC2PreparedReferences(ctx context.Context, prepared *fc2IngestPreparedItem) error {
	movie, err := s.deps.FC2MovieModel.FindOneByKey(ctx, prepared.meta.key)
	if err != nil {
		return fmt.Errorf("FC2影片资料复核失败: %w", err)
	}
	if movie.ID != prepared.fc2MovieID || strings.TrimSpace(movie.DataStatus) != "ready" {
		return fmt.Errorf("FC2影片资料在预处理后发生变化: %s", prepared.meta.key)
	}
	if err := s.ensureFC2MediaInsertable(ctx, prepared.fc2MovieID, prepared.meta.key); err != nil {
		return err
	}
	thread, err := s.deps.ShSehuatangThreadModel.FindOne(ctx, prepared.sourceThreadID)
	if err != nil {
		return fmt.Errorf("FC2来源帖子复核失败: %w", err)
	}
	if thread.FC2MovieID != prepared.fc2MovieID {
		return fmt.Errorf("FC2来源帖子影片关联在预处理后发生变化: %s", prepared.meta.key)
	}
	magnets, err := s.deps.ShMagnetModel.ListByMovieKey(ctx, "", prepared.meta.code)
	if err != nil {
		return fmt.Errorf("FC2磁力来源复核失败: %w", err)
	}
	for _, magnet := range magnets {
		if magnet != nil && strings.TrimSpace(magnet.ThreadTID) == strings.TrimSpace(thread.ThreadTID) && strings.EqualFold(strings.TrimSpace(magnet.InfoHash), prepared.sourceTorrentHash) {
			return nil
		}
	}
	return fmt.Errorf("FC2来源 Info Hash 在预处理后发生变化: %s", prepared.meta.key)
}

func (s *Service) ensureFC2MediaInsertable(ctx context.Context, movieID int64, fc2Key string) error {
	existing, err := s.deps.FC2MediaModel.FindOneByMovieID(ctx, movieID)
	if err == nil && existing != nil {
		return fmt.Errorf("FC2影片媒体已入库: %s，file=%s", fc2Key, existing.FileName)
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("FC2影片媒体状态查询失败: %w", err)
	}
	return nil
}

func (s *Service) processOneFC2IngestFile(ctx context.Context, layout rootLayout, prepared *fc2IngestPreparedItem, now time.Time) (item *FC2IngestFileItem) {
	if prepared == nil {
		return &FC2IngestFileItem{Status: ingestItemStatusFail, RootDir: layout.rootDir, Error: "nil ingest prepared item"}
	}
	item = &FC2IngestFileItem{
		Status:            ingestItemStatusPass,
		SourceDir:         prepared.root.sourceDir,
		RootDir:           prepared.root.rootDir,
		SourcePath:        prepared.sourcePath,
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
	targetDir, folderID, err := ensureFC2TargetDirectory(ctx, s.deps.FC2FolderModel, prepared.root.rootDir, now)
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
	insertedRow, err := s.insertFC2Media(ctx, row)
	if err != nil {
		processErr = err
		return item
	}
	item.FC2MediaID = insertedRow.ID
	holding = ""
	if err = s.moveFC2SourceItemToMediaAlbum(ctx, prepared.favoriteSource); err != nil {
		processErr = err
		return item
	}

	holding = ""
	return item
}

func (s *Service) insertFC2Media(ctx context.Context, row *moviex.FC2Media) (*moviex.FC2Media, error) {
	if row == nil {
		return nil, fmt.Errorf("nil fc2_media row")
	}
	if insertedRow, err := s.deps.FC2MediaModel.Insert(ctx, row); err == nil {
		if err := s.syncFC2MediaSnapshots(ctx, insertedRow.FC2MovieID); err != nil {
			return nil, err
		}
		if err := s.refreshSHMagnetDownloadStates(ctx, insertedRow.SourceTorrentHash); err != nil {
			return nil, err
		}
		return insertedRow, nil
	} else if isDuplicateEntryErr(err) {
		return nil, fmt.Errorf("FC2影片媒体已存在，拒绝覆盖: movie_id=%d: %w", row.FC2MovieID, err)
	} else {
		return nil, err
	}
}

func fc2CommitFailedItem(root fc2MediaRoot, entry *fc2IngestPlanEntry, err error) *FC2IngestFileItem {
	item := &FC2IngestFileItem{Status: ingestItemStatusFail, SourceDir: root.sourceDir, RootDir: root.rootDir, Error: err.Error()}
	if entry != nil {
		item.SourcePath = entry.SourcePath
		item.FC2Code = entry.FC2Code
		item.FC2Key = entry.FC2Key
		item.FC2MovieID = entry.FC2MovieID
		item.SourceThreadID = entry.SourceThreadID
		item.SourceTorrentHash = entry.SourceTorrentHash
	}
	return item
}
