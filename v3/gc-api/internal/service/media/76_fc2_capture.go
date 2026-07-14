package media

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/model/modelx/moviex"
)

const (
	FC2MediaImageTypeScreenshot   = "screenshot"
	FC2MediaImageTypeContactSheet = "contact_sheet"
)

var ErrInvalidFC2CaptureRequest = errors.New("FC2视频图片生成参数无效")

type FC2ScreenshotOptions struct {
	TimeSeconds float64
}

type FC2ContactSheetOptions struct {
	Columns   int
	Rows      int
	CellWidth int
}

func (s *Service) GenerateFC2Screenshot(ctx context.Context, mediaRow *moviex.FC2Media, options FC2ScreenshotOptions) (*moviex.FC2MediaImage, error) {
	if mediaRow == nil || mediaRow.ID <= 0 {
		return nil, fmt.Errorf("%w: FC2媒体记录为空", ErrInvalidFC2CaptureRequest)
	}
	if math.IsNaN(options.TimeSeconds) || math.IsInf(options.TimeSeconds, 0) || options.TimeSeconds < 0 {
		return nil, fmt.Errorf("%w: 截图时间必须是非负数", ErrInvalidFC2CaptureRequest)
	}
	if mediaRow.Duration <= 0 {
		return nil, fmt.Errorf("%w: 视频时长无效", ErrInvalidFC2CaptureRequest)
	}
	if options.TimeSeconds >= float64(mediaRow.Duration) {
		return nil, fmt.Errorf("%w: 截图时间必须小于视频时长", ErrInvalidFC2CaptureRequest)
	}

	sourcePath, captureDir, err := s.prepareFC2Capture(mediaRow)
	if err != nil {
		return nil, err
	}
	tempPath, finalPath, err := reserveFC2CaptureOutput(captureDir, FC2MediaImageTypeScreenshot)
	if err != nil {
		return nil, err
	}
	keepTemp := true
	defer func() {
		if keepTemp {
			_ = os.Remove(tempPath)
		}
	}()

	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin", "-y",
		"-i", sourcePath,
		"-ss", formatFFmpegSeconds(options.TimeSeconds),
		"-map", "0:v:0", "-frames:v", "1", "-an",
		tempPath,
	}
	if err := runFFmpeg(ctx, args...); err != nil {
		return nil, err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return nil, err
	}
	keepTemp = false
	imageRow, err := s.persistFC2MediaImage(ctx, mediaRow.ID, FC2MediaImageTypeScreenshot, finalPath)
	if err != nil {
		return nil, err
	}
	if err := s.updateFC2CoverImage(ctx, mediaRow.FC2MovieID, imageRow.ID); err != nil {
		deleteErr := s.deps.FC2MediaImageModel.DeleteByMediaIDAndID(ctx, mediaRow.ID, imageRow.ID)
		removeErr := os.Remove(imageRow.FilePath)
		if deleteErr != nil || (removeErr != nil && !os.IsNotExist(removeErr)) {
			return nil, fmt.Errorf("设置新截图为封面失败: %v；回滚图片记录失败: %v；回滚图片文件失败: %v", err, deleteErr, removeErr)
		}
		return nil, fmt.Errorf("设置新截图为封面失败: %w", err)
	}
	return imageRow, nil
}

