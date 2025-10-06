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

func (r *DirectoryRepoSqlx) GetOrCreateChain(ctx context.Context, parts []string) (int64, error) {
	// 规范化 parts
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return 0, nil
	}

	now := time.Now().Unix()
	var parentID int64 = 0 // 根=0（not null 约定）

	for i, name := range clean {
		// 1) 先查 (parent_id, name)
		if row, err := r.m.FindOneByParentIdName(ctx, parentID, name); err == nil && row != nil {
			parentID = row.Id
			continue
		}

		// 2) 不存在则插入
		depth := int64(i + 1)
		path := "/" + strings.Join(clean[:i+1], "/")
		sum := md5.Sum([]byte(path))
		//pathHashHex := hex.EncodeToString(sum[:]) // 你的 PathHash 是 string

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
				continue
			}
			return 0, err
		}

		// 取回新ID（通过唯一键再查一次，避免依赖 LastInsertId）
		row3, err := r.m.FindOneByParentIdName(ctx, parentID, name)
		if err != nil || row3 == nil {
			return 0, err
		}
		parentID = row3.Id
	}

	return parentID, nil
}
