package modelx

import (
	"strings"
	"testing"

	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func TestListMovieIDsQueryUsesGroupedSortCompatibleWithMySQLDistinctRules(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	sqlText, _, err := repo.baseCardQuery(&types.CardsListRequest{}).
		Column("m.jav_id AS movie_jav_id").
		GroupBy("m.jav_id").
		OrderBy(cardGroupedOrderClause(consts.OrderByReleasingDate, "desc")).
		Limit(18).
		ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	if strings.Contains(sqlText, "DISTINCT m.jav_id") {
		t.Fatalf("list query should not use DISTINCT with external order columns: %s", sqlText)
	}
	if !strings.Contains(sqlText, "GROUP BY m.jav_id") {
		t.Fatalf("list query should group by movie jav id: %s", sqlText)
	}
	if !strings.Contains(sqlText, "ORDER BY MAX(m.releasing_date) DESC") {
		t.Fatalf("list query should order by grouped releasing date: %s", sqlText)
	}
}

func TestCardTodayIDPlanUsesLegacyAMovieFastPath(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:     "cardstoday",
		OrderBy:  consts.OrderByReleasingDate,
		Order:    "desc",
		Page:     1,
		PageSize: 18,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if !ok || plan == nil {
		t.Fatalf("cardstoday should use fast path, ok=%v plan=%#v", ok, plan)
	}
	if plan.source != "a_movie" {
		t.Fatalf("cardstoday should use a_movie source, got %q", plan.source)
	}
	if strings.Contains(plan.countSQL, "JOIN") || strings.Contains(plan.countSQL, "COUNT(DISTINCT") {
		t.Fatalf("cardstoday count should not join or count distinct: %s", plan.countSQL)
	}
	if !strings.Contains(plan.listSQL, "FROM a_movie") ||
		!strings.Contains(plan.listSQL, "releasing_date <= ?") ||
		!strings.Contains(plan.listSQL, "ORDER BY releasing_date DESC") {
		t.Fatalf("cardstoday list sql should match legacy a_movie path: %s", plan.listSQL)
	}
}

func TestRankedIDPlanUsesLegacyMinfoFastPath(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	days := int64(1)
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:          "cardshasrank",
		DaysInRankMin: &days,
		OrderBy:       consts.OrderByRankDate,
		Order:         "desc",
		Page:          1,
		PageSize:      18,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if !ok || plan == nil {
		t.Fatalf("cardshasrank should use fast path, ok=%v plan=%#v", ok, plan)
	}
	if plan.source != "bm_minfo" {
		t.Fatalf("cardshasrank should use bm_minfo source, got %q", plan.source)
	}
	if strings.Contains(plan.listSQL, "JOIN") || !strings.Contains(plan.listSQL, "days_in_rank > ?") {
		t.Fatalf("cardshasrank list sql should stay on bm_minfo: %s", plan.listSQL)
	}
}

func TestMediaOwnedIDPlanUsesLegacyWMediaFastPath(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:       "cardsmediamowned",
		MediaOwned: consts.OwnedAllNotRemoved,
		OrderBy:    consts.OrderByMediaBirthTime,
		Order:      "desc",
		Page:       1,
		PageSize:   18,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if !ok || plan == nil {
		t.Fatalf("cardsmediamowned should use fast path, ok=%v plan=%#v", ok, plan)
	}
	if plan.source != "w_media" {
		t.Fatalf("cardsmediamowned should use w_media source, got %q", plan.source)
	}
	if strings.Contains(plan.listSQL, "JOIN") ||
		!strings.Contains(plan.listSQL, "source_type = ?") ||
		!strings.Contains(plan.listSQL, "is_removed = ?") {
		t.Fatalf("cardsmediamowned list sql should stay on w_media: %s", plan.listSQL)
	}
}

func TestNeedDownloadIDPlanUsesLegacyAlbumFastPath(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:         "cardsneeddownload",
		NeedDownload: consts.MovieNeedDownloadOK,
		OrderBy:      consts.OrderByReleasingDate,
		Order:        "desc",
		Page:         1,
		PageSize:     18,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if !ok || plan == nil {
		t.Fatalf("cardsneeddownload should use fast path, ok=%v plan=%#v", ok, plan)
	}
	if plan.source != "c_movie_album_item" {
		t.Fatalf("cardsneeddownload should use album item source, got %q", plan.source)
	}
	if !strings.Contains(plan.listSQL, "FROM c_movie_album_item cai") ||
		!strings.Contains(plan.listSQL, "JOIN c_movie_album ca") ||
		!strings.Contains(plan.listSQL, "ca.name = ?") {
		t.Fatalf("cardsneeddownload list sql should match album path: %s", plan.listSQL)
	}
}

func TestComplexCardFiltersFallbackToGenericPlan(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:      "cardstoday",
		CastNames: "Alice",
		OrderBy:   consts.OrderByReleasingDate,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if ok || plan != nil {
		t.Fatalf("m2m filters should fallback to generic query, ok=%v plan=%#v", ok, plan)
	}
}

func TestCardFiltersCoverLegacyCardsQueryParameters(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	castAgeMin := 18.5
	castAgeMax := 30.0
	viewWatchedMin := int64(0)
	viewWatchedMax := int64(1000)
	scoreMin := 7.5
	scoreMax := 9.9
	daysInRankMin := int64(1)
	scTimesMin := int64(0)
	maxSc := int64(0)
	comeTimesMin := int64(0)
	maxCome := int64(0)
	req := &types.CardsListRequest{
		PersonIds:      "100",
		AlbumName:      "电影稍后下载",
		CastAgeMin:     &castAgeMin,
		CastAgeMax:     &castAgeMax,
		ViewWatchedMin: &viewWatchedMin,
		ViewWatchedMax: &viewWatchedMax,
		ScoreMin:       &scoreMin,
		ScoreMax:       &scoreMax,
		LastScTimeMin:  "2024-01-01",
		LastScTimeMax:  "2024-12-31",
		ScTimesMin:     &scTimesMin,
		ScTimesMax:     &maxSc,
		ComeTimesMin:   &comeTimesMin,
		ComeTimesMax:   &maxCome,
		MediaDir1:      "media",
		MediaDir2:      "watched",
		DaysInRankMin:  &daysInRankMin,
		NeedDownload:   consts.MovieNeedDownloadOK,
		MediaOwned:     consts.OwnedAllNotRemoved,
	}

	sqlText, _, err := repo.baseCardQuery(req).
		Column("m.jav_id AS movie_jav_id").
		GroupBy("m.jav_id").
		OrderBy(cardGroupedOrderClause(consts.OrderByCastAgeAsc, "asc")).
		Limit(18).
		ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}

	required := []string{
		"c.person_id = ?",
		"ca.name = ?",
		"m.cast_average_age >= ?",
		"m.cast_average_age <= ?",
		"m.viewers_number_watched >= ?",
		"m.viewers_number_watched <= ?",
		"m.score >= ?",
		"m.score <= ?",
		"gs.last_sc_time >= ?",
		"gs.last_sc_time <= ?",
		"gs.sc_times >= ?",
		"gs.sc_times <= ?",
		"gs.come_times >= ?",
		"gs.come_times <= ?",
		"CONCAT('/', wm.full_dir, '/') LIKE ?",
		"mi.days_in_rank >= ?",
		"wm.is_removed = ?",
		"ORDER BY COALESCE(MIN(NULLIF(m.cast_average_age, 0)), 999999) ASC",
	}
	for _, item := range required {
		if !strings.Contains(sqlText, item) {
			t.Fatalf("cards query missing %q in sql: %s", item, sqlText)
		}
	}
}

