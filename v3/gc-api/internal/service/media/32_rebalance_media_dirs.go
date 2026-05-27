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

type MediaDirRebalanceResult struct {
	Roots           int      `json:"roots"`
	ScannedRoots    []string `json:"scanned_roots"`
	Moved           int      `json:"moved"`
	RemovedLeafDirs int      `json:"removed_leaf_dirs"`
	RemovedYearDirs int      `json:"removed_year_dirs"`
}

type mediaDirRebalanceRootResult struct {
	RootDir         string
	Moved           int
	RemovedLeafDirs int
	RemovedYearDirs int
}

type mediaLeafBucketState struct {
	path     string
	folderID int64
	rows     []*moviex.WMedia
}

func (s *Service) RebalanceMediaLeafDirs(ctx context.Context) (*MediaDirRebalanceResult, error) {
	roots := s.mediaRoots()
	result := &MediaDirRebalanceResult{
		Roots:        len(roots),
		ScannedRoots: make([]string, 0, len(roots)),
	}
	for _, root := range roots {
		layout := buildRootLayout(root)
		rootResult, err := s.rebalanceMediaLeafDirsForLayout(ctx, layout)
		if err != nil {
			return result, err
		}
		result.ScannedRoots = append(result.ScannedRoots, root)
		result.Moved += rootResult.Moved
		result.RemovedLeafDirs += rootResult.RemovedLeafDirs
		result.RemovedYearDirs += rootResult.RemovedYearDirs
	}
	sort.Strings(result.ScannedRoots)
	return result, nil
}

func (s *Service) rebalanceMediaLeafDirsForLayout(ctx context.Context, layout rootLayout) (*mediaDirRebalanceRootResult, error) {
	baseDir := filepath.Clean(strings.TrimSpace(layout.media))
	if baseDir == "" {
		return &mediaDirRebalanceRootResult{RootDir: layout.rootDir}, nil
	}
	if err := os.MkdirAll(baseDir, defaultFilePerm); err != nil {
		return nil, err
	}

	rows, err := s.deps.WMediaModel.ListByFullDirPrefixes(ctx, []string{baseDir})
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]*moviex.WMedia)
	for _, row := range rows {
		if row == nil || row.SourceType != consts.WMediaSourceNative || row.IsRemoved == consts.FilmIsRemoved {
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
		if path == "" {
			continue
		}
		bucketPaths = append(bucketPaths, path)
	}
	sort.Strings(bucketPaths)

	buckets := make([]*mediaLeafBucketState, 0, len(bucketPaths))
	for _, path := range bucketPaths {
		state := &mediaLeafBucketState{
			path: path,
			rows: grouped[path],
		}
		sort.SliceStable(state.rows, func(i, j int) bool {
			if state.rows[i].BirthTime != state.rows[j].BirthTime {
				return state.rows[i].BirthTime < state.rows[j].BirthTime
			}
			if state.rows[i].FileName != state.rows[j].FileName {
				return state.rows[i].FileName < state.rows[j].FileName
			}
			return state.rows[i].Id < state.rows[j].Id
		})
		folderRow, findErr := s.deps.WFolderModel.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
		if findErr == nil && folderRow != nil {
			state.folderID = folderRow.Id
		} else if findErr != nil && !errors.Is(findErr, moviex.ErrNotFound) {
			return nil, findErr
		}
		buckets = append(buckets, state)
	}

	nowUnix := time.Now().Unix()
	invalidMovieIDs := make([]string, 0, len(rows))
	result := &mediaDirRebalanceRootResult{RootDir: layout.rootDir}

	for i := 0; i < len(buckets); i++ {
		target := buckets[i]
		for len(target.rows) < maxFilesPerLeafDir {
			donorIdx := nextDonorBucketIndex(buckets, i+1)
			if donorIdx < 0 {
				break
			}
			if target.folderID <= 0 {
				targetFolder, err := ensureFolderChainForDir(ctx, s, target.path, nowUnix)
				if err != nil {
					return result, err
				}
				target.folderID = targetFolder.Id
			}

			donor := buckets[donorIdx]
			row := donor.rows[0]
			updated, err := s.moveMediaRowToDir(ctx, row, layout.rootDir, target.path, target.folderID, nowUnix)
			if err != nil {
				return result, err
			}
			donor.rows = donor.rows[1:]
			target.rows = append(target.rows, updated)
			invalidMovieIDs = append(invalidMovieIDs, updated.MovieJavId)
			result.Moved++
		}
	}

	if len(invalidMovieIDs) > 0 {
		s.invalidateMovieTypeCaches(ctx, invalidMovieIDs...)
	}

	removedLeafDirs, removedYearDirs, err := s.cleanupEmptyMediaDirs(ctx, baseDir)
	if err != nil {
		return result, err
	}
	result.RemovedLeafDirs = removedLeafDirs
	result.RemovedYearDirs = removedYearDirs
	return result, nil
}

func nextDonorBucketIndex(buckets []*mediaLeafBucketState, start int) int {
	for i := start; i < len(buckets); i++ {
		if len(buckets[i].rows) > 0 {
			return i
		}
	}
	return -1
}

func (s *Service) moveMediaRowToDir(ctx context.Context, row *moviex.WMedia, rootDir, targetDir string, targetFolderID, nowUnix int64) (*moviex.WMedia, error) {
	if row == nil {
		return nil, fmt.Errorf("nil w_media row")
	}
	sourcePath := joinMediaPath(row.FullDir, row.FileName)
	if strings.TrimSpace(sourcePath) == "" {
		return nil, fmt.Errorf("empty source path, movie_jav_id=%s", row.MovieJavId)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source file missing, movie_jav_id=%s, path=%s: %w", row.MovieJavId, sourcePath, err)
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

	updated := cloneWMedia(row)
	updateRowLocation(updated, rootDir, targetDir, targetName, targetFolderID, nowUnix)
	if err := s.deps.WMediaModel.Update(ctx, updated); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) cleanupEmptyMediaDirs(ctx context.Context, baseDir string) (int, int, error) {
	removedLeafDirs := 0
	removedYearDirs := 0

	leafBuckets, err := listAllDayBucketsUnderBase(baseDir)
	if err != nil {
		return 0, 0, err
	}
	for i := len(leafBuckets) - 1; i >= 0; i-- {
		bucket := leafBuckets[i]
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
		if err := s.deleteFolderRowByPathIfExists(ctx, bucket.path); err != nil {
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
	for i := len(yearBuckets) - 1; i >= 0; i-- {
		bucket := yearBuckets[i]
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
		if err := s.deleteFolderRowByPathIfExists(ctx, bucket.path); err != nil {
			return removedLeafDirs, removedYearDirs, err
		}
		removedYearDirs++
	}

	return removedLeafDirs, removedYearDirs, nil
}

func (s *Service) deleteFolderRowByPathIfExists(ctx context.Context, path string) error {
	row, err := s.deps.WFolderModel.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return err
	}
	if row == nil || row.Id <= 0 {
		return nil
	}
	return s.deps.WFolderModel.Delete(ctx, row.Id)
}
