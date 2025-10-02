// internal/spider/logic/a_FD_parse_details_build_raw.go
package logic

import (
	"bytes"
	"fmt"
	"strings"

	"rudy_gc/internal/types"

	"github.com/PuerkitoBio/goquery"
)

// RawJavMovie 承载从详情页解析出的结构化数据
type RawJavMovie struct {
	JavId         string
	Title         string
	Designation   string
	Prefix        string
	Number        string
	Date          string
	Length        string
	Subscribed    string
	Watched       string
	Owned         string
	Director      *RawItem
	Maker         *RawItem
	Label         *RawItem
	Score         string
	Genres        []*RawItem
	Casts         []*RawItem
	ImgUrl        string
	SmallImgUrl   string
	BirthTime     int64
	LastQueryTime int64
}

type RawItem struct {
	Name  string
	JavId string
}

func (l *CrawlLogic) buildRawMovieByDetail(it *types.Item) (*RawJavMovie, error) {
	// 1. 查询 detail
	detail, err := l.deps.DetailRepo.FindOneByJavId(l.ctx, it.JavId)
	if err != nil {
		return nil, fmt.Errorf("根据 JavId=%s 查询 detail 失败: %w", it.JavId, err)
	}

	// 2. 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(detail.Content)))
	if err != nil {
		return nil, fmt.Errorf("解析详情 HTML 失败: %w", err)
	}

	// 3. 提取字段
	casts := extractItems(doc, "div[id=video_cast] [class=cast]", 14)
	genres := extractItems(doc, "div[id=video_genres] [class=genre]", 15)

	mv := &RawJavMovie{
		Casts:         casts,
		Genres:        genres,
		Title:         parseText(doc, "div[id=video_title] h3"),
		Designation:   parseText(doc, "div[id=video_id] td[class=text]"),
		Date:          parseText(doc, "div[id=video_date] td[class=text]"),
		Length:        parseText(doc, "div[id=video_length] span[class=text]"),
		Score:         parseText(doc, "div[id=video_review] span[class=score]"),
		Subscribed:    parseText(doc, "div[id=video_favorite_edit] span[id=subscribed] a"),
		Watched:       parseText(doc, "div[id=video_favorite_edit] span[id=watched] a"),
		Owned:         parseText(doc, "div[id=video_favorite_edit] span[id=owned] a"),
		JavId:         it.JavId,
		Director:      extractSingleItem(doc, "div[id=video_director]", 18, "nil"),
		Maker:         extractSingleItem(doc, "div[id=video_maker]", 15, "プレステージ"),
		Label:         extractSingleItem(doc, "div[id=video_label]", 15, "nil"),
		ImgUrl:        parseImageURL(doc),
		SmallImgUrl:   it.CoverUrl, // 小图直接用 item 表里的
		BirthTime:     detail.CreatedOn,
		LastQueryTime: detail.LastQueryTime,
	}

	// 4. 拆 prefix-number
	prefix, number := prefixAndNumber(mv.Designation)
	mv.Prefix = prefix
	mv.Number = number

	return mv, nil
}

// ====== 辅助函数（跟老逻辑一样） ======

func extractItems(doc *goquery.Document, selector string, idOffset int) []*RawItem {
	var items []*RawItem
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		itemName := strings.TrimSpace(s.Find("a").Text())
		itemId, _ := s.Find("a").Attr("href")
		if len(itemId) > idOffset {
			items = append(items, &RawItem{
				Name:  itemName,
				JavId: itemId[idOffset:],
			})
		}
	})
	return items
}

func extractSingleItem(doc *goquery.Document, baseSelector string, idOffset int, defaultName string) *RawItem {
	name := strings.TrimSpace(doc.Find(baseSelector + " td[class=text]").Text())
	if name == "" {
		name = defaultName
	}
	href, _ := doc.Find(baseSelector + " a").Attr("href")
	if len(href) > idOffset {
		return &RawItem{Name: name, JavId: href[idOffset:]}
	}
	return &RawItem{Name: defaultName, JavId: "nil"}
}

func parseText(doc *goquery.Document, selector string) string {
	return strings.TrimSpace(doc.Find(selector).Text())
}

func parseImageURL(doc *goquery.Document) string {
	pic, _ := doc.Find("img[id=video_jacket_img]").Attr("src")
	if pic != "" && !strings.HasPrefix(pic, "https:") {
		return "https:" + pic
	}
	return pic
}

func prefixAndNumber(name string) (string, string) {
	if idx := strings.Index(name, "-"); idx != -1 {
		return name[:idx], name[idx+1:]
	}
	return name, ""
}
