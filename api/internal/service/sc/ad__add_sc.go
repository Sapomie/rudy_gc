package sc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
	fc2domain "rudy_gc/internal/domain/fc2"
	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/taskctx"
	"rudy_gc/internal/types"
)

type AddScInput struct {
	Dir             string
	ComeMovieType   string
	ComeMovieJavId  string
	ComeFC2MovieId  int64
	MovieCast       string
	Kind            string
	DurationMinutes int64
	Fg              string
	Vessel          string
	Remarks         string
	Movies          []AddScInputMovie
}

type AddScInputMovie struct {
	MovieType  string
	MovieJavId string
	FC2MovieId int64
	IsSc       int64
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
	MovieType  string
	MovieJavId string
	FC2MovieId int64
	MovieKey   string
	MovieName  string
	Title      string
	MovieHref  string
	JacketImg  string
	Casts      []string
	JAVCard    *types.MovieType
	FC2Movie   *moviex.FC2Movie
}

func (l *ScService) AddSc(ctx context.Context, in AddScInput) error {
	log := l.deps.Log.WithContext(ctx)
	dir := strings.TrimSpace(in.Dir)
	if dir == "" {
		return fmt.Errorf("dir is required")
	}
	if in.DurationMinutes < 0 {
		return fmt.Errorf("duration must not be negative")
	}

	preview, imageFilePath, err := l.buildAddScPreview(ctx, dir)
	if err != nil {
		return err
	}
	if preview == nil {
		return fmt.Errorf("empty sc preview")
	}
	if err := l.validateAddScNameAvailable(ctx, preview.ScName); err != nil {
		return err
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:   "sc_add_preview_ready",
		Message: fmt.Sprintf("预览完成，影片数=%d", preview.MovieCount),
	})

	comeRef := normalizeAddScMovieRef(in.ComeMovieType, in.ComeMovieJavId, in.ComeFC2MovieId)
	if strings.TrimSpace(in.MovieCast) != "" && comeRef == "" {
		return fmt.Errorf("movie cast requires come movie")
	}
	selectedIsScMap := buildAddScMovieSelectionMap(in.Movies)

	var (
		totalCount    int64
		scCount       int64
		comeMovie     string
		movieJavIdMap = make(map[string]struct{})
		fc2MovieIDMap = make(map[int64]struct{})
		comeFound     bool
		gListRows     = make([]*moviex.GList, 0, len(preview.Movies))
	)
	for _, movie := range preview.Movies {
		if err := taskctx.WaitIfPaused(ctx); err != nil {
			return err
		}
		if movie == nil {
			continue
		}
		movieRef := addScPreviewMovieRef(movie)
		if movieRef == "" {
			return fmt.Errorf("preview movie identity is invalid: %s", movie.MovieName)
		}
		if movie.MovieType == moviedomain.SehuatangThreadMovieTypeJAV {
			movieJavIdMap[movie.MovieJavId] = struct{}{}
		} else if movie.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV {
			fc2MovieIDMap[movie.FC2MovieId] = struct{}{}
		}

		isCome := movieRef == comeRef
		isSc := resolveAddScMovieIsSc(selectedIsScMap, movieRef)
		delete(selectedIsScMap, movieRef)
		if isCome && isSc != consts.GListIsSc {
			return fmt.Errorf("come movie must be selected from is_sc=2 movies")
		}
		if isCome && movie.MovieType == moviedomain.SehuatangThreadMovieTypeFC2PPV && strings.TrimSpace(in.MovieCast) != "" {
			return fmt.Errorf("FC2 come movie lead cast must be empty")
		}
		if isCome && movie.MovieType == moviedomain.SehuatangThreadMovieTypeJAV && strings.TrimSpace(in.MovieCast) != "" && !containsExactString(movie.Casts, in.MovieCast) {
			return fmt.Errorf("movie cast must be selected from come movie casts")
		}
		gl := createGList(preview.ScName, movie, isCome, isSc)
		if isCome {
			comeMovie = movie.MovieName
			comeFound = true
		}
		if isSc == consts.GListIsSc {
			scCount++
		}

		gListRows = append(gListRows, mapTypesGListToModel(gl))
		totalCount++
	}
	if len(selectedIsScMap) > 0 {
		return fmt.Errorf("submitted movie is not present in current preview")
	}
	if comeRef != "" && !comeFound {
		return fmt.Errorf("selected come movie not found in preview movies")
	}

	scName := preview.ScName
	if len(scName) != 16 {
		log.Error("Invalid directory name length")
		return fmt.Errorf("directory name %s does not have the expected length of 16 characters", scName)
	}

	now := time.Now().Unix()
	sc := &moviex.GSc{
		Name:          scName,
		ScTime:        preview.ScTime,
		Kind:          strings.TrimSpace(in.Kind),
		ComeMovieName: comeMovie,
		MovieNumber:   scCount,
		Cooldown:      0,
		Duration:      in.DurationMinutes,
		Fg:            strings.TrimSpace(in.Fg),
		Vessel:        strings.TrimSpace(in.Vessel),
		Remarks:       strings.TrimSpace(in.Remarks),
		MovieCast:     strings.TrimSpace(in.MovieCast),
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	var copiedImagePath string
	if imageFilePath != "" {
		imagePath, err := l.copyScImage(imageFilePath, scName)
		if err != nil {
			return fmt.Errorf("failed to copy sc image: %w", err)
		}
		sc.ImagePath = imagePath
		copiedImagePath = imagePath
	}

	prev, err := l.scFindNearest(ctx, preview.ScTime)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		if copiedImagePath != "" {
			_ = os.Remove(filepath.Clean(copiedImagePath))
		}
		return fmt.Errorf("failed to find previous sc: %w", err)
	}
	if prev != nil {
		sc.Cooldown = preview.ScTime - prev.ScTime
	}

	if err = l.deps.ScEventWriteModel.CreateEvent(ctx, sc, gListRows); err != nil {
		if copiedImagePath != "" {
			_ = os.Remove(filepath.Clean(copiedImagePath))
		}
		return fmt.Errorf("failed to create sc event transaction: %w", err)
	}
	if err = l.refreshPersonScSnapshotsByScNames(ctx, now, scName); err != nil {
		return fmt.Errorf("failed to refresh person sc snapshots: %w", err)
	}
	if err = l.AddMovieAndCastScInfo(ctx, movieJavIdMap); err != nil {
		return fmt.Errorf("failed to add movie and cast sc: %w", err)
	}
	if err = l.AddFC2MovieScInfo(ctx, fc2MovieIDMap); err != nil {
		return fmt.Errorf("failed to add fc2 movie sc: %w", err)
	}
	for _, movie := range preview.Movies {
		if movie == nil {
			continue
		}
		taskctx.ReportProgress(ctx, taskctx.Progress{
			Stage:        "sc_add_movie_done",
			Message:      fmt.Sprintf("已写入影片：%s", movie.MovieName),
			HandledCount: int(totalCount),
			SuccessCount: int(totalCount),
		})
	}
	taskctx.ReportProgress(ctx, taskctx.Progress{
		Stage:        "sc_add_done",
		Message:      fmt.Sprintf("SC 新增完成：%s", scName),
		HandledCount: int(totalCount),
		SuccessCount: int(totalCount),
	})

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
		l.deps.Log.WithContext(ctx).Error("Invalid directory name length")
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
	videoInfos := make([]os.FileInfo, 0, len(entries))
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
		videoInfos = append(videoInfos, info)
	}
	if len(videoInfos) == 0 {
		return preview, imageFilePath, nil
	}

	fileNames := make([]string, 0, len(videoInfos))
	movieNames := make([]string, 0, len(videoInfos))
	fc2Keys := make([]string, 0, len(videoInfos))
	fc2KeyByFileName := make(map[string]string, len(videoInfos))
	for _, info := range videoInfos {
		fileNames = append(fileNames, info.Name())
		movieNames = append(movieNames, extractMovieName(info.Name()))
		if key, ok := normalizeScAddFC2Key(info.Name()); ok {
			fc2Keys = append(fc2Keys, key)
			fc2KeyByFileName[info.Name()] = key
		}
	}
	fc2ByFileRows, err := l.deps.FC2MediaModel.ListByFileNames(ctx, fileNames)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list fc2 media by file names: %w", err)
	}
	fc2ByKeyRows, err := l.deps.FC2MediaModel.ListByKeys(ctx, fc2Keys)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list fc2 media by keys: %w", err)
	}
	javRows, err := l.deps.WMediaModel.ListByMovieNames(ctx, movieNames)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list jav media by movie names: %w", err)
	}

	fc2MediaByFileName := make(map[string]*moviex.FC2Media, len(fc2ByFileRows))
	fc2MediaByKey := make(map[string]*moviex.FC2Media, len(fc2ByKeyRows))
	fc2MovieIDs := make([]int64, 0, len(fc2ByFileRows)+len(fc2ByKeyRows))
	for _, row := range append(fc2ByFileRows, fc2ByKeyRows...) {
		if row == nil || row.FC2MovieID <= 0 {
			continue
		}
		fc2MovieIDs = append(fc2MovieIDs, row.FC2MovieID)
		if strings.TrimSpace(row.FileName) != "" {
			fc2MediaByFileName[strings.TrimSpace(row.FileName)] = row
		}
		if strings.TrimSpace(row.FC2Key) != "" {
			fc2MediaByKey[strings.TrimSpace(row.FC2Key)] = row
		}
	}
	fc2MovieRows, err := l.deps.FC2MovieModel.ListByIDs(ctx, fc2MovieIDs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list fc2 movies: %w", err)
	}
	fc2MovieByID := make(map[int64]*moviex.FC2Movie, len(fc2MovieRows))
	for _, row := range fc2MovieRows {
		if row != nil {
			fc2MovieByID[row.ID] = row
		}
	}
	javByMovieName := make(map[string]*moviex.WMedia, len(javRows))
	for _, row := range javRows {
		if row != nil {
			javByMovieName[strings.TrimSpace(row.MovieName)] = row
		}
	}

	seenRefs := make(map[string]string, len(videoInfos))
	for _, info := range videoInfos {
		candidates := make(map[string]*AddScPreviewMovie)
		addFC2PreviewCandidate(candidates, fc2MediaByFileName[info.Name()], fc2MovieByID)
		if fc2Key := fc2KeyByFileName[info.Name()]; fc2Key != "" {
			addFC2PreviewCandidate(candidates, fc2MediaByKey[fc2Key], fc2MovieByID)
		}
		movieName := extractMovieName(info.Name())
		if wm := javByMovieName[movieName]; wm != nil {
			mt, movieErr := l.movieSvc.GetMovieType(ctx, wm.MovieJavId)
			if movieErr != nil || mt == nil {
				return nil, "", fmt.Errorf("failed to get movie type: %w,%s", movieErr, wm.MovieJavId)
			}
			ref := normalizeAddScMovieRef(moviedomain.SehuatangThreadMovieTypeJAV, wm.MovieJavId, 0)
			candidates[ref] = &AddScPreviewMovie{
				MovieType:  moviedomain.SehuatangThreadMovieTypeJAV,
				MovieJavId: wm.MovieJavId,
				MovieKey:   wm.MovieName,
				MovieName:  wm.MovieName,
				Title:      mt.Title,
				MovieHref:  buildMovieHref(wm.MovieName),
				JacketImg:  strings.TrimSpace(mt.JacketImg),
				Casts:      collectCastNames(mt),
				JAVCard:    mt,
			}
		}
		if len(candidates) == 0 {
			return nil, "", fmt.Errorf("failed to resolve sc movie file: %s", info.Name())
		}
		if len(candidates) > 1 {
			refs := make([]string, 0, len(candidates))
			for ref := range candidates {
				refs = append(refs, ref)
			}
			return nil, "", fmt.Errorf("sc movie file has conflicting owners: %s => %s", info.Name(), strings.Join(refs, ","))
		}
		for ref, movie := range candidates {
			if firstFile, ok := seenRefs[ref]; ok {
				return nil, "", fmt.Errorf("sc movie appears multiple times: %s, files=%s,%s", ref, firstFile, info.Name())
			}
			seenRefs[ref] = info.Name()
			preview.Movies = append(preview.Movies, movie)
			preview.MovieCount++
		}
	}

	return preview, imageFilePath, nil
}

