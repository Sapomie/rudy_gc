// internal/domain/film/a__service.go
package film

import (
	"context"
	"path/filepath"
	"strings"
)

type processFilmDirectorResponse struct {
	DirectoryID int64  // 叶子目录ID
	FileName    string // 仅文件名
	FilePath    string // 绝对全路径
	Dir1Id      int64
	Dir2Id      int64
	Dir3Id      int64
	Dir4Id      int64
}

func (s *Service) processFilmDirectory(ctx context.Context, fullPath string) (*processFilmDirectorResponse, error) {
	parts, file := extractParts(fullPath)

	// GetOrCreateChainWithLevels 返回从顶层到叶子的 ID 列表（例如 [root, ..., leaf]）
	levels, err := s.deps.DirectoryRepo.GetOrCreateChainWithLevels(ctx, parts)
	if err != nil {
		return nil, err
	}

	// 叶子目录（最接近文件的那层）
	var leafID int64
	if len(levels) > 0 {
		leafID = levels[len(levels)-1]
	}

	// 从尾部开始回填（不足则为 0）
	var dir1, dir2, dir3, dir4 int64
	n := len(levels)
	if n >= 1 {
		dir1 = levels[n-1] // 叶子
	}
	if n >= 2 {
		dir2 = levels[n-2]
	}
	if n >= 3 {
		dir3 = levels[n-3]
	}
	if n >= 4 {
		dir4 = levels[n-4]
	}

	return &processFilmDirectorResponse{
		DirectoryID: leafID,
		FileName:    file,
		FilePath:    fullPath,
		Dir1Id:      dir1, // 文件所在的最后一层目录（叶子）
		Dir2Id:      dir2, // 上层
		Dir3Id:      dir3, // 再上一层
		Dir4Id:      dir4, // 再上一层（靠近根）
	}, nil
}

func extractParts(fullPath string) ([]string, string) {
	fullPath = filepath.Clean(fullPath)

	dir := filepath.Dir(fullPath)
	file := filepath.Base(fullPath)

	// 去掉首个 "/" 再 split
	dir = strings.TrimPrefix(dir, string(filepath.Separator))
	parts := strings.Split(dir, string(filepath.Separator))

	// 过滤空字符串
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			res = append(res, p)
		}
	}
	return res, file
}
