package logic

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"rudy_gc/internal/types"
	"rudy_gc/pkg/ptr"
	"strconv"
	"time"
)

const formatYearMonth = "2006-01"

func (l *CrawlLogic) DownLoadAllPicture() error {
	items, err := l.deps.ItemRepo.FindByDownloadCoverStatus(l.ctx, types.ItemCoverNone)
	if err != nil {
		return fmt.Errorf("FindByDownloadCoverStatus err: %w", err)
	}

	var count int64
	for _, item := range items {
		err := l.DownloadPictureOfMovie(item)
		if err != nil {
			l.deps.Log.Error("图片下载错误：", item.Name, err.Error())
			time.Sleep(getRandomSleepDuration())
			continue
		}
		count++
		l.deps.Log.Infof("图片下载已完成 %d/%d --- %s", count, len(items), item.Name)
		time.Sleep(getRandomSleepDuration())

		if count >= 10 {
			break
		}
	}

	return nil
}

func (l *CrawlLogic) DownloadPictureOfMovie(item *types.Item) error {
	movie, err := l.deps.MovieRepo.FindOneByJavId(l.ctx, item.JavId)
	if err != nil {
		return fmt.Errorf("FindOneByJavId err: %w", err)
	}
	imagePath, err := l.createDirectoriesForMovie(movie)
	if err != nil {
		return errors.New("createDirectoriesForMovie err: " + err.Error())
	}

	murl, err := l.deps.MurlRepo.FindOneByJavId(l.ctx, movie.JavId)
	if err != nil {
		return fmt.Errorf("FindOneByJavId err: %w", err)
	}

	err = l.downloadAndSaveImages(murl, imagePath)
	if err != nil {
		return errors.New("downloadAndSaveImages err: " + err.Error())
	}

	err = l.updateMovieAndMurlInfo(murl, imagePath)
	return err
}

func (l *CrawlLogic) createDirectoriesForMovie(m *types.Movie) (string, error) {
	t := time.Unix(m.ReleasingDate, 0)
	firstDir := strconv.Itoa(t.Year())
	secondDir := t.Format(formatYearMonth)
	imagePath := filepath.Join(firstDir, secondDir)
	fullPath := filepath.Join(l.deps.Config.Fetcher.LocalImageDir, imagePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(fullPath, 0744); err != nil {
			return "", err
		}
	}
	return imagePath, nil
}

func (l *CrawlLogic) downloadAndSaveImages(murl *types.Murl, imagePath string) error {
	jacketNameFull := getFullPath(l.deps.Config.Fetcher.LocalImageDir, imagePath, murl.Name, "Jacket.jpg")

	if err := l.downloadPicture(murl.JacketImg, jacketNameFull); err != nil {
		l.deps.Log.Error(murl.Name, err)
		return errors.New("downloadPicture err:" + err.Error())
	}

	murl.JacketImgLocal = jacketNameFull
	err := l.deps.MurlRepo.UpsertByJavIdPreserveLocal(l.ctx, murl)
	if err != nil {
		return errors.New("murlRepo.UpsertByJavIdPreserveLocal err:" + err.Error())
	}
	l.movieSvc.InvalidateMovieType(l.ctx, murl.JavId)

	return nil
}

func getFullPath(basePath, imagePath, title, suffix string) string {
	return filepath.Join(basePath, imagePath, fmt.Sprintf("%s_%s", title, suffix))
}

func (l *CrawlLogic) downloadPicture(url, filepath string) error {
	resp, err := l.deps.Fetcher.Get(l.ctx, url)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(resp.Body); err != nil {
		return err
	}
	return nil
}

func (l *CrawlLogic) updateMovieAndMurlInfo(murl *types.Murl, imagePath string) error {
	murl.JacketImgLocal = filepath.Join(imagePath, fmt.Sprintf("%s_Jacket.jpg", murl.Name))

	now := time.Now().Unix()
	err := l.deps.ItemRepo.UpdatePartialByJavId(l.ctx, murl.JavId, types.ItemPatch{
		HasDownloadCover: ptr.Int64(types.ItemCoverOK),
		UpdatedOn:        &now,
	})
	if err != nil {
		return errors.New("UpdatePartialByJavId err:" + err.Error())
	}

	return nil
}
