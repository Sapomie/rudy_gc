// internal/infra/film_infra/directory_repo_sqlx.go
package film_infra

import (
	"context"
	"crypto/md5"
	"strings"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/repo/film_repo"
)

var _ film_repo.DirectoryRepo = (*DirectoryRepoSqlx)(nil)

type DirectoryRepoSqlx struct {
	m moviex.VDirectoryModel
}

func NewDirectoryRepoSqlx(m moviex.VDirectoryModel) *DirectoryRepoSqlx {
	return &DirectoryRepoSqlx{m: m}
}

func (r *DirectoryRepoSqlx) GetOrCreateChainWithLevels(ctx context.Context, parts []string) ([4]int64, error) {
	// 规范化 parts
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			clean = append(clean, s)
		}
	}

	var levels [4]int64
	if len(clean) == 0 {
		return levels, nil
	}

	now := time.Now().Unix()
	var parentID int64 = 0 // 根=0（NOT NULL 约定）

	for i, name := range clean {
		// 1) 先查 (parent_id, name)
		if row, err := r.m.FindOneByParentIdName(ctx, parentID, name); err == nil && row != nil {
			parentID = row.Id
			if i < 4 {
				levels[i] = parentID
			}
			continue
		}

		// 2) 不存在则插入
		depth := int64(i + 1)
		path := "/" + strings.Join(clean[:i+1], "/")
		sum := md5.Sum([]byte(path)) // BINARY(16)：存原始 16 字节

		vd := &moviex.VDirectory{
			ParentId:  parentID,
			Name:      name,
			Depth:     depth,
			Path:      path,
			PathHash:  string(sum[:]),
			CreatedOn: now,
			UpdatedOn: now,
		}

		if _, err := r.m.Insert(ctx, vd); err != nil {
			// 并发兜底：再查一次
			if row2, err2 := r.m.FindOneByParentIdName(ctx, parentID, name); err2 == nil && row2 != nil {
				parentID = row2.Id
				if i < 4 {
					levels[i] = parentID
				}
				continue
			}
			return levels, err
		}

		// 3) 回读拿到 id（避免依赖 LastInsertId）
		row3, err := r.m.FindOneByParentIdName(ctx, parentID, name)
		if err != nil || row3 == nil {
			return levels, err
		}
		parentID = row3.Id
		if i < 4 {
			levels[i] = parentID
		}
	}

	return levels, nil
}
