package vfilm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/convert"
	"strconv"
	"strings"

	"github.com/xfrr/goffmpeg/models"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	FilmNameHasSub     = "sub"
	FilmNameNoSub      = "nos"
	FilmNameCompress   = "comp"
	FilmNameNoCompress = "nop"
	FilmNameErased     = "era"
	FilmNameNOMosaic   = "nomsk"
	FilmNameNoErased   = "noe"
	VideoExt           = ".mp4"
	MinFileSize        = 20000
)

func (s *FilmService) RenameFilm() error {
	ctx := context.Background()
	dir := s.deps.Config.Film.RenamePath
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	filmMap, err := s.getExistingFilmNames(ctx)
	if err != nil {
		return err
	}

	count := 0
	for _, f := range fs {
		if !isValidMovieFile(f) {
			continue
		}

		movieName := getMovieNameByFileRawName(f.Name())
		if _, exists := filmMap[movieName]; exists {
			s.deps.Log.Warn("已存在film:", movieName)
		}

		newName, err := s.generateNewFileName(ctx, dir, f.Name(), movieName)
		if err != nil {
			s.deps.Log.Warnf("generateNewFileName err:%v", err.Error())
			continue
		}

		oldPath := filepath.Join(dir, f.Name())
		newPath := filepath.Join(dir, newName)
		if err = os.Rename(oldPath, newPath); err != nil {
			return err
		}

		count++
		s.deps.Log.Infof("重命名第%d部: %s", count, newName)
	}

	return nil
}

func (s *FilmService) getExistingFilmNames(ctx context.Context) (map[string]int, error) {
	films, err := s.deps.FilmRepo.FindAll(ctx, consts.FilmIsNotRemoved)
	if err != nil {
		return nil, err
	}

	filmMap := make(map[string]int, len(films))
	for _, film := range films {
		filmMap[film.MovieName] = 1
	}
	return filmMap, nil
}

func isValidMovieFile(f os.FileInfo) bool {
	return strings.HasSuffix(f.Name(), VideoExt) && f.Size() >= MinFileSize
}

func (s *FilmService) generateNewFileName(ctx context.Context, dir, fileName, movieName string) (string, error) {
	movies, err := s.deps.MovieRepo.FindMoviesByName(ctx, movieName)
	if err != nil {
		return "", err
	}
	if len(movies) < 1 {
		s.deps.Log.Error("No record :", movieName)
		return "", sqlx.ErrNotFound
	}
	if len(movies) > 1 {
		s.deps.Log.Warn("More than one record :", movieName)
	}
	movie := movies[0]

	movieType, err := s.movieSvc.GetMovieType(ctx, movie.JavId)
	if err != nil {
		return "", err
	}

	fullPathName := dir + "/" + fileName
	metaData, err := s.getMetadataForFile(fullPathName)
	if err != nil {
		return "", err
	}

	subPart := getSubPart(fileName)
	compressPart := getCompressPart(fileName)
	erasedPart := getErasedPart(fileName)
	castPart, genrePart := getMovieParts(movieType)

	heightPart, bitRatePart := s.getMovieTechnicalDetails(metaData)
	movieNameUpper := strings.ToUpper(movieName)

	fullName := fmt.Sprintf("%s_%s_%s_%s_%s_%s_%s_%s_%s_%s%s",
		movieNameUpper, castPart, genrePart, movieType.Director,
		heightPart, bitRatePart, subPart, compressPart,
		erasedPart, movieType.Title, VideoExt)

	return fullName, nil
}

func (s *FilmService) getMetadataForFile(name string) (*models.Metadata, error) {
	_, metaData, err := getMetadata(name)
	return metaData, err
}

func getSubPart(fileName string) string {
	if strings.Contains(fileName, "-C") {
		return FilmNameHasSub
	}
	return FilmNameNoSub
}

func getCompressPart(fileName string) string {
	if strings.Contains(fileName, "~1") {
		return FilmNameCompress
	}
	return FilmNameNoCompress
}

func getErasedPart(fileName string) string {
	if strings.Contains(fileName, "~E") {
		return FilmNameErased
	} else if strings.Contains(fileName, "~P") {
		return FilmNameNOMosaic
	}
	return FilmNameNoErased
}

func getMovieParts(movieType *types.MovieType) (string, string) {
	var (
		castPart  strings.Builder
		genrePart strings.Builder
	)

	for i, cast := range movieType.Cast {
		if i > 0 {
			castPart.WriteString("-")
		}
		castPart.WriteString(cast.Name)
	}

	for i, genre := range movieType.Genre {
		if i > 0 {
			genrePart.WriteString("-")
		}
		genrePart.WriteString(genre)
	}

	return castPart.String(), genrePart.String()
}

func (s *FilmService) getMovieTechnicalDetails(metaData *models.Metadata) (string, string) {
	var heightPart string
	for _, stream := range metaData.Streams {
		if stream.CodecType == "video" {
			heightPart = strconv.Itoa(stream.Height)
			break
		}
	}

	bitRate, err := strconv.ParseFloat(metaData.Format.BitRate, 64)
	if err != nil {
		s.deps.Log.Errorf("Error parsing bit rate: %v", err)
		return heightPart, ""
	}
	bitRatePart := convert.FloatTo(bitRate / 1e5).DecimalStr(0)

	return heightPart, bitRatePart
}

func getMovieNameByFileRawName(fileName string) string {
	name := strings.TrimSuffix(fileName, VideoExt)
	extraSuffixes := []string{"~E", "~P", "~1", "-C"}

	for _, suffix := range extraSuffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	return name
}
