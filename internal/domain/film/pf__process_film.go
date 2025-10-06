package film

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"rudy_gc/internal/types"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessFilmResponse struct {
	Total   int                            // 命中的视频文件数
	Items   []*processFilmDirectorResponse // 每个文件解析出的目录ID/文件名/路径
	Skipped int                            // 跳过的小文件计数
}

var videoExts = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".wmv": {}, ".flv": {}, ".ts": {},
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := videoExts[ext]
	return ok
}

// 小于该值则跳过（字节）
const minFileSize int64 = 50 * 1024 * 1024 // 50MB

func (s *Service) ProcessFilm(ctx context.Context) (*ProcessFilmResponse, error) {
	var items []*processFilmDirectorResponse
	skipped := 0

	for _, root := range s.deps.Config.Film.FilmPath {
		root = filepath.Clean(root)

		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, walkErr error) error {
			// 1) 支持取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 2) 遍历错误立即停止
			if walkErr != nil {
				return walkErr
			}

			// 3) 目录跳过
			if d.IsDir() {
				return nil
			}

			// 4) 非视频立即报错
			if !isVideo(p) {
				logx.Alert(fmt.Errorf("non-video file encountered: %s", p).Error())
				return nil
			}

			// 5) 文件大小检查（<50MB 跳过）
			info, err := os.Stat(p)
			if err != nil {
				return err // 文件信息获取失败算严重错误
			}
			if info.Size() < minFileSize {
				skipped++
				return nil
			}

			// 6) 处理目录（任何错误立即中止）
			resp, err := s.processFilmDirectory(ctx, p)
			if err != nil {
				return fmt.Errorf("processFilmDirectory failed for %s: %w", p, err)
			}

			in := types.Film{
				MovieJavId:  resp.FileName,
				MovieName:   resp.FileName,
				FileName:    resp.FileName,
				DirectoryId: resp.DirectoryID,
				FilePath:    resp.FilePath,
			}

			_, walkErr = s.deps.FilmRepo.UpsertFilm(ctx, &in)
			if walkErr != nil {
				return fmt.Errorf("processFilmRepository.UpsertFilm failed for %s: %w", p, walkErr)
			}

			items = append(items, resp)
			return nil
		})

		// 任意错误立即中止整个流程
		if err != nil {
			return nil, err
		}
	}

	return &ProcessFilmResponse{
		Total:   len(items),
		Items:   items,
		Skipped: skipped,
	}, nil
}