func addFC2PreviewCandidate(candidates map[string]*AddScPreviewMovie, media *moviex.FC2Media, movies map[int64]*moviex.FC2Movie) {
	if media == nil || media.FC2MovieID <= 0 {
		return
	}
	movie := movies[media.FC2MovieID]
	if movie == nil {
		return
	}
	ref := normalizeAddScMovieRef(moviedomain.SehuatangThreadMovieTypeFC2PPV, "", movie.ID)
	candidates[ref] = &AddScPreviewMovie{
		MovieType:  moviedomain.SehuatangThreadMovieTypeFC2PPV,
		FC2MovieId: movie.ID,
		MovieKey:   movie.FC2Key,
		MovieName:  movie.FC2Key,
		Title:      movie.Title,
		MovieHref:  "/fcdetail/" + movie.FC2Key,
		Casts:      []string{},
		FC2Movie:   movie,
	}
}

func createGList(scName string, movie *AddScPreviewMovie, isCome bool, isSc int64) *types.GList {
	comeFlag := consts.GListIsNotCome
	if isCome {
		comeFlag = consts.GListIsCome
	}
	if isSc != consts.GListIsSc {
		isSc = consts.GListIsNotSc
	}
	return &types.GList{
		Name:       fmt.Sprintf("%s__%s", scName, movie.MovieKey),
		ScName:     scName,
		MovieType:  movie.MovieType,
		MovieJavId: movie.MovieJavId,
		FC2MovieId: movie.FC2MovieId,
		IsCome:     comeFlag,
		IsSc:       isSc,
	}
}

