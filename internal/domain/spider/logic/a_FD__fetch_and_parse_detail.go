// internal/spiderx/logic/a_FD__fetch_and_parse_detail.go
package logic

func (l *CrawlLogic) FetchAndParseDetails() (int, error) {
	detailNum, err := l.FetchDetailsByItemDetailStatus()
	if err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("FetchDetailsByItemDetailStatus: %v", err)
		return 0, err
	}

	err = l.ParseDetails()
	if err != nil {
		l.deps.Log.WithContext(l.ctx).Errorf("ParseDetails: %v", err)
		return 0, err
	}
	return detailNum, nil
}
