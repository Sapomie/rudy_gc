package sc

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

type deleteEventModelStub struct {
	moviex.ScEventWriteModel
	result *moviex.ScEventDeleteResult
	err    error
	name   string
}

func (s *deleteEventModelStub) DeleteEvent(_ context.Context, scName string, _ int64) (*moviex.ScEventDeleteResult, error) {
	s.name = scName
	return s.result, s.err
}

func TestDeleteEventReturnsDeletedEventSummary(t *testing.T) {
	model := &deleteEventModelStub{
		result: &moviex.ScEventDeleteResult{
			Event:  &moviex.GSc{Name: "2026-07-28-12-00"},
			Movies: []*moviex.GList{{Id: 1}, {Id: 2}},
		},
	}
	service := NewService(&dep.Dep{ScEventWriteModel: model})

	result, err := service.DeleteEvent(context.Background(), " 2026-07-28-12-00 ")
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if model.name != "2026-07-28-12-00" {
		t.Fatalf("normalized name = %q", model.name)
	}
	if result == nil || result.Name != model.name || result.DeletedMovieCount != 2 || !result.DerivedDataFresh {
		t.Fatalf("result = %#v", result)
	}
}

func TestDeleteEventMapsNotFound(t *testing.T) {
	model := &deleteEventModelStub{err: moviex.ErrNotFound}
	service := NewService(&dep.Dep{ScEventWriteModel: model})

	result, err := service.DeleteEvent(context.Background(), "missing")
	if !errors.Is(err, types.ErrNotFound) {
		t.Fatalf("DeleteEvent error = %v", err)
	}
	if result != nil {
		t.Fatalf("DeleteEvent result = %#v", result)
	}
}

func TestDeleteEventArchivesEventImage(t *testing.T) {
	activeDir := t.TempDir()
	deletedDir := t.TempDir()
	sourcePath := filepath.Join(activeDir, "2026-07-28-12-00.jpg")
	if err := os.WriteFile(sourcePath, []byte("event-image"), 0o644); err != nil {
		t.Fatalf("write event image: %v", err)
	}
	model := &deleteEventModelStub{result: &moviex.ScEventDeleteResult{
		Event: &moviex.GSc{Id: 42, Name: "2026-07-28-12-00", ImagePath: sourcePath},
	}}
	service := NewService(&dep.Dep{
		ScEventWriteModel: model,
		Config: config.Config{SCEvent: config.SCEventConfig{
			ActiveImageDir:  activeDir,
			DeletedImageDir: deletedDir,
		}},
	})

	result, err := service.DeleteEvent(context.Background(), "2026-07-28-12-00")
	if err != nil {
		t.Fatalf("DeleteEvent: %v", err)
	}
	if result == nil || !result.DerivedDataFresh {
		t.Fatalf("DeleteEvent result = %#v", result)
	}
	if _, err := os.Stat(filepath.Join(deletedDir, filepath.Base(sourcePath))); err != nil {
		t.Fatalf("stat archived event image: %v", err)
	}
}

func TestDeleteEventReportsCommittedDeleteWithCacheFailure(t *testing.T) {
	model := &deleteEventModelStub{
		result: &moviex.ScEventDeleteResult{Event: &moviex.GSc{Name: "2026-07-28-12-00"}},
		err:    errors.New("cache unavailable"),
	}
	service := NewService(&dep.Dep{ScEventWriteModel: model})

	result, err := service.DeleteEvent(context.Background(), "2026-07-28-12-00")
	if err == nil || !strings.Contains(err.Error(), "已删除") {
		t.Fatalf("DeleteEvent error = %v", err)
	}
	if result == nil || result.Name != "2026-07-28-12-00" || result.DerivedDataFresh {
		t.Fatalf("DeleteEvent result = %#v", result)
	}
}

