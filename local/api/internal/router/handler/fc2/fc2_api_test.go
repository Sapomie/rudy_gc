package fc2

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	mediaservice "rudy_gc/local/api/internal/service/media"
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

func TestBuildFC2ScEvents(t *testing.T) {
	rows := []*moviex.GList{
		{ScName: "SC 旧", IsSc: consts.GListIsSc, IsCome: consts.GListIsNotCome},
		{ScName: "SC 新/1", IsSc: consts.GListIsSc, IsCome: consts.GListIsCome},
		{ScName: "SC 旧", IsSc: consts.GListIsSc, IsCome: consts.GListIsCome},
		{ScName: "忽略", IsSc: consts.GListIsNotSc},
		{ScName: "  ", IsSc: consts.GListIsSc},
	}
	events := []*moviex.GSc{
		{Name: "SC 旧", ScTime: 100, Cooldown: 5},
		{Name: "SC 新/1", ScTime: 200, Cooldown: 10},
	}

	got := buildFC2ScEvents(rows, events)
	if len(got) != 2 {
		t.Fatalf("SC 事件数量 = %d，期望 2", len(got))
	}
	if got[0].ScName != "SC 新/1" || got[0].ScTime != 200 || got[0].Cooldown != 10 || !got[0].IsCome {
		t.Fatalf("第一条 SC 事件异常: %+v", got[0])
	}
	if got[0].Href != "/scevent/SC%20%E6%96%B0%2F1" {
		t.Fatalf("SC 事件链接 = %q", got[0].Href)
	}
	if got[1].ScName != "SC 旧" || got[1].ScTime != 100 || got[1].IsCome {
		t.Fatalf("第二条 SC 事件异常: %+v", got[1])
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

func TestBuildFC2DetailCoverImageViewSelectsDetailSheet(t *testing.T) {
	dir := t.TempDir()
	cardPath := filepath.Join(dir, "contact-sheet.png")
	detailPath := filepath.Join(dir, "detail-contact-sheet.png")
	for _, path := range []string{cardPath, detailPath} {
		if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
			t.Fatalf("创建测试图片失败: %v", err)
		}
	}
	view := buildFC2DetailCoverImageView(7, []*moviex.FC2MediaImage{
		{ID: 9, FC2MediaID: 7, ImageType: mediaservice.FC2MediaImageTypeContactSheet, FilePath: cardPath, CreatedOn: 200},
		{ID: 10, FC2MediaID: 7, ImageType: mediaservice.FC2MediaImageTypeDetailContactSheet, FilePath: detailPath, CreatedOn: 100},
	})
	if view == nil || view.ID != 10 || view.ImageURL != filepath.ToSlash(detailPath) {
		t.Fatalf("详情封面视图 = %#v", view)
	}
	if view.IsCover {
		t.Fatal("详情联系表不应标记为卡片封面")
	}
}

func TestNormalizeFC2Sort(t *testing.T) {
	for _, field := range []string{"fc2_key", "title", "performer_text", "post_time", "play", "birth_time", "sc_times", "come_times", "last_selected_sc_time"} {
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

func TestNormalizeFC2CardSort(t *testing.T) {
	for _, field := range []string{"post_time", "fc2_key", "last_selected_sc_time", "birth_time", "sc_times"} {
		if got := normalizeFC2CardSort(field, "birth_time"); got != field {
			t.Fatalf("normalizeFC2CardSort(%q) = %q", field, got)
		}
	}
	if got := normalizeFC2CardSort("title", "birth_time"); got != "birth_time" {
		t.Fatalf("normalizeFC2CardSort(title) = %q", got)
	}
}

func TestBuildFC2CardsSortQueryTogglesDirectionAndResetsPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/fclistcard?sort=sc_times&order=desc&p=3&ps=18", nil)

	query := buildFC2CardsSortQuery(ctx, "sc_times", "desc")
	if !query.ByScTimes.Active || !query.ByScTimes.Desc {
		t.Fatalf("active SC sort = %#v", query.ByScTimes)
	}
	if query.ByScTimes.Href != "/fclistcard?order=asc&p=1&ps=18&sort=sc_times" {
		t.Fatalf("SC sort href = %q", query.ByScTimes.Href)
	}
	if query.ByFC2Key.Active || query.ByFC2Key.Desc {
		t.Fatalf("inactive FC2 key sort = %#v", query.ByFC2Key)
	}
	if query.ByFC2Key.Href != "/fclistcard?order=desc&p=1&ps=18&sort=fc2_key" {
		t.Fatalf("FC2 key sort href = %q", query.ByFC2Key.Href)
	}
}

func TestBuildFC2CardRowsUsesSelectedCover(t *testing.T) {
	dir := t.TempDir()
	paths := map[int64]string{}
	for _, id := range []int64{9, 10, 11, 12, 13, 14} {
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
	statRows := []*moviex.FC2ScStat{{FC2MovieID: 28, ScTimes: 3, ComeTimes: 2, LastSelectedScTime: 200}}
	imageRows := []*moviex.FC2MediaImage{
		{ID: 9, FC2MediaID: 7, ImageType: "legacy", FilePath: paths[9], CreatedOn: 100},
		{ID: 10, FC2MediaID: 7, ImageType: "contact_sheet", FilePath: paths[10], CreatedOn: 300},
		{ID: 11, FC2MediaID: 7, ImageType: "contact_sheet", FilePath: paths[11], CreatedOn: 200},
		{ID: 12, FC2MediaID: 8, ImageType: "contact_sheet", FilePath: paths[12], CreatedOn: 300},
		{ID: 13, FC2MediaID: 9, ImageType: "contact_sheet", FilePath: paths[13], CreatedOn: 400},
		{ID: 14, FC2MediaID: 7, ImageType: "detail_contact_sheet", FilePath: paths[14], CreatedOn: 500},
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
	if rows[0].ScTimes != 3 || rows[0].ComeTimes != 2 || rows[0].LastSelectedScTime != 200 {
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

func TestFC2MediaIngestResponseIncludesReplacementDetails(t *testing.T) {
	item := &mediaservice.FC2IngestFileItem{
		Status:       "pass",
		Action:       "replace",
		Notice:       "活动 FC2影片已存在，将替换旧版本",
		FC2Key:       "FC2PPV-4862980-1",
		ExistingPath: "/old/media.mp4",
		RemovedPath:  "/root/001_process/005_removed/media.mp4",
	}
	precheck := buildFC2MediaPrecheckResponse("precheck", &mediaservice.FC2IngestResult{
		Precheck: mediaservice.IngestPrecheckStat{Total: 1, Passed: 1},
		Items:    []*mediaservice.FC2IngestFileItem{item},
	})
	if precheck.Replace != 1 || len(precheck.PrecheckPass) != 1 {
		t.Fatalf("预处理替换统计异常: %+v", precheck)
	}
	row := precheck.PrecheckPass[0]
	if row.ActionLabel != "替换活动旧版本" || row.ExistingPath != item.ExistingPath || row.Notice != item.Notice {
		t.Fatalf("预处理替换字段异常: %+v", row)
	}

	commit := buildFC2MediaCommitResponse("commit", &mediaservice.FC2IngestResult{
		Total: 1, Success: 1, Items: []*mediaservice.FC2IngestFileItem{item},
	})
	if commit.Replace != 1 || len(commit.CommitSuccess) != 1 || commit.CommitSuccess[0].RemovedPath != item.RemovedPath {
		t.Fatalf("提交替换字段异常: %+v", commit)
	}
}
