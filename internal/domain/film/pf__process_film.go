package film

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"
	"syscall"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessFilmResponse struct {
	Total   int                            // 命中的视频文件数
	Items   []*processFilmDirectorResponse // 每个文件解析出的目录ID/文件名/路径
	Skipped int                            // 跳过的小文件计数
}

// 单个影片在“本轮扫描”中的状态镜像
type filmExistInfo struct {
	Film       *types.Film
	NeedRemove int64 // 本轮扫描是否在磁盘上看到
	NeedScan   int64 // 是否需要解析元数据（由 DB 的 NeedScanMeta 映射）
}

type sameMovieInfo struct {
	MovieName string
	MoviePath string
}

type filmContext struct {
	// 以“movieName（例如 IPX-529）”为 key 的索引
	// 说明：老逻辑就是用 movieName 作为唯一键；你仍可继续沿用
	FilmExistMap map[string]*filmExistInfo
	NameMap      map[string]*sameMovieInfo
	FilePathMap  map[string][]string
	Processed    int64 // 成功处理（入库/更新）的数量
	Removed      int64 // 本轮标记删除的数量
}

const (
	needRemove = iota + 1
	noNeedRemove
)

var errNoMovie = errors.New("no movie found")

func (s *Service) ProcessFilm(ctx context.Context) error {
	fimCtx, err := s.buildFilmContext(ctx)
	if err != nil {
		return err
	}

	_, err = s.ProcessFiles(ctx, fimCtx)
	if err != nil {
		return err
	}

	err = s.removeMissingFilmFiles(ctx, fimCtx)
	if err != nil {
		return err
	}

	//todo:
	//l.asyncUpdateOwnedMovieNumber()
	//time.Sleep(time.Second * 10)

	return nil
}

func (s *Service) buildFilmContext(ctx context.Context) (*filmContext, error) {
	films, err := s.deps.FilmRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("FilmRepo.FindAll failed for  %s", err)
	}

	filmExistMap := map[string]*filmExistInfo{}
	allRootFilmMap := map[string][]string{}
	for _, film := range films {
		filmInfo := &filmExistInfo{
			Film:       film,
			NeedRemove: needRemove,
			NeedScan:   film.NeedScanMeta,
		}
		filmExistMap[film.MovieName] = filmInfo
		allRootFilmMap[film.RootDir] = append(allRootFilmMap[film.RootDir], film.MovieName)
	}

	fCtx := &filmContext{
		FilmExistMap: filmExistMap,
		NameMap:      make(map[string]*sameMovieInfo),

		Processed: 0,
		Removed:   0,
	}

	return fCtx, nil
}

func (s *Service) markDirFilmsAsExisting(dir string, fCtx *filmContext) {
	for _, filmName := range fCtx.FilePathMap[dir] {
		fCtx.FilmExistMap[filmName].NeedRemove = noNeedRemove
	}
}

var videoExts = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".wmv": {}, ".flv": {}, ".ts": {},
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := videoExts[ext]
	return ok
}

func dirDepth(root, dir string) int {
	// 统一规范
	root = strings.TrimRight(filepath.Clean(root), string(filepath.Separator))
	dir = filepath.Clean(dir)

	if dir == root {
		return 0 // 根目录
	}
	rel := strings.TrimPrefix(dir, root+string(filepath.Separator))
	if rel == dir || rel == "" {
		// 不在root下，或异常
		return 0
	}
	// 第一层=1，所以分隔符个数+1
	return strings.Count(rel, string(filepath.Separator)) + 1
}

const minFileSize int64 = 50 * 1024 * 1024 // 50MB

