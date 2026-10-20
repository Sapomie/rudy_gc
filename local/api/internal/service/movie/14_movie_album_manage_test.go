package movie

import (
	"context"
	"errors"
	"testing"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func TestIsProtectedMovieAlbumName(t *testing.T) {
	tests := []struct {
		name      string
		protected bool
	}{
		{name: "待下载电影", protected: true},
		{name: " 待删除 ", protected: true},
		{name: "v3", protected: true},
		{name: "vx", protected: true},
		{name: "自定义相册", protected: false},
	}

	for _, tt := range tests {
		if got := IsProtectedMovieAlbumName(tt.name); got != tt.protected {
			t.Fatalf("IsProtectedMovieAlbumName(%q) = %v, want %v", tt.name, got, tt.protected)
		}
	}
}

func TestValidateMovieAlbumDeletion(t *testing.T) {
	if err := validateMovieAlbumDeletion(0); err != nil {
		t.Fatalf("validateMovieAlbumDeletion(0) returned error: %v", err)
	}
	if err := validateMovieAlbumDeletion(1); !errors.Is(err, ErrMovieAlbumNotEmpty) {
		t.Fatalf("validateMovieAlbumDeletion(1) error = %v, want %v", err, ErrMovieAlbumNotEmpty)
	}
}

func TestUpdateMovieAlbumRejectsInvalidID(t *testing.T) {
	service := &Service{}
	_, err := service.UpdateMovieAlbum(context.Background(), 0, "自定义相册")
	if !errors.Is(err, ErrMovieAlbumInvalidID) {
		t.Fatalf("UpdateMovieAlbum error = %v, want %v", err, ErrMovieAlbumInvalidID)
	}
}

func TestDeleteMovieAlbumRejectsInvalidID(t *testing.T) {
	service := &Service{}
	_, err := service.DeleteMovieAlbum(context.Background(), 0)
	if !errors.Is(err, ErrMovieAlbumInvalidID) {
		t.Fatalf("DeleteMovieAlbum error = %v, want %v", err, ErrMovieAlbumInvalidID)
	}
}

func TestMovieAlbumItemTypeUsesFC2KeyIdentity(t *testing.T) {
	row := &moviex.CMovieAlbumItem{FC2Key: "FC2PPV-4556413-1"}
	if got := movieAlbumItemType(row); got != moviedomain.SehuatangThreadMovieTypeFC2PPV {
		t.Fatalf("movieAlbumItemType() = %q, want %q", got, moviedomain.SehuatangThreadMovieTypeFC2PPV)
	}
}

func TestMovieAlbumItemTypeDefaultsLegacyRowsToJAV(t *testing.T) {
	row := &moviex.CMovieAlbumItem{MovieJavId: "ABP-123"}
	if got := movieAlbumItemType(row); got != moviedomain.SehuatangThreadMovieTypeJAV {
		t.Fatalf("movieAlbumItemType() = %q, want %q", got, moviedomain.SehuatangThreadMovieTypeJAV)
	}
}
