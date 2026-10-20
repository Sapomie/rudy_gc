package movie

import (
	"errors"
	"testing"
)

func TestNormalizeFC2ResourceAlbumName(t *testing.T) {
	for _, albumName := range []string{defaultAlbumName, pendingAlbumName} {
		got, err := normalizeFC2ResourceAlbumName("  " + albumName + "  ")
		if err != nil {
			t.Fatalf("normalizeFC2ResourceAlbumName(%q) returned error: %v", albumName, err)
		}
		if got != albumName {
			t.Fatalf("normalizeFC2ResourceAlbumName(%q) = %q, want %q", albumName, got, albumName)
		}
	}

	if _, err := normalizeFC2ResourceAlbumName("其他相册"); !errors.Is(err, ErrInvalidFC2Album) {
		t.Fatalf("normalizeFC2ResourceAlbumName error = %v, want %v", err, ErrInvalidFC2Album)
	}
}

func TestNormalizeFC2ResourceKey(t *testing.T) {
	got, err := normalizeFC2ResourceKey("fc2ppv- 004845383-1")
	if err != nil {
		t.Fatalf("normalizeFC2ResourceKey returned error: %v", err)
	}
	if got != "FC2PPV-4845383-1" {
		t.Fatalf("normalizeFC2ResourceKey = %q, want %q", got, "FC2PPV-4845383-1")
	}

	if _, err := normalizeFC2ResourceKey("invalid"); !errors.Is(err, ErrInvalidFC2Key) {
		t.Fatalf("normalizeFC2ResourceKey error = %v, want %v", err, ErrInvalidFC2Key)
	}
}
