package wfoldertree

import (
	"context"
	"crypto/md5"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

type Store interface {
	FindOneByPathSourceType(ctx context.Context, path string, sourceType int64) (*moviex.WFolder, error)
	Insert(ctx context.Context, data *moviex.WFolder) (sql.Result, error)
	Update(ctx context.Context, data *moviex.WFolder) error
	ListAllBySourceType(ctx context.Context, sourceType int64) ([]*moviex.WFolder, error)
}

func NormalizeAll(ctx context.Context, store Store, sourceType int64, nowUnix int64) error {
	rows, err := store.ListAllBySourceType(ctx, sourceType)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if row == nil {
			continue
		}
		if _, err := EnsurePathChain(ctx, store, sourceType, row.Path, nowUnix); err != nil {
			return err
		}
	}
	return nil
}

func EnsurePathChain(ctx context.Context, store Store, sourceType int64, fullPath string, nowUnix int64) (*moviex.WFolder, error) {
	parts := splitPathParts(fullPath)
	if len(parts) == 0 {
		return nil, nil
	}

	var (
		parentID int64
		leaf     *moviex.WFolder
	)
	for i, name := range parts {
		curPath := joinPath(parts[:i+1])
		row, err := store.FindOneByPathSourceType(ctx, curPath, sourceType)
		switch {
		case err == nil && row != nil:
			if normalized, changed := normalizeFolderRow(row, parentID, sourceType, int64(i+1), name, curPath, nowUnix); changed {
				if err := store.Update(ctx, normalized); err != nil {
					return nil, err
				}
				row = normalized
			}
		case errors.Is(err, moviex.ErrNotFound):
			insert := newFolderRow(parentID, sourceType, int64(i+1), name, curPath, nowUnix)
			if _, err := store.Insert(ctx, insert); err != nil {
				row, err = store.FindOneByPathSourceType(ctx, curPath, sourceType)
				if err != nil {
					return nil, err
				}
			} else {
				row, err = store.FindOneByPathSourceType(ctx, curPath, sourceType)
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, err
		}

		parentID = row.Id
		leaf = row
	}

	return leaf, nil
}

func normalizeFolderRow(row *moviex.WFolder, parentID, sourceType, depth int64, name, path string, nowUnix int64) (*moviex.WFolder, bool) {
	if row == nil {
		return nil, false
	}

	hash := folderPathHash(path)
	changed := false
	if row.ParentId != parentID {
		row.ParentId = parentID
		changed = true
	}
	if row.SourceType != sourceType {
		row.SourceType = sourceType
		changed = true
	}
	if row.Depth != depth {
		row.Depth = depth
		changed = true
	}
	if row.Name != name {
		row.Name = name
		changed = true
	}
	if row.Path != path {
		row.Path = path
		changed = true
	}
	if row.PathHash != hash {
		row.PathHash = hash
		changed = true
	}
	if row.CreatedOn <= 0 {
		row.CreatedOn = nowUnix
		changed = true
	}
	if changed {
		row.UpdatedOn = nowUnix
	}
	return row, changed
}

func newFolderRow(parentID, sourceType, depth int64, name, path string, nowUnix int64) *moviex.WFolder {
	return &moviex.WFolder{
		ParentId:   parentID,
		Name:       name,
		SourceType: sourceType,
		Depth:      depth,
		Path:       path,
		PathHash:   folderPathHash(path),
		CreatedOn:  nowUnix,
		UpdatedOn:  nowUnix,
	}
}

func folderPathHash(path string) string {
	sum := md5.Sum([]byte(path))
	return string(sum[:])
}

func splitPathParts(path string) []string {
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
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return string(filepath.Separator) + filepath.Join(parts...)
}
