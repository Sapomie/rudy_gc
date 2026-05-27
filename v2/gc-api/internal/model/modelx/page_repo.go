package modelx

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rudy-gc-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const pageKindList = "list"
const pageKindOperation = "operation"

type PageRepo struct {
	conn sqlx.SqlConn
}

type PageConfig struct {
	Key                    string
	Title                  string
	Description            string
	Group                  string
	LegacyPath             string
	Kind                   string
	BaseSQL                string
	StaticWhere            string
	Columns                []*types.PageColumn
	SearchColumns          []string
	SortColumns            map[string]string
	OrderColumns           map[string]string
	DefaultOrderBy         string
	DefaultOrder           string
	DefaultPageSize        int64
	DefaultFilters         map[string]string
	DefaultFiltersSkipWhen map[string][]string
	Filters                []*types.PageFilter
	FilterRules            []*PageFilterRule
	Actions                []*types.PageAction
	Links                  []*types.PageLink
}

type PageFilterRule struct {
	Key       string
	Column    string
	Operator  string
	ValueType string
}

func NewPageRepo(conn sqlx.SqlConn) *PageRepo {
	return &PageRepo{conn: conn}
}

func PageSummaries() []*types.PageSummary {
	configs := pageConfigs()
	out := make([]*types.PageSummary, 0, len(configs))
	for _, cfg := range configs {
		out = append(out, &types.PageSummary{
			Key:         cfg.Key,
			Title:       cfg.Title,
			Description: cfg.Description,
			LegacyPath:  cfg.LegacyPath,
			Kind:        cfg.Kind,
			Group:       cfg.Group,
		})
	}
	return out
}

func FindPageConfig(key string) (*PageConfig, bool) {
	for _, cfg := range pageConfigs() {
		if cfg.Key == key {
			return cfg, true
		}
	}
	return nil, false
}

func (r *PageRepo) Load(ctx context.Context, key string, req *types.PageListRequest) (*types.PageListResponse, error) {
	if req == nil {
		req = &types.PageListRequest{}
	}
	cfg, ok := FindPageConfig(key)
	if !ok {
		return nil, fmt.Errorf("unknown page key: %s", key)
	}

	page := normalizePositive(req.Page, 1)
	pageSize := normalizePositive(req.PageSize, firstPositive(cfg.DefaultPageSize, 24))
	orderBy := normalizeOrderByKey(cfg, firstNonEmptyText(req.OrderBy, queryValue(req, "sort"), queryValue(req, "od"), cfg.DefaultOrderBy))
	order := normalizeOrder(firstNonEmptyText(req.Order, queryValue(req, "dir"), cfg.DefaultOrder))
	query := firstNonEmptyText(req.Query, queryValue(req, "keyword"))

	resp := &types.PageListResponse{
		Key:         cfg.Key,
		Title:       cfg.Title,
		Description: cfg.Description,
		LegacyPath:  cfg.LegacyPath,
		Kind:        cfg.Kind,
		Group:       cfg.Group,
		Page:        page,
		PageSize:    pageSize,
		OrderBy:     orderBy,
		Order:       order,
		Query:       query,
		Columns:     cfg.Columns,
		Rows:        []types.PageRow{},
		Filters:     cloneFilters(cfg, req),
		Actions:     cfg.Actions,
		Links:       cfg.Links,
		Stats:       buildPageStats(cfg),
	}
	if cfg.Kind == pageKindOperation || cfg.BaseSQL == "" {
		return resp, nil
	}
	if cfg.Key == "movie-agg-all-release" {
		return r.loadMovieReleaseAggPage(ctx, cfg, req, resp)
	}
	if r.conn == nil {
		return resp, nil
	}

	db, err := r.conn.RawDB()
	if err != nil {
		return nil, err
	}

	whereSQL, args := buildWhere(cfg, req)
	countSQL := "SELECT COUNT(*) " + cfg.BaseSQL + whereSQL
	if err := db.QueryRowContext(ctx, countSQL, args...).Scan(&resp.Total); err != nil {
		return nil, err
	}
	if resp.Total == 0 {
		return resp, nil
	}

	selectSQL := buildSelect(cfg)
	orderSQL := buildOrder(cfg, orderBy, order)
	offset := (page - 1) * pageSize
	querySQL := selectSQL + " " + cfg.BaseSQL + whereSQL + orderSQL + " LIMIT ? OFFSET ?"
	queryArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapped, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	resp.Rows = mapped
	return resp, nil
}

