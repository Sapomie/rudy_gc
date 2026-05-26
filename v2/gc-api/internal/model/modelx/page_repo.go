package modelx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"rudy-gc-api/internal/types"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const pageKindList = "list"
const pageKindOperation = "operation"

type PageRepo struct {
	conn sqlx.SqlConn
}

type PageConfig struct {
	Key            string
	Title          string
	Description    string
	Group          string
	LegacyPath     string
	Kind           string
	BaseSQL        string
	Columns        []*types.PageColumn
	SearchColumns  []string
	SortColumns    map[string]string
	DefaultOrderBy string
	DefaultOrder   string
	Filters        []*types.PageFilter
	Actions        []*types.PageAction
	Links          []*types.PageLink
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
	cfg, ok := FindPageConfig(key)
	if !ok {
		return nil, fmt.Errorf("unknown page key: %s", key)
	}

	page := normalizePositive(req.Page, 1)
	pageSize := normalizePositive(req.PageSize, 24)
	orderBy := firstNonEmptyText(req.OrderBy, cfg.DefaultOrderBy)
	order := normalizeOrder(firstNonEmptyText(req.Order, cfg.DefaultOrder))

	resp := &types.PageListResponse{
		Key:         cfg.Key,
		Title:       cfg.Title,
		Description: cfg.Description,
		LegacyPath:  cfg.LegacyPath,
		Kind:        cfg.Kind,
		Page:        page,
		PageSize:    pageSize,
		OrderBy:     orderBy,
		Order:       order,
		Query:       strings.TrimSpace(req.Query),
		Columns:     cfg.Columns,
		Rows:        []types.PageRow{},
		Filters:     cloneFilters(cfg.Filters, req),
		Actions:     cfg.Actions,
		Links:       cfg.Links,
		Stats:       buildPageStats(cfg),
	}
	if cfg.Kind == pageKindOperation || cfg.BaseSQL == "" {
		return resp, nil
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
	query := strings.TrimSpace(req.Query)
	if query != "" && len(cfg.SearchColumns) > 0 {
		searchParts := make([]string, 0, len(cfg.SearchColumns))
		for _, column := range cfg.SearchColumns {
			searchParts = append(searchParts, column+" LIKE ?")
			args = append(args, "%"+query+"%")
		}
		clauses = append(clauses, "("+strings.Join(searchParts, " OR ")+")")
	}
	if req.ID != nil {
		if column, ok := cfg.SortColumns["id"]; ok {
			clauses = append(clauses, column+" = ?")
			args = append(args, *req.ID)
		}
	}
	if strings.TrimSpace(req.Name) != "" {
		if column, ok := cfg.SortColumns["name"]; ok {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Name))
		}
	}
	if strings.TrimSpace(req.Level) != "" {
		if column, ok := cfg.SortColumns["level"]; ok {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Level))
		}
	}
	if strings.TrimSpace(req.Status) != "" {
		if column, ok := cfg.SortColumns["status"]; ok {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Status))
		}
	}
	if strings.TrimSpace(req.Album) != "" {
		if column, ok := cfg.SortColumns["album_name"]; ok {
			clauses = append(clauses, column+" = ?")
			args = append(args, strings.TrimSpace(req.Album))
		}
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func buildOrder(cfg *PageConfig, orderBy, order string) string {
	column, ok := cfg.SortColumns[orderBy]
	if !ok || column == "" {
		column = cfg.SortColumns[cfg.DefaultOrderBy]
	}
	if column == "" {
		return ""
	}
	return " ORDER BY " + column + " " + normalizeOrder(order)
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

func cloneFilters(filters []*types.PageFilter, req *types.PageListRequest) []*types.PageFilter {
	out := make([]*types.PageFilter, 0, len(filters)+1)
	out = append(out, &types.PageFilter{Key: "q", Label: "关键词", Type: "text", Placeholder: "按当前页面字段搜索", Value: strings.TrimSpace(req.Query)})
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		copied := *filter
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
		case "album":
			copied.Value = req.Album
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

func normalizeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
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
