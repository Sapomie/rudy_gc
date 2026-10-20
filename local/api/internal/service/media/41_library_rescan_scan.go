package media

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
)

func scanLibraryRescanScopes(ctx context.Context, scopes []libraryRescanScope, result *LibraryRescanResult) (map[string]libraryDiskFile, error) {
	files := make(map[string]libraryDiskFile)
	for _, scope := range scopes {
		err := filepath.WalkDir(scope.Path, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				result.Errors++
				libraryTypeStats(result, scope.MediaType).Errors++
				result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{MediaType: scope.MediaType, Path: path, RootDir: scope.Root, Error: walkErr.Error()})
				return nil
			}
			if entry == nil || entry.IsDir() || shouldSkipIngestEntryName(entry.Name()) || !isVideoName(entry.Name()) {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			info, err := entry.Info()
			if err != nil {
				result.Errors++
				libraryTypeStats(result, scope.MediaType).Errors++
				result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{MediaType: scope.MediaType, FileName: entry.Name(), Path: path, RootDir: scope.Root, Error: err.Error()})
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			cleanPath := filepath.Clean(path)
			files[libraryMediaPathKey(scope.MediaType, cleanPath)] = libraryDiskFile{
				MediaType: scope.MediaType,
				Path:      cleanPath,
				FileName:  entry.Name(),
				RootDir:   scope.Root,
				Size:      info.Size(),
			}
			return nil
		})
		if err != nil {
			return files, err
		}
	}
	return files, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
