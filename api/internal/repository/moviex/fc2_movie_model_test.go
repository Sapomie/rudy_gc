package moviex

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"

	"rudy_gc/internal/types"
)

func TestFC2MovieOrderBy(t *testing.T) {
	playOrder := fc2MovieOrderBy("play", "asc")
	if len(playOrder) == 0 || playOrder[0] != "`media_present` ASC" {
		t.Fatalf("play order = %#v", playOrder)
	}
	keyOrder := fc2MovieOrderBy("fc2_key", "desc")
	if len(keyOrder) != 3 || keyOrder[0] != "`ppv_no` DESC" || keyOrder[1] != "`code_seq` DESC" {
		t.Fatalf("key order = %#v", keyOrder)
	}
	titleOrder := fc2MovieOrderBy("title", "asc")
	if len(titleOrder) == 0 || titleOrder[0] != "`title` ASC" {
		t.Fatalf("title order = %#v", titleOrder)
	}
	birthOrder := fc2MovieOrderBy("birth_time", "desc")
	if len(birthOrder) == 0 || birthOrder[0] != "`media_birth_time` DESC" {
		t.Fatalf("birth order = %#v", birthOrder)
	}
	for _, order := range append(playOrder, birthOrder...) {
		upper := strings.ToUpper(order)
		if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS") || strings.Contains(upper, "SELECT") {
			t.Fatalf("关联查询残留: %s", order)
		}
	}
}

func TestFC2MovieListQueryHasNoRelatedQuery(t *testing.T) {
	query, _, err := applyFC2MovieFilter(squirrel.Select(fc2MovieRows).From("`fc2_movie`"), "FC2PPV", "ready").
		OrderBy(fc2MovieOrderBy("birth_time", "desc")...).
		Limit(50).
		ToSql()
	if err != nil {
		t.Fatalf("build fc2 movie query: %v", err)
	}
	upper := strings.ToUpper(query)
	if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS (") || strings.Contains(upper, "(SELECT") {
		t.Fatalf("fc2 movie query contains related query: %s", query)
	}
}

func TestApplyFC2CardFilterOwned(t *testing.T) {
	query, _, err := applyFC2CardFilter(squirrel.Select(fc2MovieRows).From("`fc2_movie`"), true).ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "WHERE `media_present` = 1") {
		t.Fatalf("owned card query = %s", query)
	}

	allQuery, _, err := applyFC2CardFilter(squirrel.Select(fc2MovieRows).From("`fc2_movie`"), false).ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(allQuery, "WHERE `media_present` = 1") {
		t.Fatalf("all card query should not filter media_present: %s", allQuery)
	}
}

func TestBuildFC2SmartPickMediaQueryUsesSingleTableFilters(t *testing.T) {
	hasSub := int64(2)
	selfMake := int64(1)
	hasMask := int64(3)
	query, args, err := buildFC2SmartPickMediaQuery(FC2SmartPickMediaFilter{
		BirthTimeStart: 100,
		BirthTimeEnd:   200,
		MediaDir1:      "FC2",
		MediaDir2:      "sub",
		HasSub:         &hasSub,
		SelfMake:       &selfMake,
		HasMask:        &hasMask,
	}).ToSql()
	if err != nil {
		t.Fatalf("build query: %v", err)
	}
	lowerQuery := strings.ToLower(query)
	if strings.Contains(lowerQuery, " join ") {
		t.Fatalf("FC2 Smart Pick 查询禁止 JOIN: %s", query)
	}
	for _, fragment := range []string{"from `fc2_media`", "`is_removed` = ?", "`birth_time` >= ?", "`birth_time` <= ?", "`has_sub` = ?", "`self_make` = ?", "`has_mask` = ?"} {
		if !strings.Contains(lowerQuery, strings.ToLower(fragment)) {
			t.Fatalf("query missing %q: %s", fragment, query)
		}
	}
	if len(args) != 8 {
		t.Fatalf("args = %#v", args)
	}
}

func TestNativeMediaBaseQueryHasNoJoin(t *testing.T) {
	query, _, err := nativeMediaListBuilder("`w_media`", "wm").ToSql()
	if err != nil {
		t.Fatalf("build native media query: %v", err)
	}
	upper := strings.ToUpper(query)
	if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS") || strings.Contains(upper, "SELECT (") {
		t.Fatalf("native media query contains related query: %s", query)
	}
	if strings.Contains(strings.ToLower(query), "g_sc_stat") {
		t.Fatalf("native media query contains g_sc_stat: %s", query)
	}
}

func TestSortAndFilterNativeMediaStats(t *testing.T) {
	rows := []*NativeMediaListRow{
		{Id: 1, MovieName: "A", ScTimes: 1, LastScTime: 100},
		{Id: 2, MovieName: "B", ScTimes: 3, LastScTime: 90},
		{Id: 3, MovieName: "C", ScTimes: 3, LastScTime: 110},
	}
	filtered := filterNativeMediaStats(rows, types.MediaListFilter{HasScTimesMin: true, ScTimesMin: 2})
	if len(filtered) != 2 {
		t.Fatalf("filtered count = %d", len(filtered))
	}
	sortNativeMediaRows(filtered, "COALESCE(gss.sc_times, 0) DESC, COALESCE(gss.last_sc_time, 0) DESC")
	if filtered[0].Id != 3 || filtered[1].Id != 2 {
		t.Fatalf("sorted ids = %d,%d", filtered[0].Id, filtered[1].Id)
	}
}

func TestValidateScEventMovieIdentity(t *testing.T) {
	validRows := []*GList{
		{Name: "event__ABP-123", ScName: "event", MovieType: "jav", MovieJavId: "ABP-123"},
		{Name: "event__FC2PPV-42-1", ScName: "event", MovieType: "fc2ppv", FC2MovieId: 42},
	}
	for _, row := range validRows {
		if err := validateScEventMovieIdentity(row); err != nil {
			t.Fatalf("valid row rejected: %v", err)
		}
	}
	invalidRows := []*GList{
		{Name: "event__bad", ScName: "event", MovieType: "jav", MovieJavId: "ABP-123", FC2MovieId: 42},
		{Name: "event__bad", ScName: "event", MovieType: "fc2ppv", MovieJavId: "ABP-123", FC2MovieId: 42},
	}
	for _, row := range invalidRows {
		if err := validateScEventMovieIdentity(row); err == nil {
			t.Fatalf("invalid row accepted: %#v", row)
		}
	}
}
