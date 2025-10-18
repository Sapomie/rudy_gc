package logic

import (
	"context"
	"fmt"
	consts "rudy_gc/internal/consts"
	"rudy_gc/internal/types"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

func (l *CrawlLogic) makeAndInsertItems(ctx context.Context, content, searchBy string, category int64) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("parse document: %w", err)
	}

	searchType := mapInventoryCategoryToItemSearchType(category)
	var toInsert []*types.Item

	// 兼容旧站点结构：div.videos div.video
	doc.Find("div.videos div.video").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("div.title").Text())
		if isBluRayTitle(title) {
			return
		}

		name := strings.TrimSpace(s.Find("div.id").Text())
		if name == "" {
			return
		}

		javID, _ := s.Find("div.toolbar a").Attr("id")
		javID = strings.TrimSpace(javID)
		if javID == "" {
			return
		}

		coverURL, _ := s.Find("img").Attr("src")
		coverURL = strings.TrimSpace(coverURL)

		prefix, _ := splitPrefixAndNumber(name)
		if isRedundantPrefix(prefix) {
			l.deps.Log.WithFields(logrus.Fields{
				"name":   name,
				"prefix": prefix,
			}).Info("skip redundant prefix")
			return
		}

		now := time.Now().Unix()
		it := &types.Item{
			Name:                name,
			JavId:               javID,
			Prefix:              prefix,
			SearchType:          searchType,
			CoverUrl:            coverURL,
			SearchBy:            searchBy,
			CreatedOn:           now,
			UpdatedOn:           now,
			HasDetail:           consts.ItemDetailNone,
			HasDownloadCover:    consts.ItemCoverNone,
			HasChinese:          consts.ItemChineseNone,
			DetailBirthTime:     0,
			LastQueryDetailTime: 0,
			DetailNeedScan:      consts.ItemDetailStatusNeedScan,
		}
		toInsert = append(toInsert, it)
	})

	// 幂等插入
	for _, it := range toInsert {
		inserted, err := l.deps.ItemRepo.TryInsert(ctx, it)
		if err != nil {
			return fmt.Errorf("insert item name=%s javId=%s: %w", it.Name, it.JavId, err)
		}
		if inserted {
			l.deps.Log.WithFields(logrus.Fields{
				"name":  it.Name,
				"javId": it.JavId,
			}).Info("item inserted")
		}
	}

	return nil
}

// ---------- 辅助函数 ----------

// mapInventoryCategoryToItemSearchType 将 inventory 的类别映射到 item 的 SearchType
// 与旧项目行为一致：Prefix→Prefix，Label→Label；未知保持原值
func mapInventoryCategoryToItemSearchType(cat int64) int64 {
	switch cat {
	case consts.InventoryCategoryByPrefix:
		return consts.ItemSearchByPrefix
	case consts.InventoryCategoryByLabel:
		return consts.ItemSearchByLabel
	default:
		return cat
	}
}

// splitPrefixAndNumber 从 "ABC-123" 拆出 "ABC" 和 "123"
func splitPrefixAndNumber(name string) (prefix string, number string) {
	if idx := strings.Index(name, "-"); idx != -1 {
		return name[:idx], name[idx+1:]
	}
	return name, ""
}

// isBluRayTitle 过滤蓝光等不需要的条目（尽量保守，不绑死具体常量）
func isBluRayTitle(title string) bool {
	t := strings.ToLower(title)
	return strings.Contains(t, consts.MarkBlueRay) || strings.Contains(title, "蓝光")
}

// isRedundantPrefix 冗余前缀过滤（可接入你的 cons.RedundantPrefixSet）
// 这里先留个可替换点，默认不过滤
func isRedundantPrefix(prefix string) bool {
	_, ok := consts.RedundantPrefixSet[prefix]
	return ok
}