func buildAddScMovieSelectionMap(items []AddScInputMovie) map[string]int64 {
	out := make(map[string]int64, len(items))
	for _, item := range items {
		ref := normalizeAddScMovieRef(item.MovieType, item.MovieJavId, item.FC2MovieId)
		if ref == "" {
			continue
		}
		out[ref] = normalizeAddScMovieIsSc(item.IsSc)
	}
	return out
}

func resolveAddScMovieIsSc(selected map[string]int64, movieRef string) int64 {
	if selected == nil {
		return consts.GListIsSc
	}
	if v, ok := selected[strings.TrimSpace(movieRef)]; ok {
		return normalizeAddScMovieIsSc(v)
	}
	return consts.GListIsSc
}

func normalizeAddScMovieRef(movieType, movieJavID string, fc2MovieID int64) string {
	switch strings.ToLower(strings.TrimSpace(movieType)) {
	case moviedomain.SehuatangThreadMovieTypeJAV:
		movieJavID = strings.TrimSpace(movieJavID)
		if movieJavID == "" || fc2MovieID != 0 {
			return ""
		}
		return "jav:" + movieJavID
	case moviedomain.SehuatangThreadMovieTypeFC2PPV:
		if strings.TrimSpace(movieJavID) != "" || fc2MovieID <= 0 {
			return ""
		}
		return "fc2ppv:" + strconv.FormatInt(fc2MovieID, 10)
	default:
		return ""
	}
}

