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
}

func (s *Service) processFilmDirectory(ctx context.Context, fullPath string) (*processFilmDirectorResponse, error) {
	parts, fileName := extractParts(fullPath)
	dirID, err := s.deps.DirectoryRepo.GetOrCreateChain(ctx, parts)
	if err != nil {
		return nil, err
	}
	return &processFilmDirectorResponse{DirectoryID: dirID, FilePath: fullPath, FileName: fileName}, nil
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
