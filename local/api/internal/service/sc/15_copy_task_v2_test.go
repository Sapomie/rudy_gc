package sc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/types"
)

func TestCopyV2RunsIndependentJobsAndReservesDuplicateNames(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := t.TempDir()
	sourcePath := filepath.Join(sourceDir, "source.mp4")
	if err := os.WriteFile(sourcePath, []byte("scpickv2"), 0o644); err != nil {
		t.Fatal(err)
	}
	service := &Service{deps: &dep.Dep{Config: config.Config{Film: config.Film{CopyDestinationPath: destinationDir}}}}
	movie := []*types.MovieType{{Name: "ABC-123", Title: "标题", VideoUrlWMedia: sourcePath}}
	previewOne, err := service.CreateMixedCopyPreview(movie, SmartPickSourceWMedia, nil)
	if err != nil {
		t.Fatal(err)
	}
	previewTwo, err := service.CreateMixedCopyPreview(movie, SmartPickSourceWMedia, nil)
	if err != nil {
		t.Fatal(err)
	}
	jobOne, err := service.StartCopyV2(previewOne)
	if err != nil {
		t.Fatal(err)
	}
	jobTwo, err := service.StartCopyV2(previewTwo)
	if err != nil {
		t.Fatal(err)
	}
	finalOne := waitCopyV2(t, service, jobOne.JobID)
	finalTwo := waitCopyV2(t, service, jobTwo.JobID)
	if finalOne.State != CopyTaskV2StateCompleted || finalTwo.State != CopyTaskV2StateCompleted {
		t.Fatalf("states = %q, %q", finalOne.State, finalTwo.State)
	}
	if finalOne.Success != 1 || finalTwo.Success != 1 || finalOne.Failed != 0 || finalTwo.Failed != 0 {
		t.Fatalf("statuses = %#v, %#v", finalOne, finalTwo)
	}
	entries, err := os.ReadDir(destinationDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("destination entries = %d", len(entries))
	}
}

func waitCopyV2(t *testing.T, service *Service, jobID string) CopyTaskV2Status {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		status, ok := service.CopyV2Status(jobID)
		if !ok {
			t.Fatalf("job %s disappeared", jobID)
		}
		if status.State != CopyTaskV2StateRunning && status.State != CopyTaskV2StateStopping {
			return status
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish", jobID)
	return CopyTaskV2Status{}
}
