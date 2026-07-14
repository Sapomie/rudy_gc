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
	"rudy_gc/internal/repository/moviex"
)

type fc2MediaLeafBucketState struct {
	path     string
	folderID int64
	rows     []*moviex.FC2Media
}

func (s *Service) rebalanceFC2MediaLeafDirsForLayout(ctx context.Context, layout rootLayout) (*mediaDirRebalanceRootResult, error) {
	baseDir := filepath.Clean(strings.TrimSpace(layout.media))
	if baseDir == "" {
		return &mediaDirRebalanceRootResult{RootDir: layout.rootDir}, nil
	}
	if err := os.MkdirAll(baseDir, defaultFilePerm); err != nil {
		return nil, err
	}

	rows, err := s.deps.FC2MediaModel.ListByFullDirPrefixes(ctx, []string{baseDir})
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]*moviex.FC2Media)
	for _, row := range rows {
		if row == nil || row.IsRemoved == int64(consts.FilmIsRemoved) {
			continue
		}
		fullDir := filepath.Clean(strings.TrimSpace(row.FullDir))
		if !isPathWithin(fullDir, baseDir) {
			continue
		}
		grouped[fullDir] = append(grouped[fullDir], row)
	}

	existingBuckets, err := listAllDayBucketsUnderBase(baseDir)
	if err != nil {
		return nil, err
	}
	pathSet := make(map[string]struct{}, len(existingBuckets)+len(grouped))
	for _, bucket := range existingBuckets {
		pathSet[bucket.path] = struct{}{}
	}
	for path := range grouped {
		pathSet[path] = struct{}{}
	}
	bucketPaths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		path = filepath.Clean(strings.TrimSpace(path))
		if path != "" {
			bucketPaths = append(bucketPaths, path)
		}
	}
	sort.Strings(bucketPaths)

	buckets := make([]*fc2MediaLeafBucketState, 0, len(bucketPaths))
	for _, path := range bucketPaths {
		state := &fc2MediaLeafBucketState{path: path, rows: grouped[path]}
		sort.SliceStable(state.rows, func(i, j int) bool {
			if state.rows[i].BirthTime != state.rows[j].BirthTime {
				return state.rows[i].BirthTime < state.rows[j].BirthTime
			}
			if state.rows[i].FileName != state.rows[j].FileName {
				return state.rows[i].FileName < state.rows[j].FileName
			}
			return state.rows[i].ID < state.rows[j].ID
		})
		folder, findErr := s.deps.FC2FolderModel.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
		if findErr == nil && folder != nil {
			state.folderID = folder.ID
		} else if findErr != nil && !errors.Is(findErr, moviex.ErrNotFound) {
			return nil, findErr
		}
		buckets = append(buckets, state)
	}

	nowUnix := time.Now().Unix()
	result := &mediaDirRebalanceRootResult{RootDir: layout.rootDir}
	for index := 0; index < len(buckets); index++ {
		target := buckets[index]
		for len(target.rows) < maxFilesPerLeafDir {
			donorIndex := nextFC2DonorBucketIndex(buckets, index+1)
			if donorIndex < 0 {
				break
			}
			if target.folderID <= 0 {
				folder, err := ensureFC2FolderChain(ctx, s.deps.FC2FolderModel, target.path, nowUnix)
				if err != nil {
					return result, err
				}
				target.folderID = folder.ID
			}
			donor := buckets[donorIndex]
			row := donor.rows[0]
			updated, err := s.moveFC2MediaRowToDir(ctx, row, layout.rootDir, target.path, target.folderID, nowUnix)
			if err != nil {
				return result, err
			}
			donor.rows = donor.rows[1:]
			target.rows = append(target.rows, updated)
			result.Moved++
		}
	}

	removedLeafDirs, removedYearDirs, err := s.cleanupEmptyFC2MediaDirs(ctx, baseDir)
	if err != nil {
		return result, err
	}
	result.RemovedLeafDirs = removedLeafDirs
	result.RemovedYearDirs = removedYearDirs
	return result, nil
}

