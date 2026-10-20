package media

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type auditWMediaModel struct {
	moviex.WMediaModel
	rows        []*moviex.WMedia
	updateCount int
}

type auditFC2MediaModel struct {
	moviex.FC2MediaModel
	rows        []*moviex.FC2Media
	updateCount int
}

func (m *auditFC2MediaModel) ListAll(context.Context) ([]*moviex.FC2Media, error) {
	out := make([]*moviex.FC2Media, 0, len(m.rows))
	for _, row := range m.rows {
		if row == nil {
			continue
		}
		copyRow := *row
		out = append(out, &copyRow)
	}
	return out, nil
}

func (m *auditFC2MediaModel) Update(context.Context, *moviex.FC2Media) error {
	m.updateCount++
	return nil
}

func (m *auditWMediaModel) ListNativeMediaAll(context.Context) ([]*moviex.WMedia, error) {
	out := make([]*moviex.WMedia, 0, len(m.rows))
	for _, row := range m.rows {
		if row == nil {
			continue
		}
		copyRow := *row
		out = append(out, &copyRow)
	}
	return out, nil
}

func (m *auditWMediaModel) Update(context.Context, *moviex.WMedia) error {
	m.updateCount++
	return nil
}

func TestRescanLibraryReportsConsistencyProblemsWithoutWrites(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := os.MkdirAll(layout.media, 0o755); err != nil {
		t.Fatal(err)
	}

	trackedPath := writeAuditTestFile(t, layout.media, "tracked.mp4", []byte("tracked"))
	mismatchPath := writeAuditTestFile(t, layout.media, "mismatch.mp4", []byte("mismatch"))
	writeAuditTestFile(t, layout.media, "untracked.mp4", []byte("untracked"))
	writeAuditTestFile(t, layout.media, "._sidecar.mp4", []byte("sidecar"))

	model := &auditWMediaModel{rows: []*moviex.WMedia{
		{Id: 1, MovieJavId: "AUDIT-001", MovieName: "AUDIT-001", FullDir: filepath.Dir(trackedPath), FileName: filepath.Base(trackedPath), Size: 7, IsRemoved: consts.FilmIsNotRemoved},
		{Id: 2, MovieJavId: "AUDIT-002", MovieName: "AUDIT-002", FullDir: filepath.Dir(mismatchPath), FileName: filepath.Base(mismatchPath), Size: 1, IsRemoved: consts.FilmIsNotRemoved},
		{Id: 3, MovieJavId: "AUDIT-003", MovieName: "AUDIT-003", FullDir: layout.media, FileName: "missing.mp4", Size: 10, IsRemoved: consts.FilmIsNotRemoved},
		{Id: 4, MovieJavId: "AUDIT-004", MovieName: "AUDIT-004", FullDir: layout.media, FileName: "removed.mp4", Size: 10, IsRemoved: consts.FilmIsRemoved},
	}}
	service := NewService(&dep.Dep{
		Config:      config.Config{Media: config.Media{RootDirs: []string{rootDir}}},
		WMediaModel: model,
	})

	result, err := service.RescanLibrary(context.Background(), []LibraryRescanSelection{{MediaType: libraryMediaTypeJAV, Root: rootDir, Scope: scopeMedia}})
	if err != nil {
		t.Fatal(err)
	}
	if result.ActiveRecords != 3 || result.DiskFiles != 3 {
		t.Fatalf("unexpected totals: active=%d disk=%d", result.ActiveRecords, result.DiskFiles)
	}
	if result.Matched != 2 || result.Healthy != 1 || result.Missing != 1 || result.Untracked != 1 || result.SizeMismatch != 1 {
		t.Fatalf("unexpected audit result: %+v", result)
	}
	if result.DuplicateActive != 0 || result.Errors != 0 {
		t.Fatalf("unexpected duplicate/error result: %+v", result)
	}
	if len(result.MissingItems) != 1 || len(result.UntrackedItems) != 1 || len(result.SizeItems) != 1 {
		t.Fatalf("unexpected item counts: missing=%d untracked=%d size=%d", len(result.MissingItems), len(result.UntrackedItems), len(result.SizeItems))
	}
	if model.updateCount != 0 {
		t.Fatalf("consistency check performed %d database updates", model.updateCount)
	}
}