type movieReleaseAggScope struct {
	Key        string
	Level      string
	Year       int
	Quarter    int
	Month      int
	Day        int
	RangeStart string
	RangeEnd   string
}

type movieReleaseBucketRow struct {
	ID                  int64
	AggMode             string
	ScopeKey            string
	Level               string
	Year                int
	Quarter             int
	Month               int
	Day                 int
	CountAll            int64
	CountOwned          int64
	SizeBytes           int64
	LatestReleasingDate int64
}

type movieReleaseTopRow struct {
	ID         int64
	AggMode    string
	ScopeKey   string
	Level      string
	Year       int
	Quarter    int
	Month      int
	AggType    string
	AggID      int64
	AggKey     string
	AggName    string
	CountAll   int64
	CountOwned int64
	RankNo     int64
}

func (r *PageRepo) loadMovieReleaseAggPage(ctx context.Context, cfg *PageConfig, req *types.PageListRequest, resp *types.PageListResponse) (*types.PageListResponse, error) {
	mode := normalizeMovieReleaseAggMode(queryValueWithDefault(cfg, req, "agg_mode"))
	year := parseOptionalInt(queryValue(req, "year"))
	quarter := parseOptionalInt(queryValue(req, "quarter"))
	month := parseOptionalInt(queryValue(req, "month"))
	day := parseOptionalInt(queryValue(req, "day"))
	topN := parseMovieReleaseTopN(req)
	scope := buildMovieReleaseAggScope(year, quarter, month, day)

	resp.Title = movieReleaseAggTitle(mode, scope)
	resp.Page = 1
	resp.OrderBy = "default"
	resp.Order = "ASC"
	resp.Stats = append(resp.Stats,
		&types.PageStat{Label: "level", Value: scope.Level},
		&types.PageStat{Label: "range_start", Value: scope.RangeStart},
		&types.PageStat{Label: "range_end", Value: scope.RangeEnd},
		&types.PageStat{Label: "scope_key", Value: scope.Key},
	)
	if r.conn == nil {
		return resp, nil
	}
	db, err := r.conn.RawDB()
	if err != nil {
		return nil, err
	}

	rows := make([]types.PageRow, 0, 160)
	switch scope.Level {
	case "root":
		buckets, err := queryMovieReleaseBuckets(ctx, db, mode, "year", 0, 0, 0)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseBucketRowsToPageRows(buckets, mode, scope, "Year")...)
	case "year":
		quarterBuckets, err := queryMovieReleaseBuckets(ctx, db, mode, "quarter", scope.Year, 0, 0)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseBucketRowsToPageRows(quarterBuckets, mode, scope, "Quarter")...)
		monthBuckets, err := queryMovieReleaseBuckets(ctx, db, mode, "month", scope.Year, 0, 0)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseBucketRowsToPageRows(monthBuckets, mode, scope, "Month")...)
	case "quarter":
		buckets, err := queryMovieReleaseBuckets(ctx, db, mode, "month", scope.Year, scope.Quarter, 0)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseBucketRowsToPageRows(buckets, mode, scope, "Month")...)
	case "month":
		buckets, err := queryMovieReleaseBuckets(ctx, db, mode, "day", scope.Year, 0, scope.Month)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseBucketRowsToPageRows(buckets, mode, scope, "Day")...)
	}
	if scope.Level != "day" {
		topRows, err := queryMovieReleaseTopRows(ctx, db, mode, scope.Key, topN)
		if err != nil {
			return nil, err
		}
		rows = append(rows, movieReleaseTopRowsToPageRows(topRows, mode, scope)...)
	}

	resp.Rows = rows
	resp.Total = int64(len(rows))
	if len(rows) > 0 {
		resp.PageSize = int64(len(rows))
	}
	return resp, nil
}

