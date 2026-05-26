package fetchsehuatang

import (
	"bytes"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"rudy_gc/internal/service/fetchsite"
)

var (
	safeIDPattern               = regexp.MustCompile(`var\s+safeid='([^']+)'`)
	magnetPattern               = regexp.MustCompile(`(?i)magnet:\?xt=urn:btih:[0-9a-z]+(?:&[a-z0-9._~!$&'()*+,;=:@%/-]+)*`)
	javIDPattern                = regexp.MustCompile(`(?i)\b([a-z0-9]{2,12}-\d{2,7})\b`)
	forumHoursAgoPattern        = regexp.MustCompile(`(?i)(\d+)\s*小时[前前]`)
	forumMinutesAgoPattern      = regexp.MustCompile(`(?i)(\d+)\s*分钟[前前]`)
	forumSecondsAgoPattern      = regexp.MustCompile(`(?i)(\d+)\s*秒[前前]`)
	forumYesterdayPattern       = regexp.MustCompile(`(?i)昨天\s*(\d{1,2}):(\d{2})`)
	forumBeforeYesterdayPattern = regexp.MustCompile(`(?i)前天\s*(\d{1,2}):(\d{2})`)
)

type topicLink struct {
	Title      string
	DetailURL  string
	ListPostAt string
}

type parseListTopicsResult struct {
	Topics          []*topicLink
	SkippedTopCount int
}

type parsedMagnet struct {
	InfoHash string
}

func parseSafeID(body []byte) string {
	matches := safeIDPattern.FindSubmatch(body)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(string(matches[1]))
}

func parseListTopics(listHTML []byte, listURL string, keyword string) (*parseListTopicsResult, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(listHTML))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, 64)
	result := &parseListTopicsResult{
		Topics:          make([]*topicLink, 0, 64),
		SkippedTopCount: 0,
	}

	doc.Find("table#threadlisttableid > tbody").Each(func(_ int, tbody *goquery.Selection) {
		if isStickyThreadRow(tbody) {
			result.SkippedTopCount++
			return
		}
		if !isNormalThreadRow(tbody) {
			return
		}
		sel := tbody.Find("a.s.xst").First()
		if sel.Length() == 0 {
			return
		}
		title := strings.TrimSpace(sel.Text())
		if title == "" {
			return
		}
		if keyword != "" && !strings.Contains(title, keyword) {
			return
		}

		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		detailURL := resolveURL(listURL, href)
		if detailURL == "" {
			return
		}
		if _, ok := seen[detailURL]; ok {
			return
		}
		postAt := extractListPostAt(sel)
		seen[detailURL] = struct{}{}
		result.Topics = append(result.Topics, &topicLink{
			Title:      title,
			DetailURL:  detailURL,
			ListPostAt: postAt,
		})
	})

	return result, nil
}

func isNormalThreadRow(tbody *goquery.Selection) bool {
	if tbody == nil || tbody.Length() == 0 {
		return false
	}
	rowID := strings.TrimSpace(tbody.AttrOr("id", ""))
	return strings.HasPrefix(rowID, "normalthread_")
}

func isStickyThreadRow(tbody *goquery.Selection) bool {
	if tbody == nil || tbody.Length() == 0 {
		return false
	}
	if isNormalThreadRow(tbody) {
		return false
	}
	if tbody.Find(`img[src*="pin_"]`).Length() > 0 {
		return true
	}
	cellText := normalizeForumTimeText(tbody.Find("th.common, th.new").First().Text())
	return strings.Contains(cellText, "置顶")
}

func extractListPostAt(anchor *goquery.Selection) string {
	if anchor == nil {
		return ""
	}
	byCell := anchor.Closest("tbody").Find("td.by").First()
	if byCell.Length() == 0 {
		return ""
	}
	text := strings.TrimSpace(byCell.Find("em span[title]").First().AttrOr("title", ""))
	if text != "" {
		return text
	}
	text = strings.TrimSpace(byCell.Find("em span").First().Text())
	if text != "" {
		return normalizeForumTimeText(text)
	}
	text = strings.TrimSpace(byCell.Find("em a").First().Text())
	return normalizeForumTimeText(text)
}

func parseThreadTitle(detailHTML []byte, fallback string) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(detailHTML))
	if err != nil {
		return strings.TrimSpace(fallback)
	}
	title := strings.TrimSpace(doc.Find("#thread_subject").First().Text())
	if title != "" {
		return title
	}

	pageTitle := strings.TrimSpace(doc.Find("title").First().Text())
	if pageTitle == "" {
		return strings.TrimSpace(fallback)
	}
	pageTitle = strings.Split(pageTitle, " - ")[0]
	pageTitle = strings.TrimSpace(pageTitle)
	if pageTitle == "" {
		return strings.TrimSpace(fallback)
	}
	return pageTitle
}

func parseMovieName(title string) string {
	text := strings.TrimSpace(title)
	if text == "" {
		return ""
	}
	match := javIDPattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(match[1]))
}

func parseThreadTag(title string) string {
	text := strings.ToUpper(strings.TrimSpace(title))
	if text == "" {
		return ""
	}
	if strings.Contains(text, "FC2PPV") {
		return "FC2PPV"
	}
	if strings.Contains(text, "[自译征用]") || strings.Contains(text, "[自提征用]") {
		return "自提征用"
	}
	return ""
}

