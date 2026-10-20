package sc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestScAddJAVMediaMatchesFileNameCaseInsensitively(t *testing.T) {
	for _, storedName := range []string{"OLM-044N", "OLM-044n", " olm-044N "} {
		t.Run(storedName, func(t *testing.T) {
			media := &moviex.WMedia{MovieName: storedName, MovieJavId: "OLM-044N"}
			index := indexScAddJAVMedia([]*moviex.WMedia{nil, media})
			for _, fileName := range []string{
				"OLM-044n_武藤绫香(42.9)-七瀬いおり_以年下男子为目标的美味人妻身材苗条的巨乳！！.mp4",
				"OLM-044N_影片标题.mp4",
				"olm-044n_影片标题.mp4",
			} {
				key := normalizeScAddJAVKey(extractMovieName(fileName))
				if key != "OLM-044N" {
					t.Fatalf("query key for %q = %q, want OLM-044N", fileName, key)
				}
				if got := index[key]; got != media {
					t.Fatalf("media for %q = %v, want %v", fileName, got, media)
				}
			}
			if index[normalizeScAddJAVKey("OLM-044")] != nil {
				t.Fatal("different serial without letter suffix matched")
			}
			if media.MovieName != storedName {
				t.Fatalf("display name changed: %q, want %q", media.MovieName, storedName)
			}
		})
	}
}

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

	svc := &Service{}
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

func TestCopyScImageUsesSCEventImageDir(t *testing.T) {
	sourceDir := t.TempDir()
	destinationDir := filepath.Join(t.TempDir(), "sc", "event_image")
	sourcePath := filepath.Join(sourceDir, "source.PNG")
	if err := os.WriteFile(sourcePath, []byte("event-image"), 0o644); err != nil {
		t.Fatalf("write source image: %v", err)
	}

	service := &Service{deps: &dep.Dep{Config: config.Config{
		SCEvent: config.SCEventConfig{ActiveImageDir: destinationDir},
	}}}
	got, err := service.copyScImage(sourcePath, "2026-07-28-14-45")
	if err != nil {
		t.Fatalf("copyScImage returned error: %v", err)
	}
	want := filepath.Join(destinationDir, "2026-07-28-14-45.png")
	if got != filepath.ToSlash(want) {
		t.Fatalf("copyScImage path = %q, want %q", got, filepath.ToSlash(want))
	}
	payload, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("read copied image: %v", err)
	}
	if string(payload) != "event-image" {
		t.Fatalf("copied image payload = %q", payload)
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
