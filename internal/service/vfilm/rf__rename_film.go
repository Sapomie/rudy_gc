package vfilm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/convert"
	"strconv"
	"strings"

	"github.com/xfrr/goffmpeg/models"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 文件内使用的小写常量
const (
	filmNameHasSub     = "sub"
	filmNameNoSub      = "nos"
	filmNameCompress   = "comp"
	filmNameNoCompress = "nop"
	filmNameErased     = "era"
	filmNameNOMosaic   = "nomsk"
	filmNameNoErased   = "noe"
	videoExt           = ".mp4"
)

// RenameFilm 扫描配置目录下的 mp4 文件并按规则重命名。
func (s *FilmService) RenameFilm(ctx context.Context) error {
	log := s.deps.Log.WithContext(ctx)

	dirs := s.deps.Config.Film.RenamePaths
	if len(dirs) == 0 {
		return nil
	}

	filmSet, err := s.getExistingFilmNameSet(ctx)
	if err != nil {
		return fmt.Errorf("get existing films: %w", err)
	}

	var count, total int
	for _, dir := range dirs {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				log.Warnf("目录不存在，跳过：%s", dir)
				continue
			}
			return fmt.Errorf("read dir %q: %w", dir, readErr)
		}

		for _, e := range entries {
			if err := taskctx.WaitIfPaused(ctx); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if !e.Type().IsRegular() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(strings.ToLower(name), videoExt) {
				continue
			}
			info, statErr := e.Info()
			if statErr != nil {
				log.Warnf("stat %s failed: %v", filepath.Join(dir, name), statErr)
				continue
			}
			if info.Size() < minFileSize {
				continue
			}
			total++

			movieName := movieNameFromRawFile(name)
			if _, exists := filmSet[movieName]; exists {
				log.Warnf("已存在 film: %s", movieName)
			}

			newName, genErr := s.generateNewFileName(ctx, dir, name, movieName)
			if genErr != nil {
				log.Warnf("generateNewFileName(%s) err: %v", filepath.Join(dir, name), genErr)
				continue
			}
			if newName == name {
				log.Infof("跳过：文件名相同（%s）", filepath.Join(dir, name))
				continue
			}

			oldPath := filepath.Join(dir, name)
			newPath := filepath.Join(dir, newName)
			if _, existErr := os.Stat(newPath); existErr == nil {
				log.Warnf("目标已存在，跳过：%s", newPath)
				continue
			}
			if err = os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("rename %q -> %q: %w", oldPath, newPath, err)
			}

			count++
			log.Infof("重命名第 %d/%d 个: %s -> %s", count, total, oldPath, newPath)
			taskctx.ReportProgress(ctx, taskctx.Progress{
				Stage:        "film_rename_done",
				Message:      fmt.Sprintf("已重命名：%s", oldPath),
				HandledCount: total,
				SuccessCount: count,
			})
		}
	}
	log.Infof("重命名完成：共扫描 %d，成功 %d", total, count)
	return nil
}

func (s *FilmService) getExistingFilmNameSet(ctx context.Context) (map[string]struct{}, error) {
	films, err := s.filmFindAll(ctx, consts.FilmIsNotRemoved)
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(films))
	for _, f := range films {
		set[f.MovieName] = struct{}{}
	}
	return set, nil
}

func (s *FilmService) generateNewFileName(ctx context.Context, dir, fileName, movieName string) (string, error) {
	log := s.deps.Log.WithContext(ctx)
	movies, err := s.movieFindByName(ctx, movieName)
	if err != nil {
		return "", fmt.Errorf("FindMoviesByName(%s): %w", movieName, err)
	}
	if len(movies) == 0 {
		log.Errorf("No record: %s", movieName)
		return "", sqlx.ErrNotFound
	}
	if len(movies) > 1 {
		log.Warnf("More than one record: %s (选用第一条)", movieName)
	}
	movie := movies[0]

	movieType, err := s.movieSvc.GetMovieType(ctx, movie.JavId)
	if err != nil {
		return "", fmt.Errorf("GetMovieType(%s): %w", movie.JavId, err)
	}

	fullPath := filepath.Join(dir, fileName)
	meta, err := s.getMetadataForFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("getMetadataForFile(%s): %w", fullPath, err)
	}

	subPart := parseSubPart(fileName)
	compressPart := parseCompressPart(fileName)
	erasedPart := parseErasedPart(fileName)
	castPart, genrePart := buildMovieParts(movieType)
	heightPart, bitRatePart := s.extractTech(ctx, meta)

	movieNameUpper := strings.ToUpper(movieName)
	newBase := fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s_%s_%s_%s",
		movieNameUpper, castPart, genrePart, movieType.Director,
		heightPart, bitRatePart, subPart, compressPart,
		erasedPart, movieType.Title)

	return newBase + videoExt, nil
}

func (s *FilmService) getMetadataForFile(name string) (*models.Metadata, error) {
	_, md, err := getMetadata(name)
	return md, err
}

// --- 文件名标记解析 ---

func parseSubPart(fileName string) string {
	if strings.Contains(fileName, "-C") {
		return filmNameHasSub
	}
	return filmNameNoSub
}

func parseCompressPart(fileName string) string {
	if strings.Contains(fileName, "~1") {
		return filmNameCompress
	}
	return filmNameNoCompress
}

func parseErasedPart(fileName string) string {
	switch {
	case strings.Contains(fileName, "~E"):
		return filmNameErased
	case strings.Contains(fileName, "~P"):
		return filmNameNOMosaic
	default:
		return filmNameNoErased
	}
}

func buildMovieParts(mt *types.MovieType) (cast string, genre string) {
	var castB, genreB strings.Builder
	for i, c := range mt.Cast {
		if i > 0 {
			castB.WriteByte('-')
		}
		castB.WriteString(c.Name)
	}
	for i, g := range mt.Genre {
		if i > 0 {
			genreB.WriteByte('-')
		}
		genreB.WriteString(g)
	}
	return castB.String(), genreB.String()
}

func (s *FilmService) extractTech(ctx context.Context, md *models.Metadata) (heightPart, bitRatePart string) {
	log := s.deps.Log.WithContext(ctx)
	// 取视频流的高度
	for _, st := range md.Streams {
		if strings.EqualFold(st.CodecType, "video") && st.Height > 0 {
			heightPart = strconv.Itoa(st.Height)
			break
		}
	}

	// 解析比特率（models.Metadata.Format 为值类型，不是指针）
	if md.Format.BitRate != "" {
		if br, err := strconv.ParseFloat(md.Format.BitRate, 64); err == nil && br > 0 {
			// 你的原逻辑：除以 1e5 后取整
			bitRatePart = convert.FloatTo(br / 1e5).DecimalStr(0)
		} else if err != nil {
			log.Warnf("parse bitrate failed: %v", err)
		}
	}

	return
}

func movieNameFromRawFile(fileName string) string {
	if !strings.HasSuffix(strings.ToLower(fileName), videoExt) {
		return fileName
	}
	name := strings.TrimSuffix(fileName, videoExt)
	for _, suf := range []string{"~E", "~P", "~1", "-C"} {
		name = strings.TrimSuffix(name, suf)
	}
	return name
}
