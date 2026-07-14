package sc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"rudy_gc/internal/dep"
	"rudy_gc/internal/platform/config"
	"rudy_gc/internal/repository/moviex"
	"rudy_gc/internal/types"
)

func TestStartFC2CopyAsyncCopiesWithFC2FileName(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	sourcePath := filepath.Join(sourceDir, "source.mp4")
	if err := os.WriteFile(sourcePath, []byte("fc2-copy"), 0o644); err != nil {
		t.Fatal(err)
	}
	service := &Service{deps: &dep.Dep{Config: config.Config{Film: config.Film{CopyDestinationPath: destinationDir}}}}
	started, _ := service.StartFC2CopyAsync([]*FC2SmartPickMovie{{
		FC2Movie: &moviex.FC2Movie{FC2Key: "FC2PPV-1", PerformerText: "演员甲", Title: "标题"},
		VideoURL: sourcePath,
	}})
	if !started {
		t.Fatal("FC2 copy task did not start")
	}
	deadline := time.Now().Add(3 * time.Second)
	for service.CopyStatus().Running && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	status := service.CopyStatus()
	if status.Running || status.Done != 1 || status.LastError != "" {
		t.Fatalf("copy status = %#v", status)
	}
	wantPath := filepath.Join(destinationDir, "FC2PPV-1_演员甲_标题.mp4")
	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(content) != "fc2-copy" {
		t.Fatalf("copied content = %q", content)
	}
}

func TestStartMixedCopyAsyncCopiesJAVAndFC2InOneTask(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	javSourcePath := filepath.Join(sourceDir, "jav-source.mp4")
	fc2SourcePath := filepath.Join(sourceDir, "fc2-source.mp4")
	if err := os.WriteFile(javSourcePath, []byte("jav-copy"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fc2SourcePath, []byte("fc2-copy"), 0o644); err != nil {
		t.Fatal(err)
	}

	service := &Service{deps: &dep.Dep{Config: config.Config{Film: config.Film{CopyDestinationPath: destinationDir}}}}
	started, initial := service.StartMixedCopyAsync(
		[]*types.MovieType{{
			Name:           "ABP-123",
			Title:          "JAV标题",
			VideoUrlWMedia: javSourcePath,
			Cast:           []*types.CastInfo{{NameShow: "演员乙"}},
		}},
		SmartPickSourceWMedia,
		[]*FC2SmartPickMovie{{
			FC2Movie: &moviex.FC2Movie{FC2Key: "FC2PPV-2", PerformerText: "演员甲", Title: "FC2标题"},
			VideoURL: fc2SourcePath,
		}},
	)
	if !started || initial.Total != 2 {
		t.Fatalf("mixed copy start = %v, status = %#v", started, initial)
	}

	deadline := time.Now().Add(3 * time.Second)
	for service.CopyStatus().Running && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	status := service.CopyStatus()
	if status.Running || status.Done != 2 || status.Total != 2 || status.LastError != "" {
		t.Fatalf("mixed copy status = %#v", status)
	}

	assertCopiedContent(t, filepath.Join(destinationDir, "ABP-123_演员乙_JAV标题.mp4"), "jav-copy")
	assertCopiedContent(t, filepath.Join(destinationDir, "FC2PPV-2_演员甲_FC2标题.mp4"), "fc2-copy")
}

func assertCopiedContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read copied file %s: %v", path, err)
	}
	if string(content) != expected {
		t.Fatalf("copied content %s = %q", path, content)
	}
}
