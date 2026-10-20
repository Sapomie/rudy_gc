package sc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/local/api/internal/types"
	"rudy_gc/local/api/pkg/filetool"
)

type copyTask struct {
	running    bool
	total      int
	done       int
	current    string
	startedAt  int64
	finishedAt int64
	stopped    bool
	lastErr    string
	cancel     context.CancelFunc
}

type CopyStatus struct {
	Running    bool   `json:"running"`
	Total      int    `json:"total"`
	Done       int    `json:"done"`
	Current    string `json:"current"`
	StartedAt  int64  `json:"started_at"`
	FinishedAt int64  `json:"finished_at"`
	Stopped    bool   `json:"stopped"`
	LastError  string `json:"last_error"`
}

type SmartPickCopyItem struct {
	SourcePath string
	FileName   string
}

func (l *Service) StartCopyAsync(movies []*types.MovieType, source string) (bool, CopyStatus) {
	return l.StartCopyItemsAsync(l.smartPickCopyItems(movies, source))
}

func (l *Service) smartPickCopyItems(movies []*types.MovieType, source string) []SmartPickCopyItem {
	items := make([]SmartPickCopyItem, 0, len(movies))
	for _, movie := range movies {
		videoURL := SmartPickMovieVideoURL(movie, source)
		items = append(items, SmartPickCopyItem{
			SourcePath: videoURL,
			FileName:   l.smartPickCopyFileName(movie, videoURL, source),
		})
	}
	return items
}

func (l *Service) StartFC2CopyAsync(movies []*FC2SmartPickMovie) (bool, CopyStatus) {
	return l.StartCopyItemsAsync(l.smartPickFC2CopyItems(movies))
}

func (l *Service) smartPickFC2CopyItems(movies []*FC2SmartPickMovie) []SmartPickCopyItem {
	items := make([]SmartPickCopyItem, 0, len(movies))
	for _, movie := range movies {
		sourcePath := ""
		if movie != nil {
			sourcePath = movie.VideoURL
		}
		items = append(items, SmartPickCopyItem{
			SourcePath: sourcePath,
			FileName:   smartPickFC2CopyFileName(movie, sourcePath),
		})
	}
	return items
}

func (l *Service) StartMixedCopyAsync(javMovies []*types.MovieType, source string, fc2Movies []*FC2SmartPickMovie) (bool, CopyStatus) {
	items := l.smartPickCopyItems(javMovies, source)
	items = append(items, l.smartPickFC2CopyItems(fc2Movies)...)
	return l.StartCopyItemsAsync(items)
}

func (l *Service) StartCopyItemsAsync(items []SmartPickCopyItem) (bool, CopyStatus) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()

	if l.copyTask != nil && l.copyTask.running {
		return false, l.copyStatusLocked()
	}
	if len(items) == 0 {
		return false, CopyStatus{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.copyTask = &copyTask{
		running:   true,
		total:     len(items),
		startedAt: time.Now().Unix(),
		cancel:    cancel,
	}
	go l.runCopyTask(ctx, items)
	return true, l.copyStatusLocked()
}

func (l *Service) StopCopy() bool {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()

	if l.copyTask == nil || !l.copyTask.running {
		return false
	}
	if l.copyTask.cancel != nil {
		l.copyTask.cancel()
	}
	return true
}

func (l *Service) CopyStatus() CopyStatus {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	return l.copyStatusLocked()
}

func (l *Service) copyStatusLocked() CopyStatus {
	if l.copyTask == nil {
		return CopyStatus{}
	}
	t := l.copyTask
	return CopyStatus{
		Running:    t.running,
		Total:      t.total,
		Done:       t.done,
		Current:    t.current,
		StartedAt:  t.startedAt,
		FinishedAt: t.finishedAt,
		Stopped:    t.stopped,
		LastError:  t.lastErr,
	}
}

func (l *Service) runCopyTask(ctx context.Context, items []SmartPickCopyItem) {
	for _, item := range items {
		if ctx.Err() != nil {
			l.finishCopy(true)
			return
		}
		videoURL := strings.TrimSpace(item.SourcePath)
		current := strings.TrimSpace(item.FileName)
		if strings.TrimSpace(current) == "" {
			current = filepath.Base(videoURL)
		}
		l.setCopyCurrent(current)
		if videoURL == "" {
			l.setCopyError("缺少可复制的视频路径")
			l.incCopyDone("")
			continue
		}
		if err := l.copyFileToDestinationCtx(ctx, item); err != nil {
			if errors.Is(err, context.Canceled) {
				l.finishCopy(true)
				return
			}
			l.deps.Log.Error("copy err: ", err)
			l.setCopyError(err.Error())
		}
		l.incCopyDone("")
	}
	l.finishCopy(false)
}

func (l *Service) copyFileToDestinationCtx(ctx context.Context, item SmartPickCopyItem) error {
	destDir := strings.TrimSpace(l.deps.Config.Film.CopyDestinationPath)
	if destDir == "" {
		return fmt.Errorf("copy destination path is empty")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create copy destination dir failed: %w", err)
	}

	srcFilePath := strings.TrimSpace(item.SourcePath)
	fileName := strings.TrimSpace(item.FileName)
	if fileName == "" {
		fileName = filepath.Base(srcFilePath)
	}
	destFilePath, err := nextAvailableTargetPath(destDir, fileName)
	if err != nil {
		return err
	}
	return filetool.CopyFileWithProgressCtx(ctx, srcFilePath, destFilePath)
}

func nextAvailableTargetPath(destDir, fileName string) (string, error) {
	destDir = filepath.Clean(strings.TrimSpace(destDir))
	fileName = strings.TrimSpace(fileName)
	if destDir == "" {
		return "", fmt.Errorf("empty destination dir")
	}
	if fileName == "" {
		return "", fmt.Errorf("empty file name")
	}

	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)
	for i := 1; i <= 10000; i++ {
		candidateName := fileName
		if i > 1 {
			candidateName = fmt.Sprintf("%s_%d%s", base, i, ext)
		}
		candidatePath := filepath.Join(destDir, candidateName)
		_, err := os.Stat(candidatePath)
		if err == nil {
			continue
		}
		if os.IsNotExist(err) {
			return candidatePath, nil
		}
		return "", fmt.Errorf("stat destination file failed: %w", err)
	}
	return "", fmt.Errorf("too many duplicated target file names: %s", fileName)
}

func (l *Service) setCopyCurrent(name string) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	if l.copyTask == nil {
		return
	}
	l.copyTask.current = name
}

func (l *Service) incCopyDone(errMsg string) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	if l.copyTask == nil {
		return
	}
	l.copyTask.done++
	if errMsg != "" {
		l.copyTask.lastErr = errMsg
	}
}

func (l *Service) setCopyError(errMsg string) {
	if errMsg == "" {
		return
	}
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	if l.copyTask == nil {
		return
	}
	l.copyTask.lastErr = errMsg
}

func (l *Service) finishCopy(stopped bool) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	if l.copyTask == nil {
		return
	}
	l.copyTask.running = false
	l.copyTask.stopped = stopped
	l.copyTask.finishedAt = time.Now().Unix()
	l.copyTask.current = ""
}