func queryMovieReleaseBuckets(ctx context.Context, db *sql.DB, mode, level string, year, quarter, month int) ([]movieReleaseBucketRow, error) {
	clauses := []string{"agg_mode = ?", "level = ?", "count_all > 0"}
	args := []any{mode, level}
	if year > 0 {
		clauses = append(clauses, "year = ?")
		args = append(args, year)
	}
	if quarter > 0 {
		clauses = append(clauses, "quarter = ?")
		args = append(args, quarter)
	}
	if month > 0 {
		clauses = append(clauses, "month = ?")
		args = append(args, month)
	}
	query := `
SELECT id, agg_mode, scope_key, level, year, quarter, month, day, count_all, count_owned, size_bytes, latest_releasing_date
FROM movie_release_bucket_stat
WHERE ` + strings.Join(clauses, " AND ") + `
ORDER BY year DESC, quarter DESC, month DESC, day DESC, scope_key DESC`
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]movieReleaseBucketRow, 0)
	for rows.Next() {
		var row movieReleaseBucketRow
		if err := rows.Scan(&row.ID, &row.AggMode, &row.ScopeKey, &row.Level, &row.Year, &row.Quarter, &row.Month, &row.Day, &row.CountAll, &row.CountOwned, &row.SizeBytes, &row.LatestReleasingDate); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func queryMovieReleaseTopRows(ctx context.Context, db *sql.DB, mode, scopeKey string, topN int) ([]movieReleaseTopRow, error) {
	out := make([]movieReleaseTopRow, 0, topN*4)
	for _, aggType := range []string{"cast", "director", "label", "prefix"} {
		rows, err := db.QueryContext(ctx, `
SELECT id, agg_mode, scope_key, level, year, quarter, month, agg_type, agg_id, agg_key, agg_name, count_all, count_owned, rank_no
FROM movie_release_top_stat
WHERE agg_mode = ? AND scope_key = ? AND agg_type = ? AND rank_no <= ?
ORDER BY rank_no ASC, agg_name ASC`, mode, scopeKey, aggType, topN)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var row movieReleaseTopRow
			if err := rows.Scan(&row.ID, &row.AggMode, &row.ScopeKey, &row.Level, &row.Year, &row.Quarter, &row.Month, &row.AggType, &row.AggID, &row.AggKey, &row.AggName, &row.CountAll, &row.CountOwned, &row.RankNo); err != nil {
				rows.Close()
				return nil, err
			}
			out = append(out, row)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return out, nil
}

func movieReleaseBucketRowsToPageRows(rows []movieReleaseBucketRow, mode string, scope movieReleaseAggScope, section string) []types.PageRow {
	out := make([]types.PageRow, 0, len(rows))
	for _, row := range rows {
		hrefQuarter := row.Quarter
		if row.Level == "month" || row.Level == "day" {
			hrefQuarter = 0
		}
		out = append(out, types.PageRow{
			"id":          fmt.Sprintf("%s:bucket:%s", mode, row.ScopeKey),
			"href":        buildMovieReleaseAggHref(mode, row.Year, hrefQuarter, row.Month, row.Day),
			"agg_mode":    mode,
			"section":     section,
			"scope_key":   row.ScopeKey,
			"level":       row.Level,
			"page_level":  scope.Level,
			"year":        strconv.Itoa(row.Year),
			"quarter":     strconv.Itoa(row.Quarter),
			"month":       strconv.Itoa(row.Month),
			"day":         strconv.Itoa(row.Day),
			"label":       movieReleaseBucketLabel(row),
			"count_all":   strconv.FormatInt(row.CountAll, 10),
			"count_owned": strconv.FormatInt(row.CountOwned, 10),
			"size_bytes":  strconv.FormatInt(row.SizeBytes, 10),
			"latest_time": strconv.FormatInt(row.LatestReleasingDate, 10),
			"range_start": scope.RangeStart,
			"range_end":   scope.RangeEnd,
			"rank_no":     "",
			"agg_type":    "",
		})
	}
	return out
}

func movieReleaseTopRowsToPageRows(rows []movieReleaseTopRow, mode string, scope movieReleaseAggScope) []types.PageRow {
	out := make([]types.PageRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, types.PageRow{
			"id":          fmt.Sprintf("%s:%s:%s", mode, row.AggType, row.AggKey),
			"href":        buildMovieReleaseTopHref(row, mode, scope),
			"agg_mode":    mode,
			"section":     movieReleaseTopSection(row.AggType),
			"scope_key":   row.ScopeKey,
			"level":       row.Level,
			"page_level":  scope.Level,
			"year":        strconv.Itoa(row.Year),
			"quarter":     strconv.Itoa(row.Quarter),
			"month":       strconv.Itoa(row.Month),
			"day":         "",
			"label":       row.AggName,
			"count_all":   strconv.FormatInt(row.CountAll, 10),
			"count_owned": strconv.FormatInt(row.CountOwned, 10),
			"size_bytes":  "0",
			"latest_time": "0",
			"range_start": scope.RangeStart,
			"range_end":   scope.RangeEnd,
			"rank_no":     strconv.FormatInt(row.RankNo, 10),
			"agg_type":    row.AggType,
		})
	}
	return out
}

func normalizeMovieReleaseAggMode(mode string) string {
	if strings.TrimSpace(mode) == "owned" {
		return "owned"
	}
	return "all"
}

func parseOptionalInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return value
}

