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
