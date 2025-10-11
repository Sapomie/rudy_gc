package logic

import (
	"context"
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

// ---- 顶层：按条下载，失败继续，限速保持 ----
func (l *CrawlLogic) DownLoadAllPicture() error {
	items, err := l.deps.ItemRepo.FindByDownloadCoverStatus(l.ctx, types.ItemCoverNone)
	if err != nil {
		return fmt.Errorf("FindByDownloadCoverStatus: %w", err)
	}

	var done int64
	for _, it := range items {
		// 支持取消
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		default:
		}

		if err := l.DownloadPictureOfMovie(it); err != nil {
			l.deps.Log.Errorf("图片下载错误: %s (%s): %v", it.Name, it.JavId, err)
			time.Sleep(getRandomSleepDuration())
			continue
		}
		done++
		l.deps.Log.Infof("图片下载已完成 %d/%d --- %s", done, len(items), it.Name)
		time.Sleep(getRandomSleepDuration())

	}
	return nil
}

func (l *CrawlLogic) DownloadPictureOfMovie(item *types.Item) error {
	// 1) 查 Movie
	movie, err := l.deps.MovieRepo.FindOneByJavId(l.ctx, item.JavId)
	if err != nil {
		return fmt.Errorf("MovieRepo.FindOneByJavId(%s): %w", item.JavId, err)
	}

	// 2) 目录（幂等）
	imageRel, err := l.createDirectoriesForMovie(movie)
	if err != nil {
		return fmt.Errorf("createDirectoriesForMovie: %w", err)
	}

	// 3) 查 Murl
	murl, err := l.deps.MurlRepo.FindOneByJavId(l.ctx, movie.JavId)
	if err != nil {
		return fmt.Errorf("MurlRepo.FindOneByJavId(%s): %w", movie.JavId, err)
	}

	// 4) 目标文件路径（相对 & 绝对）
	base := l.deps.Config.Fetcher.LocalImageDir
	dstRel := filepath.Join(imageRel, fmt.Sprintf("%s_Jacket.jpg", murl.Name))
	dstAbs := filepath.Join(base, dstRel)

	// 即使已存在也要覆盖下载
	if st, statErr := os.Stat(dstAbs); statErr == nil && st.Size() > 0 {
		// 如果 DB 存的本地路径不一致，先对齐（不早退）
		if murl.JacketImgLocal != dstRel {
			now := time.Now().Unix()
			if err := l.deps.MurlRepo.UpdatePartialByJavId(l.ctx, movie.JavId, types.MurlPatch{
				JacketImgLocal: &dstRel,
				UpdatedOn:      &now,
			}); err != nil {
				return fmt.Errorf("MurlRepo.UpdatePartialByJavId (align existing file): %w", err)
			}
			l.movieSvc.InvalidateMovieType(l.ctx, murl.JavId)
		}
	}

	// 无论存在与否，都重新下载并原子覆盖
	if err := l.downloadToFileAtomic(l.ctx, murl.JacketImg, dstAbs); err != nil {
		return fmt.Errorf("downloadToFileAtomic: %w", err)
	}

	// 覆盖后再次对齐 DB（防止上面未对齐的情况）
	{
		now := time.Now().Unix()
		if err := l.deps.MurlRepo.UpdatePartialByJavId(l.ctx, movie.JavId, types.MurlPatch{
			JacketImgLocal: &dstRel,
			UpdatedOn:      &now,
		}); err != nil {
			return fmt.Errorf("MurlRepo.UpdatePartialByJavId: %w", err)
		}
		l.movieSvc.InvalidateMovieType(l.ctx, murl.JavId)
	}

	// 标记 item 封面 OK
	if err := l.markItemCoverOK(movie.JavId); err != nil {
		return err
	}

	return nil
}

func (l *CrawlLogic) markItemCoverOK(javId string) error {
	now := time.Now().Unix()
	if err := l.deps.ItemRepo.UpdatePartialByJavId(l.ctx, javId, types.ItemPatch{
		HasDownloadCover: ptr.Int64(types.ItemCoverOK),
		UpdatedOn:        &now,
	}); err != nil {
		return fmt.Errorf("ItemRepo.UpdatePartialByJavId(%s): %w", javId, err)
	}
	return nil
}

// 目录结构：<yyyy>/<yyyy-mm>
func (l *CrawlLogic) createDirectoriesForMovie(m *types.Movie) (string, error) {
	t := time.Unix(m.ReleasingDate, 0)
	rel := filepath.Join(strconv.Itoa(t.Year()), t.Format(formatYearMonth))
	abs := filepath.Join(l.deps.Config.Fetcher.LocalImageDir, rel)

	// 幂等创建
	if err := os.MkdirAll(abs, 0o744); err != nil {
		return "", fmt.Errorf("MkdirAll(%s): %w", abs, err)
	}
	return rel, nil
}

// 原子落盘：写入 .tmp 再 rename
func (l *CrawlLogic) downloadToFileAtomic(ctx context.Context, url, dstAbs string) error {
	resp, err := l.deps.Fetcher.Get(ctx, url)
	if err != nil {
		return fmt.Errorf("Fetcher.Get: %w", err)
	}
	if len(resp.Body) == 0 {
		return errors.New("empty response body")
	}

	tmp := dstAbs + ".tmp"
	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o744); err != nil {
		return fmt.Errorf("MkdirAll(parent): %w", err)
	}

	// 写临时文件
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("Create tmp: %w", err)
	}
	// 尽量避免延迟关闭导致泄露
	_, werr := f.Write(resp.Body)
	cerr := f.Close()
	if werr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("write tmp: %w", werr)
	}
	if cerr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close tmp: %w", cerr)
	}

	// 原子替换
	if err := os.Rename(tmp, dstAbs); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename tmp->dst: %w", err)
	}
	return nil
}
