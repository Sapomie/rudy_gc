package source

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	"rudy_gc/remote/api/internal/platform/sourceclient"
	contracts "rudy_gc/shared/dailybest"

	"github.com/PuerkitoBio/goquery"
)

const bestRatedMonthPath = "vl_bestrated.php?&mode="

type Source struct {
	client     *sourceclient.Client
	javAddress string
}

func New(client *sourceclient.Client, javAddress string) *Source {
	return &Source{client: client, javAddress: strings.TrimSpace(javAddress)}
}

func (s *Source) FetchPage(ctx context.Context, page int64) (contracts.DailyBestSnapshotPage, error) {
	if s.client == nil {
		return contracts.DailyBestSnapshotPage{}, fmt.Errorf("DailyBest SourceClient 为空")
	}
	query := fmt.Sprintf("/%s&page=%d", bestRatedMonthPath, page)
	targetURL := fmt.Sprintf("https://%s/cn%s", s.javAddress, query)
	resp, err := s.client.Get(ctx, targetURL)
	if err != nil {
		return contracts.DailyBestSnapshotPage{}, fmt.Errorf("请求 DailyBest 第 %d 页失败: %w", page, err)
	}
	if resp.StatusCode != http.StatusOK {
		return contracts.DailyBestSnapshotPage{}, fmt.Errorf("请求 DailyBest 第 %d 页返回状态码 %d", page, resp.StatusCode)
	}
	content, err := extractRightColumn(resp.Body)
	if err != nil {
		return contracts.DailyBestSnapshotPage{}, fmt.Errorf("解析 DailyBest 第 %d 页容器失败: %w", page, err)
	}
	if isBlankContent(content) {
		return contracts.DailyBestSnapshotPage{}, dailybest.ErrBlankPage
	}
	return contracts.DailyBestSnapshotPage{
		PageNumber: page,
		SourceURL:  targetURL,
		Content:    content,
		Checksum:   contracts.ComputeDailyBestPageChecksum(content),
	}, nil
}

func extractRightColumn(body []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return doc.Find("div[id=rightcolumn]").Html()
}

func isBlankContent(content string) bool {
	if !strings.Contains(content, "videothumblist") {
		return true
	}
	return strings.Contains(content, "搜寻没有结果") || strings.Contains(content, "空的列表")
}
