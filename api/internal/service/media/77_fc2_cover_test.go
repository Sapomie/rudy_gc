package media

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"rudy_gc/internal/repository/moviex"
)

func TestResolveFC2CoverImageUsesSelectedContactSheet(t *testing.T) {
	dir := t.TempDir()
	selected := newFC2CoverTestImage(t, dir, 9, 7, FC2MediaImageTypeContactSheet, 100)
	newest := newFC2CoverTestImage(t, dir, 10, 7, FC2MediaImageTypeContactSheet, 200)

	got := ResolveFC2CoverImage(7, selected.ID, []*moviex.FC2MediaImage{newest, selected})
	if got == nil || got.ID != selected.ID {
		t.Fatalf("人工选择封面 = %#v，期望图片 %d", got, selected.ID)
	}
}

func TestResolveFC2CoverImageKeepsSelectedContactSheet(t *testing.T) {
	dir := t.TempDir()
	selected := newFC2CoverTestImage(t, dir, 9, 7, FC2MediaImageTypeContactSheet, 100)
	newest := newFC2CoverTestImage(t, dir, 10, 7, FC2MediaImageTypeContactSheet, 200)

	got := ResolveFC2CoverImage(7, selected.ID, []*moviex.FC2MediaImage{newest, selected})
	if got == nil || got.ID != selected.ID {
		t.Fatalf("人工选择联系表封面 = %#v，期望图片 %d", got, selected.ID)
	}
}

func TestResolveFC2CoverImageFallsBackToNewestContactSheet(t *testing.T) {
	dir := t.TempDir()
	older := newFC2CoverTestImage(t, dir, 9, 7, FC2MediaImageTypeContactSheet, 100)
	newest := newFC2CoverTestImage(t, dir, 10, 7, FC2MediaImageTypeContactSheet, 200)

	for _, selectedImageID := range []int64{0, 999} {
		got := ResolveFC2CoverImage(7, selectedImageID, []*moviex.FC2MediaImage{older, newest})
		if got == nil || got.ID != newest.ID {
			t.Fatalf("选择 ID %d 时的默认封面 = %#v，期望图片 %d", selectedImageID, got, newest.ID)
		}
	}
}

func TestResolveFC2CoverImageFallsBackWhenSelectedFileIsMissing(t *testing.T) {
	dir := t.TempDir()
	fallback := newFC2CoverTestImage(t, dir, 9, 7, FC2MediaImageTypeContactSheet, 100)
	missingSelected := &moviex.FC2MediaImage{
		ID: 10, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet,
		FilePath: filepath.Join(dir, "missing-selected.png"), CreatedOn: 200,
	}

	got := ResolveFC2CoverImage(7, missingSelected.ID, []*moviex.FC2MediaImage{missingSelected, fallback})
	if got == nil || got.ID != fallback.ID {
		t.Fatalf("失效人工封面回退结果 = %#v，期望图片 %d", got, fallback.ID)
	}
}

func TestResolveFC2CoverImageRejectsUnusableRows(t *testing.T) {
	dir := t.TempDir()
	wrongMedia := newFC2CoverTestImage(t, dir, 10, 8, FC2MediaImageTypeContactSheet, 200)
	missingFile := &moviex.FC2MediaImage{ID: 11, FC2MediaID: 7, ImageType: FC2MediaImageTypeContactSheet, FilePath: filepath.Join(dir, "missing.png"), CreatedOn: 100}
	unknownType := newFC2CoverTestImage(t, dir, 12, 7, "unknown", 300)

	if got := ResolveFC2CoverImage(7, 0, []*moviex.FC2MediaImage{wrongMedia, missingFile, unknownType}); got != nil {
		t.Fatalf("无有效封面图片时应返回空，实际 %#v", got)
	}
}

func TestResolveFC2DetailCoverImageUsesNewestDetailSheetOnly(t *testing.T) {
	dir := t.TempDir()
	card := newFC2CoverTestImage(t, dir, 9, 7, FC2MediaImageTypeContactSheet, 300)
	older := newFC2CoverTestImage(t, dir, 10, 7, FC2MediaImageTypeDetailContactSheet, 100)
	newest := newFC2CoverTestImage(t, dir, 11, 7, FC2MediaImageTypeDetailContactSheet, 200)

	got := ResolveFC2DetailCoverImage(7, []*moviex.FC2MediaImage{card, older, newest})
	if got == nil || got.ID != newest.ID {
		t.Fatalf("详情联系表 = %#v，期望图片 %d", got, newest.ID)
	}
	if cardCover := ResolveFC2CoverImage(7, 0, []*moviex.FC2MediaImage{newest, card}); cardCover == nil || cardCover.ID != card.ID {
		t.Fatalf("详情联系表影响了卡片封面选择: %#v", cardCover)
	}
}

func newFC2CoverTestImage(t *testing.T, dir string, id, mediaID int64, imageType string, createdOn int64) *moviex.FC2MediaImage {
	t.Helper()
	filePath := filepath.Join(dir, imageType+"-"+strconv.FormatInt(id, 10)+".png")
	if err := os.WriteFile(filePath, []byte("image"), 0o644); err != nil {
		t.Fatalf("创建测试图片失败: %v", err)
	}
	return &moviex.FC2MediaImage{ID: id, FC2MediaID: mediaID, ImageType: imageType, FilePath: filePath, CreatedOn: createdOn}
}
