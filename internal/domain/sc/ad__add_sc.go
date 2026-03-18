package sc

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"
	"time"
)

type AddScInput struct {
	Dir            string
	ComeMovieJavId string
	MovieCast      string
	Duration       int64
	Fg             string
	Vessel         string
	Remarks        string
}

type AddScPreview struct {
	Dir        string
	ScName     string
	ScTime     int64
	MovieCount int64
	ImageFound bool
	ImageName  string
	Movies     []*AddScPreviewMovie
}

type AddScPreviewMovie struct {
	MovieName  string
	MovieJavId string
	Casts      []string
}

func (l *ScService) AddSc(ctx context.Context, in AddScInput) error {
	dir := strings.TrimSpace(in.Dir)
	if dir == "" {
		return fmt.Errorf("dir is required")
	}

	preview, imageFilePath, err := l.buildAddScPreview(ctx, dir)
	if err != nil {
		return err
	}
	if preview == nil {
		return fmt.Errorf("empty sc preview")
	}

	comeMovieJavId := strings.TrimSpace(in.ComeMovieJavId)
	if comeMovieJavId == "" {
		return fmt.Errorf("come movie is required")
	}

	var (
		count         int64
		comeMovie     string
		movieJavIdMap = make(map[string]struct{})
		comeFound     bool
	)
	for _, movie := range preview.Movies {
		if movie == nil {
			continue
		}
		movieJavIdMap[movie.MovieJavId] = struct{}{}

		isCome := movie.MovieJavId == comeMovieJavId
		gl := createGList(preview.ScName, movie.MovieName, movie.MovieJavId, isCome)
		if isCome {
			comeMovie = movie.MovieName
			comeFound = true
		}

		_, err = l.deps.GListRepo.Upsert(ctx, gl)
		if err != nil {
			return fmt.Errorf("failed to upsert glist: %w", err)
		}
		count++
	}
	if !comeFound {
		return fmt.Errorf("selected come movie not found in preview movies")
	}

	scName := preview.ScName
	if len(scName) != 16 {
		l.deps.Log.Error("Invalid directory name length")
		return fmt.Errorf("directory name %s does not have the expected length of 16 characters", scName)
	}

	sc := &types.GSc{
		Name:          scName,
		ScTime:        preview.ScTime,
		ComeMovieName: comeMovie,
		MovieNumber:   count,
		Cooldown:      0,
		Duration:      in.Duration,
		Fg:            strings.TrimSpace(in.Fg),
		Vessel:        strings.TrimSpace(in.Vessel),
		Remarks:       strings.TrimSpace(in.Remarks),
		MovieCast:     strings.TrimSpace(in.MovieCast),
	}
	if imageFilePath != "" {
		imagePath, err := l.copyScImage(imageFilePath, scName)
		if err != nil {
			return fmt.Errorf("failed to copy sc image: %w", err)
		}
		sc.ImagePath = imagePath
	}

	if sc.MovieCast == "" {
		return fmt.Errorf("movie cast is required")
	}

	prev, err := l.deps.ScRepo.FindNearest(ctx, preview.ScTime)
	if err != nil {
		return fmt.Errorf("failed to find previous sc: %w", err)
	}
	sc.Cooldown = preview.ScTime - prev.ScTime

	_, err = l.deps.ScRepo.Upsert(ctx, sc)
	if err != nil {
		return fmt.Errorf("failed to upsert sc: %w", err)
	}

	err = l.AddMovieAndCastScInfo(ctx, movieJavIdMap)
	if err != nil {
		return fmt.Errorf("failed to add movie and cast sc: %w", err)
	}

	return nil
}

func (l *ScService) BuildAddScPreview(ctx context.Context, dir string) (*AddScPreview, error) {
	preview, _, err := l.buildAddScPreview(ctx, dir)
	return preview, err
}