func TestCardAgeFiltersKeepExplicitZeroWithoutHiddenExclusion(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	zero := 0.0
	req := &types.CardsListRequest{CastAgeMin: &zero, CastAgeMax: &zero}

	sqlText, args, err := repo.baseCardQuery(req).
		Column("m.jav_id AS movie_jav_id").
		GroupBy("m.jav_id").
		ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	if !strings.Contains(sqlText, "m.cast_average_age >= ?") || !strings.Contains(sqlText, "m.cast_average_age <= ?") {
		t.Fatalf("explicit zero age bounds should stay in sql: %s", sqlText)
	}
	if strings.Contains(sqlText, "cast_average_age <> ?") || strings.Contains(sqlText, "cast_average_age != ?") {
		t.Fatalf("age filters should not add hidden zero exclusion: %s", sqlText)
	}
	if len(args) != 2 || args[0] != int64(0) || args[1] != int64(0) {
		t.Fatalf("explicit zero age args should be preserved, got %#v", args)
	}
}

func TestAMovieFastPathAgeFiltersKeepExplicitZeroWithoutHiddenExclusion(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	zero := 0.0
	plan, ok, err := repo.buildCardIDFastPathPlan(&types.CardsListRequest{
		View:       "cards",
		CastAgeMin: &zero,
		CastAgeMax: &zero,
		OrderBy:    consts.OrderByReleasingDate,
		Page:       1,
		PageSize:   18,
	})
	if err != nil {
		t.Fatalf("build fast path failed: %v", err)
	}
	if !ok || plan == nil {
		t.Fatalf("a_movie age filter should keep fast path, ok=%v plan=%#v", ok, plan)
	}
	if !strings.Contains(plan.listSQL, "cast_average_age >= ?") || !strings.Contains(plan.listSQL, "cast_average_age <= ?") {
		t.Fatalf("explicit zero age bounds should stay in fast path sql: %s", plan.listSQL)
	}
	if strings.Contains(plan.listSQL, "cast_average_age <> ?") || strings.Contains(plan.listSQL, "cast_average_age != ?") {
		t.Fatalf("fast path should not add hidden zero exclusion: %s", plan.listSQL)
	}
}

