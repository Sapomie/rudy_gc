// internal/infra/film_infra/directory_repo_sqlx.go
package film_infra

import (
	"context"
	"crypto/md5"
	"database/sql"
	"rudy_gc/internal/types"
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

	// 先完整拿到从顶层到叶子的全部 ID
	fullIDs := make([]int64, 0, len(clean))
	for i, name := range clean {
		// 1) 先查 (parent_id, name)
		if row, err := r.m.FindOneByParentIdName(ctx, parentID, name); err == nil && row != nil {
			parentID = row.Id
			fullIDs = append(fullIDs, parentID)
			continue
		}

		// 2) 不存在则插入
		depth := int64(i + 1)
		path := "/" + strings.Join(clean[:i+1], "/")
		sum := md5.Sum([]byte(path)) // BINARY(16)

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
				fullIDs = append(fullIDs, parentID)
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
		fullIDs = append(fullIDs, parentID)
	}

	// 只返回最后 4 层：右对齐到 levels
	// levels[3] = 叶子(dir1)，levels[2] = 父(dir2)，levels[1] = 再父(dir3)，levels[0] = 最上层(dir4)
	k := len(fullIDs)
	if k > 4 {
		k = 4
	}
	for i := 0; i < k; i++ {
		levels[4-1-i] = fullIDs[len(fullIDs)-1-i]
	}

	return levels, nil
}

// 省略已有结构：type DirectoryRepoSqlx struct { m moviex.VDirectoryModel }

func (r *DirectoryRepoSqlx) FindOneByID(ctx context.Context, id int64) (*types.Directory, error) {
	v, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return toTypesDirectory(v), nil
}

func (r *DirectoryRepoSqlx) FindOneByName(ctx context.Context, name string) (*types.Directory, error) {
	v, err := r.m.FindOneByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return toTypesDirectory(v), nil
}

func toTypesDirectory(v *moviex.VDirectory) *types.Directory {
	if v == nil {
		return nil
	}
	return &types.Directory{
		Id:        v.Id,
		ParentId:  v.ParentId,
		Name:      v.Name,
		Depth:     v.Depth,
		Path:      v.Path,
		CreatedOn: v.CreatedOn,
		UpdatedOn: v.UpdatedOn,
	}
}

func (r *DirectoryRepoSqlx) ListSubtreeIDs(ctx context.Context, id int64) ([]int64, error) {
	// 先查到这个目录，拿到它的 path
	d, err := r.m.FindOne(ctx, id)
	if err != nil {
		if err == moviex.ErrNotFound {
			return []int64{}, nil
		}
		return nil, err
	}
	if d == nil {
		return []int64{}, nil
	}

	// 再用 path 前缀在 modelx 层拉整棵子树
	ids, err := r.m.ListSubtreeIDsByPath(ctx, d.Path)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		// 至少返回自身
		return []int64{id}, nil
	}
	return ids, nil
}

// ====== 列表（根/子） ======
func (r *DirectoryRepoSqlx) ListRoots(ctx context.Context, page, size int) ([]*types.DirSummary, int64, error) {
	return r.listByParent(ctx, 0, page, size)
}

func (r *DirectoryRepoSqlx) ListChildren(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {
	return r.listByParent(ctx, parentID, page, size)
}

func (r *DirectoryRepoSqlx) listByParent(ctx context.Context, parentID int64, page, size int) ([]*types.DirSummary, int64, error) {

	rows, total, err := r.m.ListByParent(ctx, parentID, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*types.DirSummary, 0, len(rows))
	for _, it := range rows {
		out = append(out, &types.DirSummary{
			Id: it.Id, ParentId: it.ParentId, Name: it.Name, Depth: it.Depth, Path: it.Path, UpdatedOn: it.UpdatedOn,
		})
	}
	return out, total, nil
}

// ====== 同级 / 面包屑 ======
func (r *DirectoryRepoSqlx) ListSiblings(ctx context.Context, id int64) ([]*types.DirSummary, error) {
	cur, err := r.m.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return []*types.DirSummary{}, nil
	}
	rows, err := r.m.ListSiblings(ctx, cur.ParentId, 300)
	if err != nil {
		return nil, err
	}
	out := make([]*types.DirSummary, 0, len(rows))
	for _, it := range rows {
		out = append(out, &types.DirSummary{
			Id: it.Id, ParentId: it.ParentId, Name: it.Name, Depth: it.Depth, Path: it.Path, UpdatedOn: it.UpdatedOn,
		})
	}
	return out, nil
}

func (r *DirectoryRepoSqlx) BuildBreadcrumbs(ctx context.Context, id int64) ([]types.Breadcrumb, error) {
	cur, err := r.m.FindOne(ctx, id)
	if err != nil || cur == nil {
		return []types.Breadcrumb{}, err
	}
	// 通过 path 逐级查 (parent_id,name)
	parts := stringsSplitPath(cur.Path)
	var parentID int64
	var out []types.Breadcrumb
	for _, name := range parts {
		row, err := r.m.FindOneByParentIdName(ctx, parentID, name)
		if err != nil {
			return nil, err
		}
		out = append(out, types.Breadcrumb{Id: row.Id, Name: row.Name, Path: row.Path})
		parentID = row.Id
	}
	return out, nil
}

func stringsSplitPath(p string) []string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func nullI64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}
