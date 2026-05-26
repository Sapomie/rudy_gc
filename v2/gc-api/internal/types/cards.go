package types

type CardsListRequest struct {
	View string `form:"view"`

	CastNames    string `form:"cn"`
	PersonIds    string `form:"pid"`
	GenreNames   string `form:"gn"`
	DirectorName string `form:"dn"`
	PrefixName   string `form:"pn"`
	MakerName    string `form:"mn"`
	LabelName    string `form:"ln"`
	LabelJavID   string `form:"lj"`
	AlbumName    string `form:"an"`

	ReleasingDateStart  string `form:"rs"`
	ReleasingDateEnd    string `form:"re"`
	MediaBirthTimeStart string `form:"mbs"`
	MediaBirthTimeEnd   string `form:"mbe"`

	CastAgeMin *float64 `form:"cay"`
	CastAgeMax *float64 `form:"cao"`

	StartRankingDateStart string `form:"srds"`
	StartRankingDateEnd   string `form:"srde"`

	DaysInRankMin *int64 `form:"drkmin"`
	NeedDownload  int64  `form:"nd"`
	Word          string `form:"wd"`
	MediaOwned    int64  `form:"mowned"`

	ViewWatchedMin *int64   `form:"vwmin"`
	ViewWatchedMax *int64   `form:"vwmax"`
	ScoreMin       *float64 `form:"smin"`
	ScoreMax       *float64 `form:"smax"`

	LastScTimeMin string `form:"lsctmin"`
	LastScTimeMax string `form:"lsctmax"`
	ScTimesMin    *int64 `form:"scmin"`
	ScTimesMax    *int64 `form:"scmax"`
	ComeTimesMin  *int64 `form:"comin"`
	ComeTimesMax  *int64 `form:"comax"`

	MediaDir1 string `form:"md1"`
	MediaDir2 string `form:"md2"`
	MediaDir3 string `form:"md3"`
	MediaDir4 string `form:"md4"`

	OrderBy  string `form:"od"`
	Order    string `form:"order"`
	Page     int64  `form:"p"`
	PageSize int64  `form:"ps"`
	RandomN  int64  `form:"n"`
}

type CardsListResponse struct {
	View            string        `json:"view"`
	Title           string        `json:"title"`
	Items           []*MovieCard  `json:"items"`
	Total           int64         `json:"total"`
	Page            int64         `json:"page"`
	PageSize        int64         `json:"page_size"`
	OrderBy         string        `json:"order_by"`
	Order           string        `json:"order"`
	Views           []*ViewOption `json:"views"`
	FilterSnapshot  *CardsFilter  `json:"filter_snapshot"`
	RandomRequested int64         `json:"random_requested,omitempty"`
}

type ViewOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Href  string `json:"href"`
}

type CardsFilter struct {
	CastNames             string   `json:"cast_names"`
	PersonIds             string   `json:"person_ids"`
	GenreNames            string   `json:"genre_names"`
	DirectorName          string   `json:"director_name"`
	PrefixName            string   `json:"prefix_name"`
	MakerName             string   `json:"maker_name"`
	LabelName             string   `json:"label_name"`
	LabelJavID            string   `json:"label_jav_id"`
	AlbumName             string   `json:"album_name"`
	ReleasingDateStart    string   `json:"releasing_date_start"`
	ReleasingDateEnd      string   `json:"releasing_date_end"`
	MediaBirthTimeStart   string   `json:"media_birth_time_start"`
	MediaBirthTimeEnd     string   `json:"media_birth_time_end"`
	CastAgeMin            *float64 `json:"cast_age_min,omitempty"`
	CastAgeMax            *float64 `json:"cast_age_max,omitempty"`
	StartRankingDateStart string   `json:"start_ranking_date_start"`
	StartRankingDateEnd   string   `json:"start_ranking_date_end"`
	DaysInRankMin         *int64   `json:"days_in_rank_min,omitempty"`
	NeedDownload          int64    `json:"need_download"`
	Word                  string   `json:"word"`
	MediaOwned            int64    `json:"media_owned"`
	ViewWatchedMin        *int64   `json:"view_watched_min,omitempty"`
	ViewWatchedMax        *int64   `json:"view_watched_max,omitempty"`
	ScoreMin              *float64 `json:"score_min,omitempty"`
	ScoreMax              *float64 `json:"score_max,omitempty"`
	LastScTimeMin         string   `json:"last_sc_time_min"`
	LastScTimeMax         string   `json:"last_sc_time_max"`
	ScTimesMin            *int64   `json:"sc_times_min,omitempty"`
	ScTimesMax            *int64   `json:"sc_times_max,omitempty"`
	ComeTimesMin          *int64   `json:"come_times_min,omitempty"`
	ComeTimesMax          *int64   `json:"come_times_max,omitempty"`
	MediaDir1             string   `json:"media_dir_1"`
	MediaDir2             string   `json:"media_dir_2"`
	MediaDir3             string   `json:"media_dir_3"`
	MediaDir4             string   `json:"media_dir_4"`
}

type MovieCard struct {
	MovieName            string     `json:"movie_name"`
	MovieJavID           string     `json:"movie_jav_id"`
	MovieHref            string     `json:"movie_href"`
	Title                string     `json:"title"`
	JacketImg            string     `json:"jacket_img"`
	ComeTimes            int64      `json:"come_times"`
	ScTimes              int64      `json:"sc_times"`
	HighestRank          int64      `json:"highest_rank"`
	Score                float64    `json:"score"`
	ViewersNumberWatched int64      `json:"viewers_number_watched"`
	OwnedWMedia          int64      `json:"owned_w_media"`
	VideoUrlWMedia       string     `json:"video_url_w_media"`
	FilmBirthDateWMedia  string     `json:"film_birth_date_w_media"`
	Genre                []string   `json:"genre"`
	Cast                 []*CastTag `json:"cast"`
	Director             string     `json:"director"`
	DirectorHref         string     `json:"director_href"`
	Prefix               string     `json:"prefix"`
	PrefixHref           string     `json:"prefix_href"`
	BusUrl               string     `json:"bus_url"`
	SearchUrl            string     `json:"search_url"`
	JavUrl               string     `json:"jav_url"`
	ReleasingDate        string     `json:"releasing_date"`
	FirstRankingDay      string     `json:"first_ranking_day"`
	Label                string     `json:"label"`
	LabelHref            string     `json:"label_href"`
	Maker                string     `json:"maker"`
	MakerHref            string     `json:"maker_href"`
	NeedDownload         int64      `json:"need_download"`
}

type CastTag struct {
	PersonID    int64  `json:"person_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	NameShow    string `json:"name_show"`
	Href        string `json:"href"`
}
