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
	"rudy_gc/internal/taskctx"
)

type ingestRootResult struct {
	Precheck IngestPrecheckStat
	Items    []*IngestFileItem
}

func (s *Service) ingestPrecheckOneRoot(ctx context.Context, root string) (*ingestRootResult, error) {
	layout := buildRootLayout(root)
	if err := ensureRootLayout(layout); err != nil {
		return nil, err
	}

	files, err := listIngestNewFiles(layout)
	if err != nil {
		return nil, err
	}

	result := &ingestRootResult{
		Items: make([]*IngestFileItem, 0, len(files)),
	}

	if len(files) == 0 {
		emptyPlan := &ingestPrecheckPlan{
			Version:     ingestPrecheckPlanVersion,
			RootDir:     layout.rootDir,
			GeneratedAt: time.Now().Unix(),
			Entries:     []*ingestPrecheckPlanEntry{},
			Checks:      []*ingestPrecheckPlanCheck{},
		}
		if err = s.saveIngestPrecheckPlan(layout, emptyPlan); err != nil {
			return result, err
		}
		return result, nil
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:       "media_precheck_start",
		Message:     fmt.Sprintf("开始预处理校验，共 %d 个文件", len(files)),
		QueuedCount: len(files),
	})

	decisions, err := s.precheckIngestFiles(ctx, layout, files)
	passPrepared := make([]*ingestPreparedItem, 0, len(decisions))
	now := time.Now()

	for _, decision := range decisions {
		if decision == nil {
			continue
		}

		result.Precheck.Total++
		if decision.prepared != nil {
			previewItem, previewErr := buildPrecheckPreviewItem(layout, decision.prepared, now)
			if previewErr != nil {
				result.Precheck.Failed++
				result.Items = append(result.Items, &IngestFileItem{
					Status:     ingestItemStatusFail,
					RootDir:    layout.rootDir,
					SourcePath: decision.prepared.sourcePath,
					MovieName:  decision.prepared.meta.movieName,
					Error:      "precheck preview build failed: " + previewErr.Error(),
				})
				continue
			}

			result.Precheck.Passed++
			result.Items = append(result.Items, previewItem)
			passPrepared = append(passPrepared, decision.prepared)
			continue
		}

		result.Precheck.Failed++
		if decision.failedItem != nil {
			decision.failedItem.Status = ingestItemStatusFail
			result.Items = append(result.Items, decision.failedItem)
			continue
		}
		result.Items = append(result.Items, &IngestFileItem{
			Status:  ingestItemStatusFail,
			RootDir: layout.rootDir,
			Error:   "precheck failed",
		})
	}

	plan := buildIngestPrecheckPlan(layout, passPrepared, result.Items)
	if saveErr := s.saveIngestPrecheckPlan(layout, plan); saveErr != nil {
		return result, saveErr
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) ingestCommitOneRoot(ctx context.Context, root string, now time.Time) (*ingestRootResult, error) {
	layout := buildRootLayout(root)
	if err := ensureRootLayout(layout); err != nil {
		return nil, err
	}

	plan, err := s.loadIngestPrecheckPlan(layout)
	if err != nil {
		if errors.Is(err, ErrIngestPrecheckPlanNotFound) {
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:   "media_commit_skip",
				Message: "没有预处理计划，跳过该 root",
			})
			return &ingestRootResult{
				Precheck: IngestPrecheckStat{},
				Items:    []*IngestFileItem{},
			}, nil
		}
		return nil, err
	}

	result := &ingestRootResult{
		Precheck: IngestPrecheckStat{
			Total:  plan.Total,
			Passed: plan.Passed,
			Failed: plan.Failed,
		},
		Items: make([]*IngestFileItem, 0, len(plan.Entries)),
	}

	if len(plan.Entries) == 0 {
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "media_commit_skip",
			Message: "没有预处理通过的文件，请先执行第一段预处理",
		})
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

		prepared, buildErr := buildPreparedFromPlanEntry(entry)
		if buildErr != nil {
			item := &IngestFileItem{
				Status:     ingestItemStatusFail,
				RootDir:    layout.rootDir,
				SourcePath: safePlanSourcePath(entry),
				MovieName:  safePlanMovieName(entry),
				Error:      buildErr.Error(),
			}
			result.Items = append(result.Items, item)
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:   "media_commit_item_failed",
				Message: fmt.Sprintf("执行失败：%s，err=%s", filepath.Base(item.SourcePath), item.Error),
			})
			continue
		}

		item := s.processOneIngestFile(ctx, layout, prepared, now)
		result.Items = append(result.Items, item)

		stage := "media_commit_item_done"
		if item.Error != "" {
			stage = "media_commit_item_failed"
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   stage,
			Message: fmt.Sprintf("执行插入：%s", filepath.Base(prepared.sourcePath)),
		})
	}

	if clearErr := s.clearIngestPrecheckPlan(layout); clearErr != nil {
		return result, clearErr
	}
	return result, nil
}

