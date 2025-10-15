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
)

/* =========================
   对外响应结构（可选）
========================= */

type ProcessFilmResponse struct {
	Total   int
	Items   []*processFilmDirectorResponse // 目前未写入，如需返回明细可在 handleOneVideo 中append
	Skipped int
}

/* =========================
   本轮扫描状态镜像
========================= */

type RemoveFlag int8

const (
	RemoveUnknown RemoveFlag = iota
	RemoveYes
	RemoveNo
)

type filmExistInfo struct {
	Film       *types.Film
	RemoveFlag RemoveFlag
	NeedScan   int64
}

type sameMovieInfo struct {
	MovieName string
	MoviePath string
}

type filmContext struct {
	FilmExistMap map[string]*filmExistInfo
	NameMap      map[string]*sameMovieInfo
	FilePathMap  map[string][]string // rootDir -> []movieName（用于根目录不可读兜底）
	Processed    int64
	Removed      int64
}

var (
	errNoMovie        = errors.New("no movie found")
	videoExts         = map[string]struct{}{".mp4": {}, ".mkv": {}, ".avi": {}, ".mov": {}, ".wmv": {}, ".flv": {}, ".ts": {}}
	minFileSize int64 = 50 * 1024 * 1024 // 50MB
)

/* =========================
   顶层流程
========================= */

func (s *Service) ProcessFilm(ctx context.Context) error {
	fctx, err := s.buildFilmContext(ctx)
	if err != nil {
		return err
	}

	if _, err := s.scanRoots(ctx, fctx); err != nil {
		return err
	}

	if err := s.removeMissingFilmFiles(ctx, fctx); err != nil {
		return err
	}
	return nil
}

/* =========================
   Context 构建
========================= */

func (s *Service) buildFilmContext(ctx context.Context) (*filmContext, error) {
	films, err := s.deps.FilmRepo.FindAll(ctx, consts.FilmIsNotRemoved)
	if err != nil {
		return nil, fmt.Errorf("FilmRepo.FindAll failed: %w", err)
	}

	filmExistMap := make(map[string]*filmExistInfo, len(films))
	allRootFilmMap := make(map[string][]string)

	for _, film := range films {
		filmExistMap[film.MovieName] = &filmExistInfo{
			Film:       film,
			RemoveFlag: RemoveYes, // 默认“需要删除”，扫描到文件后会标记为 RemoveNo
			NeedScan:   film.NeedScanMeta,
		}
		allRootFilmMap[film.RootDir] = append(allRootFilmMap[film.RootDir], film.MovieName)
	}

	return &filmContext{
		FilmExistMap: filmExistMap,
		NameMap:      make(map[string]*sameMovieInfo),
		FilePathMap:  allRootFilmMap,
		Processed:    0,
		Removed:      0,
	}, nil
}

/* =========================
   遍历根目录
========================= */

func (s *Service) scanRoots(ctx context.Context, fctx *filmContext) (*ProcessFilmResponse, error) {
	var (
		items   []*processFilmDirectorResponse
		skipped int
	)
	for _, root := range s.deps.Config.Film.RootDirs {
		if err := s.walkOneRoot(ctx, filepath.Clean(root), fctx, &items, &skipped); err != nil {
			return nil, err
		}
	}
	return &ProcessFilmResponse{
		Total:   len(items),
		Items:   items,
		Skipped: skipped,
	}, nil
}

func (s *Service) walkOneRoot(
	ctx context.Context,
	root string,
	fctx *filmContext,
	items *[]*processFilmDirectorResponse,
	skipped *int,
) error {
	root = strings.TrimRight(root, string(filepath.Separator))
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, walkErr error) error {
		// 1) 支持取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 2) 根级不可读：视作“目录存在”，避免误删
		if walkErr != nil {
			if dirDepth(root, p) == 0 {
				s.markDirFilmsAsExisting(p, fctx)
				return nil
			}
			return walkErr
		}

		// 3) 目录跳过
		if d.IsDir() {
			return nil
		}

		// 4) 文件大小阈值
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() < minFileSize {
			*skipped++
			return nil
		}

		// 5) 非视频：仅告警
		if !isVideo(p) {
			s.deps.Log.Warnf("non-video file encountered: %s", p)
			return nil
		}

		// 6) 处理单个视频
		if _, err := s.handleOneVideo(ctx, root, p, fctx, info); err != nil {
			return err
		}

		// 7) 进度日志
		fctx.Processed++
		if fctx.Processed%100 == 0 {
			s.deps.Log.Infof("处理完成第 %v 部 film", fctx.Processed)
		}
		return nil
	})
}