func TestRescanLibraryReportsDuplicateActiveRecords(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := os.MkdirAll(layout.media, 0o755); err != nil {
		t.Fatal(err)
	}
	firstPath := writeAuditTestFile(t, layout.media, "first.mp4", []byte("first"))
	secondPath := writeAuditTestFile(t, layout.media, "second.mp4", []byte("second"))

	model := &auditWMediaModel{rows: []*moviex.WMedia{
		{Id: 11, MovieJavId: "AUDIT-011", MovieName: "AUDIT-011", FullDir: filepath.Dir(firstPath), FileName: filepath.Base(firstPath), Size: 5, IsRemoved: consts.FilmIsNotRemoved},
		{Id: 12, MovieJavId: "AUDIT-011", MovieName: "AUDIT-011", FullDir: filepath.Dir(secondPath), FileName: filepath.Base(secondPath), Size: 6, IsRemoved: consts.FilmIsNotRemoved},
	}}
	service := NewService(&dep.Dep{
		Config:      config.Config{Media: config.Media{RootDirs: []string{rootDir}}},
		WMediaModel: model,
	})

	result, err := service.RescanLibrary(context.Background(), []LibraryRescanSelection{{MediaType: libraryMediaTypeJAV, Root: rootDir, Scope: scopeMedia}})
	if err != nil {
		t.Fatal(err)
	}
	if result.DuplicateActive != 1 || len(result.DuplicateItems) != 2 {
		t.Fatalf("unexpected duplicate result: groups=%d items=%d", result.DuplicateActive, len(result.DuplicateItems))
	}
	if result.Healthy != 2 || result.Missing != 0 || result.Untracked != 0 || result.SizeMismatch != 0 || result.Errors != 0 {
		t.Fatalf("unexpected consistency result: %+v", result)
	}
	if model.updateCount != 0 {
		t.Fatalf("consistency check performed %d database updates", model.updateCount)
	}
}

func TestRescanLibraryReportsFC2ConsistencyProblemsWithoutWrites(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := os.MkdirAll(layout.media, 0o755); err != nil {
		t.Fatal(err)
	}

	trackedPath := writeAuditTestFile(t, layout.media, "fc2-tracked.mp4", []byte("tracked"))
	mismatchPath := writeAuditTestFile(t, layout.media, "fc2-mismatch.mp4", []byte("mismatch"))
	writeAuditTestFile(t, layout.media, "fc2-untracked.mp4", []byte("untracked"))
	writeAuditTestFile(t, layout.media, "._fc2-sidecar.mp4", []byte("sidecar"))

	model := &auditFC2MediaModel{rows: []*moviex.FC2Media{
		{ID: 21, FC2MovieID: 201, FC2Key: "FC2-PPV-201", FC2Code: "FC2-201", FullDir: filepath.Dir(trackedPath), FileName: filepath.Base(trackedPath), Size: 7, IsRemoved: consts.FilmIsNotRemoved},
		{ID: 22, FC2MovieID: 202, FC2Key: "FC2-PPV-202", FC2Code: "FC2-202", FullDir: filepath.Dir(mismatchPath), FileName: filepath.Base(mismatchPath), Size: 1, IsRemoved: consts.FilmIsNotRemoved},
		{ID: 23, FC2MovieID: 203, FC2Key: "FC2-PPV-203", FC2Code: "FC2-203", FullDir: layout.media, FileName: "fc2-missing.mp4", Size: 10, IsRemoved: consts.FilmIsNotRemoved},
		{ID: 24, FC2MovieID: 204, FC2Key: "FC2-PPV-204", FC2Code: "FC2-204", FullDir: layout.media, FileName: "fc2-removed.mp4", Size: 10, IsRemoved: consts.FilmIsRemoved},
	}}
	service := NewService(&dep.Dep{
		Config:        config.Config{FC2Media: config.FC2MediaConfig{RootDirs: []string{rootDir}}},
		FC2MediaModel: model,
	})

	result, err := service.RescanLibrary(context.Background(), []LibraryRescanSelection{{MediaType: libraryMediaTypeFC2, Root: rootDir, Scope: scopeMedia}})
	if err != nil {
		t.Fatal(err)
	}
	if result.ActiveRecords != 3 || result.DiskFiles != 3 || result.FC2Stats.ActiveRecords != 3 || result.FC2Stats.DiskFiles != 3 {
		t.Fatalf("unexpected FC2 totals: result=%+v stats=%+v", result, result.FC2Stats)
	}
	if result.Matched != 2 || result.Healthy != 1 || result.Missing != 1 || result.Untracked != 1 || result.SizeMismatch != 1 || result.DuplicateActive != 0 || result.Errors != 0 {
		t.Fatalf("unexpected FC2 audit result: %+v", result)
	}
	if len(result.MissingItems) != 1 || result.MissingItems[0].MediaType != libraryMediaTypeFC2 || result.MissingItems[0].FC2Key != "FC2-PPV-203" {
		t.Fatalf("unexpected FC2 missing items: %+v", result.MissingItems)
	}
	if model.updateCount != 0 {
		t.Fatalf("FC2 consistency check performed %d database updates", model.updateCount)
	}
}

