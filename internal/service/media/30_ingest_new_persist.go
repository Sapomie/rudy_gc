package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"github.com/xfrr/goffmpeg/models"
)

var (
	favoriteAlbumName = "下载中"
	mediaAlbumName    = "Media"

	videoExts = map[string]struct{}{
		".mp4": {},
		".mkv": {},
		".avi": {},
		".mov": {},
		".wmv": {},
		".flv": {},
		".ts":  {},
	}

	movieNameRe = regexp.MustCompile(`^([A-Z]+)-?(\d+)$`)
)

type movieInfo struct {
	javID        string
	releasingDay int64
}

type rawMovieMeta struct {
	movieName string
	ext       string
	hasSub    int64
	selfMake  int64
	hasMask   int64
}

type probedMeta struct {
	width        int64
	height       int64
	bitRate      int64
	duration     int64
	frameAverage float64
}

type mediaRowInput struct {
	MovieInfo       movieInfo
	MovieName       string
	FileName        string
	RootDir         string
	FullDir         string
	DirectoryID     int64
	Alias           string
	Size            int64
	VideoMeta       probedMeta
	NeedScanMeta    int64
	HasSub          int64
	SelfMake        int64
	HasMask         int64
	BirthTime       int64
	SourceTorrentID string
	NowUnix         int64
}

type favoriteAlbumSourceInfo struct {
	favoriteAlbumID int64
	item            *moviex.TmAlbumItem
	infoHash        string
}

func (s *Service) findMovieInfoByName(ctx context.Context, movieName string) (movieInfo, error) {
	list, err := s.deps.MovieModel.FindMoviesByName(ctx, movieName)
	if err != nil {
		return movieInfo{}, err
	}
	if len(list) == 0 {
		return movieInfo{}, fmt.Errorf("movie not found: %s", movieName)
	}
	return movieInfo{
		javID:        list[0].JavId,
		releasingDay: list[0].ReleasingDate,
	}, nil
}

func (s *Service) findFavoriteAlbumSourceInfo(ctx context.Context, movieJavID string) (*favoriteAlbumSourceInfo, bool, error) {
	movieJavID = strings.TrimSpace(movieJavID)
	if movieJavID == "" {
		return nil, false, nil
	}

	album, err := s.deps.AlbumModel.FindOneByName(ctx, favoriteAlbumName)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	items, err := s.deps.AlbumItemModel.ListByAlbumIdMovieJavId(ctx, album.Id, movieJavID)
	if err != nil {
		return nil, false, err
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		hash := strings.TrimSpace(item.InfoHash)
		if hash == "" {
			continue
		}
		return &favoriteAlbumSourceInfo{
			favoriteAlbumID: album.Id,
			item:            item,
			infoHash:        hash,
		}, true, nil
	}
	return nil, false, nil
}

func (s *Service) moveFavoriteItemToMediaAlbum(ctx context.Context, source *favoriteAlbumSourceInfo) error {
	if source == nil || source.item == nil || source.favoriteAlbumID <= 0 {
		return nil
	}
	item := source.item

	mediaAlbumID, err := s.ensureAlbumByName(ctx, mediaAlbumName, "media ingest 自动迁移")
	if err != nil {
		return err
	}

	_, err = s.deps.AlbumItemModel.FindOneByAlbumIdSourceTypeSourceRowId(ctx, mediaAlbumID, item.SourceType, item.SourceRowId)
	switch {
	case err == nil:
		// 已存在则直接从下载中移除，保证迁移幂等。
	case errors.Is(err, moviex.ErrNotFound):
		now := time.Now().Unix()
		_, insertErr := s.deps.AlbumItemModel.Insert(ctx, &moviex.TmAlbumItem{
			AlbumId:     mediaAlbumID,
			SourceType:  item.SourceType,
			SourceRowId: item.SourceRowId,
			MovieJavId:  item.MovieJavId,
			InfoHash:    item.InfoHash,
			MovieName:   item.MovieName,
			Size:        item.Size,
			PublishTime: item.PublishTime,
			CreatedOn:   now,
			UpdatedOn:   now,
		})
		if insertErr != nil {
			return insertErr
		}
	default:
		return err
	}

	_, err = s.deps.AlbumItemModel.DeleteByAlbumIdSourceTypeSourceRowId(ctx, source.favoriteAlbumID, item.SourceType, item.SourceRowId)
	return err
}

