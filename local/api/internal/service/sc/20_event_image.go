package sc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rudy_gc/local/api/internal/types"
)

func (s *Service) saveScEventImage(ctx context.Context, eventID int64, image io.Reader) (path string, err error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	reader := bufio.NewReader(image)
	header, readErr := reader.Peek(512)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", &types.BadRequestError{Message: "读取图片失败"}
	}
	var extension string
	switch http.DetectContentType(header) {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		return "", &types.BadRequestError{Message: "请选择 JPG、PNG 或 WEBP 图片"}
	}
	baseDir := strings.TrimSpace(s.deps.Config.SCEvent.ActiveImageDir)
	if baseDir == "" {
		return "", errors.New("sc_event.active_image_dir is empty")
	}
	if err := os.MkdirAll(baseDir, 0o744); err != nil {
		return "", fmt.Errorf("创建 SC 图片目录失败: %w", err)
	}
	file, err := os.CreateTemp(baseDir, fmt.Sprintf("sc-%d-*%s", eventID, extension))
	if err != nil {
		return "", fmt.Errorf("创建 SC 图片失败: %w", err)
	}
	defer func() {
		err = errors.Join(err, file.Close())
		if err != nil {
			err = errors.Join(err, discardScEventImage(file.Name()))
			path = ""
		}
	}()
	if err := file.Chmod(0o644); err != nil {
		return "", fmt.Errorf("设置 SC 图片权限失败: %w", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("保存 SC 图片失败: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("同步 SC 图片失败: %w", err)
	}
	return filepath.ToSlash(file.Name()), nil
}

func discardScEventImage(path string) error {
	if err := os.Remove(filepath.FromSlash(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("清理本次上传图片失败: %w", err)
	}
	return nil
}