func TestMapCardBaseRowTreatsZeroWMediaIDAsNoMedia(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	card := repo.mapCardBaseRow(&cardBaseRow{
		MovieJavID:   "jav-id",
		MovieName:    "HMN-321",
		NeedDownload: consts.MovieNeedDownloadNone,
	})
	if card.MovieHref != "/movie/HMN-321" {
		t.Fatalf("movie href should use movie name, got %q", card.MovieHref)
	}
	if card.OwnedWMedia != 0 || card.VideoUrlWMedia != "" || card.FilmBirthDateWMedia != "" {
		t.Fatalf("zero w_media id should be treated as no media: %#v", card)
	}
}

func TestBuildLocalJacketPathMatchesLegacyVolumeURL(t *testing.T) {
	got := buildLocalJacketPath("/Volumes/T7/data/jacket_cover/", "HMN/HMN-321.jpg")
	want := "/Volumes/T7/data/jacket_cover/HMN/HMN-321.jpg"
	if got != want {
		t.Fatalf("local jacket path should keep legacy volume URL, got %q want %q", got, want)
	}
}

func TestMapCardBaseRowPrefersLocalJacketThenRemoteFallback(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{
		Fetcher: config.FetcherConf{LocalImageDir: "/Volumes/T7/data/jacket_cover/"},
	})
	localCard := repo.mapCardBaseRow(&cardBaseRow{
		MovieJavID:     "jav-id",
		MovieName:      "HMN-321",
		JacketImg:      "https://example.test/HMN-321.jpg",
		JacketImgLocal: "HMN/HMN-321.jpg",
	})
	if localCard.JacketImg != "/Volumes/T7/data/jacket_cover/HMN/HMN-321.jpg" {
		t.Fatalf("local jacket should be exposed through legacy volume path, got %q", localCard.JacketImg)
	}

	remoteCard := repo.mapCardBaseRow(&cardBaseRow{
		MovieJavID: "jav-id",
		MovieName:  "HMN-322",
		JacketImg:  "https://example.test/HMN-322.jpg",
	})
	if remoteCard.JacketImg != "https://example.test/HMN-322.jpg" {
		t.Fatalf("remote jacket should be fallback when local path is absent, got %q", remoteCard.JacketImg)
	}
}