func (s *Service) ensureAlbumByName(ctx context.Context, albumName, remark string) (int64, error) {
	albumName = strings.TrimSpace(albumName)
	if albumName == "" {
		return 0, fmt.Errorf("empty album name")
	}

	album, err := s.deps.AlbumModel.FindOneByName(ctx, albumName)
	if err == nil && album != nil {
		return album.Id, nil
	}
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return 0, err
	}

	now := time.Now().Unix()
	result, err := s.deps.AlbumModel.Insert(ctx, &moviex.TAlbum{
		Name:      albumName,
		Remark:    strings.TrimSpace(remark),
		CreatedOn: now,
		UpdatedOn: now,
	})
	if err != nil {
		again, againErr := s.deps.AlbumModel.FindOneByName(ctx, albumName)
		if againErr == nil && again != nil {
			return again.Id, nil
		}
		return 0, err
	}

	if insertID, idErr := result.LastInsertId(); idErr == nil && insertID > 0 {
		return insertID, nil
	}

	again, againErr := s.deps.AlbumModel.FindOneByName(ctx, albumName)
	if againErr != nil {
		return 0, againErr
	}
	return again.Id, nil
}

func (s *Service) buildMediaRow(in mediaRowInput) *moviex.WMedia {
	sourceHash := strings.TrimSpace(in.SourceTorrentID)
	if sourceHash == "" {
		sourceHash = strings.Repeat("0", defaultSourceHashLen)
	}

	return &moviex.WMedia{
		MovieJavId:        in.MovieInfo.javID,
		MovieName:         in.MovieName,
		FileName:          in.FileName,
		SourceType:        consts.WMediaSourceNative,
		SourceTorrentHash: sourceHash,
		DirectoryId:       in.DirectoryID,
		RootDir:           filepath.Clean(in.RootDir),
		FullDir:           filepath.Clean(in.FullDir),
		Alias:             in.Alias,
		Size:              in.Size,
		Width:             in.VideoMeta.width,
		Height:            in.VideoMeta.height,
		BitRate:           in.VideoMeta.bitRate,
		Duration:          in.VideoMeta.duration,
		FrameAverage:      in.VideoMeta.frameAverage,
		HasSub:            in.HasSub,
		SelfMake:          in.SelfMake,
		HasMask:           in.HasMask,
		NeedScanMeta:      in.NeedScanMeta,
		IsRemoved:         consts.FilmIsNotRemoved,
		RemoveTime:        0,
		BirthTime:         in.BirthTime,
		ReleasingDate:     in.MovieInfo.releasingDay,
		CreatedOn:         in.NowUnix,
		UpdatedOn:         in.NowUnix,
	}
}

func (s *Service) upsertMedia(ctx context.Context, row *moviex.WMedia) error {
	existing, err := s.findMediaForUpsert(ctx, row)
	if err != nil {
		return err
	}
	if existing != nil {
		row.Id = existing.Id
		row.CreatedOn = existing.CreatedOn
		if err := s.deps.WMediaModel.Update(ctx, row); err != nil {
			return err
		}
		s.markMediaAggDirty(ctx, existing, row)
		s.invalidateMovieTypeCaches(ctx, existing.MovieJavId, row.MovieJavId)
		return nil
	}

	if _, err := s.deps.WMediaModel.Insert(ctx, row); err != nil {
		// 唯一键冲突时，再尝试一次回查并更新。
		if !isDuplicateEntryErr(err) {
			return err
		}
		dup, findErr := s.findMediaForUpsert(ctx, row)
		if findErr != nil {
			return err
		}
		if dup == nil {
			return err
		}
		row.Id = dup.Id
		row.CreatedOn = dup.CreatedOn
		if err := s.deps.WMediaModel.Update(ctx, row); err != nil {
			return err
		}
		s.markMediaAggDirty(ctx, dup, row)
		s.invalidateMovieTypeCaches(ctx, dup.MovieJavId, row.MovieJavId)
		return nil
	}
	s.markMediaAggDirty(ctx, row)
	s.invalidateMovieTypeCaches(ctx, row.MovieJavId)
	return nil
}