func TestArchiveDeletedScEventImageMovesFileAndCompanion(t *testing.T) {
	activeDir := filepath.Join(t.TempDir(), "event_image")
	deletedDir := filepath.Join(t.TempDir(), "deleted_event_image")
	if err := os.MkdirAll(activeDir, 0o755); err != nil {
		t.Fatalf("mkdir active image dir: %v", err)
	}
	sourcePath := filepath.Join(activeDir, "2026-07-28-12-00.png")
	companionPath := filepath.Join(activeDir, "._2026-07-28-12-00.png")
	if err := os.WriteFile(sourcePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("write event image: %v", err)
	}
	if err := os.WriteFile(companionPath, []byte("metadata"), 0o644); err != nil {
		t.Fatalf("write event image companion: %v", err)
	}
	service := NewService(&dep.Dep{Config: config.Config{SCEvent: config.SCEventConfig{
		ActiveImageDir:  activeDir,
		DeletedImageDir: deletedDir,
	}}})

	archivedPath, err := service.archiveDeletedScEventImage(&moviex.GSc{Id: 42, ImagePath: sourcePath})
	if err != nil {
		t.Fatalf("archiveDeletedScEventImage: %v", err)
	}
	wantPath := filepath.Join(deletedDir, filepath.Base(sourcePath))
	if archivedPath != filepath.ToSlash(wantPath) {
		t.Fatalf("archived path = %q, want %q", archivedPath, filepath.ToSlash(wantPath))
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("stat archived image: %v", err)
	}
	if _, err := os.Stat(filepath.Join(deletedDir, filepath.Base(companionPath))); err != nil {
		t.Fatalf("stat archived companion: %v", err)
	}
	if _, err := os.Stat(sourcePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source image still exists: %v", err)
	}
}

func TestArchiveDeletedScEventImageUsesEventIDOnNameConflict(t *testing.T) {
	activeDir := t.TempDir()
	deletedDir := t.TempDir()
	sourcePath := filepath.Join(activeDir, "2026-07-28-12-00.jpg")
	if err := os.WriteFile(sourcePath, []byte("new"), 0o644); err != nil {
		t.Fatalf("write source image: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deletedDir, filepath.Base(sourcePath)), []byte("old"), 0o644); err != nil {
		t.Fatalf("write existing archived image: %v", err)
	}
	service := NewService(&dep.Dep{Config: config.Config{SCEvent: config.SCEventConfig{
		ActiveImageDir:  activeDir,
		DeletedImageDir: deletedDir,
	}}})

	archivedPath, err := service.archiveDeletedScEventImage(&moviex.GSc{Id: 42, ImagePath: sourcePath})
	if err != nil {
		t.Fatalf("archiveDeletedScEventImage: %v", err)
	}
	wantPath := filepath.Join(deletedDir, "2026-07-28-12-00__gsc_42.jpg")
	if archivedPath != filepath.ToSlash(wantPath) {
		t.Fatalf("archived path = %q, want %q", archivedPath, filepath.ToSlash(wantPath))
	}
}

func TestArchiveDeletedScEventImageRejectsPathOutsideActiveDir(t *testing.T) {
	activeDir := t.TempDir()
	deletedDir := t.TempDir()
	sourcePath := filepath.Join(t.TempDir(), "outside.jpg")
	if err := os.WriteFile(sourcePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("write outside image: %v", err)
	}
	service := NewService(&dep.Dep{Config: config.Config{SCEvent: config.SCEventConfig{
		ActiveImageDir:  activeDir,
		DeletedImageDir: deletedDir,
	}}})

	if _, err := service.archiveDeletedScEventImage(&moviex.GSc{Id: 42, ImagePath: sourcePath}); err == nil || !strings.Contains(err.Error(), "不在当前图片目录") {
		t.Fatalf("archiveDeletedScEventImage error = %v", err)
	}
	if _, err := os.Stat(sourcePath); err != nil {
		t.Fatalf("outside image changed: %v", err)
	}
}