func (l *ScService) buildAddScPreview(ctx context.Context, dir string) (*AddScPreview, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	scName := filepath.Base(dir)
	if len(scName) != 16 {
		l.deps.Log.Error("Invalid directory name length")
		return nil, "", fmt.Errorf("directory name %s does not have the expected length of 16 characters", scName)
	}

	scTime, err := getScTime(scName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse SC time from %s: %w", scName, err)
	}

	preview := &AddScPreview{
		Dir:    dir,
		ScName: scName,
		ScTime: scTime,
		Movies: make([]*AddScPreviewMovie, 0),
	}

	var imageFilePath string
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get info for entry %s: %w", entry.Name(), err)
		}

		if isScImageFile(info) {
			if imageFilePath != "" {
				return nil, "", fmt.Errorf("multiple image files found in %s", dir)
			}
			imageFilePath = filepath.Join(dir, entry.Name())
			preview.ImageFound = true
			preview.ImageName = entry.Name()
			continue
		}

		if !isMp4File(info) {
			continue
		}

		movieName := extractMovieName(info.Name())
		vf, err := l.deps.FilmRepo.FindOneByMovieName(ctx, movieName)
		if err != nil {
			return nil, "", fmt.Errorf("failed to find film by name %s: %w", movieName, err)
		}

		mt, err := l.movieSvc.GetMovieType(ctx, vf.MovieJavId)
		if err != nil || mt == nil {
			return nil, "", fmt.Errorf("failed to get movie type: %w,%s", err, vf.MovieJavId)
		}

		previewMovie := &AddScPreviewMovie{
			MovieName:  vf.MovieName,
			MovieJavId: vf.MovieJavId,
			Casts:      collectCastNames(mt),
		}
		preview.Movies = append(preview.Movies, previewMovie)
		preview.MovieCount++
	}

	if preview.MovieCount == 0 {
		return nil, "", fmt.Errorf("no movies found in sc dir %s", dir)
	}

	return preview, imageFilePath, nil
}

func createGList(scName, movieName, movieJavId string, isCome bool) *types.GList {
	comeFlag := consts.GListIsNotCome
	if isCome {
		comeFlag = consts.GListIsCome
	}
	return &types.GList{
		Name:       fmt.Sprintf("%s__%s", scName, movieName),
		ScName:     scName,
		MovieJavId: movieJavId,
		IsCome:     comeFlag,
	}
}

func getScTime(scName string) (int64, error) {
	// 定义时间布局，Go 的时间布局以示例时间为基准
	const layout = "2006-01-02-15-04"

	t, err := time.ParseInLocation(layout, scName, time.Local)
	if err != nil {
		return 0, fmt.Errorf("error parsing time from %s: %w", scName, err)
	}

	return t.Unix(), nil
}

func isMp4File(info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), ".mp4") && info.Size() >= 20000
}

func isScImageFile(info os.FileInfo) bool {
	if info == nil || info.IsDir() {
		return false
	}
	switch strings.ToLower(filepath.Ext(info.Name())) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func (l *ScService) copyScImage(srcPath, scName string) (string, error) {
	baseDir := strings.TrimSpace(l.deps.Config.Fetcher.LocalImageDir)
	if baseDir == "" {
		return "", fmt.Errorf("fetcher.local_image_dir is empty")
	}

	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		return "", fmt.Errorf("image file has no extension: %s", srcPath)
	}

	dstDir := filepath.Join(baseDir, "sc_event")
	if err := os.MkdirAll(dstDir, 0o744); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dstDir, err)
	}

	dstPath := filepath.Join(dstDir, scName+ext)
	if err := copyFile(srcPath, dstPath); err != nil {
		return "", err
	}

	return filepath.ToSlash(dstPath), nil
}

func copyFile(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src %s: %w", srcPath, err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create dst %s: %w", dstPath, err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(dstPath)
		return fmt.Errorf("copy %s -> %s: %w", srcPath, dstPath, err)
	}
	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		_ = os.Remove(dstPath)
		return fmt.Errorf("sync dst %s: %w", dstPath, err)
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(dstPath)
		return fmt.Errorf("close dst %s: %w", dstPath, err)
	}

	return nil
}

func extractMovieName(fileName string) string {
	strs := strings.Split(fileName, "_")
	return strs[0]
}
