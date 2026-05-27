package fetchsukebei

import (
	"regexp"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"

	"github.com/PuerkitoBio/goquery"
)

var sukebeiViewIDPattern = regexp.MustCompile(`/view/(\d+)`)

func parseTorrentRows(movieJavID, queryText, html string) ([]*moviex.TSukebeiTorrent, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	rows := make([]*moviex.TSukebeiTorrent, 0)
	doc.Find("table.torrent-list tbody tr").Each(func(_ int, sel *goquery.Selection) {
		tds := sel.Find("td")
		if tds.Length() < 8 {
			return
		}

		nameCell := tds.Eq(1)
		titleAnchor := nameCell.Find(`a[href^="/view/"]`).FilterFunction(func(_ int, a *goquery.Selection) bool {
			href, ok := a.Attr("href")
			if !ok {
				return false
			}
			return !strings.Contains(href, "#comments")
		}).First()
		if titleAnchor.Length() == 0 {
			return
		}
		title := strings.TrimSpace(titleAnchor.Text())
		viewHref, _ := titleAnchor.Attr("href")
		viewID := parseViewID(viewHref)
		if viewID == 0 {
			return
		}

		sizeText := strings.TrimSpace(tds.Eq(3).Text())
		publishText := strings.TrimSpace(tds.Eq(4).Text())
		if dataTs, ok := tds.Eq(4).Attr("data-timestamp"); ok && strings.TrimSpace(dataTs) != "" {
			publishText = dataTs
		}

		row := &moviex.TSukebeiTorrent{
			MovieJavId:   movieJavID,
			QueryText:    queryText,
			TorrentTitle: title,
			ViewId:       viewID,
			InfoHash:     extractInfoHash(sel),
			SizeBytes:    fetchsite.ParseSizeBytes(sizeText),
			PublishTime:  fetchsite.ParseDateTime(publishText),
			Seeders:      fetchsite.ParseInt64(tds.Eq(5).Text()),
			Leechers:     fetchsite.ParseInt64(tds.Eq(6).Text()),
			Completed:    fetchsite.ParseInt64(tds.Eq(7).Text()),
			LastSeenTime: 0,
			CreatedOn:    0,
			UpdatedOn:    0,
		}
		rows = append(rows, row)
	})

	return rows, nil
}

func extractInfoHash(sel *goquery.Selection) string {
	infoHash := ""
	sel.Find("a").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, ok := a.Attr("href")
		if !ok || !strings.HasPrefix(href, "magnet:") {
			return true
		}
		infoHash = fetchsite.ParseInfoHash(href)
		return false
	})
	return infoHash
}

func parseViewID(viewHref string) int64 {
	trimmed := strings.TrimSpace(viewHref)
	if trimmed == "" {
		return 0
	}
	matches := sukebeiViewIDPattern.FindStringSubmatch(trimmed)
	if len(matches) < 2 {
		return 0
	}
	return fetchsite.ParseInt64(matches[1])
}
