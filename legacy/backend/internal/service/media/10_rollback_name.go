package media

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/taskctx"
)

func (s *Service) RollbackName(ctx context.Context) (*IngestNewResult, error) {
	roots := s.mediaRoots()
	result := &IngestNewResult{
		Roots: len(roots),
		Items: make([]*IngestFileItem, 0, 64),
	}
	if len(roots) == 0 {
		return result, nil
	}

	for _, root := range roots {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   "media_rollback_root_begin",
			Message: "开始回滚 root: " + root,
		})

		rootItems, err := s.rollbackOneRoot(ctx, root)
		result.Total += len(rootItems)
		for _, item := range rootItems {
			if item.Error == "" {
				result.Success++
			} else {
				result.Failed++
			}
		}
		result.Items = append(result.Items, rootItems...)
		if err != nil {
			return result, err
		}
	}

	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "media_rollback_done",
		Message:      "rollback_name 处理完成",
		HandledCount: result.Total,
		SuccessCount: result.Success,
		FailedCount:  result.Failed,
	})
	return result, nil
}

func (s *Service) rollbackOneRoot(ctx context.Context, root string) ([]*IngestFileItem, error) {
	log := s.deps.Log.WithContext(ctx)

	layout := buildRootLayout(root)
	if err := ensureRootLayout(layout); err != nil {
		return nil, err
	}

	files, err := listRollbackFiles(layout)
	if err != nil {
		return nil, err
	}

	items := make([]*IngestFileItem, 0, len(files))
	for _, sourcePath := range files {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return items, err
		}
		select {
		case <-ctx.Done():
			return items, ctx.Err()
		default:
		}

		item := s.processOneRollbackFile(ctx, layout, sourcePath)
		items = append(items, item)

		stage := "media_rollback_item_done"
		message := fmt.Sprintf("回滚完成：%s -> %s", filepath.Base(sourcePath), filepath.Base(item.TargetPath))
		if item.Error != "" {
			stage = "media_rollback_item_failed"
			message = fmt.Sprintf("回滚失败：%s，err=%s", filepath.Base(sourcePath), item.Error)
			log.Errorf("media rollback failed: source=%s err=%s failed_path=%s", sourcePath, item.Error, item.FailedPath)
		}

		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:   stage,
			Message: message,
		})
	}
	return items, nil
}

func (s *Service) processOneRollbackFile(ctx context.Context, layout rootLayout, sourcePath string) (item *IngestFileItem) {
	item = &IngestFileItem{
		RootDir:    layout.rootDir,
		SourcePath: sourcePath,
	}

	var (
		err     error
		holding = sourcePath
	)
	defer func() {
		if err == nil {
			return
		}
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

	fileName := filepath.Base(sourcePath)
	movieName, targetName, err := buildRollbackTargetFromFileName(fileName)
	if err != nil {
		return item
	}
	item.MovieName = movieName

	targetPath, err := moveIntoDirWithNameUnique(sourcePath, layout.ingestNew, targetName)
	if err != nil {
		return item
	}

	holding = ""
	item.TargetPath = targetPath
	return item
}

func buildRollbackTargetFromFileName(fileName string) (movieName, targetName string, err error) {
	encoded := extractEncodedCodeToken(fileName)
	movieName, ok := decodeEncodedMovieCode(encoded)
	if !ok {
		return "", "", fmt.Errorf("cannot decode rollback file name: %s", fileName)
	}

	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	parts := strings.Split(base, "_")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid rollback file name: %s", fileName)
	}

	hasSub := int64(consts.FilmNoSub)
	selfMake := int64(consts.FilmNoSelfMake)
	hasMask := int64(consts.FilmNotErased)

	tokens := parts[1:]
	for i, raw := range tokens {
		token := strings.ToLower(strings.TrimSpace(raw))
		if token == "" {
			continue
		}

		switch token {
		case "sub":
			hasSub = consts.FilmHasSub
		case "self":
			selfMake = consts.FilmSelfMake
		case "era":
			if hasMask == consts.FilmNoMosaic {
				return "", "", fmt.Errorf("rollback token conflict (era/nomsk): %s", fileName)
			}
			hasMask = consts.FilmErased
		case "nomsk":
			if hasMask == consts.FilmErased {
				return "", "", fmt.Errorf("rollback token conflict (era/nomsk): %s", fileName)
			}
			hasMask = consts.FilmNoMosaic
		default:
			// 兼容冲突重命名产生的尾部序号，如 _001。
			if i == len(tokens)-1 && isAllDigits(token) {
				continue
			}
			return "", "", fmt.Errorf("unknown rollback token %q in %s", token, fileName)
		}
	}

	targetName, err = buildRollbackFileName(movieName, hasSub, selfMake, hasMask, ext)
	if err != nil {
		return "", "", err
	}
	return movieName, targetName, nil
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