func TestRescanLibraryReportsFC2DuplicateActiveRecords(t *testing.T) {
	rootDir := t.TempDir()
	layout := buildRootLayout(rootDir)
	if err := os.MkdirAll(layout.media, 0o755); err != nil {
		t.Fatal(err)
	}
	firstPath := writeAuditTestFile(t, layout.media, "fc2-first.mp4", []byte("first"))
	secondPath := writeAuditTestFile(t, layout.media, "fc2-second.mp4", []byte("second"))

	model := &auditFC2MediaModel{rows: []*moviex.FC2Media{
		{ID: 31, FC2MovieID: 301, FC2Key: "FC2-PPV-301", FC2Code: "FC2-301", FullDir: filepath.Dir(firstPath), FileName: filepath.Base(firstPath), Size: 5, IsRemoved: consts.FilmIsNotRemoved},
		{ID: 32, FC2MovieID: 302, FC2Key: "FC2-PPV-301", FC2Code: "FC2-302", FullDir: filepath.Dir(secondPath), FileName: filepath.Base(secondPath), Size: 6, IsRemoved: consts.FilmIsNotRemoved},
	}}
	service := NewService(&dep.Dep{
		Config:        config.Config{FC2Media: config.FC2MediaConfig{RootDirs: []string{rootDir}}},
		FC2MediaModel: model,
	})

	result, err := service.RescanLibrary(context.Background(), []LibraryRescanSelection{{MediaType: libraryMediaTypeFC2, Root: rootDir, Scope: scopeMedia}})
	if err != nil {
		t.Fatal(err)
	}
	if result.DuplicateActive != 1 || result.FC2Stats.DuplicateActive != 1 || len(result.DuplicateItems) != 2 {
		t.Fatalf("unexpected FC2 duplicate result: result=%+v stats=%+v", result, result.FC2Stats)
	}
	if result.Healthy != 2 || result.Missing != 0 || result.Untracked != 0 || result.SizeMismatch != 0 || result.Errors != 0 {
		t.Fatalf("unexpected FC2 consistency result: %+v", result)
	}
	if model.updateCount != 0 {
		t.Fatalf("FC2 consistency check performed %d database updates", model.updateCount)
	}
}

func TestListLibraryRescanRootOptionsIncludesJAVAndFC2(t *testing.T) {
	javRoot := t.TempDir()
	fc2Root := t.TempDir()
	for _, rootDir := range []string{javRoot, fc2Root} {
		layout := buildRootLayout(rootDir)
		if err := os.MkdirAll(layout.media, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(layout.watched, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	service := NewService(&dep.Dep{Config: config.Config{
		Media:    config.Media{RootDirs: []string{javRoot}},
		FC2Media: config.FC2MediaConfig{RootDirs: []string{fc2Root}},
	}})

	options, err := service.ListLibraryRescanRootOptions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 2 {
		t.Fatalf("unexpected root option count: %d", len(options))
	}
	if options[0].MediaType != libraryMediaTypeJAV || options[0].Root != javRoot {
		t.Fatalf("unexpected JAV option: %+v", options[0])
	}
	if options[1].MediaType != libraryMediaTypeFC2 || options[1].Root != fc2Root {
		t.Fatalf("unexpected FC2 option: %+v", options[1])
	}
	for _, option := range options {
		if len(option.Scopes) != 2 || !option.Scopes[0].Exists || !option.Scopes[1].Exists {
			t.Fatalf("unexpected scopes: %+v", option)
		}
	}
}

func writeAuditTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCurrentLibraryConsistencyCheck(t *testing.T) {
	if os.Getenv("RUDY_RUN_MEDIA_AUDIT") != "1" {
		t.Skip("set RUDY_RUN_MEDIA_AUDIT=1 to run against current configured media library")
	}
	cfg := config.MustLoad("../../../config/config.yaml")
	deps, err := dep.BuildDep(cfg)
	if err != nil {
		t.Fatal(err)
	}
	selections := make([]LibraryRescanSelection, 0, len(cfg.Media.RootDirs)*2)
	for _, rootDir := range cfg.Media.RootDirs {
		selections = append(selections,
			LibraryRescanSelection{MediaType: libraryMediaTypeJAV, Root: rootDir, Scope: scopeMedia},
			LibraryRescanSelection{MediaType: libraryMediaTypeJAV, Root: rootDir, Scope: scopeWatched},
		)
	}
	for _, rootDir := range cfg.FC2Media.RootDirs {
		selections = append(selections,
			LibraryRescanSelection{MediaType: libraryMediaTypeFC2, Root: rootDir, Scope: scopeMedia},
			LibraryRescanSelection{MediaType: libraryMediaTypeFC2, Root: rootDir, Scope: scopeWatched},
		)
	}
	result, err := NewService(deps).RescanLibrary(context.Background(), selections)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("active=%d disk=%d matched=%d healthy=%d missing=%d untracked=%d size_mismatch=%d duplicate_active=%d errors=%d jav=%+v fc2=%+v",
		result.ActiveRecords,
		result.DiskFiles,
		result.Matched,
		result.Healthy,
		result.Missing,
		result.Untracked,
		result.SizeMismatch,
		result.DuplicateActive,
		result.Errors,
		result.JAVStats,
		result.FC2Stats,
	)
}
