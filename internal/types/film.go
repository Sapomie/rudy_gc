// internal/types/film.go
package types

// Film 与 internal/model/modelx/moviex.VFilm 字段一一对应（用于 Domain/Repo 传参与返回）
type Film struct {
	Id          int64
	MovieJavId  string
	MovieName   string
	FileName    string
	DirectoryId int64
	RootDir     string
	FullDir     string

	// 新增的多层目录 ID（不足层为 0）
	Dir1Id int64
	Dir2Id int64
	Dir3Id int64
	Dir4Id int64

	Alias        string
	Size         int64
	Width        int64
	Height       int64
	BitRate      int64
	Duration     int64
	FrameAverage float64

	HasSub   int64
	SelfMake int64
	HasMask  int64

	NeedScanMeta  int64
	IsRemoved     int64
	RemoveTime    int64
	ScTimes       int64
	ComeTimes     int64
	LastScTime    int64
	BirthTime     int64
	ReleasingDate int64
	CreatedOn     int64
	UpdatedOn     int64
}
