package handler

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	mediaservice "rudy_gc/internal/service/media"
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
		ImageType:  "contact_sheet",
		FilePath:   "/Volumes/Expansion/f1/captures/7/contact-sheet-test.png",
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
	dir := t.TempDir()
	paths := map[int64]string{}
	for _, id := range []int64{9, 10, 11, 12, 13} {
		paths[id] = filepath.Join(dir, "contact-sheet-"+strconv.FormatInt(id, 10)+".png")
		if err := os.WriteFile(paths[id], []byte("image"), 0o644); err != nil {
			t.Fatalf("创建测试联系表失败: %v", err)
		}
	}
	movieRows := []*moviex.FC2Movie{
		{ID: 28, FC2Key: "FC2PPV-4556413-1", CoverImageID: 10},
		{ID: 29, FC2Key: "FC2PPV-4664062-1"},
		{ID: 30, FC2Key: "FC2PPV-4771711-1", CoverImageID: 999},
		{ID: 31, FC2Key: "FC2PPV-4888888-1"},
	}
	mediaRows := []*moviex.FC2Media{
		{ID: 7, FC2MovieID: 28, FullDir: "/media", FileName: "4556413.mp4", BirthTime: 100, HasSub: int64(consts.FilmHasSub)},
		{ID: 8, FC2MovieID: 29, FullDir: "/removed", FileName: "4664062.mp4", BirthTime: 101, IsRemoved: int64(consts.FilmIsRemoved)},
		{ID: 9, FC2MovieID: 30},
	}
	statRows := []*moviex.FC2ScStat{{FC2MovieID: 28, ScTimes: 3, ComeTimes: 2, LastScTime: 200}}
	imageRows := []*moviex.FC2MediaImage{
		{ID: 9, FC2MediaID: 7, ImageType: "legacy", FilePath: paths[9], CreatedOn: 100},
		{ID: 10, FC2MediaID: 7, ImageType: "contact_sheet", FilePath: paths[10], CreatedOn: 300},
		{ID: 11, FC2MediaID: 7, ImageType: "contact_sheet", FilePath: paths[11], CreatedOn: 200},
		{ID: 12, FC2MediaID: 8, ImageType: "contact_sheet", FilePath: paths[12], CreatedOn: 300},
		{ID: 13, FC2MediaID: 9, ImageType: "contact_sheet", FilePath: paths[13], CreatedOn: 400},
	}

	rows := mediaservice.AssembleFC2CardViews(movieRows, mediaRows, statRows, imageRows)
	if len(rows) != 4 {
		t.Fatalf("card rows = %d", len(rows))
	}
	if rows[0].CoverImageURL != filepath.ToSlash(paths[10]) || rows[0].EffectiveCoverImageID != 10 {
		t.Fatalf("cover image url = %q", rows[0].CoverImageURL)
	}
	if rows[1].CoverImageURL != filepath.ToSlash(paths[12]) || rows[1].EffectiveCoverImageID != 12 {
		t.Fatalf("默认封面 = %#v", rows[1])
	}
	if rows[2].CoverImageURL != filepath.ToSlash(paths[13]) || rows[2].EffectiveCoverImageID != 13 {
		t.Fatalf("失效关联回退封面 = %#v", rows[2])
	}
	if rows[0].ScTimes != 3 || rows[0].ComeTimes != 2 || rows[0].LastScTime != 200 {
		t.Fatalf("sc stats = %#v", rows[0])
	}
	if !rows[0].ShowPlayButton || rows[0].PlayButtonClass != "btn-success" || rows[0].DownloadTime != 100 {
		t.Fatalf("字幕媒体播放状态 = %#v", rows[0])
	}
	if !rows[1].ShowPlayButton || rows[1].PlayButtonClass != "btn-secondary" || rows[1].PlayButtonText != "Removed" || rows[1].DownloadTime != 101 {
		t.Fatalf("已移除媒体播放状态 = %#v", rows[1])
	}
	if rows[3].CoverImageURL != "" {
		t.Fatalf("movie without media should have empty cover: %q", rows[3].CoverImageURL)
	}
}

func TestSuccessfulFC2MediaIDsOnlyReturnsCommittedRows(t *testing.T) {
	result := &mediaservice.FC2IngestResult{
		Success: 2,
		Items: []*mediaservice.FC2IngestFileItem{
			{FC2MediaID: 7},
			{FC2MediaID: 8, Error: "failed"},
			{FC2MediaID: 9},
			{FC2MediaID: 0},
		},
	}
	ids := successfulFC2MediaIDs(result)
	if len(ids) != 2 || ids[0] != 7 || ids[1] != 9 {
		t.Fatalf("成功媒体 ID = %#v，期望 [7 9]", ids)
	}
}
