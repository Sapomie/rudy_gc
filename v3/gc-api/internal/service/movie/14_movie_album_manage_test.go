package movie

import (
	"context"
	"errors"
	"testing"
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