func parsePostTime(detailHTML []byte, fallback string, now time.Time) int64 {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(detailHTML))
	if err != nil {
		return parseForumTimeText(fallback, now)
	}

	firstPost := doc.Find(`div[id^="post_"]`).First()
	if firstPost.Length() == 0 {
		return parseForumTimeText(fallback, now)
	}

	auth := firstPost.Find(".pti .authi").First()
	if auth.Length() == 0 {
		auth = firstPost.Find(".authi").First()
	}

	titleText := strings.TrimSpace(auth.Find("em span[title]").First().AttrOr("title", ""))
	if titleText != "" {
		if ts := fetchsite.ParseDateTime(titleText); ts > 0 {
			return ts
		}
	}

	timeText := strings.TrimSpace(auth.Find("em span").First().Text())
	if timeText == "" {
		timeText = strings.TrimSpace(auth.Find("em").First().Text())
	}
	if ts := parseForumTimeText(timeText, now); ts > 0 {
		return ts
	}

	if ts := parseForumTimeText(fallback, now); ts > 0 {
		return ts
	}
	return 0
}

func parsePostDate(postTime int64, now time.Time) int64 {
	if postTime <= 0 {
		postTime = now.Unix()
	}
	t := time.Unix(postTime, 0).In(time.Local)
	day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return day.Unix()
}

func parseForumTimeText(raw string, now time.Time) int64 {
	text := normalizeForumTimeText(raw)
	if text == "" {
		return 0
	}

	if ts := fetchsite.ParseDateTime(text); ts > 0 {
		return ts
	}
	if text == "半小时前" {
		return now.Add(-30 * time.Minute).Unix()
	}
	if matches := forumHoursAgoPattern.FindStringSubmatch(text); len(matches) > 1 {
		hours := fetchsite.ParseInt64(matches[1])
		if hours <= 0 {
			return now.Unix()
		}
		return now.Add(-time.Duration(hours) * time.Hour).Unix()
	}
	if matches := forumMinutesAgoPattern.FindStringSubmatch(text); len(matches) > 1 {
		minutes := fetchsite.ParseInt64(matches[1])
		if minutes <= 0 {
			return now.Unix()
		}
		return now.Add(-time.Duration(minutes) * time.Minute).Unix()
	}
	if matches := forumSecondsAgoPattern.FindStringSubmatch(text); len(matches) > 1 {
		seconds := fetchsite.ParseInt64(matches[1])
		if seconds <= 0 {
			return now.Unix()
		}
		return now.Add(-time.Duration(seconds) * time.Second).Unix()
	}
	if matches := forumYesterdayPattern.FindStringSubmatch(text); len(matches) > 2 {
		hour := int(fetchsite.ParseInt64(matches[1]))
		minute := int(fetchsite.ParseInt64(matches[2]))
		t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local).AddDate(0, 0, -1)
		return t.Unix()
	}
	if matches := forumBeforeYesterdayPattern.FindStringSubmatch(text); len(matches) > 2 {
		hour := int(fetchsite.ParseInt64(matches[1]))
		minute := int(fetchsite.ParseInt64(matches[2]))
		t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.Local).AddDate(0, 0, -2)
		return t.Unix()
	}
	return 0
}

func normalizeForumTimeText(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = strings.ReplaceAll(text, "发表于", "")
	return strings.TrimSpace(text)
}

func parseMagnets(detailHTML []byte) []parsedMagnet {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(detailHTML))
	if err != nil {
		return parseMagnetsFromText(string(detailHTML))
	}

	content := strings.TrimSpace(doc.Find(`td[id^="postmessage_"]`).First().Text())
	if content == "" {
		content = strings.TrimSpace(doc.Text())
	}
	return parseMagnetsFromText(content)
}

func parseMagnetsFromText(text string) []parsedMagnet {
	matches := magnetPattern.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	out := make([]parsedMagnet, 0, len(matches))
	for _, magnet := range matches {
		magnet = strings.TrimSpace(magnet)
		if magnet == "" {
			continue
		}
		infoHash := fetchsite.ParseInfoHash(magnet)
		if infoHash == "" {
			continue
		}
		if _, ok := seen[infoHash]; ok {
			continue
		}
		seen[infoHash] = struct{}{}
		out = append(out, parsedMagnet{InfoHash: infoHash})
	}
	return out
}

func buildListPageURL(baseListURL string, pageNo int64) string {
	baseListURL = strings.TrimSpace(baseListURL)
	if baseListURL == "" || pageNo <= 0 {
		return ""
	}

	parsed, err := url.Parse(baseListURL)
	if err != nil {
		return ""
	}

	if parsed.RawQuery != "" {
		values := parsed.Query()
		values.Set("page", strconv.FormatInt(pageNo, 10))
		parsed.RawQuery = values.Encode()
		return parsed.String()
	}

	pathPattern := regexp.MustCompile(`^(.*-)(\d+)(\.html)$`)
	matches := pathPattern.FindStringSubmatch(parsed.Path)
	if len(matches) == 4 {
		parsed.Path = matches[1] + strconv.FormatInt(pageNo, 10) + matches[3]
		return parsed.String()
	}
	return ""
}

func resolveURL(baseURL, href string) string {
	baseURL = strings.TrimSpace(baseURL)
	href = strings.TrimSpace(href)
	if baseURL == "" || href == "" {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}
