package dailybestcollector

import (
	"fmt"
	"strings"

	contracts "rudy_gc/shared/dailybest"

	"github.com/PuerkitoBio/goquery"
)

const markBlueRay = "（ブルーレイディスク）"

func ParsePages(pages []contracts.DailyBestSnapshotPage) ([]contracts.DailyBestSnapshotItem, error) {
	items := make([]contracts.DailyBestSnapshotItem, 0)
	seenJavIDs := make(map[string]struct{})
	seenRanks := make(map[int64]struct{})
	for _, page := range pages {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page.Content))
		if err != nil {
			return nil, fmt.Errorf("解析 DailyBest 第 %d 页失败: %w", page.PageNumber, err)
		}
		rankPos := (page.PageNumber-1)*20 + 1
		accepted := int64(0)
		doc.Find("div.videos div.video").Each(func(index int, selection *goquery.Selection) {
			title := strings.TrimSpace(selection.Find("div.title").Text())
			if strings.Contains(title, markBlueRay) || strings.Contains(title, "蓝光") {
				return
			}
			movieName := strings.TrimSpace(selection.Find("div.id").Text())
			javID, _ := selection.Find("div.toolbar a").Attr("id")
			javID = strings.TrimSpace(javID)
			if movieName == "" || javID == "" {
				return
			}
			coverURL, _ := selection.Find("img").Attr("src")
			items = append(items, contracts.DailyBestSnapshotItem{
				RankPos:     rankPos + accepted,
				SourcePage:  page.PageNumber,
				SourceIndex: int64(index + 1),
				JavID:       javID,
				MovieName:   movieName,
				MovieCode:   movieName,
				Title:       title,
				CoverURL:    strings.TrimSpace(coverURL),
			})
			accepted++
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("DailyBest 快照没有可发布条目")
	}
	for _, item := range items {
		if _, ok := seenJavIDs[item.JavID]; ok {
			return nil, fmt.Errorf("DailyBest 快照存在重复 JavID: %s", item.JavID)
		}
		seenJavIDs[item.JavID] = struct{}{}
		if _, ok := seenRanks[item.RankPos]; ok {
			return nil, fmt.Errorf("DailyBest 快照存在重复排名: %d", item.RankPos)
		}
		seenRanks[item.RankPos] = struct{}{}
	}
	return items, nil
}