/* =========================
   单文件流水线
========================= */

func (s *Service) handleOneVideo(ctx context.Context, root, fullPath string, fctx *filmContext, info os.FileInfo) (*types.Film, error) {
	fileName := info.Name()
	movieName := extractMovieName(fileName)

	// 同名告警
	s.handleMovieNameConflict(movieName, fctx, fullPath)

	// 取 Movie
	movie, err := s.pickMovieByName(ctx, movieName)
	if err != nil {
		return nil, err
	}

	// 目录链
	dirMeta, err := s.processFilmDirectory(ctx, fullPath)
	if err != nil {
		return nil, err
	}

	// 组装基础 Film
	film := s.buildFilmSkeleton(
		movie, movieName, fileName, root, fullPath, info.Size(), dirMeta,
	)

	// 补全元数据（继承 or 扫描）
	if err := s.fillFilmMeta(ctx, film, fullPath, fctx); err != nil {
		return nil, err
	}

	// Upsert
	_, status, err := s.deps.FilmRepo.UpsertFilm(ctx, film)
	if err != nil {
		return nil, err
	}
	if status == consts.UpsertInserted {
		s.deps.Log.Info("Added Film:", film.MovieName)
	}

	// 失效缓存
	s.movieSvc.InvalidateMovieType(ctx, film.MovieJavId)
	return film, nil
}

func (s *Service) pickMovieByName(ctx context.Context, name string) (*types.Movie, error) {
	movies, err := s.deps.MovieRepo.FindMoviesByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(movies) == 0 {
		return nil, errNoMovie
	}
	if len(movies) > 1 {
		s.deps.Log.Warnf("存在相同名字的电影：%s", name)
	}
	return movies[0], nil
}

func (s *Service) buildFilmSkeleton(
	m *types.Movie,
	movieName, fileName, root, fullPath string,
	size int64,
	d *processFilmDirectorResponse,
) *types.Film {
	birth := getFileBirthTime(fullPath)
	return &types.Film{
		MovieJavId:    m.JavId,
		MovieName:     movieName,
		FileName:      fileName,
		DirectoryId:   d.DirectoryID,
		RootDir:       root,
		FullDir:       filepath.Dir(fullPath),
		Dir1Id:        d.Dir1Id,
		Dir2Id:        d.Dir2Id,
		Dir3Id:        d.Dir3Id,
		Dir4Id:        d.Dir4Id,
		Alias:         s.filmAlias(m, size, birth),
		Size:          size,
		HasSub:        determineFilmSubStatus(fullPath),
		SelfMake:      determineFilmSelfMakeStatus(fullPath),
		HasMask:       determineFilmEraseStatus(fullPath),
		NeedScanMeta:  consts.FilmMetaDataNoNeedScan,
		IsRemoved:     consts.FilmIsNotRemoved,
		RemoveTime:    0,
		ReleasingDate: m.ReleasingDate,
		BirthTime:     birth,
	}
}

func (s *Service) fillFilmMeta(ctx context.Context, f *types.Film, fullPath string, fctx *filmContext) error {
	// 继承历史计数
	if old, ok := fctx.FilmExistMap[f.MovieName]; ok {
		f.ScTimes = old.Film.ScTimes
		f.LastScTime = old.Film.LastScTime
		f.ComeTimes = old.Film.ComeTimes
	}

	// 决定是否扫描
	if s.shouldScanMetadata(f.MovieName, fctx) {
		return s.scanAndAttachMetadata(f, fullPath)
	}

	// 不需要扫描：复用旧值
	if old, ok := fctx.FilmExistMap[f.MovieName]; ok {
		f.Width = old.Film.Width
		f.Height = old.Film.Height
		f.BitRate = old.Film.BitRate
		f.Duration = old.Film.Duration
		f.FrameAverage = old.Film.FrameAverage
	}
	return nil
}

/* =========================
   清理：标记缺失文件
========================= */

