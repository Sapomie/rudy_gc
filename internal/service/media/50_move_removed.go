package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/pkg/filetool"
)

type MoveWMediaResult struct {
	MovieJavId string `json:"movie_jav_id"`
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Ok         bool   `json:"ok"`
	Error      string `json:"error"`
}

func (s *Service) MoveWMediaToRemoved(ctx context.Context, javId string) (*MoveWMediaResult, error) {
	result, row, updated, err := s.moveWMediaToRemoved(ctx, javId)
	if err != nil || result == nil || !result.Ok {
		return result, err
	}

	s.markMediaAggDirty(ctx, row, updated)
	s.invalidateMovieTypeCaches(ctx, javId)
	if err := s.rebuildMediaAggsAfterFlow(ctx, "move_removed"); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) MoveWMediaToRemovedDeferRefresh(ctx context.Context, javId string) (*MoveWMediaResult, error) {
	result, _, _, err := s.moveWMediaToRemoved(ctx, javId)
	return result, err
}

func (s *Service) moveWMediaToRemoved(ctx context.Context, javId string) (*MoveWMediaResult, *moviex.WMedia, *moviex.WMedia, error) {
	javId = strings.TrimSpace(javId)
	if javId == "" {
		return nil, nil, nil, fmt.Errorf("movie_jav_id 为空")
	}

	row, err := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, javId, consts.WMediaSourceNative)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return &MoveWMediaResult{MovieJavId: javId, Ok: false, Error: "w_media 不存在"}, nil, nil, nil
		}
		return nil, nil, nil, err
	}
	if row == nil {
		return &MoveWMediaResult{MovieJavId: javId, Ok: false, Error: "w_media 不存在"}, nil, nil, nil
	}

	rootDir := filepath.Clean(strings.TrimSpace(row.RootDir))
	if rootDir == "" {
		return &MoveWMediaResult{MovieJavId: javId, Ok: false, Error: "root_dir 为空"}, row, nil, nil
	}

	srcPath := joinMediaPath(row.FullDir, row.FileName)
	if strings.TrimSpace(srcPath) == "" {
		return &MoveWMediaResult{MovieJavId: javId, Ok: false, Error: "源路径为空"}, row, nil, nil
	}

	removedDir := buildRemovedDir(rootDir)
	if err := os.MkdirAll(removedDir, 0o755); err != nil {
		return &MoveWMediaResult{MovieJavId: javId, SourcePath: srcPath, Ok: false, Error: "创建 removed 目录失败: " + err.Error()}, row, nil, nil
	}

	targetPath, err := nextAvailableTargetPath(removedDir, filepath.Base(srcPath))
	if err != nil {
		return &MoveWMediaResult{MovieJavId: javId, SourcePath: srcPath, Ok: false, Error: err.Error()}, row, nil, nil
	}

	if err := moveFileWithFallback(ctx, srcPath, targetPath); err != nil {
		return &MoveWMediaResult{MovieJavId: javId, SourcePath: srcPath, TargetPath: targetPath, Ok: false, Error: err.Error()}, row, nil, nil
	}

	now := time.Now().Unix()
	updated := cloneWMedia(row)
	updated.IsRemoved = consts.FilmIsRemoved
	updated.RemoveTime = now
	updated.UpdatedOn = now

	if err := s.deps.WMediaModel.Update(ctx, updated); err != nil {
		return &MoveWMediaResult{MovieJavId: javId, SourcePath: srcPath, TargetPath: targetPath, Ok: false, Error: err.Error()}, row, updated, nil
	}

	return &MoveWMediaResult{
		MovieJavId: javId,
		SourcePath: srcPath,
		TargetPath: targetPath,
		Ok:         true,
	}, row, updated, nil
}

func (s *Service) FinalizeMoveRemovedBatch(ctx context.Context, javIds ...string) error {
	seen := make(map[string]struct{}, len(javIds))
	rows := make([]*moviex.WMedia, 0, len(javIds))
	dirtyJavIDs := make([]string, 0, len(javIds))
	for _, javId := range javIds {
		javId = strings.TrimSpace(javId)
		if javId == "" {
			continue
		}
		if _, ok := seen[javId]; ok {
			continue
		}
		seen[javId] = struct{}{}
		dirtyJavIDs = append(dirtyJavIDs, javId)

		row, err := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, javId, consts.WMediaSourceNative)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return err
		}
		if row != nil {
			rows = append(rows, row)
		}
	}

	if len(rows) > 0 {
		s.markMediaAggDirty(ctx, rows...)
	}
	if len(dirtyJavIDs) > 0 {
		s.invalidateMovieTypeCaches(ctx, dirtyJavIDs...)
	}
	if len(rows) == 0 && len(dirtyJavIDs) == 0 {
		return nil
	}
	return s.rebuildMediaAggsAfterFlow(ctx, "move_removed_batch")
}

func buildRemovedDir(rootDir string) string {
	rootDir = filepath.Clean(strings.TrimSpace(rootDir))
	if rootDir == "" {
		return ""
	}
	return filepath.Join(rootDir, "001_process", "005_removed")
}

func nextAvailableTargetPath(destDir, fileName string) (string, error) {
	destDir = filepath.Clean(strings.TrimSpace(destDir))
	fileName = strings.TrimSpace(fileName)
	if destDir == "" {
		return "", fmt.Errorf("目标目录为空")
	}
	if fileName == "" {
		return "", fmt.Errorf("目标文件名为空")
	}

	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)
	for i := 1; i <= 10000; i++ {
		candidate := fileName
		if i > 1 {
			candidate = fmt.Sprintf("%s_%d%s", base, i, ext)
		}
		path := filepath.Join(destDir, candidate)
		_, err := os.Stat(path)
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			return path, nil
		}
		return "", fmt.Errorf("检查目标文件失败: %w", err)
	}
	return "", fmt.Errorf("目标文件名冲突过多: %s", fileName)
}

func moveFileWithFallback(ctx context.Context, srcPath, targetPath string) error {
	if err := os.Rename(srcPath, targetPath); err == nil {
		return nil
	} else if errors.Is(err, syscall.EXDEV) {
		if copyErr := filetool.CopyFileWithProgressCtx(ctx, srcPath, targetPath); copyErr != nil {
			return copyErr
		}
		return os.Remove(srcPath)
	} else {
		return err
	}
}