func nextFC2DonorBucketIndex(buckets []*fc2MediaLeafBucketState, start int) int {
	for index := start; index < len(buckets); index++ {
		if len(buckets[index].rows) > 0 {
			return index
		}
	}
	return -1
}

func (s *Service) moveFC2MediaRowToDir(ctx context.Context, row *moviex.FC2Media, rootDir, targetDir string, targetFolderID, nowUnix int64) (*moviex.FC2Media, error) {
	if row == nil {
		return nil, fmt.Errorf("nil fc2_media row")
	}
	sourcePath := joinMediaPath(row.FullDir, row.FileName)
	if strings.TrimSpace(sourcePath) == "" {
		return nil, fmt.Errorf("empty source path, fc2_code=%s", row.FC2Code)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file missing, fc2_code=%s, path=%s: %w", row.FC2Code, sourcePath, err)
	}
	if err := os.MkdirAll(targetDir, defaultFilePerm); err != nil {
		return nil, err
	}
	targetName, err := ensureUniqueFileName(targetDir, row.FileName)
	if err != nil {
		return nil, err
	}
	targetPath := filepath.Join(targetDir, targetName)
	if err := moveFileWithFallback(ctx, sourcePath, targetPath); err != nil {
		return nil, err
	}

	updated := *row
	updated.RootDir = filepath.Clean(rootDir)
	updated.FullDir = filepath.Clean(targetDir)
	updated.FileName = targetName
	updated.DirectoryID = targetFolderID
	updated.UpdatedOn = nowUnix
	if err := s.deps.FC2MediaModel.Update(ctx, &updated); err != nil {
		return nil, err
	}
	if err := s.syncFC2MediaSnapshots(ctx, updated.FC2MovieID); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *Service) cleanupEmptyFC2MediaDirs(ctx context.Context, baseDir string) (int, int, error) {
	removedLeafDirs := 0
	removedYearDirs := 0
	leafBuckets, err := listAllDayBucketsUnderBase(baseDir)
	if err != nil {
		return 0, 0, err
	}
	for index := len(leafBuckets) - 1; index >= 0; index-- {
		bucket := leafBuckets[index]
		entries, err := os.ReadDir(bucket.path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removedLeafDirs, removedYearDirs, err
		}
		if len(entries) > 0 {
			continue
		}
		if err := os.Remove(bucket.path); err != nil && !os.IsNotExist(err) {
			return removedLeafDirs, removedYearDirs, err
		}
		if err := s.deleteFC2FolderRowByPathIfExists(ctx, bucket.path); err != nil {
			return removedLeafDirs, removedYearDirs, err
		}
		removedLeafDirs++
	}
	yearBuckets, err := listYearBuckets(baseDir, "")
	if err != nil {
		if os.IsNotExist(err) {
			return removedLeafDirs, removedYearDirs, nil
		}
		return removedLeafDirs, removedYearDirs, err
	}
	for index := len(yearBuckets) - 1; index >= 0; index-- {
		bucket := yearBuckets[index]
		entries, err := os.ReadDir(bucket.path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removedLeafDirs, removedYearDirs, err
		}
		if len(entries) > 0 {
			continue
		}
		if err := os.Remove(bucket.path); err != nil && !os.IsNotExist(err) {
			return removedLeafDirs, removedYearDirs, err
		}
		if err := s.deleteFC2FolderRowByPathIfExists(ctx, bucket.path); err != nil {
			return removedLeafDirs, removedYearDirs, err
		}
		removedYearDirs++
	}
	return removedLeafDirs, removedYearDirs, nil
}

func (s *Service) deleteFC2FolderRowByPathIfExists(ctx context.Context, path string) error {
	row, err := s.deps.FC2FolderModel.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	if row == nil || row.ID <= 0 {
		return nil
	}
	return s.deps.FC2FolderModel.Delete(ctx, row.ID)
}
