package types

type Media struct {
	Id            int64
	MovieJavId    string
	MovieName     string
	FileName      string
	DirectoryId   int64
	RootDir       string
	FullDir       string
	Alias         string
	Size          int64
	Width         int64
	Height        int64
	BitRate       int64
	Duration      int64
	FrameAverage  float64
	HasSub        int64
	SelfMake      int64
	HasMask       int64
	NeedScanMeta  int64
	IsRemoved     int64
	RemoveTime    int64
	BirthTime     int64
	ReleasingDate int64
	CreatedOn     int64
	UpdatedOn     int64
}
