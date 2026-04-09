package moviex

import (
	"strconv"
	"strings"

	"rudy_gc/internal/consts"
)

func wMediaSourceTypeSQL(sourceType int64) string {
	return strconv.FormatInt(sourceType, 10)
}

func nativeWMediaSourceTypeSQL() string {
	return wMediaSourceTypeSQL(consts.WMediaSourceNative)
}

func legacyWMediaSourceTypeSQL() string {
	return wMediaSourceTypeSQL(consts.WMediaSourceLegacyVFilm)
}

func buildWMediaJoin(tableExpr, alias, outerMovieJavID string, sourceType int64) string {
	return tableExpr + " " + alias + " ON " + alias + ".movie_jav_id = " + outerMovieJavID + " AND " + alias + ".source_type = " + wMediaSourceTypeSQL(sourceType)
}

func buildNativeWMediaJoin(tableExpr, alias, outerMovieJavID string) string {
	return buildWMediaJoin(tableExpr, alias, outerMovieJavID, consts.WMediaSourceNative)
}

func buildLegacyWMediaJoin(tableExpr, alias, outerMovieJavID string) string {
	return buildWMediaJoin(tableExpr, alias, outerMovieJavID, consts.WMediaSourceLegacyVFilm)
}

func buildWMediaExists(tableExpr, alias, outerMovieJavID string, sourceType int64, extraConditions ...string) string {
	conditions := []string{
		alias + ".movie_jav_id = " + outerMovieJavID,
		alias + ".source_type = " + wMediaSourceTypeSQL(sourceType),
	}
	for _, condition := range extraConditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}
		conditions = append(conditions, condition)
	}
	return "EXISTS (SELECT 1 FROM " + tableExpr + " " + alias + " WHERE " + strings.Join(conditions, " AND ") + ")"
}

func buildNativeWMediaExists(tableExpr, alias, outerMovieJavID string, extraConditions ...string) string {
	return buildWMediaExists(tableExpr, alias, outerMovieJavID, consts.WMediaSourceNative, extraConditions...)
}

func buildLegacyWMediaExists(tableExpr, alias, outerMovieJavID string, extraConditions ...string) string {
	return buildWMediaExists(tableExpr, alias, outerMovieJavID, consts.WMediaSourceLegacyVFilm, extraConditions...)
}

func buildWMediaNotExists(tableExpr, alias, outerMovieJavID string, sourceType int64, extraConditions ...string) string {
	conditions := []string{
		alias + ".movie_jav_id = " + outerMovieJavID,
		alias + ".source_type = " + wMediaSourceTypeSQL(sourceType),
	}
	for _, condition := range extraConditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}
		conditions = append(conditions, condition)
	}
	return "NOT EXISTS (SELECT 1 FROM " + tableExpr + " " + alias + " WHERE " + strings.Join(conditions, " AND ") + ")"
}

func buildNativeWMediaNotExists(tableExpr, alias, outerMovieJavID string, extraConditions ...string) string {
	return buildWMediaNotExists(tableExpr, alias, outerMovieJavID, consts.WMediaSourceNative, extraConditions...)
}

func buildLegacyWMediaNotExists(tableExpr, alias, outerMovieJavID string, extraConditions ...string) string {
	return buildWMediaNotExists(tableExpr, alias, outerMovieJavID, consts.WMediaSourceLegacyVFilm, extraConditions...)
}
