package movie

import (
	"context"
	"errors"
	"testing"
)

func TestIsProtectedTorrentAlbumName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		albumName string
		want      bool
	}{
		{name: "下载中", albumName: "下载中", want: true},
		{name: "待下载", albumName: "待下载", want: true},
		{name: "Media", albumName: "Media", want: true},
		{name: "FcMedia", albumName: "FcMedia", want: true},
		{name: "自动清理空格", albumName: "  Media  ", want: true},
		{name: "自定义相册", albumName: "收藏", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsProtectedTorrentAlbumName(tt.albumName); got != tt.want {
				t.Fatalf("IsProtectedTorrentAlbumName(%q) = %v, want %v", tt.albumName, got, tt.want)
			}
		})
	}
}

func TestDeleteAlbumRejectsInvalidID(t *testing.T) {
	t.Parallel()

	service := &Service{}
	result, err := service.DeleteAlbum(context.Background(), 0)
	if result != nil {
		t.Fatalf("DeleteAlbum() result = %#v, want nil", result)
	}
	if !errors.Is(err, ErrAlbumInvalidID) {
		t.Fatalf("DeleteAlbum() error = %v, want %v", err, ErrAlbumInvalidID)
	}
}

func TestUpdateAlbumRejectsInvalidID(t *testing.T) {
	t.Parallel()

	service := &Service{}
	result, err := service.UpdateAlbum(context.Background(), 0, "新名称")
	if result != nil {
		t.Fatalf("UpdateAlbum() result = %#v, want nil", result)
	}
	if !errors.Is(err, ErrAlbumInvalidID) {
		t.Fatalf("UpdateAlbum() error = %v, want %v", err, ErrAlbumInvalidID)
	}
}

func TestValidateAlbumDeletion(t *testing.T) {
	t.Parallel()

	if err := validateAlbumDeletion(0); err != nil {
		t.Fatalf("validateAlbumDeletion(0) error = %v, want nil", err)
	}
	if err := validateAlbumDeletion(1); !errors.Is(err, ErrAlbumNotEmpty) {
		t.Fatalf("validateAlbumDeletion(1) error = %v, want %v", err, ErrAlbumNotEmpty)
	}
}
