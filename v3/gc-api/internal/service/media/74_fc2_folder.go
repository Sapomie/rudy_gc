package media

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

func ensureFC2TargetDirectory(ctx context.Context, model moviex.FC2FolderModel, rootDir string, now time.Time) (string, int64, error) {
	layout := buildRootLayout(rootDir)
	targetDir, err := allocateTargetDirectoryPathUnder(layout.media, now)
	if err != nil {
		return "", 0, err
	}
	folder, err := ensureFC2FolderChain(ctx, model, targetDir, now.Unix())
	if err != nil {
		return "", 0, err
	}
	return targetDir, folder.ID, nil
}

func ensureFC2FolderChain(ctx context.Context, model moviex.FC2FolderModel, fullPath string, now int64) (*moviex.FC2Folder, error) {
	parts := splitFC2FolderPathParts(fullPath)
	if len(parts) == 0 {
		return nil, fmt.Errorf("FC2目录路径为空: %s", fullPath)
	}
	var (
		parentID int64
		leaf     *moviex.FC2Folder
	)
	for index, name := range parts {
		path := string(filepath.Separator) + filepath.Join(parts[:index+1]...)
		row, err := ensureFC2FolderRow(ctx, model, parentID, name, int64(index+1), path, now)
		if err != nil {
			return nil, err
		}
		parentID = row.ID
		leaf = row
	}
	return leaf, nil
}

func ensureFC2FolderRow(ctx context.Context, model moviex.FC2FolderModel, parentID int64, name string, depth int64, path string, now int64) (*moviex.FC2Folder, error) {
	path = filepath.Clean(path)
	existing, err := model.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative)
	if err == nil {
		hash := md5.Sum([]byte(path))
		changed := existing.ParentID != parentID || existing.Name != name || existing.SourceType != consts.WFolderSourceNative || existing.Depth != depth || existing.Path != path || string(existing.PathHash) != string(hash[:])
		if existing.CreatedOn <= 0 {
			existing.CreatedOn = now
			changed = true
		}
		if changed {
			existing.ParentID = parentID
			existing.Name = name
			existing.SourceType = consts.WFolderSourceNative
			existing.Depth = depth
			existing.Path = path
			existing.PathHash = hash[:]
			existing.UpdatedOn = now
			if err := model.Update(ctx, existing); err != nil {
				return nil, err
			}
		}
		return existing, nil
	}
	if !errors.Is(err, moviex.ErrNotFound) {
		return nil, err
	}
	hash := md5.Sum([]byte(path))
	created, err := model.Insert(ctx, &moviex.FC2Folder{
		ParentID:   parentID,
		Name:       name,
		SourceType: consts.WFolderSourceNative,
		Depth:      depth,
		Path:       path,
		PathHash:   hash[:],
		CreatedOn:  now,
		UpdatedOn:  now,
	})
	if err == nil {
		return created, nil
	}
	if existing, findErr := model.FindOneByPathSourceType(ctx, path, consts.WFolderSourceNative); findErr == nil {
		return existing, nil
	}
	return nil, err
}

func splitFC2FolderPathParts(path string) []string {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if cleaned == "" || cleaned == "." || cleaned == string(filepath.Separator) {
		return nil
	}
	cleaned = strings.TrimPrefix(cleaned, string(filepath.Separator))
	if cleaned == "" {
		return nil
	}
	raw := strings.Split(cleaned, string(filepath.Separator))
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