func (s *Service) GenerateFC2ContactSheet(ctx context.Context, mediaRow *moviex.FC2Media, options FC2ContactSheetOptions) (*moviex.FC2MediaImage, error) {
	frameCount, err := validateFC2ContactSheetOptions(mediaRow, options)
	if err != nil {
		return nil, err
	}
	sourcePath, captureDir, err := s.prepareFC2Capture(mediaRow)
	if err != nil {
		return nil, err
	}
	buildDir, err := os.MkdirTemp(captureDir, ".contact-sheet-build-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(buildDir)

	timestamps := buildFC2ContactSheetTimestamps(float64(mediaRow.Duration), frameCount)
	for index, timestamp := range timestamps {
		framePath := filepath.Join(buildDir, fmt.Sprintf("frame-%06d.png", index))
		args := []string{
			"-hide_banner", "-loglevel", "error", "-nostdin", "-y",
			"-ss", formatFFmpegSeconds(timestamp),
			"-i", sourcePath,
			"-map", "0:v:0", "-frames:v", "1", "-an",
			"-vf", fmt.Sprintf("scale=%d:-2", options.CellWidth),
			framePath,
		}
		if err := runFFmpeg(ctx, args...); err != nil {
			return nil, fmt.Errorf("联系表第 %d 帧生成失败: %w", index+1, err)
		}
	}

	tempPath, finalPath, err := reserveFC2CaptureOutput(captureDir, FC2MediaImageTypeContactSheet)
	if err != nil {
		return nil, err
	}
	keepTemp := true
	defer func() {
		if keepTemp {
			_ = os.Remove(tempPath)
		}
	}()
	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin", "-y",
		"-framerate", "1", "-start_number", "0",
		"-i", filepath.Join(buildDir, "frame-%06d.png"),
		"-vf", fmt.Sprintf("tile=%dx%d:nb_frames=%d", options.Columns, options.Rows, frameCount),
		"-frames:v", "1", "-an",
		tempPath,
	}
	if err := runFFmpeg(ctx, args...); err != nil {
		return nil, err
	}
	if err := os.Rename(tempPath, finalPath); err != nil {
		return nil, err
	}
	keepTemp = false
	return s.persistFC2MediaImage(ctx, mediaRow.ID, FC2MediaImageTypeContactSheet, finalPath)
}

func validateFC2ContactSheetOptions(mediaRow *moviex.FC2Media, options FC2ContactSheetOptions) (int, error) {
	if mediaRow == nil || mediaRow.ID <= 0 {
		return 0, fmt.Errorf("%w: FC2媒体记录为空", ErrInvalidFC2CaptureRequest)
	}
	if mediaRow.Duration <= 0 {
		return 0, fmt.Errorf("%w: 视频时长无效", ErrInvalidFC2CaptureRequest)
	}
	if options.Columns <= 0 || options.Rows <= 0 || options.CellWidth <= 0 {
		return 0, fmt.Errorf("%w: 行数、列数和单格宽度必须是正整数", ErrInvalidFC2CaptureRequest)
	}
	if options.Columns > math.MaxInt/options.Rows {
		return 0, fmt.Errorf("%w: 行列乘积溢出", ErrInvalidFC2CaptureRequest)
	}
	return options.Columns * options.Rows, nil
}

func (s *Service) prepareFC2Capture(mediaRow *moviex.FC2Media) (string, string, error) {
	if s == nil || s.deps == nil {
		return "", "", fmt.Errorf("FC2媒体图片服务依赖为空")
	}
	rootDir := filepath.Clean(strings.TrimSpace(mediaRow.RootDir))
	if rootDir == "" || rootDir == "." {
		return "", "", fmt.Errorf("FC2媒体 root_dir 为空")
	}
	configured := false
	for _, rawRoot := range s.deps.Config.FC2Media.RootDirs {
		configuredRoot := filepath.Clean(strings.TrimSpace(rawRoot))
		if configuredRoot != "" && configuredRoot != "." && rootDir == configuredRoot {
			configured = true
			break
		}
	}
	if !configured {
		return "", "", fmt.Errorf("FC2媒体 root_dir 未配置: %s", rootDir)
	}
	sourcePath := joinMediaPath(mediaRow.FullDir, mediaRow.FileName)
	if sourcePath == "" || !isPathWithin(sourcePath, rootDir) {
		return "", "", fmt.Errorf("FC2媒体文件路径无效: %s", sourcePath)
	}
	stat, err := os.Stat(sourcePath)
	if err != nil {
		return "", "", fmt.Errorf("FC2媒体文件不可用: %s: %w", sourcePath, err)
	}
	if !stat.Mode().IsRegular() {
		return "", "", fmt.Errorf("FC2媒体路径不是普通文件: %s", sourcePath)
	}
	localImageDir := filepath.Clean(strings.TrimSpace(s.deps.Config.FC2Media.LocalImageDir))
	if localImageDir == "" || localImageDir == "." {
		return "", "", fmt.Errorf("FC2媒体图片 local_image_dir 未配置")
	}
	captureDir := filepath.Join(localImageDir, strconv.FormatInt(mediaRow.ID, 10))
	if err := os.MkdirAll(captureDir, defaultFilePerm); err != nil {
		return "", "", err
	}
	return sourcePath, captureDir, nil
}

