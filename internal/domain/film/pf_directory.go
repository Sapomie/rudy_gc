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

	// 逐级创建/获取目录链（只拿前4层ID）
	levels, err := s.deps.DirectoryRepo.GetOrCreateChainWithLevels(ctx, parts)
	if err != nil {
		return nil, err
	}

	// 叶子目录ID：levels 中最后一个非 0
	var leafID int64
	for i := len(levels) - 1; i >= 0; i-- {
		if levels[i] != 0 {
			leafID = levels[i]
			break
		}
	}

	return &processFilmDirectorResponse{
		DirectoryID: leafID,
		FileName:    file,
		FilePath:    fullPath,
		Dir1Id:      levels[0],
		Dir2Id:      levels[1],
		Dir3Id:      levels[2],
		Dir4Id:      levels[3],
	}, nil
}

// extractParts 拆分完整路径为目录层级 parts 和文件名。
// 例如：
//
//	/Volumes/Expansion/v3_watched/sub/MARKV3_2024-001/2024-12-19-004/good_movie.mp4
//	parts = ["Volumes","Expansion","v3_watched","sub","MARKV3_2024-001","2024-12-19-004"]
//	file  = "good_movie.mp4"
func extractParts(fullPath string) ([]string, string) {
	// 清理路径（去掉多余的分隔符）
	fullPath = filepath.Clean(fullPath)

	// 拆出目录和文件名
	dir := filepath.Dir(fullPath)
	file := filepath.Base(fullPath)

	// 去掉首个 "/" 再 split
	dir = strings.TrimPrefix(dir, string(filepath.Separator))
	parts := strings.Split(dir, string(filepath.Separator))

	// 过滤空字符串（防止多余 /）
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			res = append(res, p)
		}
	}

	return res, file
}