func (s *Service) removeMissingFilmFiles(ctx context.Context, fctx *filmContext) error {
	for _, fi := range fctx.FilmExistMap {
		if fi.RemoveFlag == RemoveNo {
			continue
		}
		film, err := s.deps.FilmRepo.FindOne(ctx, fi.Film.Id)
		if err != nil {
			return fmt.Errorf("查找 Film 条目失败: %w", err)
		}
		film.IsRemoved = consts.FilmIsRemoved
		film.RemoveTime = time.Now().Unix()
		if _, _, err := s.deps.FilmRepo.UpsertFilm(ctx, film); err != nil {
			return fmt.Errorf("更新 Film 条目失败: %w", err)
		}
		s.movieSvc.InvalidateMovieType(ctx, film.MovieJavId)
		s.deps.Log.Infof("Film 条目删除成功: %s", fi.Film.MovieName)
	}
	return nil
}

/* =========================
   小工具 & 规则
========================= */

func (s *Service) markDirFilmsAsExisting(dir string, fctx *filmContext) {
	for _, name := range fctx.FilePathMap[dir] {
		if node, ok := fctx.FilmExistMap[name]; ok {
			node.RemoveFlag = RemoveNo
		}
	}
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := videoExts[ext]
	return ok
}

func dirDepth(root, path string) int {
	root = strings.TrimRight(filepath.Clean(root), string(filepath.Separator))
	path = filepath.Clean(path)
	if path == root {
		return 0
	}
	if !strings.HasPrefix(path, root+string(filepath.Separator)) {
		return 0
	}
	rel := path[len(root)+1:]
	if rel == "" {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func extractMovieName(fileName string) string {
	head, _, _ := strings.Cut(fileName, "_")
	return head
}

func (s *Service) handleMovieNameConflict(movieName string, fctx *filmContext, fullPath string) {
	if existing, ok := fctx.NameMap[movieName]; !ok {
		fctx.NameMap[movieName] = &sameMovieInfo{MovieName: movieName, MoviePath: fullPath}
	} else {
		s.deps.Log.Warnf("Same movie %v, Path 1: %s", existing, existing.MoviePath)
		s.deps.Log.Warnf("Same movie %v, Path 2: %s", existing, fullPath)
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
	}
	if strings.Contains(filmPath, "_nomsk") {
		return consts.FilmNoMosaic
	}
	return consts.FilmNotErased
}

func getFileBirthTime(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now().Unix()
	}
	// macOS/BSD: Ctimespec 通常是创建时间；Linux上多为“状态变更时间”，已加修改时间兜底
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if stat.Ctimespec.Sec > 0 {
			return stat.Ctimespec.Sec
		}
	}
	return info.ModTime().Unix()
}

func (s *Service) filmAlias(movie *types.Movie, filmSize, birthTime int64) string {
	parts := strings.Split(movie.Name, "-")
	if len(parts) < 2 {
		s.deps.Log.Warnf("MovieName 错误：%s", movie.Name)
		return ""
	}
	return fmt.Sprintf("%04d-%v_%s_%v",
		movie.PrefixId, parts[1],
		time.Unix(birthTime, 0).Format(time.DateOnly),
		filmSize,
	)
}

func (s *Service) shouldScanMetadata(movieName string, fctx *filmContext) bool {
	if fm, ok := fctx.FilmExistMap[movieName]; ok {
		fm.RemoveFlag = RemoveNo
		return fm.NeedScan != consts.FilmMetaDataNoNeedScan
	}
	return true
}

func (s *Service) scanAndAttachMetadata(f *types.Film, filmPath string) error {
	vm, err := filmMetaData(filmPath)
	if err != nil {
		return fmt.Errorf("解析元数据失败: %w", err)
	}
	f.Width = vm.Width
	f.Height = vm.Height
	f.BitRate = vm.BitRate
	f.Duration = vm.Duration
	f.FrameAverage = vm.FrameAverage
	return nil
}

/* =========================
   注意：以下外部依赖
   - processFilmDirectory(ctx, fullPath) (*processFilmDirectorResponse, error)
   - processFilmDirectorResponse{ DirectoryID, Dir1Id..Dir4Id }
   以上函数/结构体在你现有工程的目录模块中
========================= */
