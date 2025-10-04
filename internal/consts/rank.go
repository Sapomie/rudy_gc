package consts

import (
	"time"
)

var BestGenesisDay = time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)

func GetRankDayNumber(date string) int64 {
	t1, _ := time.ParseInLocation(time.DateOnly, date, time.Local)
	d := t1.Sub(BestGenesisDay).Hours() / 24.0
	return int64(d + 1)
}
