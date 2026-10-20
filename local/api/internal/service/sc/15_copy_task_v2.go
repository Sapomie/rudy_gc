package sc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"rudy_gc/local/api/internal/types"
	"rudy_gc/local/api/pkg/filetool"
)

const (
	CopyTaskV2StateRunning   = "running"
	CopyTaskV2StateStopping  = "stopping"
	CopyTaskV2StateCompleted = "completed"
	CopyTaskV2StateFailed    = "failed"
	CopyTaskV2StateStopped   = "stopped"
)

type CopyTaskV2Failure struct {
	FileName string `json:"file_name"`
	Error    string `json:"error"`
}

type CopyTaskV2Status struct {
	JobID         string              `json:"job_id"`
	State         string              `json:"state"`
	Total         int                 `json:"total"`
	Done          int                 `json:"done"`
	Success       int                 `json:"success"`
	Failed        int                 `json:"failed"`
	Skipped       int                 `json:"skipped"`
	Current       string              `json:"current"`
	TotalBytes    int64               `json:"total_bytes"`
	CopiedBytes   int64               `json:"copied_bytes"`
	StartedAt     int64               `json:"started_at"`
	FinishedAt    int64               `json:"finished_at"`
	StopRequested bool                `json:"stop_requested"`
	Failures      []CopyTaskV2Failure `json:"failures"`
}

type copyTaskV2 struct {
	status CopyTaskV2Status
	cancel context.CancelFunc
}

func (l *Service) CreateMixedCopyPreview(javMovies []*types.MovieType, source string, fc2Movies []*FC2SmartPickMovie) (string, error) {
	items := l.smartPickCopyItems(javMovies, source)
	items = append(items, l.smartPickFC2CopyItems(fc2Movies)...)
	if len(items) == 0 {
		return "", errors.New("copy preview is empty")
	}
	previewID, err := newCopyV2ID()
	if err != nil {
		return "", err
	}
	l.copyV2Mu.Lock()
	l.ensureCopyV2MapsLocked()
	l.copyV2Preview[previewID] = append([]SmartPickCopyItem(nil), items...)
	l.copyV2Mu.Unlock()
	return previewID, nil
}

func (l *Service) StartCopyV2(previewID string) (CopyTaskV2Status, error) {
	previewID = strings.TrimSpace(previewID)
	if previewID == "" {
		return CopyTaskV2Status{}, errors.New("preview_id is required")
	}

	l.copyV2Mu.Lock()
	l.ensureCopyV2MapsLocked()
	items, ok := l.copyV2Preview[previewID]
	if !ok {
		l.copyV2Mu.Unlock()
		return CopyTaskV2Status{}, errors.New("copy preview not found or already started")
	}
	jobID, err := newCopyV2ID()
	if err != nil {
		l.copyV2Mu.Unlock()
		return CopyTaskV2Status{}, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	status := CopyTaskV2Status{
		JobID:      jobID,
		State:      CopyTaskV2StateRunning,
		Total:      len(items),
		TotalBytes: copyV2TotalBytes(items),
		StartedAt:  time.Now().Unix(),
		Failures:   []CopyTaskV2Failure{},
	}
	task := &copyTaskV2{status: status, cancel: cancel}
	l.copyV2Tasks[jobID] = task
	delete(l.copyV2Preview, previewID)
	l.copyV2Mu.Unlock()

	go l.runCopyTaskV2(ctx, jobID, append([]SmartPickCopyItem(nil), items...))
	return status, nil
}

func (l *Service) CopyV2Status(jobID string) (CopyTaskV2Status, bool) {
	l.copyV2Mu.RLock()
	defer l.copyV2Mu.RUnlock()
	task, ok := l.copyV2Tasks[strings.TrimSpace(jobID)]
	if !ok || task == nil {
		return CopyTaskV2Status{}, false
	}
	return cloneCopyTaskV2Status(task.status), true
}

func (l *Service) StopCopyV2(jobID string) (CopyTaskV2Status, bool) {
	l.copyV2Mu.Lock()
	defer l.copyV2Mu.Unlock()
	task, ok := l.copyV2Tasks[strings.TrimSpace(jobID)]
	if !ok || task == nil {
		return CopyTaskV2Status{}, false
	}
	if task.status.State == CopyTaskV2StateRunning {
		task.status.State = CopyTaskV2StateStopping
		task.status.StopRequested = true
		if task.cancel != nil {
			task.cancel()
		}
	}
	return cloneCopyTaskV2Status(task.status), true
}

func (l *Service) runCopyTaskV2(ctx context.Context, jobID string, items []SmartPickCopyItem) {
	var completedBytes int64
	for _, item := range items {
		if ctx.Err() != nil {
			l.finishCopyTaskV2(jobID, CopyTaskV2StateStopped)
			return
		}

		fileName := strings.TrimSpace(item.FileName)
		if fileName == "" {
			fileName = filepath.Base(strings.TrimSpace(item.SourcePath))
		}
		l.updateCopyTaskV2(jobID, func(status *CopyTaskV2Status) {
			status.Current = fileName
			status.CopiedBytes = completedBytes
		})

		if strings.TrimSpace(item.SourcePath) == "" {
			l.failCopyTaskV2Item(jobID, fileName, "缺少可复制的视频路径", completedBytes)
			continue
		}

		targetPath, release, err := l.reserveCopyV2Target(fileName)
		if err != nil {
			l.failCopyTaskV2Item(jobID, fileName, err.Error(), completedBytes)
			continue
		}
		err = filetool.CopyFileAtomicCtx(ctx, item.SourcePath, targetPath, func(copied, _ int64) {
			l.updateCopyTaskV2(jobID, func(status *CopyTaskV2Status) {
				status.CopiedBytes = completedBytes + copied
			})
		})
		release()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				l.finishCopyTaskV2(jobID, CopyTaskV2StateStopped)
				return
			}
			l.deps.Log.Error("scpickv2 copy err: ", err)
			l.failCopyTaskV2Item(jobID, fileName, err.Error(), completedBytes)
			continue
		}

		if info, statErr := os.Stat(targetPath); statErr == nil {
			completedBytes += info.Size()
		}
		l.updateCopyTaskV2(jobID, func(status *CopyTaskV2Status) {
			status.Success++
			status.Done++
			status.CopiedBytes = completedBytes
		})
	}

	state := CopyTaskV2StateCompleted
	if status, ok := l.CopyV2Status(jobID); ok && status.Failed > 0 {
		state = CopyTaskV2StateFailed
	}
	l.finishCopyTaskV2(jobID, state)
}