func (s *Service) ProcessFiles(ctx context.Context, fCtx *filmContext) (*ProcessFilmResponse, error) {
	var items []*processFilmDirectorResponse
	skipped := 0

	for _, root := range s.deps.Config.Film.RootDirs {
		root = filepath.Clean(root)
		root = strings.TrimRight(root, string(filepath.Separator))
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, walkErr error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			depth := dirDepth(root, p)

			if walkErr != nil {
				if depth == 0 {
					s.markDirFilmsAsExisting(p, fCtx)
					logx.Info(p, "------------", depth, "-------", root)
					return nil
				} else {
					return walkErr
				}
			}

			if d.IsDir() {
				return nil
			}

			if !isVideo(p) {
				logx.Alert(fmt.Errorf("non-video file encountered: %s", p).Error())
				return nil
			}

			// 文件大小检查（<50MB 跳过）
			info, err := os.Stat(p)
			if err != nil {
				return err // 文件信息获取失败算严重错误
			}
			if info.Size() < minFileSize {
				skipped++
				return nil
			}

			_, err = s.makeAndInsertFilm(ctx, d, root, p, fCtx)
			if err != nil {
				return err
			}

			fCtx.Processed++
			if fCtx.Processed%100 == 0 {
				s.deps.Log.Infof("处理完成第 %v部film", fCtx.Processed)
			}
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

func (s *Service) makeAndInsertFilm(ctx context.Context, e os.DirEntry, rootDir, fullPath string, fCtx *filmContext) (interface{}, error) {
	fileInfo, err := e.Info()
	if err != nil {
		return nil, err
	}
	fileName := fileInfo.Name()
	filmSize := fileInfo.Size()

	movieName := extractMovieName(fileName)
	s.handleMovieNameConflict(movieName, fCtx, fullPath)
	movies, err := s.deps.MovieRepo.FindMoviesByName(ctx, movieName)
	if err != nil {
		return 0, err
	}
	if len(movies) == 0 {
		return 0, errNoMovie
	}
	if len(movies) > 1 {
		s.deps.Log.Warnf("存在相同名字的电影：%s", movieName)
	}
	movie := movies[0]

	resp, err := s.processFilmDirectory(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	filmBirthTime := getFileBirthTime(fullPath)

	film := &types.Film{
		MovieJavId:   movie.JavId,
		MovieName:    movieName,
		FileName:     fileName,
		DirectoryId:  resp.DirectoryID,
		RootDir:      rootDir,
		Dir1Id:       resp.Dir1Id,
		Dir2Id:       resp.Dir2Id,
		Dir3Id:       resp.Dir3Id,
		Dir4Id:       resp.Dir4Id,
		Alias:        s.filmAlias(movie, filmSize, filmBirthTime),
		Size:         filmSize,
		HasSub:       determineFilmSubStatus(fullPath),
		SelfMake:     determineFilmSelfMakeStatus(fullPath),
		HasMask:      determineFilmEraseStatus(fullPath),
		NeedScanMeta: consts.FilmMetaDataNeedScan,
		IsRemoved:    consts.FilmIsNotRemoved,
		RemoveTime:   0,
		BirthTime:    filmBirthTime,
	}

	err = s.processFilmMetadata(film, fullPath, fCtx)
	if err != nil {
		return nil, err
	}

	_, err = s.deps.FilmRepo.UpsertFilm(ctx, film)
	if err != nil {
		return nil, err
	}

	s.movieSvc.InvalidateMovieType(ctx, film.MovieJavId)

	return nil, nil
}

func (s *Service) removeMissingFilmFiles(ctx context.Context, fCtx *filmContext) error {

	for _, filmInfo := range fCtx.FilmExistMap {
		if filmInfo.NeedRemove == noNeedRemove {
			continue
		}

		film, err := s.deps.FilmRepo.FindOne(ctx, filmInfo.Film.Id)
		if err != nil {
			return fmt.Errorf("查找 Film 条目失败: %w", err)
		}

		film.IsRemoved = consts.FilmIsRemoved
		film.RemoveTime = time.Now().Unix()
		_, err = s.deps.FilmRepo.UpsertFilm(ctx, film)
		if err != nil {
			return fmt.Errorf("更新 Film 条目失败: %w", err)
		}

		s.movieSvc.InvalidateMovieType(ctx, film.MovieJavId)

		s.deps.Log.Infof("Film 条目删除成功: %s", filmInfo.Film.MovieName)
	}
	return nil
}

func extractMovieName(fileName string) string {
	strs := strings.Split(fileName, "_")
	return strs[0]
}

func (s *Service) handleMovieNameConflict(movieName string, fCtx *filmContext, fullPath string) {

	if existingName, exists := fCtx.NameMap[movieName]; !exists {
		fCtx.NameMap[movieName] = &sameMovieInfo{MovieName: movieName, MoviePath: fullPath}
	} else {
		s.deps.Log.Warnf("Same movie %v, Path 1: %s", existingName, fCtx.NameMap[movieName].MoviePath)
		s.deps.Log.Warnf("Same movie %v, Path 2: %s", existingName, fullPath)
	}
}

func determineFilmSubStatus(filmPath string) int64 {
	if strings.Contains(filmPath, "_sub_") {
		return consts.FilmHasSub
	}
	return consts.FilmNoSub
}

func determineFilmSelfMakeStatus(filmPath string) int64 {
	if strings.Contains(filmPath, "_comp_") {
		return consts.FilmSelfMake
	}
	return consts.FilmNoSelfMake
}

func determineFilmEraseStatus(filmPath string) int64 {
	if strings.Contains(filmPath, "_era") {
		return consts.FilmErased
	} else if strings.Contains(filmPath, "_nomsk") {
		return consts.FilmNoMosaic
	}
	return consts.FilmNotErased
}

func getFileBirthTime(path string) int64 {
	fileInfo, _ := os.Stat(path)
	ss := fileInfo.Sys().(*syscall.Stat_t)
	tCreate := ss.Ctimespec.Sec
	return tCreate
}

func (s *Service) filmAlias(movie *types.Movie, filmSize int64, birthTime int64) string {
	strs := strings.Split(movie.Name, "-")
	if len(strs) < 2 {
		s.deps.Log.Warnf("MovieName 错误：%s", movie.Name)
		return ""
	}

	return fmt.Sprintf("%04d-%v_%s_%v",
		movie.PrefixId, strs[1],
		time.Unix(birthTime, 0).Format(time.DateOnly),
		filmSize)
}

func (s *Service) shouldScanMetadata(movieName string, fCtx *filmContext) bool {
	fm, exists := fCtx.FilmExistMap[movieName]
	if exists {
		fm.NeedRemove = noNeedRemove
	}
	return !exists || fm.NeedScan != consts.FilmMetaDataNoNeedScan
}

func (s *Service) scanAndAttachMetadata(film *types.Film, filmPath string) error {
	// 调用 filmMetaData 函数以解析元数据
	vm, err := filmMetaData(filmPath)
	if err != nil {
		return fmt.Errorf("解析元数据失败: %w", err)
	}

	// 赋值解析到的元数据
	film.Width = vm.Width
	film.Height = vm.Height
	film.BitRate = vm.BitRate
	film.Duration = vm.Duration
	film.FrameAverage = vm.FrameAverage

	return nil
}

func (s *Service) processFilmMetadata(film *types.Film, filmPath string, fCtx *filmContext) error {
	filmOld, exists := fCtx.FilmExistMap[film.MovieName]
	if !exists {
		film.ScTimes = filmOld.Film.ScTimes
		film.LastScTime = filmOld.Film.LastScTime
		film.ComeTimes = filmOld.Film.ComeTimes
	}

	if s.shouldScanMetadata(film.MovieName, fCtx) {
		return s.scanAndAttachMetadata(film, filmPath)
	} else if exists {
		film.Width = filmOld.Film.Width
		film.Height = filmOld.Film.Height
		film.BitRate = filmOld.Film.BitRate
		film.Duration = filmOld.Film.Duration
		film.FrameAverage = filmOld.Film.FrameAverage
	}

	return nil
}
