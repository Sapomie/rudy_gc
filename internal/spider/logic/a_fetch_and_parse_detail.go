// internal/spiderx/logic/a_fetch_and_parse_detail.go
package logic

func (l *CrawlLogic) FetchAndParseDetails() (int, error) {
	detailNum, err := l.FetchDetails()
	if err != nil {
		return 0, err
	}

	//err = l.ParseDetails()
	//if err != nil {
	//	return 0, err
	//}
	return detailNum, nil
}
