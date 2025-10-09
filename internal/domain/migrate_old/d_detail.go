package migrate

import (
	"bytes"
	"context"
	"fmt"
	"rudy_gc/internal/types"
	"rudy_gc/oldmodel/modelx"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func (s *Service) MigrateDetail() error {
	ctx := context.Background()

	allJavIds, err := s.xModel.DetailModel.AllJavIds(ctx)
	if err != nil {
		return err
	}
	javIdsNew, err := s.deps.DetailRepo.AllJavIds(ctx)
	if err != nil {
		return err
	}

	// 1) 去重 + 差集：从 allJavIds 中剔除已在新库中的 javId
	newSet := make(map[string]struct{}, len(javIdsNew))
	for _, id := range javIdsNew {
		if id == "" {
			continue
		}
		newSet[id] = struct{}{}
	}

	seen := make(map[string]struct{}, len(allJavIds))
	javIdsNeedUpsert := make([]string, 0, len(allJavIds))
	for _, id := range allJavIds {
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		if _, exists := newSet[id]; !exists {
			javIdsNeedUpsert = append(javIdsNeedUpsert, id)
		}
	}
	var count int
	for _, javId := range javIdsNeedUpsert {
		xDetail, err := s.xModel.DetailModel.FindOneByJavId(ctx, javId)
		if err != nil {
			return err
		}
		content, err := filterDetailContent([]byte(xDetail.Content))
		if err != nil {
			return err
		}

		now := time.Now().Unix()
		detail := types.Detail{
			Name:          xDetail.Name,
			JavId:         xDetail.JavId,
			Prefix:        xDetail.Prefix,
			QueryUrl:      xDetail.QueryUrl,
			Content:       content,
			LastQueryTime: xDetail.LastQueryTime,
			CreatedOn:     now,
			UpdatedOn:     now,
		}

		err = s.deps.DetailRepo.Upsert(ctx, &detail)
		if err != nil {
			return err
		}

		item := &types.Item{
			Name:             xDetail.Name,
			JavId:            xDetail.JavId,
			Prefix:           xDetail.Prefix,
			SearchType:       types.ItemSearchByOld,
			CoverUrl:         "migrate",
			SearchBy:         "migrate",
			HasDetail:        types.ItemDetailOK,
			HasDownloadCover: types.ItemCoverNone,
			HasChinese:       types.ItemChineseNone,
			DetailNeedScan:   modelx.DetailStatusNeedScan,
			DetailBirthTime:  now,
			DetailUpdateTime: now,
			CreatedOn:        now,
			UpdatedOn:        now,
		}

		_, err = s.deps.ItemRepo.TryInsert(ctx, item)
		if err != nil {
			return err
		}

		count++
		s.deps.Log.Infof("完成%v/%v", count, len(javIdsNeedUpsert))

	}

	return nil
}

func filterDetailContent(body []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("解析 HTML 出错: %w", err)
	}

	// 新建一个干净的 DOM 作为输出
	out := &strings.Builder{}
	out.WriteString("<html><body>\n")

	// 只挑选需要的节点复制出来
	keepIDs := []string{
		"video_title",
		"video_id",
		"video_date",
		"video_length",
		"video_director",
		"video_maker",
		"video_label",
		"video_review",
		"video_genres",
		"video_cast",
		"video_favorite_edit",
		"video_jacket",
	}

	for _, id := range keepIDs {
		if sel := doc.Find(fmt.Sprintf("#%s", id)); sel.Length() > 0 {
			html, _ := sel.Html()
			out.WriteString(fmt.Sprintf("<div id=\"%s\">%s</div>\n", id, html))
		}
	}

	out.WriteString("</body></html>")
	return out.String(), nil
}
