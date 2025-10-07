package modelg

type Movie struct {
	//base info
	Id                   int64   `gorm:"primaryKey"`
	Name                 string  `gorm:"not null;type:varchar(191);index;index:idx_owned_release_name,priority:3,sort:desc;index:idx_owned_birth_name,priority:3,sort:desc;index:idx_owned_rank_release_name,priority:4,sort:desc;index:idx_owned_viewed_name,priority:3,sort:desc;index:idx_owned_lastsc_name,priority:3,sort:desc;index:idx_owned_sctimes_name,priority:3,sort:desc;index:idx_owned_cometimes_name,priority:3,sort:desc;index:idx_owned_highestrank_name,priority:3,sort:desc"`
	JavId                string  `gorm:"not null;unique"`
	Title                string  `gorm:"not null;type:varchar(300)"`
	ReleasingDate        int64   `gorm:"not null;index;index:idx_owned_release_name,priority:2,sort:desc;index:idx_owned_rank_release_name,priority:3,sort:desc"`
	Length               int64   `gorm:"not null;type:smallint"`
	Score                int64   `gorm:"not null;type:tinyint"`
	ViewersNumberWant    int64   `gorm:"not null;type:MEDIUMINT"`
	ViewersNumberOwned   int64   `gorm:"not null;type:MEDIUMINT"`
	ViewersNumberWatched int64   `gorm:"not null;type:MEDIUMINT;index;index:idx_owned_viewed_name,priority:2,sort:desc"`
	PrefixId             int64   `gorm:"not null"`
	MakerId              int64   `gorm:"not null"`
	LabelId              int64   `gorm:"not null"`
	DirectorId           int64   `gorm:"not null"`
	CastNumber           int64   `gorm:"not null;type:smallint"`
	LastQueryJavTime     int64   `gorm:"not null;index"`
	DetailBirthTime      int64   `gorm:"not null"`
	CastAverageAge       float64 `gorm:"not null;index"`
	EncodeName           string  `gorm:"not null;type:varchar(191);index"`
	//translation
	HasChinese int64  `gorm:"not null;type:tinyint"`
	Chinese    string `gorm:"not null;type:varchar(300)"`
	//rank info
	FirstRankDayNumber int64 `gorm:"not null;index;index:idx_owned_rank_release_name,priority:2,sort:desc"`
	HighestRank        int64 `gorm:"not null;index;index:idx_owned_highestrank_name,priority:2,sort:asc"`
	DaysInRank         int64 `gorm:"not null;index"`
	//to item
	HasDownloadedCover int64 `gorm:"not null;type:tinyint"`
	NeedDownload       int64 `gorm:"not null;type:tinyint;index"`
	//owned
	FilmBirthTime int64 `gorm:"not null;index;index:idx_owned_birth_name,priority:2,sort:desc"`
	MovieOwned    int64 `gorm:"not null;type:tinyint;index:idx_owned_release_name,priority:1;index:idx_owned_birth_name,priority:1;index:idx_owned_rank_release_name,priority:1;index:idx_owned_viewed_name,priority:1;index:idx_owned_lastsc_name,priority:1;index:idx_owned_sctimes_name,priority:1;index:idx_owned_cometimes_name,priority:1;index:idx_owned_highestrank_name,priority:1"`
	ScTimes       int64 `gorm:"not null;type:MEDIUMINT;index;index:idx_owned_sctimes_name,priority:2,sort:desc"`
	ComeTimes     int64 `gorm:"not null;type:MEDIUMINT;index;index:idx_owned_cometimes_name,priority:2,sort:desc"`
	LastScTime    int64 `gorm:"not null;index;index:idx_owned_lastsc_name,priority:2,sort:desc"`

	CreatedOn int64 `gorm:"not null"`
	UpdatedOn int64 `gorm:"not null"`
}

const movieTableName = "a_movie"

func (i *Movie) TableName() string {
	return movieTableName
}
