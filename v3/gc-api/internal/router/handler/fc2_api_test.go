package handler

import (
	"path/filepath"
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
)

func TestApplyFC2PlayButton(t *testing.T) {
	tests := []struct {
		name      string
		media     *moviex.FC2Media
		show      bool
		className string
		text      string
		download  int64
	}{
		{name: "no media"},
		{name: "download time without path", media: &moviex.FC2Media{BirthTime: 100}, download: 100},
		{name: "no subtitle", media: &moviex.FC2Media{FullDir: "/media", FileName: "a.mp4", HasSub: int64(consts.FilmNoSub), IsRemoved: int64(consts.FilmIsNotRemoved), BirthTime: 101}, show: true, className: "btn-info", text: "Play", download: 101},
		{name: "subtitle", media: &moviex.FC2Media{FullDir: "/media", FileName: "b.mp4", HasSub: int64(consts.FilmHasSub), IsRemoved: int64(consts.FilmIsNotRemoved), BirthTime: 102}, show: true, className: "btn-success", text: "Play", download: 102},
		{name: "removed", media: &moviex.FC2Media{FullDir: "/removed", FileName: "c.mp4", IsRemoved: int64(consts.FilmIsRemoved), BirthTime: 103}, show: true, className: "btn-secondary", text: "Removed", download: 103},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := &fc2ListRow{}
			applyFC2PlayButton(item, test.media)
			if item.ShowPlayButton != test.show || item.PlayButtonClass != test.className || item.PlayButtonText != test.text {
				t.Fatalf("play button = show:%v class:%q text:%q", item.ShowPlayButton, item.PlayButtonClass, item.PlayButtonText)
			}
			if test.show && item.VideoURL != filepath.Join(test.media.FullDir, test.media.FileName) {
				t.Fatalf("video url = %q", item.VideoURL)
			}
			if item.DownloadTime != test.download {
				t.Fatalf("download time = %d", item.DownloadTime)
			}
		})
	}
}

func TestBuildFC2MediaView(t *testing.T) {
	if buildFC2MediaView(nil) != nil {
		t.Fatal("nil media should produce nil view")
	}
	media := &moviex.FC2Media{
		ID:         7,
		FC2MovieID: 28,
		FC2Key:     "FC2PPV-4556413-1",
		FullDir:    "/media/fc2",
		FileName:   "4556413_sub.mp4",
		HasSub:     int64(consts.FilmHasSub),
		BirthTime:  1783851946,
	}
	view := buildFC2MediaView(media)
	if view == nil || view.FC2Media != media {
		t.Fatal("media view should retain the media row")
	}
	if !view.ShowPlayButton || view.PlayButtonClass != "btn-success" || view.PlayButtonText != "Play" {
		t.Fatalf("unexpected play view: %#v", view)
	}
	if view.VideoURL != filepath.Join(media.FullDir, media.FileName) {
		t.Fatalf("video url = %q", view.VideoURL)
	}
}

func TestBuildFC2MediaImageView(t *testing.T) {
	if buildFC2MediaImageView(nil) != nil {
		t.Fatal("空图片记录应返回空视图")
	}
	image := &moviex.FC2MediaImage{
		ID:         9,
		FC2MediaID: 7,
		ImageType:  "screenshot",
		FilePath:   "/Volumes/Expansion/f1/captures/7/screenshot-test.png",
		CreatedOn:  1784016000,
	}
	view := buildFC2MediaImageView(image, image.ID)
	if view == nil || view.FC2MediaImage != image {
		t.Fatal("图片视图应保留原记录")
	}
	if view.ImageURL != image.FilePath {
		t.Fatalf("图片地址 = %q", view.ImageURL)
	}
	if !view.IsCover {
		t.Fatal("封面图片应标记为当前封面")
	}
}

func TestNormalizeFC2Sort(t *testing.T) {
	for _, field := range []string{"fc2_key", "title", "performer_text", "post_time", "play", "birth_time"} {
		if got := normalizeFC2Sort(field); got != field {
			t.Fatalf("normalizeFC2Sort(%q) = %q", field, got)
		}
	}
	if got := normalizeFC2Sort("invalid"); got != "post_time" {
		t.Fatalf("normalizeFC2Sort(invalid) = %q", got)
	}
	if normalizeFC2Order("asc") != "asc" || normalizeFC2Order("anything") != "desc" {
		t.Fatal("normalizeFC2Order returned unexpected value")
	}
}

func TestBuildFC2CardRowsUsesSelectedCover(t *testing.T) {
	movieRows := []*moviex.FC2Movie{{ID: 28, FC2Key: "FC2PPV-4556413-1", CoverImageID: 9}, {ID: 29, FC2Key: "FC2PPV-4664062-1"}}
	mediaRows := []*moviex.FC2Media{{ID: 7, FC2MovieID: 28, FullDir: "/media", FileName: "4556413.mp4", BirthTime: 100}}
	statRows := []*moviex.FC2ScStat{{FC2MovieID: 28, ScTimes: 3, ComeTimes: 2, LastScTime: 200}}
	imageRows := []*moviex.FC2MediaImage{
		{ID: 9, FC2MediaID: 7, ImageType: "screenshot", FilePath: "/capture/old.png", CreatedOn: 100},
		{ID: 10, FC2MediaID: 7, ImageType: "contact_sheet", FilePath: "/capture/sheet.png", CreatedOn: 300},
		{ID: 11, FC2MediaID: 7, ImageType: "screenshot", FilePath: "/capture/new.png", CreatedOn: 200},
	}

	rows := buildFC2CardRows(movieRows, mediaRows, statRows, imageRows)
	if len(rows) != 2 {
		t.Fatalf("card rows = %d", len(rows))
	}
	if rows[0].ScreenshotURL != "/capture/old.png" {
		t.Fatalf("screenshot url = %q", rows[0].ScreenshotURL)
	}
	if rows[0].ScTimes != 3 || rows[0].ComeTimes != 2 || rows[0].LastScTime != 200 {
		t.Fatalf("sc stats = %#v", rows[0])
	}
	if rows[1].ScreenshotURL != "" {
		t.Fatalf("movie without media should have empty screenshot: %q", rows[1].ScreenshotURL)
	}
}
