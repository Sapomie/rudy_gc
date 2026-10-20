package moviex

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"

	"rudy_gc/local/api/internal/types"
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
	statOrders := map[string]string{
		"sc_times":              "COALESCE(`fss`.`sc_times`, 0) ASC",
		"come_times":            "COALESCE(`fss`.`come_times`, 0) ASC",
		"last_selected_sc_time": "COALESCE(`fss`.`last_selected_sc_time`, 0) ASC",
	}
	for field, expected := range statOrders {
		order := fc2MovieOrderBy(field, "asc")
		if len(order) == 0 || order[0] != expected {
			t.Fatalf("%s order = %#v", field, order)
		}
	}
	for _, order := range append(playOrder, birthOrder...) {
		upper := strings.ToUpper(order)
		if strings.Contains(upper, "JOIN") || strings.Contains(upper, "EXISTS") || strings.Contains(upper, "SELECT") {
			t.Fatalf("关联查询残留: %s", order)
		}
	}
}

func TestBuildFC2MovieListQueryJoinsStatsBeforePagination(t *testing.T) {
	query, args, err := buildFC2MovieListQuery(50, 50, "FC2PPV", "ready", "come_times", "desc").ToSql()
	if err != nil {
		t.Fatalf("build FC2 stat sort query: %v", err)
	}
	for _, fragment := range []string{
		"LEFT JOIN `fc2_sc_stat` AS `fss` ON `fss`.`fc2_movie_id` = `fm`.`id`",
		"ORDER BY COALESCE(`fss`.`come_times`, 0) DESC",
		"LIMIT 50 OFFSET 50",
	} {
		if !strings.Contains(query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, query)
		}
	}
	if len(args) != 5 {
		t.Fatalf("args = %#v", args)
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

func TestBuildFC2CardListQuerySupportsRequestedSorts(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		order     string
		ownedOnly bool
		want      []string
		forbidden []string
	}{
		{
			name:      "post time",
			field:     "post_time",
			order:     "desc",
			ownedOnly: false,
			want:      []string{"FROM `fc2_movie` FORCE INDEX (`idx_fc2_movie_post_time`)", "ORDER BY `post_time` DESC, `id` DESC", "LIMIT 18"},
			forbidden: []string{"JOIN `fc2_sc_stat`"},
		},
		{
			name:      "fc2 key natural order",
			field:     "fc2_key",
			order:     "asc",
			ownedOnly: false,
			want:      []string{"FROM `fc2_movie` FORCE INDEX (`idx_fc2_movie_card_key_sort`)", "ORDER BY `ppv_no` ASC, `code_seq` ASC, `id` ASC", "LIMIT 18 OFFSET 18"},
		},
		{
			name:      "birth time owned",
			field:     "birth_time",
			order:     "desc",
			ownedOnly: true,
			want:      []string{"FROM `fc2_movie` FORCE INDEX (`idx_fc2_movie_media_birth_time`)", "WHERE `media_present` = 1", "ORDER BY `media_birth_time` DESC, `id` DESC"},
			forbidden: []string{"JOIN `fc2_sc_stat`"},
		},
		{
			name:      "last selected sc time owned",
			field:     "last_selected_sc_time",
			order:     "desc",
			ownedOnly: true,
			want: []string{
				"LEFT JOIN `fc2_sc_stat` AS `fss` ON `fss`.`fc2_movie_id` = `fm`.`id`",
				"WHERE `fm`.`media_present` = 1",
				"ORDER BY COALESCE(`fss`.`last_selected_sc_time`, 0) DESC",
			},
		},
		{
			name:      "sc times",
			field:     "sc_times",
			order:     "asc",
			ownedOnly: false,
			want: []string{
				"LEFT JOIN `fc2_sc_stat` AS `fss` ON `fss`.`fc2_movie_id` = `fm`.`id`",
				"ORDER BY COALESCE(`fss`.`sc_times`, 0) ASC",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offset := int64(0)
			if test.field == "fc2_key" {
				offset = 18
			}
			query, _, err := buildFC2CardListQuery(offset, 18, test.ownedOnly, false, test.field, test.order).ToSql()
			if err != nil {
				t.Fatalf("build query: %v", err)
			}
			for _, fragment := range test.want {
				if !strings.Contains(query, fragment) {
					t.Fatalf("query missing %q: %s", fragment, query)
				}
			}
			for _, fragment := range test.forbidden {
				if strings.Contains(query, fragment) {
					t.Fatalf("query contains %q: %s", fragment, query)
				}
			}
		})
	}
}

func TestBuildFC2CardListQueryKeepsRandomOrderIndependent(t *testing.T) {
	query, _, err := buildFC2CardListQuery(0, 18, false, true, "sc_times", "desc").ToSql()
	if err != nil {
		t.Fatalf("build random query: %v", err)
	}
	if !strings.Contains(query, "ORDER BY RAND()") {
		t.Fatalf("random query = %s", query)
	}
	if strings.Contains(query, "fc2_sc_stat") {
		t.Fatalf("random query should not join stats: %s", query)
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
		{Id: 1, MovieName: "A", ScTimes: 1, LastSelectedScTime: 100},
		{Id: 2, MovieName: "B", ScTimes: 3, LastSelectedScTime: 90},
		{Id: 3, MovieName: "C", ScTimes: 3, LastSelectedScTime: 110},
	}
	filtered := filterNativeMediaStats(rows, types.MediaListFilter{HasScTimesMin: true, ScTimesMin: 2})
	if len(filtered) != 2 {
		t.Fatalf("filtered count = %d", len(filtered))
	}
	sortNativeMediaRows(filtered, "COALESCE(gss.sc_times, 0) DESC, COALESCE(gss.last_selected_sc_time, 0) DESC")
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
