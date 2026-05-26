package types

type PageListRequest struct {
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	OrderBy  string `form:"od"`
	Order    string `form:"order"`
	Query    string `form:"q"`
	ID       *int64 `form:"id"`
	Name     string `form:"name"`
	Level    string `form:"level"`
	Status   string `form:"status"`
	Album    string `form:"album"`
}

type PageListResponse struct {
	Key         string        `json:"key"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	LegacyPath  string        `json:"legacy_path"`
	Kind        string        `json:"kind"`
	Page        int64         `json:"page"`
	PageSize    int64         `json:"page_size"`
	Total       int64         `json:"total"`
	OrderBy     string        `json:"order_by"`
	Order       string        `json:"order"`
	Query       string        `json:"query"`
	Columns     []*PageColumn `json:"columns"`
	Rows        []PageRow     `json:"rows"`
	Filters     []*PageFilter `json:"filters"`
	Actions     []*PageAction `json:"actions"`
	Links       []*PageLink   `json:"links"`
	Stats       []*PageStat   `json:"stats"`
}

type PageSummary struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	LegacyPath  string `json:"legacy_path"`
	Kind        string `json:"kind"`
	Group       string `json:"group"`
}

type PageColumn struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Sortable bool   `json:"sortable"`
	Link     string `json:"link,omitempty"`
}

type PageFilter struct {
	Key         string              `json:"key"`
	Label       string              `json:"label"`
	Type        string              `json:"type"`
	Placeholder string              `json:"placeholder,omitempty"`
	Value       string              `json:"value,omitempty"`
	Options     []*PageFilterOption `json:"options,omitempty"`
}

type PageFilterOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type PageAction struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Variant  string `json:"variant"`
	Disabled bool   `json:"disabled"`
	Reason   string `json:"reason,omitempty"`
}

type PageLink struct {
	Label string `json:"label"`
	Href  string `json:"href"`
	Kind  string `json:"kind"`
}

type PageStat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note,omitempty"`
}

type PageRow map[string]string