func addScPreviewMovieRef(movie *AddScPreviewMovie) string {
	if movie == nil {
		return ""
	}
	return normalizeAddScMovieRef(movie.MovieType, movie.MovieJavId, movie.FC2MovieId)
}

func mapTypesGListToModel(row *types.GList) *moviex.GList {
	if row == nil {
		return nil
	}
	return &moviex.GList{
		Id: row.Id, Name: row.Name, ScName: row.ScName, MovieType: row.MovieType,
		MovieJavId: row.MovieJavId, FC2MovieId: row.FC2MovieId, IsCome: row.IsCome,
		IsSc: row.IsSc, CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}

func normalizeAddScMovieIsSc(v int64) int64 {
	if v == consts.GListIsSc {
		return consts.GListIsSc
	}
	return consts.GListIsNotSc
}

func (l *ScService) validateAddScNameAvailable(ctx context.Context, scName string) error {
	scName = strings.TrimSpace(scName)
	if scName == "" {
		return fmt.Errorf("sc name is required")
	}

	row, err := l.deps.ScModel.FindOneByName(ctx, scName)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("check existing sc failed: %w", err)
	}
	if row != nil {
		return fmt.Errorf("sc name already exists: %s", scName)
	}

	rows, err := l.deps.GListModel.ListByScName(ctx, scName)
	if err != nil {
		return fmt.Errorf("check existing g_list sc failed: %w", err)
	}
	if len(rows) > 0 {
		return fmt.Errorf("sc name already exists: %s", scName)
	}

	return nil
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

func normalizeScAddFC2Key(fileName string) (string, bool) {
	movieName := extractMovieName(fileName)
	if key, _, _, _, ok := fc2domain.NormalizeKey(movieName); ok {
		return key, true
	}
	return fc2domain.NormalizeEncodedMediaKey(movieName)
}
