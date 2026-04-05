package media

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) rescanOneScope(ctx context.Context, scope libraryRescanScope, nowUnix int64, result *LibraryRescanResult, seenIDs map[int64]struct{}) error {
	return filepath.WalkDir(scope.Path, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				Path:    path,
				RootDir: scope.Root,
				Error:   walkErr.Error(),
			})
			return nil
		}
		if d == nil || d.IsDir() {
			return nil
		}
		if shouldSkipIngestEntryName(d.Name()) || !isVideoName(d.Name()) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result.TotalFiles++
		fullDir := filepath.Dir(path)
		dirRow, err := ensureFolderChainForDir(ctx, s, fullDir, nowUnix)
		if err != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				FileName: filepath.Base(path),
				Path:     path,
				RootDir:  scope.Root,
				Error:    err.Error(),
			})
			return nil
		}

		row, matchBy, err := s.findExistingMediaForRescan(ctx, filepath.Base(path))
		if err != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				FileName: filepath.Base(path),
				Path:     path,
				RootDir:  scope.Root,
				Error:    err.Error(),
			})
			return nil
		}
		if row == nil {
			result.Unmatched++
			result.UnmatchedItems = append(result.UnmatchedItems, &LibraryRescanItem{
				FileName: filepath.Base(path),
				Path:     path,
				RootDir:  scope.Root,
			})
			return nil
		}

		result.Matched++
		seenIDs[row.Id] = struct{}{}

		before := buildRescanItemFromMedia(row)
		updated := cloneWMedia(row)
		locationChanged := filepath.Clean(strings.TrimSpace(updated.FullDir)) != filepath.Clean(fullDir) ||
			filepath.Clean(strings.TrimSpace(updated.RootDir)) != filepath.Clean(scope.Root) ||
			updated.DirectoryId != dirRow.Id ||
			strings.TrimSpace(updated.FileName) != filepath.Base(path)
		restored := updated.IsRemoved == consts.FilmIsRemoved || updated.RemoveTime != 0

		changed := updateRowLocation(updated, scope.Root, fullDir, filepath.Base(path), dirRow.Id, nowUnix)
		if !changed {
			result.Unchanged++
			return nil
		}

		if err := s.deps.WMediaModel.Update(ctx, updated); err != nil {
			result.Errors++
			result.ErrorItems = append(result.ErrorItems, &LibraryRescanItem{
				MovieName:    updated.MovieName,
				MovieJavId:   updated.MovieJavId,
				FileName:     filepath.Base(path),
				MatchBy:      matchBy,
				Path:         path,
				PreviousPath: before.Path,
				RootDir:      scope.Root,
				PreviousRoot: before.RootDir,
				Error:        err.Error(),
			})
			return nil
		}
		s.markMediaAggDirty(ctx, row, updated)
		s.invalidateMovieTypeCaches(ctx, updated.MovieJavId)

		if locationChanged {
			result.Moved++
			result.MovedItems = append(result.MovedItems, &LibraryRescanItem{
				MovieName:    updated.MovieName,
				MovieJavId:   updated.MovieJavId,
				FileName:     updated.FileName,
				MatchBy:      matchBy,
				Path:         path,
				PreviousPath: before.Path,
				RootDir:      scope.Root,
				PreviousRoot: before.RootDir,
			})
		}
		if restored {
			result.Restored++
			result.RestoredItems = append(result.RestoredItems, &LibraryRescanItem{
				MovieName:    updated.MovieName,
				MovieJavId:   updated.MovieJavId,
				FileName:     updated.FileName,
				MatchBy:      matchBy,
				Path:         path,
				PreviousPath: before.Path,
				RootDir:      scope.Root,
				PreviousRoot: before.RootDir,
			})
		}
		return nil
	})
}

func (s *Service) findExistingMediaForRescan(ctx context.Context, fileName string) (*moviex.WMedia, string, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, "", nil
	}

	row, err := s.deps.WMediaModel.FindOneByFileName(ctx, fileName)
	switch {
	case err == nil && row != nil:
		return row, "file_name", nil
	case err != nil && !errors.Is(err, moviex.ErrNotFound):
		return nil, "", err
	}

	candidates := s.buildRescanMovieCandidates(ctx, fileName)
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}

		row, err = s.deps.WMediaModel.FindOneByMovieJavId(ctx, candidate)
		switch {
		case err == nil && row != nil:
			return row, "movie_jav_id", nil
		case err != nil && !errors.Is(err, moviex.ErrNotFound):
			return nil, "", err
		}

		row, err = s.deps.WMediaModel.FindOneByMovieName(ctx, candidate)
		switch {
		case err == nil && row != nil:
			return row, "movie_name", nil
		case err != nil && !errors.Is(err, moviex.ErrNotFound):
			return nil, "", err
		}
	}

	return nil, "", nil
}

func (s *Service) buildRescanMovieCandidates(ctx context.Context, fileName string) []string {
	out := make([]string, 0, 4)
	seen := map[string]struct{}{}
	appendCandidate := func(value string) {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}

	if movieName, _, err := buildRollbackTargetFromFileName(fileName); err == nil {
		appendCandidate(movieName)
		if info, infoErr := s.findMovieInfoByName(ctx, movieName); infoErr == nil {
			appendCandidate(info.javID)
		}
	}

	if meta, err := parseRawMovieMeta(fileName); err == nil {
		appendCandidate(meta.movieName)
		if info, infoErr := s.findMovieInfoByName(ctx, meta.movieName); infoErr == nil {
			appendCandidate(info.javID)
		}
	}

	return out
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