func parseMovieReleaseTopN(req *types.PageListRequest) int {
	topN := 30
	for _, key := range []string{"tn", "top", "topn", "tc"} {
		raw := strings.TrimSpace(queryValue(req, key))
		if raw != "" {
			topN = parseOptionalInt(raw)
			break
		}
	}
	if topN < 1 {
		return 1
	}
	if topN > 200 {
		return 200
	}
	return topN
}

func buildMovieReleaseAggScope(year, quarter, month, day int) movieReleaseAggScope {
	if year <= 0 {
		return movieReleaseAggScope{
			Key:        "root",
			Level:      "root",
			RangeStart: "",
			RangeEnd:   "",
		}
	}
	if day > 0 {
		start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		end := start.Add(24*time.Hour - time.Second)
		return movieReleaseAggScope{
			Key:        fmt.Sprintf("day:%04d-%02d-%02d", year, month, day),
			Level:      "day",
			Year:       year,
			Quarter:    movieReleaseMonthToQuarter(month),
			Month:      month,
			Day:        day,
			RangeStart: start.Format(time.DateOnly),
			RangeEnd:   end.Format(time.DateOnly),
		}
	}
	if month > 0 {
		start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return movieReleaseAggScope{
			Key:        fmt.Sprintf("month:%04d-%02d", year, month),
			Level:      "month",
			Year:       year,
			Quarter:    movieReleaseMonthToQuarter(month),
			Month:      month,
			RangeStart: start.Format(time.DateOnly),
			RangeEnd:   end.Format(time.DateOnly),
		}
	}
	if quarter > 0 {
		startMonth := 3*(quarter-1) + 1
		start := time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
		end := start.AddDate(0, 3, 0).Add(-time.Second)
		return movieReleaseAggScope{
			Key:        fmt.Sprintf("quarter:%04d-Q%d", year, quarter),
			Level:      "quarter",
			Year:       year,
			Quarter:    quarter,
			RangeStart: start.Format(time.DateOnly),
			RangeEnd:   end.Format(time.DateOnly),
		}
	}
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.Local).Add(-time.Second)
	return movieReleaseAggScope{
		Key:        fmt.Sprintf("year:%04d", year),
		Level:      "year",
		Year:       year,
		RangeStart: start.Format(time.DateOnly),
		RangeEnd:   end.Format(time.DateOnly),
	}
}

func movieReleaseMonthToQuarter(month int) int {
	return ((month - 1) / 3) + 1
}

func movieReleaseAggTitle(mode string, scope movieReleaseAggScope) string {
	label := "全量"
	if mode == "owned" {
		label = "已拥有"
	}
	switch scope.Level {
	case "year":
		return fmt.Sprintf("上映日（%s）· %d 年", label, scope.Year)
	case "quarter":
		return fmt.Sprintf("上映日（%s）· %d 年 Q%d", label, scope.Year, scope.Quarter)
	case "month":
		return fmt.Sprintf("上映日（%s）· %d 年 %02d 月", label, scope.Year, scope.Month)
	case "day":
		return fmt.Sprintf("上映日（%s）· %04d-%02d-%02d", label, scope.Year, scope.Month, scope.Day)
	default:
		return fmt.Sprintf("电影聚合（%s）· 上映日", label)
	}
}

