// internal/types/d_seed.go
package types

type Seed struct {
	Id                             int64
	Name                           string
	Active                         int64
	SearchType                     int64
	NameType                       int64
	PageNow                        int64
	Offset                         int64
	StartPage                      int64
	EndPage                        int64
	LastQueryTime                  int64
	LastStatus                     int64
	LastError                      string
	MovieTotal                     int64
	MovieLatestReleasingMovieJavId string
	MovieLatestReleasingMovieName  string
	MovieLastAddedTime             int64
	LastInsertCount                int64
	MovieLatestReleasingDate       int64
	CreatedOn                      int64
	UpdatedOn                      int64
}

type SeedMovieStats struct {
	MovieTotal                     int64
	MovieLatestReleasingMovieJavId string
	MovieLatestReleasingMovieName  string
	MovieLatestReleasingDate       int64
}