func (s *Service) invalidateMovieTypeCaches(ctx context.Context, javIDs ...string) {
	if s.deps.MovieTypeCache == nil || len(javIDs) == 0 {
		return
	}

	seen := make(map[string]struct{}, len(javIDs))
	log := s.deps.Log.WithContext(ctx)
	for _, javID := range javIDs {
		javID = strings.TrimSpace(javID)
		if javID == "" {
			continue
		}
		if _, ok := seen[javID]; ok {
			continue
		}
		seen[javID] = struct{}{}

		if err := s.deps.MovieTypeCache.DelMovieType(ctx, javID); err != nil {
			log.Errorf("del MovieType cache failed, javId=%s, err=%v", javID, err)
			continue
		}
		log.Infof("del MovieType cache ok, javId=%s", javID)
	}
}

func (s *Service) findMediaForUpsert(ctx context.Context, row *moviex.WMedia) (*moviex.WMedia, error) {
	if strings.TrimSpace(row.MovieJavId) != "" {
		existing, err := s.deps.WMediaModel.FindOneByMovieJavIdSourceType(ctx, row.MovieJavId, row.SourceType)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, moviex.ErrNotFound) {
			return nil, err
		}
	}

	if strings.TrimSpace(row.MovieName) != "" {
		existing, err := s.deps.WMediaModel.FindOneByMovieNameSourceType(ctx, row.MovieName, row.SourceType)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, moviex.ErrNotFound) {
			return nil, err
		}
	}

	if strings.TrimSpace(row.FileName) != "" {
		existing, err := s.deps.WMediaModel.FindOneByFileNameSourceType(ctx, row.FileName, row.SourceType)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, moviex.ErrNotFound) {
			return nil, err
		}
	}

	return nil, nil
}

func isDuplicateEntryErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

func parseRawMovieMeta(fileName string) (rawMovieMeta, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	upper := strings.ToUpper(base)

	trimmed := trimMovieSuffix(upper)
	if strings.TrimSpace(trimmed) == "" {
		return rawMovieMeta{}, fmt.Errorf("invalid movie name: %s", fileName)
	}
	if !movieNameRe.MatchString(trimmed) {
		return rawMovieMeta{}, fmt.Errorf("unsupported movie code format: %s", trimmed)
	}
	return rawMovieMeta{
		movieName: trimmed,
		ext:       ext,
		hasSub:    parseHasSub(upper),
		selfMake:  parseSelfMake(upper),
		hasMask:   parseMask(upper),
	}, nil
}

