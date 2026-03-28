// internal/types/movie_type.go
package types

type MovieType struct {
	DbId                 int64       `json:"db_id"`
	Name                 string      `json:"name"`
	JavId                string      `json:"jav_id"`
	Title                string      `json:"title"`
	ReleasingDate        string      `json:"releasing_date"`
	UpdateDate           string      `json:"file_creating_date"`
	FilmBirthDate        string      `json:"film_birth_date"`
	Length               int64       `json:"length"`
	Score                float64     `json:"score"`
	ViewersNumberWant    int64       `json:"viewers_number_want"`
	ViewersNumberOwned   int64       `json:"viewers_number_owned"`
	ViewersNumberWatched int64       `json:"viewers_number_watched"`
	Maker                string      `json:"maker"`
	Director             string      `json:"director"`
	Label                string      `json:"label"`
	Genre                []string    `json:"genre"`
	Cast                 []*CastInfo `json:"cast"`
	CastNumber           int64       `json:"cast_number"`
	JavUrl               string      `json:"jav_url"`
	VideoUrl             string      `json:"video_url"`
	SearchUrl            string      `json:"search_url"`
	BusUrl               string      `json:"bus_url"`
	JacketImg            string      `json:"jacket_img"`
	Owned                int64       `json:"owned"`
	NeedDownload         int64       `json:"need_download"`
	Prefix               string      `json:"prefix"`
	EncodeName           string      `json:"encode_name"`
	ScTimes              int64       `json:"sc_times"`
	ComeTimes            int64       `json:"come_times"`
	HighestRank          int64       `json:"highest_rank"`
	FirstRankingDay      string      `json:"first_ranking_day"`
	AMovie               *Movie      `json:"a_movie"`
	BmMinfo              *Minfo      `json:"bm_minfo"`
	BmMurl               *Murl       `json:"bm_murl"`
	VFilm                *Film       `json:"v_film"`
}

type CastInfo struct {
	PersonId    int64
	Name        string
	DisplayName string
	NameShow    string
	LastScTime  int64
	Url         string
}
