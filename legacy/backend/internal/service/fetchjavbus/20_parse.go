package fetchjavbus

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/service/fetchsite"

	"github.com/PuerkitoBio/goquery"
)

var (
	gidPattern   = regexp.MustCompile(`gid\s*=\s*(\d+)`)
	imgPattern   = regexp.MustCompile(`img\s*=\s*['"]([^'"]+)['"]`)
	ucPattern    = regexp.MustCompile(`uc\s*=\s*(\d+)`)
	floorPattern = regexp.MustCompile(`floor\s*=\s*(\d+)`)

	errJavbusAgeVerification = errors.New("javbus age verification page")
	errJavbusNotFound        = errors.New("javbus detail page not found")
	errJavbusEmptyAjaxRows   = errors.New("javbus ajax returned no magnet rows")
)

type detailPagePayload struct {
	GID   string
	IMG   string
	UC    string
	Floor string
}

func parseDetailPagePayload(html string) (*detailPagePayload, error) {
	if err := detectDetailPageError(html); err != nil {
		return nil, err
	}

	payload := &detailPagePayload{
		GID:   findMatch(gidPattern, html),
		IMG:   findMatch(imgPattern, html),
		UC:    findMatch(ucPattern, html),
		Floor: findMatch(floorPattern, html),
	}
	if payload.GID == "" {
		return nil, fmt.Errorf("javbus detail payload missing gid")
	}
	if payload.IMG == "" {
		return nil, fmt.Errorf("javbus detail payload missing img")
	}
	if payload.UC == "" {
		payload.UC = "0"
	}
	if payload.Floor == "" {
		payload.Floor = "1"
	}
	return payload, nil
}

func detectDetailPageError(html string) error {
	normalized := strings.ToLower(html)
	switch {
	case strings.Contains(normalized, "<title>age verification javbus - javbus</title>"),
		strings.Contains(normalized, "/doc/driver-verify"):
		return errJavbusAgeVerification
	case strings.Contains(normalized, "<title>404 page not found! - javbus</title>"):
		return errJavbusNotFound
	default:
		return nil
	}
}

func (p *detailPagePayload) buildAjaxURL(siteSvc *fetchsite.Service) (string, error) {
	baseURL, err := siteSvc.BuildURL(fetchsite.FetchSiteCodeJavbus, "ajax", "uncledatoolsbyajax.php")
	if err != nil {
		return "", err
	}
	values := url.Values{}
	values.Set("gid", p.GID)
	values.Set("lang", "zh")
	values.Set("img", p.IMG)
	values.Set("uc", p.UC)
	values.Set("floor", p.Floor)
	return baseURL + "?" + values.Encode(), nil
}

func parseMagnetRows(movieJavID, html string) ([]*moviex.TJavbusMagnet, error) {
	wrappedHTML := html
	if !strings.Contains(strings.ToLower(html), "<html") {
		wrappedHTML = "<table><tbody>" + html + "</tbody></table>"
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrappedHTML))
	if err != nil {
		return nil, err
	}

	rows := make([]*moviex.TJavbusMagnet, 0)
	doc.Find("tr").Each(func(i int, sel *goquery.Selection) {
		magnetURL, ok := sel.Find(`a[href^="magnet:"]`).Attr("href")
		if !ok {
			return
		}

		tds := sel.Find("td")
		if tds.Length() < 3 {
			return
		}

		name := cleanMagnetName(tds.First())
		if name == "" {
			name = strings.TrimSpace(sel.Find(`a[href^="magnet:"]`).First().Text())
		}
		sizeText := strings.TrimSpace(tds.Eq(1).Text())
		shareText := strings.TrimSpace(tds.Eq(2).Text())
		infoHash := fetchsite.ParseInfoHash(magnetURL)
		if infoHash == "" {
			return
		}

		row := &moviex.TJavbusMagnet{
			MovieJavId:   movieJavID,
			MagnetName:   name,
			InfoHash:     infoHash,
			SizeBytes:    fetchsite.ParseSizeBytes(sizeText),
			ShareDate:    fetchsite.ParseDateTime(shareText),
			HasHd:        boolToInt64(strings.Contains(sel.Text(), "高清")),
			HasSubtitle:  boolToInt64(strings.Contains(sel.Text(), "字幕")),
			RowSort:      int64(i + 1),
			LastSeenTime: 0,
			CreatedOn:    0,
			UpdatedOn:    0,
		}
		rows = append(rows, row)
	})

	if len(rows) == 0 {
		return nil, errJavbusEmptyAjaxRows
	}
	return rows, nil
}

func cleanMagnetName(cell *goquery.Selection) string {
	if cell == nil {
		return ""
	}
	clone := cell.Clone()
	clone.Find("a.btn").Remove()
	return strings.TrimSpace(clone.Text())
}

func findMatch(pattern *regexp.Regexp, text string) string {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 2
}