func (l *Service) reserveCopyV2Target(fileName string) (string, func(), error) {
	destDir := strings.TrimSpace(l.deps.Config.Film.CopyDestinationPath)
	if destDir == "" {
		return "", func() {}, errors.New("copy destination path is empty")
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", func() {}, fmt.Errorf("create copy destination dir failed: %w", err)
	}
	fileName = filepath.Base(strings.TrimSpace(fileName))
	if fileName == "" || fileName == "." {
		return "", func() {}, errors.New("empty copy file name")
	}
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	ext := filepath.Ext(fileName)
	for sequence := 1; ; sequence++ {
		candidateName := fileName
		if sequence > 1 {
			candidateName = fmt.Sprintf("%s_%d%s", base, sequence, ext)
		}
		candidatePath := filepath.Join(destDir, candidateName)
		if _, err := os.Lstat(candidatePath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", func() {}, fmt.Errorf("stat destination file failed: %w", err)
		}
		lockPath := candidatePath + ".scpickv2.lock"
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", func() {}, fmt.Errorf("reserve destination file failed: %w", err)
		}
		_ = lockFile.Close()
		if _, err := os.Lstat(candidatePath); err == nil {
			_ = os.Remove(lockPath)
			continue
		} else if !os.IsNotExist(err) {
			_ = os.Remove(lockPath)
			return "", func() {}, fmt.Errorf("stat reserved destination file failed: %w", err)
		}
		return candidatePath, func() { _ = os.Remove(lockPath) }, nil
	}
}

func (l *Service) failCopyTaskV2Item(jobID, fileName, message string, completedBytes int64) {
	l.updateCopyTaskV2(jobID, func(status *CopyTaskV2Status) {
		status.Failed++
		status.Done++
		status.CopiedBytes = completedBytes
		status.Failures = append(status.Failures, CopyTaskV2Failure{FileName: fileName, Error: message})
	})
}

func (l *Service) finishCopyTaskV2(jobID, state string) {
	l.updateCopyTaskV2(jobID, func(status *CopyTaskV2Status) {
		status.State = state
		status.Current = ""
		status.FinishedAt = time.Now().Unix()
	})
}

func (l *Service) updateCopyTaskV2(jobID string, update func(*CopyTaskV2Status)) {
	l.copyV2Mu.Lock()
	defer l.copyV2Mu.Unlock()
	task := l.copyV2Tasks[jobID]
	if task == nil {
		return
	}
	update(&task.status)
}

func (l *Service) ensureCopyV2MapsLocked() {
	if l.copyV2Preview == nil {
		l.copyV2Preview = make(map[string][]SmartPickCopyItem)
	}
	if l.copyV2Tasks == nil {
		l.copyV2Tasks = make(map[string]*copyTaskV2)
	}
}

func copyV2TotalBytes(items []SmartPickCopyItem) int64 {
	var total int64
	for _, item := range items {
		if info, err := os.Stat(strings.TrimSpace(item.SourcePath)); err == nil {
			total += info.Size()
		}
	}
	return total
}

func cloneCopyTaskV2Status(status CopyTaskV2Status) CopyTaskV2Status {
	status.Failures = append([]CopyTaskV2Failure(nil), status.Failures...)
	return status
}

func newCopyV2ID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate copy id failed: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