func (s *Service) processOneIngestFile(ctx context.Context, layout rootLayout, prepared *ingestPreparedItem, now time.Time) (item *IngestFileItem) {
	if prepared == nil {
		return &IngestFileItem{
			Status:  ingestItemStatusFail,
			RootDir: layout.rootDir,
			Error:   "nil ingest prepared item",
		}
	}

	sourcePath := prepared.sourcePath
	item = &IngestFileItem{
		Status:     ingestItemStatusPass,
		RootDir:    layout.rootDir,
		SourcePath: sourcePath,
		MovieName:  prepared.meta.movieName,
	}

	var (
		err        error
		holding    string
		movedToTmp string
	)

	defer func() {
		if err == nil {
			return
		}
		item.Status = ingestItemStatusFail
		item.Error = err.Error()
		if holding == "" {
			return
		}
		failedPath, moveErr := moveIntoDirUnique(holding, layout.failed)
		if moveErr != nil {
			item.Error = item.Error + "; move_to_failed_err=" + moveErr.Error()
			return
		}
		item.FailedPath = failedPath
	}()

	movedToTmp, err = moveIntoDirUnique(sourcePath, layout.tmp)
	if err != nil {
		return item
	}
	holding = movedToTmp

	videoMeta, err := probeVideoMeta(movedToTmp)
	if err != nil {
		return item
	}

	targetDir, folderID, err := s.allocateTargetDirectory(ctx, layout, now)
	if err != nil {
		return item
	}

	finalName := buildTargetFileName(prepared.meta)
	finalName, err = ensureUniqueFileName(targetDir, finalName)
	if err != nil {
		return item
	}
	targetPath := filepath.Join(targetDir, finalName)

	if err = os.Rename(movedToTmp, targetPath); err != nil {
		return item
	}
	holding = targetPath
	item.TargetFileName = filepath.Base(targetPath)
	item.TargetDir = filepath.Dir(targetPath)
	item.TargetPath = targetPath

	stat, err := os.Stat(targetPath)
	if err != nil {
		return item
	}

	birthTime := stat.ModTime().Unix()
	item.BirthTime = birthTime
	item.Size = stat.Size()

	row := s.buildMediaRow(mediaRowInput{
		MovieInfo:       prepared.movieInfo,
		MovieName:       prepared.meta.movieName,
		FileName:        filepath.Base(targetPath),
		RootDir:         layout.rootDir,
		FullDir:         filepath.Dir(targetPath),
		DirectoryID:     folderID,
		Alias:           buildMediaAlias(prepared.meta, birthTime, stat.Size()),
		Size:            stat.Size(),
		VideoMeta:       videoMeta,
		NeedScanMeta:    int64(consts.FilmMetaDataNoNeedScan),
		HasSub:          prepared.meta.hasSub,
		SelfMake:        prepared.meta.selfMake,
		HasMask:         prepared.meta.hasMask,
		BirthTime:       birthTime,
		SourceTorrentID: prepared.favoriteSource.infoHash,
		NowUnix:         time.Now().Unix(),
	})
	item.Alias = row.Alias
	item.SourceTorrentHash = row.SourceTorrentHash

	if err = s.upsertMedia(ctx, row); err != nil {
		return item
	}

	holding = ""
	if err = s.moveFavoriteItemToMediaAlbum(ctx, prepared.favoriteSource); err != nil {
		return item
	}

	holding = ""
	return item
}

func buildPrecheckPreviewItem(layout rootLayout, prepared *ingestPreparedItem, now time.Time) (*IngestFileItem, error) {
	if prepared == nil {
		return nil, fmt.Errorf("nil prepared item")
	}
	info, err := os.Stat(prepared.sourcePath)
	if err != nil {
		return nil, err
	}
	targetDir, err := previewTargetDirectory(layout, now)
	if err != nil {
		return nil, err
	}
	birthTime := info.ModTime().Unix()
	sourceHash := ""
	if prepared.favoriteSource != nil {
		sourceHash = strings.TrimSpace(prepared.favoriteSource.infoHash)
	}
	return &IngestFileItem{
		Status:            ingestItemStatusPass,
		RootDir:           layout.rootDir,
		SourcePath:        prepared.sourcePath,
		MovieName:         prepared.meta.movieName,
		TargetFileName:    buildTargetFileName(prepared.meta),
		TargetDir:         targetDir,
		Alias:             buildMediaAlias(prepared.meta, birthTime, info.Size()),
		SourceTorrentHash: sourceHash,
		Size:              info.Size(),
		BirthTime:         birthTime,
	}, nil
}

func moveIntoDirUnique(sourcePath, targetDir string) (string, error) {
	if err := os.MkdirAll(targetDir, defaultFilePerm); err != nil {
		return "", err
	}

	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 0; i < 1000; i++ {
		name := base
		if i > 0 {
			name = fmt.Sprintf("%s_%03d%s", stem, i, ext)
		}
		targetPath := filepath.Join(targetDir, name)
		if _, err := os.Stat(targetPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if err := os.Rename(sourcePath, targetPath); err != nil {
			return "", err
		}
		return targetPath, nil
	}
	return "", fmt.Errorf("cannot move file, too many name conflicts: %s", sourcePath)
}

func moveIntoDirWithNameUnique(sourcePath, targetDir, targetName string) (string, error) {
	if err := os.MkdirAll(targetDir, defaultFilePerm); err != nil {
		return "", err
	}

	ext := filepath.Ext(targetName)
	stem := strings.TrimSuffix(targetName, ext)
	for i := 0; i < 1000; i++ {
		name := targetName
		if i > 0 {
			name = fmt.Sprintf("%s_%03d%s", stem, i, ext)
		}
		targetPath := filepath.Join(targetDir, name)
		if _, err := os.Stat(targetPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if err := os.Rename(sourcePath, targetPath); err != nil {
			return "", err
		}
		return targetPath, nil
	}
	return "", fmt.Errorf("cannot move file with target name, too many conflicts: %s", sourcePath)
}
