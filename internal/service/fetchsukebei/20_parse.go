package fetchsukebei

import (
	"regexp"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"

	"github.com/PuerkitoBio/goquery"
)

var sukebeiViewIDPattern = regexp.MustCompile(`/view/(\d+)`)

func parseTorrentRows(movieJavID, queryText, searchURL, html string) ([]*moviex.TSukebeiTorrent, error) {
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
		titleAnchor := nameCell.Find("a").First()
		title := strings.TrimSpace(titleAnchor.Text())
		viewHref, _ := titleAnchor.Attr("href")
		viewID := parseViewID(viewHref)
		if viewID == 0 {
			return
		}

		torrentURL := ""
		magnetURL := ""
		sel.Find("a").Each(func(_ int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if !ok {
				return
			}
			if strings.HasPrefix(href, "magnet:") {
				magnetURL = href
				return
			}
			if strings.HasSuffix(href, ".torrent") || strings.Contains(href, "/download/") {
				torrentURL = href
			}
		})

		sizeText := strings.TrimSpace(tds.Eq(3).Text())
		publishText := strings.TrimSpace(tds.Eq(4).Text())
		if dataTs, ok := tds.Eq(4).Attr("data-timestamp"); ok && strings.TrimSpace(dataTs) != "" {
			publishText = dataTs
		}

		row := &moviex.TSukebeiTorrent{
			MovieJavId:   movieJavID,
			QueryText:    queryText,
			SearchUrl:    searchURL,
			TorrentTitle: title,
			ViewId:       viewID,
			ViewUrl:      absoluteSukebeiURL(viewHref),
			TorrentUrl:   absoluteSukebeiURL(torrentURL),
			MagnetUrl:    magnetURL,
			InfoHash:     fetchsite.ParseInfoHash(magnetURL),
			Dn:           fetchsite.ParseDN(magnetURL),
			SizeBytes:    fetchsite.ParseSizeBytes(sizeText),
			SizeText:     sizeText,
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

func absoluteSukebeiURL(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	if strings.HasPrefix(text, "http://") || strings.HasPrefix(text, "https://") || strings.HasPrefix(text, "magnet:") {
		return text
	}
	return "https://sukebei.nyaa.si" + text
}
