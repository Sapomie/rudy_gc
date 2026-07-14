package fetchsehuatang

import (
	"testing"
	"time"

	"rudy_gc/internal/repository/moviex"
)

func TestBuildSHMagnetRows(t *testing.T) {
	seenAt := time.Date(2026, 7, 12, 14, 0, 0, 0, time.Local).Unix()
	postTime := time.Date(2026, 7, 10, 20, 30, 0, 0, time.Local).Unix()
	rows := buildSHMagnetRows(&moviex.ShSehuatangThread{
		ThreadTID:     "3619676",
		ThreadURL:     "https://example.test/thread-3619676-1-1.html",
		ThreadTitle:   "[FC2PPV] FC2PPV-1234567",
		MovieJavId:    "FC2-1234567",
		MovieName:     "FC2PPV-1234567",
		MovieType:     "fc2ppv",
		SourceForumID: "103",
		PostTime:      postTime,
	}, []parsedMagnet{
		{MagnetURL: "magnet:?xt=urn:btih:abcdef0123456789abcdef0123456789abcdef01", InfoHash: "abcdef0123456789abcdef0123456789abcdef01"},
		{MagnetURL: "magnet:?xt=urn:btih:ABCDEF0123456789ABCDEF0123456789ABCDEF01", InfoHash: "ABCDEF0123456789ABCDEF0123456789ABCDEF01"},
		{MagnetURL: "magnet:?xt=urn:btih:1111111111111111111111111111111111111111", InfoHash: "1111111111111111111111111111111111111111"},
	}, seenAt)

	if len(rows) != 2 {
		t.Fatalf("rows length = %d, want 2", len(rows))
	}
	first := rows[0]
	if first.InfoHash != "ABCDEF0123456789ABCDEF0123456789ABCDEF01" {
		t.Fatalf("InfoHash = %q", first.InfoHash)
	}
	if first.ThreadTID != "3619676" || first.SourceForumID != "103" {
		t.Fatalf("source fields = %#v", first)
	}
	if first.MovieType != "fc2ppv" || first.MovieName != "FC2PPV-1234567" {
		t.Fatalf("movie fields = %#v", first)
	}
	if first.PostTime != postTime {
		t.Fatalf("PostTime = %d, want %d", first.PostTime, postTime)
	}
	if first.CreatedOn != seenAt || first.UpdatedOn != seenAt {
		t.Fatalf("maintain fields = %#v", first)
	}
}