func reserveFC2CaptureOutput(captureDir, imageType string) (string, string, error) {
	file, err := os.CreateTemp(captureDir, "."+imageType+"-*.png")
	if err != nil {
		return "", "", err
	}
	tempPath := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", err
	}
	finalPath := filepath.Join(captureDir, strings.TrimPrefix(filepath.Base(tempPath), "."))
	return tempPath, finalPath, nil
}

func (s *Service) persistFC2MediaImage(ctx context.Context, mediaID int64, imageType, filePath string) (*moviex.FC2MediaImage, error) {
	if s.deps.FC2MediaImageModel == nil {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("FC2媒体图片模型为空")
	}
	row, err := s.deps.FC2MediaImageModel.Insert(ctx, &moviex.FC2MediaImage{
		FC2MediaID: mediaID,
		ImageType:  imageType,
		FilePath:   filepath.Clean(filePath),
		CreatedOn:  time.Now().Unix(),
	})
	if err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}
	return row, nil
}

func (s *Service) SetFC2CoverImage(ctx context.Context, mediaRow *moviex.FC2Media, imageID int64) (*moviex.FC2MediaImage, error) {
	if s == nil || s.deps == nil || s.deps.FC2MediaImageModel == nil || s.deps.FC2MovieModel == nil {
		return nil, fmt.Errorf("FC2媒体图片服务依赖为空")
	}
	if mediaRow == nil || mediaRow.ID <= 0 || mediaRow.FC2MovieID <= 0 || imageID <= 0 {
		return nil, fmt.Errorf("%w: FC2媒体图片 ID 无效", ErrInvalidFC2CaptureRequest)
	}
	imageRow, err := s.deps.FC2MediaImageModel.FindOneByMediaIDAndID(ctx, mediaRow.ID, imageID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(imageRow.ImageType) != FC2MediaImageTypeScreenshot {
		return nil, fmt.Errorf("%w: 只有截图可以设为封面", ErrInvalidFC2CaptureRequest)
	}
	if err := s.updateFC2CoverImage(ctx, mediaRow.FC2MovieID, imageRow.ID); err != nil {
		return nil, err
	}
	return imageRow, nil
}

func (s *Service) updateFC2CoverImage(ctx context.Context, movieID, imageID int64) error {
	if s == nil || s.deps == nil || s.deps.FC2MovieModel == nil {
		return fmt.Errorf("FC2影片模型为空")
	}
	if movieID <= 0 || imageID < 0 {
		return fmt.Errorf("%w: FC2封面关联无效", ErrInvalidFC2CaptureRequest)
	}
	return s.deps.FC2MovieModel.UpdateCoverImage(ctx, movieID, imageID, time.Now().Unix())
}

func (s *Service) DeleteFC2MediaImage(ctx context.Context, mediaRow *moviex.FC2Media, imageID int64) (int64, error) {
	if s == nil || s.deps == nil || s.deps.FC2MediaImageModel == nil || s.deps.FC2MovieModel == nil {
		return 0, fmt.Errorf("FC2媒体图片服务依赖为空")
	}
	if mediaRow == nil || mediaRow.ID <= 0 || mediaRow.FC2MovieID <= 0 || imageID <= 0 {
		return 0, fmt.Errorf("%w: FC2媒体图片 ID 无效", ErrInvalidFC2CaptureRequest)
	}
	imageRow, err := s.deps.FC2MediaImageModel.FindOneByMediaIDAndID(ctx, mediaRow.ID, imageID)
	if err != nil {
		return 0, err
	}
	movieRow, err := s.deps.FC2MovieModel.FindOne(ctx, mediaRow.FC2MovieID)
	if err != nil {
		return 0, err
	}
	nextCoverImageID := movieRow.CoverImageID
	coverChanged := movieRow.CoverImageID == imageID
	if coverChanged {
		nextCoverImageID = 0
		imageRows, listErr := s.deps.FC2MediaImageModel.ListByMediaID(ctx, mediaRow.ID)
		if listErr != nil {
			return 0, listErr
		}
		for _, candidate := range imageRows {
			if candidate == nil || candidate.ID == imageID || strings.TrimSpace(candidate.ImageType) != FC2MediaImageTypeScreenshot {
				continue
			}
			nextCoverImageID = candidate.ID
			break
		}
	}
	localImageDir := filepath.Clean(strings.TrimSpace(s.deps.Config.FC2Media.LocalImageDir))
	if localImageDir == "" || localImageDir == "." {
		return 0, fmt.Errorf("FC2媒体图片 local_image_dir 未配置")
	}
	activeDir := filepath.Join(localImageDir, strconv.FormatInt(mediaRow.ID, 10))
	filePath := filepath.Clean(strings.TrimSpace(imageRow.FilePath))
	if filePath == "" || filePath == "." || filepath.Dir(filePath) != activeDir || !isPathWithin(filePath, activeDir) {
		return 0, fmt.Errorf("FC2媒体图片路径无效: %s", filePath)
	}

	archivedPath := ""
	fileInfo, statErr := os.Lstat(filePath)
	if statErr == nil {
		if !fileInfo.Mode().IsRegular() {
			return 0, fmt.Errorf("FC2媒体图片不是普通文件: %s", filePath)
		}
		deletedDir := filepath.Join(localImageDir, "deleted", strconv.FormatInt(mediaRow.ID, 10))
		if err := os.MkdirAll(deletedDir, defaultFilePerm); err != nil {
			return 0, err
		}
		archivedPath = filepath.Join(deletedDir, filepath.Base(filePath))
		if _, err := os.Lstat(archivedPath); err == nil {
			return 0, fmt.Errorf("FC2媒体图片删除归档已存在: %s", archivedPath)
		} else if !os.IsNotExist(err) {
			return 0, err
		}
	} else if !os.IsNotExist(statErr) {
		return 0, statErr
	}

	if coverChanged {
		if err := s.updateFC2CoverImage(ctx, mediaRow.FC2MovieID, nextCoverImageID); err != nil {
			return 0, err
		}
	}
	restoreCover := func() error {
		if !coverChanged {
			return nil
		}
		return s.updateFC2CoverImage(ctx, mediaRow.FC2MovieID, imageID)
	}
	if archivedPath != "" {
		if err := os.Rename(filePath, archivedPath); err != nil {
			if restoreErr := restoreCover(); restoreErr != nil {
				return 0, fmt.Errorf("移动图片到删除目录失败: %v；恢复封面关联失败: %w", err, restoreErr)
			}
			return 0, err
		}
	}

	if err := s.deps.FC2MediaImageModel.DeleteByMediaIDAndID(ctx, mediaRow.ID, imageID); err != nil {
		var restoreFileErr error
		if archivedPath != "" {
			restoreFileErr = os.Rename(archivedPath, filePath)
		}
		restoreCoverErr := restoreCover()
		if restoreFileErr != nil || restoreCoverErr != nil {
			return 0, fmt.Errorf("删除图片记录失败: %v；恢复图片失败: %v；恢复封面关联失败: %v", err, restoreFileErr, restoreCoverErr)
		}
		return 0, err
	}
	_ = os.Remove(activeDir)
	return nextCoverImageID, nil
}

func buildFC2ContactSheetTimestamps(duration float64, frameCount int) []float64 {
	if duration <= 0 || frameCount <= 0 {
		return []float64{}
	}
	segment := duration / float64(frameCount)
	out := make([]float64, frameCount)
	for index := range out {
		out[index] = (float64(index) + 0.5) * segment
	}
	return out
}

func formatFFmpegSeconds(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func runFFmpeg(ctx context.Context, args ...string) error {
	output, err := exec.CommandContext(ctx, "ffmpeg", args...).CombinedOutput()
	if err == nil {
		return nil
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("ffmpeg 执行失败: %w", err)
	}
	return fmt.Errorf("ffmpeg 执行失败: %w: %s", err, message)
}
