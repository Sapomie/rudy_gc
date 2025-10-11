// internal/types/glist.go
package types

type GList struct {
	Id         int64
	Name       string
	ScName     string
	MovieJavId string
	IsCome     int64
	CreatedOn  int64
	UpdatedOn  int64
}

type GSc struct {
	Id            int64
	Name          string
	MovieNumber   int64
	ScTime        int64
	ComeMovieName string
	Cooldown      int64
	Duration      int64
	Fg            string
	Vessel        string
	MovieCast     string
	CreatedOn     int64
	UpdatedOn     int64
}
