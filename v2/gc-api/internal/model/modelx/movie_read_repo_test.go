package modelx

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"

	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func TestCardAMovieQueryKeepsSingleSourcePath(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	sqlText, _, err := applyWhere(squirrel.Select("jav_id").From("a_movie"), repo.cardAMovieWhere(&types.CardsListRequest{
		View:     "cards",
		OrderBy:  consts.OrderByReleasingDate,
		Page:     1,
		PageSize: 18,
	})).OrderBy(cardAMovieOrderClause(consts.OrderByReleasingDate, "desc")).Limit(18).ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	for _, bad := range []string{"JOIN", "COUNT(DISTINCT", "GROUP BY", "MAX("} {
		if strings.Contains(sqlText, bad) {
			t.Fatalf("cards a_movie path should not contain %s: %s", bad, sqlText)
		}
	}
	if !strings.Contains(sqlText, "FROM a_movie") || !strings.Contains(sqlText, "ORDER BY releasing_date DESC") {
		t.Fatalf("cards default should stay on a_movie release order: %s", sqlText)
	}
}

func TestCardGScOrderUsesDedicatedSourcePath(t *testing.T) {
	sqlText, _, err := squirrel.Select("wm.movie_jav_id").
		From(fmt.Sprintf("(SELECT DISTINCT movie_jav_id FROM w_media WHERE source_type = %d) wm", consts.WMediaSourceNative)).
		LeftJoin("g_sc_stat gss ON gss.movie_jav_id = wm.movie_jav_id").
		Where(squirrel.Eq{"wm.movie_jav_id": []string{"jav-a", "jav-b"}}).
		OrderBy(cardGScStatOrderClause(consts.OrderByScTimes, "desc")).
		Limit(18).
		ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	if strings.Contains(sqlText, "a_movie") || strings.Contains(sqlText, "MAX(") || strings.Contains(sqlText, "GROUP BY") {
		t.Fatalf("od=sc should not use wide a_movie grouped sort: %s", sqlText)
	}
	if !strings.Contains(sqlText, "g_sc_stat") || !strings.Contains(sqlText, "COALESCE(gss.sc_times, 0) DESC") {
		t.Fatalf("od=sc should use g_sc_stat order path: %s", sqlText)
	}
}

func TestCardMinfoRankOrderMatchesLegacyIndex(t *testing.T) {
	order := cardMinfoOrderClause(consts.OrderByRankDate, "desc")
	if order != "first_rank_day_number DESC, name DESC" {
		t.Fatalf("rank order should match bm_minfo idx_first_rank_name, got %q", order)
	}
	if strings.Contains(order, "jav_id") {
		t.Fatalf("rank order should not use jav_id tie breaker: %q", order)
	}
}

func TestCardWMediaOnlyQueryKeepsSingleSourcePage(t *testing.T) {
	where := cardWMediaWhere(&types.CardsListRequest{
		View:       "cardsmediamowned",
		MediaOwned: consts.OwnedAllNotRemoved,
		OrderBy:    consts.OrderByMediaBirthTime,
	})
	sqlText, _, err := applyWhere(squirrel.Select("movie_jav_id").From("w_media"), where).
		OrderBy(cardWMediaOrderClause(consts.OrderByMediaBirthTime, "desc")).
		Limit(18).
		ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	for _, bad := range []string{"IN (", "DISTINCT", "JOIN", "GROUP BY"} {
		if strings.Contains(sqlText, bad) {
			t.Fatalf("cardsmediamowned single-source path should not contain %s: %s", bad, sqlText)
		}
	}
	if !strings.Contains(sqlText, "FROM w_media") || !strings.Contains(sqlText, "ORDER BY birth_time DESC, movie_name DESC") {
		t.Fatalf("cardsmediamowned should page directly on w_media birth order: %s", sqlText)
	}
	if strings.Contains(sqlText, "movie_jav_id DESC") {
		t.Fatalf("cardsmediamowned order should match idx_w_media_birth_name without movie_jav_id tail: %s", sqlText)
	}
}

func TestCardAgeFiltersKeepExplicitZeroWithoutHiddenExclusion(t *testing.T) {
	repo := NewMovieReadRepo(nil, config.Config{})
	zero := 0.0
	req := &types.CardsListRequest{CastAgeMin: &zero, CastAgeMax: &zero}

	sqlText, args, err := applyWhere(squirrel.Select("jav_id").From("a_movie"), repo.cardAMovieWhere(req)).ToSql()
	if err != nil {
		t.Fatalf("build sql failed: %v", err)
	}
	if !strings.Contains(sqlText, "cast_average_age >= ?") || !strings.Contains(sqlText, "cast_average_age <= ?") {
		t.Fatalf("explicit zero age bounds should stay in sql: %s", sqlText)
	}
	if strings.Contains(sqlText, "cast_average_age <> ?") || strings.Contains(sqlText, "cast_average_age != ?") {
		t.Fatalf("age filters should not add hidden zero exclusion: %s", sqlText)
	}
	if len(args) != 2 || args[0] != int64(0) || args[1] != int64(0) {
		t.Fatalf("explicit zero age args should be preserved, got %#v", args)
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
