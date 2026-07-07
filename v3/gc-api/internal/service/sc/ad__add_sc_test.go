package sc

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildAddScPreviewAllowsImageOnly(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "2026-06-25-18-30")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir sc dir: %v", err)
	}
	imagePath := filepath.Join(dir, "IMG_2962.jpg")
	if err := os.WriteFile(imagePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("write image: %v", err)
	}

	svc := &ScService{}
	preview, imageFilePath, err := svc.buildAddScPreview(context.Background(), dir)
	if err != nil {
		t.Fatalf("buildAddScPreview returned error: %v", err)
	}
	if preview == nil {
		t.Fatal("preview is nil")
	}
	if preview.MovieCount != 0 {
		t.Fatalf("MovieCount = %d, want 0", preview.MovieCount)
	}
	if len(preview.Movies) != 0 {
		t.Fatalf("len(Movies) = %d, want 0", len(preview.Movies))
	}
	if !preview.ImageFound {
		t.Fatal("ImageFound = false, want true")
	}
	if preview.ImageName != "IMG_2962.jpg" {
		t.Fatalf("ImageName = %q, want IMG_2962.jpg", preview.ImageName)
	}
	if imageFilePath != imagePath {
		t.Fatalf("imageFilePath = %q, want %q", imageFilePath, imagePath)
	}
}
