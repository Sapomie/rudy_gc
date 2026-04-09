package filmmove

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/internal/types"
)

func (s *Service) Preview(ctx context.Context, req *types.ListMovieFullRequest) (*PreviewResult, error) {
	if req == nil {
		return nil, fmt.Errorf("nil filter request")
	}

	resp, err := s.movieSvc.ListMovieFull(ctx, req)
	if err != nil {
		return nil, err
	}

	result := &PreviewResult{
		Items: make([]*PreviewItem, 0, len(resp.List)),
	}
	planItems := make([]*movePlanItem, 0, len(resp.List))

	for _, movieType := range resp.List {
		item := s.buildPreviewItem(movieType)
		result.Total++
		if item.CanMove {
			result.Movable++
		} else {
			result.Failed++
		}
		result.Items = append(result.Items, item)
		planItems = append(planItems, &movePlanItem{
			MovieName:  item.MovieName,
			MovieJavID: item.MovieJavID,
			SourcePath: item.SourcePath,
			TargetPath: item.TargetPath,
			CanMove:    item.CanMove,
			Error:      item.Error,
		})
	}

	if len(planItems) > 0 {
		result.PlanID = s.savePlan(planItems)
	}
	return result, nil
}

func (s *Service) buildPreviewItem(movieType *types.MovieType) *PreviewItem {
	item := &PreviewItem{}
	if movieType == nil {
		item.Error = "empty movie row"
		return item
	}

	item.MovieName = strings.TrimSpace(movieType.Name)
	item.MovieJavID = strings.TrimSpace(movieType.JavId)

	if movieType.VFilm == nil {
		item.Error = "旧媒体不存在"
		return item
	}

	film := movieType.VFilm
	fullDir := strings.TrimSpace(film.FullDir)
	fileName := strings.TrimSpace(film.FileName)
	if fullDir == "" || fileName == "" {
		item.Error = "影片路径信息不完整"
		return item
	}

	item.SourcePath = filepath.Clean(filepath.Join(fullDir, fileName))
	targetDir := s.resolveTargetDir(film.RootDir)
	if targetDir == "" {
		item.Error = "未配置该 root_dir 的移动目标目录"
		return item
	}
	item.TargetPath = filepath.Clean(filepath.Join(targetDir, fileName))

	if item.SourcePath == item.TargetPath || filepath.Clean(filepath.Dir(item.SourcePath)) == filepath.Clean(targetDir) {
		item.Error = "文件已在目标目录"
		return item
	}

	if _, err := os.Stat(item.SourcePath); err != nil {
		item.Error = "源文件不存在或不可读"
		return item
	}

	if _, err := os.Stat(item.TargetPath); err == nil {
		item.Error = "目标文件已存在"
		return item
	}

	item.CanMove = true
	return item
}
