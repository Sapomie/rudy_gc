package consts

import "testing"

func TestMovieNeedDownloadAlbumName(t *testing.T) {
	if MovieNeedDownloadAlbumName != "待下载电影" {
		t.Fatalf("unexpected need-download album name: %q", MovieNeedDownloadAlbumName)
	}
}