func trimMovieSuffix(base string) string {
	current := base
	suffixes := []string{"~E", "~P", "~1", "-C"}
	for {
		changed := false
		for _, suffix := range suffixes {
			if strings.HasSuffix(current, suffix) {
				current = strings.TrimSuffix(current, suffix)
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return current
}

func parseHasSub(name string) int64 {
	if strings.Contains(name, "-C") || strings.Contains(name, "_SUB_") {
		return consts.FilmHasSub
	}
	return consts.FilmNoSub
}

func parseSelfMake(name string) int64 {
	if strings.Contains(name, "~1") || strings.Contains(name, "_COMP_") {
		return consts.FilmSelfMake
	}
	return consts.FilmNoSelfMake
}

func parseMask(name string) int64 {
	switch {
	case strings.Contains(name, "~E"), strings.Contains(name, "_ERA"):
		return consts.FilmErased
	case strings.Contains(name, "~P"), strings.Contains(name, "_NOMSK"):
		return consts.FilmNoMosaic
	default:
		return consts.FilmNotErased
	}
}

func buildTargetFileName(meta rawMovieMeta) string {
	base := buildEncodedMovieCode(meta)
	parts := make([]string, 0, 4)
	parts = append(parts, base)
	if meta.hasSub == consts.FilmHasSub {
		parts = append(parts, "sub")
	}
	if meta.selfMake == consts.FilmSelfMake {
		parts = append(parts, "self")
	}
	if meta.hasMask == consts.FilmErased {
		parts = append(parts, "era")
	}
	if meta.hasMask == consts.FilmNoMosaic {
		parts = append(parts, "nomsk")
	}

	ext := meta.ext
	if ext == "" {
		ext = ".mp4"
	}
	return strings.Join(parts, "_") + ext
}

func buildEncodedMovieCode(meta rawMovieMeta) string {
	current := strings.ToUpper(strings.TrimSpace(meta.movieName))
	match := movieNameRe.FindStringSubmatch(current)
	if len(match) != 3 {
		return current
	}

	prefix := match[1]
	serial := match[2]

	var b strings.Builder
	for _, ch := range prefix {
		if ch < 'A' || ch > 'Z' {
			return current
		}
		b.WriteString(fmt.Sprintf("%02d", int(ch-'A')+10))
	}
	return b.String() + "-" + serial
}

func buildMediaAlias(meta rawMovieMeta, birthTime, fileSize int64) string {
	date := time.Unix(birthTime, 0).Format(time.DateOnly)
	return fmt.Sprintf("%s_%s_%d", buildEncodedMovieCode(meta), date, fileSize)
}

func buildRollbackFileName(movieName string, hasSub, selfMake, hasMask int64, ext string) (string, error) {
	base := strings.ToUpper(strings.TrimSpace(movieName))
	if base == "" || !movieNameRe.MatchString(base) {
		return "", fmt.Errorf("invalid movie name for rollback: %s", movieName)
	}

	if hasSub == consts.FilmHasSub {
		base += "-C"
	}
	if selfMake == consts.FilmSelfMake {
		base += "~1"
	}
	if hasMask == consts.FilmErased {
		base += "~E"
	} else if hasMask == consts.FilmNoMosaic {
		base += "~P"
	}

	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		ext = ".mp4"
	}
	return base + ext, nil
}

func extractEncodedCodeToken(fileName string) string {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if base == "" {
		return ""
	}
	parts := strings.Split(base, "_")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func decodeEncodedMovieCode(code string) (string, bool) {
	current := strings.ToUpper(strings.TrimSpace(code))
	if current == "" {
		return "", false
	}

	parts := strings.Split(current, "-")
	if len(parts) != 2 {
		return "", false
	}
	prefixEncoded := strings.TrimSpace(parts[0])
	serial := strings.TrimSpace(parts[1])
	if prefixEncoded == "" || serial == "" || len(prefixEncoded)%2 != 0 {
		return "", false
	}

	var prefix strings.Builder
	for i := 0; i < len(prefixEncoded); i += 2 {
		v, err := strconv.Atoi(prefixEncoded[i : i+2])
		if err != nil || v < 10 || v > 35 {
			return "", false
		}
		prefix.WriteByte(byte(v-10) + 'A')
	}
	return prefix.String() + "-" + serial, true
}

func isVideoName(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	_, ok := videoExts[ext]
	return ok
}

func probeVideoMeta(filePath string) (probedMeta, error) {
	dataStr, err := ffmpeg_go.Probe(filePath)
	if err != nil {
		return probedMeta{}, err
	}

	metaData := new(models.Metadata)
	if err := json.Unmarshal([]byte(dataStr), metaData); err != nil {
		return probedMeta{}, err
	}

	out := probedMeta{}
	videoStreamFound := false
	for _, stream := range metaData.Streams {
		if strings.EqualFold(stream.CodecType, "video") {
			videoStreamFound = true
			out.width = int64(stream.Width)
			out.height = int64(stream.Height)
			out.frameAverage = parseFrameAverage(stream.AvgFrameRate)
			break
		}
	}
	if !videoStreamFound {
		return probedMeta{}, fmt.Errorf("no video stream: %s", filePath)
	}

	if v, err := strconv.ParseInt(strings.TrimSpace(metaData.Format.BitRate), 10, 64); err == nil {
		out.bitRate = v
	}
	if v, err := strconv.ParseFloat(strings.TrimSpace(metaData.Format.Duration), 64); err == nil {
		out.duration = int64(v)
	}
	return out, nil
}

func parseFrameAverage(fraction string) float64 {
	parts := strings.Split(strings.TrimSpace(fraction), "/")
	if len(parts) != 2 {
		return 0
	}
	num, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || num <= 0 {
		return 0
	}
	den, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || den <= 0 {
		return 0
	}
	return num / den
}
