package sc

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"rudy_gc/internal/types"
	"rudy_gc/pkg/filetool"
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

func (l *ScService) StartCopyAsync(movies []*types.MovieType, source string) (bool, CopyStatus) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()

	if l.copyTask != nil && l.copyTask.running {
		return false, l.copyStatusLocked()
	}
	if len(movies) == 0 {
		return false, CopyStatus{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	l.copyTask = &copyTask{
		running:   true,
		total:     len(movies),
		startedAt: time.Now().Unix(),
		cancel:    cancel,
	}
	go l.runCopyTask(ctx, movies, source)
	return true, l.copyStatusLocked()
}

func (l *ScService) StopCopy() bool {
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

func (l *ScService) CopyStatus() CopyStatus {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	return l.copyStatusLocked()
}

func (l *ScService) copyStatusLocked() CopyStatus {
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

func (l *ScService) runCopyTask(ctx context.Context, movies []*types.MovieType, source string) {
	for _, mf := range movies {
		if ctx.Err() != nil {
			l.finishCopy(true)
			return
		}
		if mf == nil {
			l.incCopyDone("")
			continue
		}
		videoURL := SmartPickMovieVideoURL(mf, source)
		current := filepath.Base(videoURL)
		l.setCopyCurrent(current)
		if videoURL == "" {
			l.setCopyError("缺少可复制的视频路径")
			l.incCopyDone("")
			continue
		}
		if err := l.copyFileToDestinationCtx(ctx, videoURL); err != nil {
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

func (l *ScService) copyFileToDestinationCtx(ctx context.Context, srcFilePath string) error {
	destFilePath := filepath.Join(l.deps.Config.Film.CopyDestinationPath, filepath.Base(srcFilePath))
	return filetool.CopyFileWithProgressCtx(ctx, srcFilePath, destFilePath)
}

func (l *ScService) setCopyCurrent(name string) {
	l.copyMu.Lock()
	defer l.copyMu.Unlock()
	if l.copyTask == nil {
		return
	}
	l.copyTask.current = name
}

func (l *ScService) incCopyDone(errMsg string) {
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

func (l *ScService) setCopyError(errMsg string) {
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

func (l *ScService) finishCopy(stopped bool) {
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
