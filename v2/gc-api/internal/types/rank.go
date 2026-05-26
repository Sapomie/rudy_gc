package types

type RankDayResponse struct {
	Title        string       `json:"title"`
	RankDate     string       `json:"rank_date"`
	Items        []*MovieCard `json:"items"`
	Total        int64        `json:"total"`
	Page         int64        `json:"page"`
	PageSize     int64        `json:"page_size"`
	PrevDate     string       `json:"prev_date"`
	NextDate     string       `json:"next_date"`
	RandomDate   string       `json:"random_date"`
	PrevDisabled bool         `json:"prev_disabled"`
	NextDisabled bool         `json:"next_disabled"`
}

type RankPeriodResponse struct {
	Title           string            `json:"title"`
	PeriodKey       string            `json:"period_key"`
	PeriodType      string            `json:"period_type"`
	PeriodTypeLabel string            `json:"period_type_label"`
	Category        int64             `json:"category"`
	CategoryLabel   string            `json:"category_label"`
	RangeStart      string            `json:"range_start"`
	RangeEnd        string            `json:"range_end"`
	Items           []*RankPeriodCard `json:"items"`
	Total           int64             `json:"total"`
	Page            int64             `json:"page"`
	PageSize        int64             `json:"page_size"`
	PrevKey         string            `json:"prev_key"`
	NextKey         string            `json:"next_key"`
	PrevDisabled    bool              `json:"prev_disabled"`
	NextDisabled    bool              `json:"next_disabled"`
	TypeLinks       []*RankSwitchLink `json:"type_links"`
	CategoryLinks   []*RankSwitchLink `json:"category_links"`
	LatestHref      string            `json:"latest_href"`
}

type RankPeriodCard struct {
	Movie           *MovieCard `json:"movie"`
	RankPos         int64      `json:"rank_pos"`
	Score           float64    `json:"score"`
	DaysInRank      int64      `json:"days_in_rank"`
	UsedPickDays    int64      `json:"used_pick_days"`
	BestRank        int64      `json:"best_rank"`
	WorstPickedRank int64      `json:"worst_picked_rank"`
	PrevRank        int64      `json:"prev_rank"`
	PrevRankText    string     `json:"prev_rank_text"`
	RankChange      int64      `json:"rank_change"`
	RankChangeText  string     `json:"rank_change_text"`
	RankChangeClass string     `json:"rank_change_class"`
	PeakRank        int64      `json:"peak_rank"`
}

type RankSwitchLink struct {
	Label  string `json:"label"`
	Href   string `json:"href"`
	Active bool   `json:"active"`
}
