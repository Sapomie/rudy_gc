package sc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	moviedomain "rudy_gc/internal/domain/movie"
	"rudy_gc/internal/repository/moviex"
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

func TestNormalizeAddScMovieRef(t *testing.T) {
	tests := []struct {
		name        string
		movieType   string
		movieJavID  string
		fc2MovieID  int64
		expectedRef string
	}{
		{name: "jav", movieType: moviedomain.SehuatangThreadMovieTypeJAV, movieJavID: "ABP-123", expectedRef: "jav:ABP-123"},
		{name: "fc2", movieType: moviedomain.SehuatangThreadMovieTypeFC2PPV, fc2MovieID: 42, expectedRef: "fc2ppv:42"},
		{name: "mixed identity", movieType: moviedomain.SehuatangThreadMovieTypeFC2PPV, movieJavID: "ABP-123", fc2MovieID: 42},
		{name: "missing jav", movieType: moviedomain.SehuatangThreadMovieTypeJAV},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizeAddScMovieRef(test.movieType, test.movieJavID, test.fc2MovieID); got != test.expectedRef {
				t.Fatalf("normalizeAddScMovieRef() = %q, want %q", got, test.expectedRef)
			}
		})
	}
}

func TestBuildAddScMovieSelectionMapUsesTypedIdentity(t *testing.T) {
	selection := buildAddScMovieSelectionMap([]AddScInputMovie{
		{MovieType: moviedomain.SehuatangThreadMovieTypeJAV, MovieJavId: "ABP-123", IsSc: 1},
		{MovieType: moviedomain.SehuatangThreadMovieTypeFC2PPV, FC2MovieId: 42, IsSc: 2},
	})
	if selection["jav:ABP-123"] != 1 {
		t.Fatalf("JAV selection = %d", selection["jav:ABP-123"])
	}
	if selection["fc2ppv:42"] != 2 {
		t.Fatalf("FC2 selection = %d", selection["fc2ppv:42"])
	}
}

func TestNormalizeScAddFC2KeySupportsSmartPickCopyName(t *testing.T) {
	copyName := smartPickFC2CopyFileName(&FC2SmartPickMovie{
		FC2Movie: &moviex.FC2Movie{
			FC2Key:        "FC2PPV-3347918-1",
			PerformerText: "石川祐奈",
			Title:         "影片标题",
		},
	}, "/media/source.mp4")

	key, ok := normalizeScAddFC2Key(copyName)
	if !ok || key != "FC2PPV-3347918-1" {
		t.Fatalf("normalizeScAddFC2Key(%q) = %q, %v", copyName, key, ok)
	}
}

func TestNormalizeScAddFC2KeyKeepsEncodedMediaNameSupport(t *testing.T) {
	key, ok := normalizeScAddFC2Key("15122252531-3347918-1_sub_self_nomsk.mp4")
	if !ok || key != "FC2PPV-3347918-1" {
		t.Fatalf("normalizeScAddFC2Key() = %q, %v", key, ok)
	}
}

func TestAddFC2PreviewCandidateKeepsCardMovie(t *testing.T) {
	movie := &moviex.FC2Movie{ID: 42, FC2Key: "FC2PPV-4556413-1", Title: "FC2 标题"}
	candidates := make(map[string]*AddScPreviewMovie)

	addFC2PreviewCandidate(candidates, &moviex.FC2Media{FC2MovieID: movie.ID}, map[int64]*moviex.FC2Movie{movie.ID: movie})

	candidate := candidates["fc2ppv:42"]
	if candidate == nil {
		t.Fatal("FC2 preview candidate is nil")
	}
	if candidate.FC2Movie != movie {
		t.Fatal("FC2 preview candidate did not keep card movie")
	}
}
