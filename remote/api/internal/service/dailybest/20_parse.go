package dailybest

import (
	"fmt"
	"net/url"
	"strings"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"

	"github.com/PuerkitoBio/goquery"
)

const markBlueRay = "（ブルーレイディスク）"

func ParsePages(pages []dailybesttypes.SourcePage) ([]dailybesttypes.ParsedRank, error) {
	items := make([]dailybesttypes.ParsedRank, 0)
	seenJavIDs := make(map[string]struct{})
	seenRanks := make(map[int64]struct{})
	for _, page := range pages {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(page.Content))
		if err != nil {
			return nil, fmt.Errorf("解析 DailyBest 第 %d 页失败: %w", page.PageNumber, err)
		}
		rankPos := (page.PageNumber-1)*20 + 1
		accepted := int64(0)
		doc.Find("div.videos div.video").Each(func(_ int, selection *goquery.Selection) {
			title := strings.TrimSpace(selection.Find("div.title").Text())
			if strings.Contains(title, markBlueRay) || strings.Contains(title, "蓝光") {
				return
			}
			javID, _ := selection.Find("div.toolbar a").Attr("id")
			javID = strings.TrimSpace(javID)
			if javID == "" {
				return
			}
			movieName := strings.TrimSpace(selection.Find("div.id").Text())
			coverURL, _ := selection.Find("img").Attr("src")
			items = append(items, dailybesttypes.ParsedRank{
				Rank: contracts.DailyBestRank{
					RankPos:    rankPos + accepted,
					MovieJavID: javID,
				},
				Movie: dailybesttypes.MovieDraft{
					MovieJavID: javID,
					MovieName:  movieName,
					MovieCode:  movieName,
					Title:      title,
					CoverURL:   resolveSourceURL(page.SourceURL, coverURL),
				},
			})
			accepted++
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("DailyBest 排名没有可发布条目")
	}
	for _, item := range items {
		if _, ok := seenJavIDs[item.Rank.MovieJavID]; ok {
			return nil, fmt.Errorf("DailyBest 排名存在重复 JavID: %s", item.Rank.MovieJavID)
		}
		seenJavIDs[item.Rank.MovieJavID] = struct{}{}
		if _, ok := seenRanks[item.Rank.RankPos]; ok {
			return nil, fmt.Errorf("DailyBest 排名存在重复名次: %d", item.Rank.RankPos)
		}
		seenRanks[item.Rank.RankPos] = struct{}{}
	}
	return items, nil
}

func resolveSourceURL(sourceURL, rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	reference, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	base, err := url.Parse(strings.TrimSpace(sourceURL))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return rawURL
	}
	return base.ResolveReference(reference).String()
}