func buildMovieReleaseAggHref(mode string, year, quarter, month, day int) string {
	q := url.Values{}
	if normalizeMovieReleaseAggMode(mode) == "owned" {
		q.Set("agg_mode", "owned")
	}
	if year > 0 {
		q.Set("year", strconv.Itoa(year))
	}
	if quarter > 0 {
		q.Set("quarter", strconv.Itoa(quarter))
	}
	if month > 0 {
		q.Set("month", strconv.Itoa(month))
	}
	if day > 0 {
		q.Set("day", strconv.Itoa(day))
	}
	href := "/movie-agg-all/release"
	if encoded := q.Encode(); encoded != "" {
		href += "?" + encoded
	}
	return href
}

func movieReleaseBucketLabel(row movieReleaseBucketRow) string {
	switch row.Level {
	case "year":
		return fmt.Sprintf("%d 年", row.Year)
	case "quarter":
		return fmt.Sprintf("Q%d 季", row.Quarter)
	case "month":
		return fmt.Sprintf("%02d 月", row.Month)
	case "day":
		return fmt.Sprintf("%02d 日", row.Day)
	default:
		return row.ScopeKey
	}
}

func movieReleaseTopSection(aggType string) string {
	switch aggType {
	case "cast":
		return "Top Cast"
	case "director":
		return "Top Director"
	case "label":
		return "Top Label"
	default:
		return "Top Prefix"
	}
}

func buildMovieReleaseTopHref(row movieReleaseTopRow, mode string, scope movieReleaseAggScope) string {
	filter := "mowned=1"
	if mode == "owned" {
		filter = "mowned=3"
	}
	var href string
	switch row.AggType {
	case "cast":
		if row.AggID > 0 {
			href = fmt.Sprintf("/cards?pid=%d&%s", row.AggID, filter)
		} else {
			href = "/cards?cn=" + url.QueryEscape(row.AggName) + "&" + filter
		}
	case "director":
		href = "/cards?dn=" + url.QueryEscape(row.AggName) + "&" + filter
	case "label":
		href = "/cards?ln=" + url.QueryEscape(row.AggName) + "&" + filter
	case "prefix":
		href = "/cards?pn=" + url.QueryEscape(row.AggName) + "&" + filter
	default:
		return ""
	}
	if scope.RangeStart != "" {
		href += "&rs=" + url.QueryEscape(scope.RangeStart)
	}
	if scope.RangeEnd != "" {
		href += "&re=" + url.QueryEscape(scope.RangeEnd)
	}
	return href
}

func buildSelect(cfg *PageConfig) string {
	parts := make([]string, 0, len(cfg.Columns))
	for _, column := range cfg.Columns {
		parts = append(parts, fmt.Sprintf("%s AS `%s`", cfg.SortColumns[column.Key], column.Key))
	}
	return "SELECT " + strings.Join(parts, ", ")
}

