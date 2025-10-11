// internal/types/d_seed.go
package types

type Seed struct {
	Id            int64
	Name          string
	Active        int64
	SearchType    int64
	NameType      int64
	PageNow       int64
	Offset        int64
	StartPage     int64
	EndPage       int64
	LastQueryTime int64
	LastStatus    int64
	LastError     string
	CreatedOn     int64
	UpdatedOn     int64
}

const (
	SeedStatusIdle  int64 = 0 // 空闲，尚未执行
	SeedStatusOK    int64 = 1 // 正常，抓取成功
	SeedStatusEmpty int64 = 2 // 空页，提前停止
	SeedStatusError int64 = 3 // 错误
)
