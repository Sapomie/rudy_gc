package sc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

func (s *Service) DeleteEvent(ctx context.Context, scName string) (*types.ScEventDeleteResult, error) {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return nil, types.ErrNotFound
	}

	now := time.Now().Unix()
	deleted, err := s.deps.ScEventWriteModel.DeleteEvent(ctx, scName, now)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		if deleted == nil || deleted.Event == nil {
			return nil, err
		}
		_, archiveErr := s.archiveDeletedScEventImage(deleted.Event)
		if archiveErr != nil {
			return buildScEventDeleteResult(deleted, false), fmt.Errorf("SC Event 已删除，缓存刷新和图片归档失败: %w", errors.Join(err, archiveErr))
		}
		return buildScEventDeleteResult(deleted, false), fmt.Errorf("SC Event 已删除，缓存刷新失败: %w", err)
	}
	if deleted == nil || deleted.Event == nil {
		return nil, types.ErrNotFound
	}

	javIDs, fc2IDs := splitGListMovieIdentities(deleted.Movies)
	postDeleteErrors := make([]error, 0, 5)
	if _, err := s.archiveDeletedScEventImage(deleted.Event); err != nil {
		postDeleteErrors = append(postDeleteErrors, fmt.Errorf("归档事件图片: %w", err))
	}
	if err := s.AddMovieAndCastScInfo(ctx, javIDs); err != nil {
		postDeleteErrors = append(postDeleteErrors, fmt.Errorf("刷新 JAV 影片及演员统计: %w", err))
	}
	if err := s.AddFC2MovieScInfo(ctx, fc2IDs); err != nil {
		postDeleteErrors = append(postDeleteErrors, fmt.Errorf("刷新 FC2 影片统计: %w", err))
	}
	if err := s.syncPersonStatsByIDs(ctx, now, deleted.PersonIDs...); err != nil {
		postDeleteErrors = append(postDeleteErrors, fmt.Errorf("刷新人物统计: %w", err))
	}
	if deleted.NextEvent != nil && strings.TrimSpace(deleted.NextEvent.Name) != "" {
		if err := s.refreshPersonScSnapshotsByScNames(ctx, now, deleted.NextEvent.Name); err != nil {
			postDeleteErrors = append(postDeleteErrors, fmt.Errorf("刷新下一场事件人物快照: %w", err))
		}
	}

	result := buildScEventDeleteResult(deleted, len(postDeleteErrors) == 0)
	if len(postDeleteErrors) > 0 {
		return result, fmt.Errorf("SC Event 已删除，后处理失败: %w", errors.Join(postDeleteErrors...))
	}
	return result, nil
}

func (s *Service) archiveDeletedScEventImage(event *moviex.GSc) (string, error) {
	if event == nil || strings.TrimSpace(event.ImagePath) == "" {
		return "", nil
	}
	if s == nil || s.deps == nil {
		return "", errors.New("SC Event 服务依赖为空")
	}

	activeDir := filepath.Clean(strings.TrimSpace(s.deps.Config.SCEvent.ActiveImageDir))
	deletedDir := filepath.Clean(strings.TrimSpace(s.deps.Config.SCEvent.DeletedImageDir))
	if activeDir == "" || activeDir == "." {
		return "", errors.New("sc_event.active_image_dir is empty")
	}
	if deletedDir == "" || deletedDir == "." {
		return "", errors.New("sc_event.deleted_image_dir is empty")
	}

	sourcePath := filepath.Clean(strings.TrimSpace(event.ImagePath))
	if filepath.Dir(sourcePath) != activeDir {
		return "", fmt.Errorf("SC Event 图片不在当前图片目录: %s", sourcePath)
	}
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return "", fmt.Errorf("读取 SC Event 图片失败: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("SC Event 图片不是普通文件: %s", sourcePath)
	}
	if err := os.MkdirAll(deletedDir, 0o744); err != nil {
		return "", fmt.Errorf("创建 SC Event 删除图片目录失败: %w", err)
	}

	targetPath, err := deletedScEventImagePath(deletedDir, sourcePath, event.Id)
	if err != nil {
		return "", err
	}
	companionSource := filepath.Join(activeDir, "._"+filepath.Base(sourcePath))
	companionTarget := filepath.Join(deletedDir, "._"+filepath.Base(targetPath))
	companionExists, err := regularFileExists(companionSource)
	if err != nil {
		return "", fmt.Errorf("读取 SC Event 图片元数据失败: %w", err)
	}
	if companionExists {
		if exists, existsErr := pathExists(companionTarget); existsErr != nil {
			return "", fmt.Errorf("检查 SC Event 图片元数据归档路径失败: %w", existsErr)
		} else if exists {
			return "", fmt.Errorf("SC Event 图片元数据归档路径已存在: %s", companionTarget)
		}
	}

	if err := os.Rename(sourcePath, targetPath); err != nil {
		return "", fmt.Errorf("归档 SC Event 图片失败: %w", err)
	}
	if companionExists {
		if stillExists, existsErr := pathExists(companionSource); existsErr != nil {
			return "", fmt.Errorf("检查 SC Event 图片元数据源路径失败: %w", existsErr)
		} else if stillExists {
			if renameErr := os.Rename(companionSource, companionTarget); renameErr != nil {
				rollbackErr := os.Rename(targetPath, sourcePath)
				if rollbackErr != nil {
					return "", fmt.Errorf("归档 SC Event 图片元数据失败: %w; 图片回滚失败: %v", renameErr, rollbackErr)
				}
				return "", fmt.Errorf("归档 SC Event 图片元数据失败: %w", renameErr)
			}
		}
	}
	return filepath.ToSlash(targetPath), nil
}

func deletedScEventImagePath(deletedDir, sourcePath string, eventID int64) (string, error) {
	baseName := filepath.Base(sourcePath)
	preferred := filepath.Join(deletedDir, baseName)
	exists, err := pathExists(preferred)
	if err != nil {
		return "", fmt.Errorf("检查 SC Event 图片归档路径失败: %w", err)
	}
	if !exists {
		return preferred, nil
	}
	if eventID <= 0 {
		return "", fmt.Errorf("SC Event 图片归档路径已存在且事件 ID 无效: %s", preferred)
	}
	archivedName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	archivedName += fmt.Sprintf("__gsc_%d%s", eventID, filepath.Ext(baseName))
	target := filepath.Join(deletedDir, archivedName)
	exists, err = pathExists(target)
	if err != nil {
		return "", fmt.Errorf("检查 SC Event 图片备用归档路径失败: %w", err)
	}
	if exists {
		return "", fmt.Errorf("SC Event 图片备用归档路径已存在: %s", target)
	}
	return target, nil
}

func regularFileExists(pathValue string) (bool, error) {
	info, err := os.Lstat(pathValue)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.Mode().IsRegular(), nil
}

func pathExists(pathValue string) (bool, error) {
	_, err := os.Lstat(pathValue)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func buildScEventDeleteResult(deleted *moviex.ScEventDeleteResult, derivedDataFresh bool) *types.ScEventDeleteResult {
	result := &types.ScEventDeleteResult{DerivedDataFresh: derivedDataFresh}
	if deleted == nil {
		return result
	}
	if deleted.Event != nil {
		result.Name = strings.TrimSpace(deleted.Event.Name)
	}
	result.DeletedMovieCount = int64(len(deleted.Movies))
	return result
}