func buildWhere(cfg *PageConfig, req *types.PageListRequest) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 8)
	if strings.TrimSpace(cfg.StaticWhere) != "" {
		clauses = append(clauses, "("+strings.TrimSpace(cfg.StaticWhere)+")")
	}
	query := firstNonEmptyText(req.Query, queryValue(req, "keyword"))
	if query != "" && len(cfg.SearchColumns) > 0 {
		searchParts := make([]string, 0, len(cfg.SearchColumns))
		for _, column := range cfg.SearchColumns {
			searchParts = append(searchParts, column+" LIKE ?")
			args = append(args, "%"+query+"%")
		}
		clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
	}
	if req.ID != nil {
		if column, ok := cfg.SortColumns["id"]; ok && !hasFilterRule(cfg, "id") {
			clauses = append(clauses, column+" = ?")
			args = append(args, *req.ID)
		}
	}
	if strings.TrimSpace(req.Name) != "" {
		if column, ok := cfg.SortColumns["name"]; ok && !hasFilterRule(cfg, "name") {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Name))
		}
	}
	if strings.TrimSpace(req.Level) != "" {
		if column, ok := cfg.SortColumns["level"]; ok && !hasFilterRule(cfg, "level") {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Level))
		}
	}
	if strings.TrimSpace(req.Status) != "" {
		if column, ok := cfg.SortColumns["status"]; ok && !hasFilterRule(cfg, "status") {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Status))
		}
	}
	if strings.TrimSpace(req.Album) != "" {
		if column, ok := cfg.SortColumns["album_name"]; ok && !hasFilterRule(cfg, "album") && !hasFilterRule(cfg, "album_name") {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Album))
		}
	}
	for _, rule := range cfg.FilterRules {
		if rule == nil || strings.TrimSpace(rule.Key) == "" || strings.TrimSpace(rule.Column) == "" {
			continue
		}
		raw := queryValueWithDefault(cfg, req, rule.Key)
		if strings.TrimSpace(raw) == "" {
			continue
		}
		clause, valueArgs, ok, err := buildFilterClause(rule, raw)
		if err != nil {
			clauses = append(clauses, "1 = 0")
			continue
		}
		if !ok {
			continue
		}
		clauses = append(clauses, clause)
		args = append(args, valueArgs...)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func buildFilterClause(rule *PageFilterRule, raw string) (string, []any, bool, error) {
	value := strings.TrimSpace(raw)
	operator := strings.TrimSpace(rule.Operator)
	if operator == "" {
		operator = "="
	}
	switch rule.ValueType {
	case "like":
		return rule.Column + " LIKE ?", []any{"%" + value + "%"}, true, nil
	case "int":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	case "float":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	case "gb":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{int64(v * 1024 * 1024 * 1024)}, true, nil
	case "minutes":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{v * 60}, true, nil
	case "date_start":
		ts, err := parsePageDateStart(value)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{ts}, true, nil
	case "date_end":
		ts, err := parsePageDateEnd(value)
		if err != nil {
			return "", nil, false, err
		}
		return rule.Column + " " + operator + " ?", []any{ts}, true, nil
	case "item_detail_status":
		v, ok := mapTextValue(value, map[string]int64{"none": 1, "ok": 2})
		if !ok {
			return "", nil, false, fmt.Errorf("invalid item detail status")
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	case "item_cover_status":
		v, ok := mapTextValue(value, map[string]int64{"none": 1, "ok": 2, "wrong": 3})
		if !ok {
			return "", nil, false, fmt.Errorf("invalid item cover status")
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	case "item_chinese_status":
		v, ok := mapTextValue(value, map[string]int64{"none": 1, "ok": 2, "error": 3, "sensitive": 4})
		if !ok {
			return "", nil, false, fmt.Errorf("invalid item chinese status")
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	case "item_need_scan_status":
		v, ok := mapTextValue(value, map[string]int64{"need": 1, "no_need": 2, "wrong": 3})
		if !ok {
			return "", nil, false, fmt.Errorf("invalid item need scan status")
		}
		return rule.Column + " " + operator + " ?", []any{v}, true, nil
	default:
		return rule.Column + " " + operator + " ?", []any{value}, true, nil
	}
}

func parsePageDateStart(raw string) (int64, error) {
	t, err := time.ParseInLocation(time.DateOnly, raw, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func parsePageDateEnd(raw string) (int64, error) {
	t, err := time.ParseInLocation(time.DateOnly, raw, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Add(24*time.Hour - time.Second).Unix(), nil
}

func buildOrder(cfg *PageConfig, orderBy, order string) string {
	column := cfg.OrderColumns[orderBy]
	if column == "" {
		column = cfg.SortColumns[orderBy]
	}
	if column == "" {
		column = cfg.OrderColumns[cfg.DefaultOrderBy]
		if column == "" {
			column = cfg.SortColumns[cfg.DefaultOrderBy]
		}
	}
	if column == "" {
		return ""
	}
	normalized := normalizeOrder(order)
	if strings.Contains(column, "{order}") {
		return " ORDER BY " + strings.ReplaceAll(column, "{order}", normalized)
	}
	return " ORDER BY " + column + " " + normalized
}

func scanRows(rows *sql.Rows) ([]types.PageRow, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make([]types.PageRow, 0)
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		dest := make([]any, len(columns))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		row := make(types.PageRow, len(columns))
		for i, column := range columns {
			if values[i].Valid {
				row[column] = values[i].String
			} else {
				row[column] = ""
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func cloneFilters(cfg *PageConfig, req *types.PageListRequest) []*types.PageFilter {
	filters := cfg.Filters
	out := make([]*types.PageFilter, 0, len(filters)+1)
	out = append(out, &types.PageFilter{Key: "q", Label: "关键词", Type: "text", Placeholder: "按当前页面字段搜索", Value: firstNonEmptyText(req.Query, queryValue(req, "keyword"))})
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		copied := *filter
		if value := queryValueWithDefault(cfg, req, copied.Key); value != "" {
			copied.Value = value
		} else {
			switch copied.Key {
			case "id":
				if req.ID != nil {
					copied.Value = fmt.Sprintf("%d", *req.ID)
				}
			case "name":
				copied.Value = req.Name
			case "level":
				copied.Value = req.Level
			case "status":
				copied.Value = req.Status
			case "album", "album_name":
				copied.Value = firstNonEmptyText(req.Album, queryValue(req, "album_name"))
			}
		}
		out = append(out, &copied)
	}
	return out
}

func buildPageStats(cfg *PageConfig) []*types.PageStat {
	return []*types.PageStat{
		{Label: "legacy", Value: cfg.LegacyPath},
		{Label: "kind", Value: cfg.Kind},
	}
}

func normalizePositive(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}

func firstPositive(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func normalizeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func normalizeOrderByKey(cfg *PageConfig, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return cfg.DefaultOrderBy
	}
	if cfg.SortColumns[value] != "" || cfg.OrderColumns[value] != "" {
		return value
	}
	return cfg.DefaultOrderBy
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func col(key, label string, sortable bool) *types.PageColumn {
	return &types.PageColumn{Key: key, Label: label, Sortable: sortable}
}

func hiddenCol(key string) *types.PageColumn {
	return &types.PageColumn{Key: key, Label: key, Hidden: true}
}

func linkCol(key, label, link string, sortable bool) *types.PageColumn {
	return &types.PageColumn{Key: key, Label: label, Link: link, Sortable: sortable}
}

func link(label, href, kind string) *types.PageLink {
	return &types.PageLink{Label: label, Href: href, Kind: kind}
}

func textFilter(key, label, placeholder string) *types.PageFilter {
	return &types.PageFilter{Key: key, Label: label, Type: "text", Placeholder: placeholder}
}

func selectFilter(key, label string, options ...*types.PageFilterOption) *types.PageFilter {
	return &types.PageFilter{Key: key, Label: label, Type: "select", Options: options}
}

func opt(label, value string) *types.PageFilterOption {
	return &types.PageFilterOption{Label: label, Value: value}
}

func numberFilter(key, label, placeholder string) *types.PageFilter {
	return &types.PageFilter{Key: key, Label: label, Type: "number", Placeholder: placeholder}
}

func dateFilter(key, label string) *types.PageFilter {
	return &types.PageFilter{Key: key, Label: label, Type: "date"}
}

func rule(key, column, operator, valueType string) *PageFilterRule {
	return &PageFilterRule{Key: key, Column: column, Operator: operator, ValueType: valueType}
}

func hasFilterRule(cfg *PageConfig, key string) bool {
	for _, rule := range cfg.FilterRules {
		if rule != nil && rule.Key == key {
			return true
		}
	}
	return false
}

func queryValueWithDefault(cfg *PageConfig, req *types.PageListRequest, key string) string {
	if value := queryValue(req, key); value != "" {
		return value
	}
	if cfg == nil || len(cfg.DefaultFilters) == 0 {
		return ""
	}
	if skipKeys := cfg.DefaultFiltersSkipWhen[key]; len(skipKeys) > 0 {
		for _, skipKey := range skipKeys {
			if queryValue(req, skipKey) != "" {
				return ""
			}
		}
	}
	return strings.TrimSpace(cfg.DefaultFilters[key])
}

func queryValue(req *types.PageListRequest, key string) string {
	if req == nil || len(req.Extra) == 0 {
		return ""
	}
	values := req.Extra[key]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[len(values)-1])
}

func mapTextValue(raw string, values map[string]int64) (int64, bool) {
	value, ok := values[strings.TrimSpace(raw)]
	return value, ok
}
