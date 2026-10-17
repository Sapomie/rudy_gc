package dailybestsource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type Source struct {
	client     *Client
	javAddress string
}

func NewSource(client *Client, javAddress string) *Source {
	return &Source{client: client, javAddress: strings.TrimSpace(javAddress)}
}

func (s *Source) WithAddress(javAddress string) *Source {
	return NewSource(s.client, javAddress)
}

func (s *Source) FetchPage(ctx context.Context, fetchRunID, page int64) (dailybesttypes.SourcePage, error) {
	if s.client == nil {
		return dailybesttypes.SourcePage{}, fmt.Errorf("DailyBest SourceClient 为空")
	}
	targetURL := dailybesttypes.SourcePageURL(s.javAddress, page)
	response, err := s.client.Get(ctx, fetchRunID, page, targetURL)
	if err != nil {
		return dailybesttypes.SourcePage{}, fmt.Errorf("请求 DailyBest 第 %d 页失败: %w", page, err)
	}
	if response.StatusCode != http.StatusOK {
		return dailybesttypes.SourcePage{}, fmt.Errorf("请求 DailyBest 第 %d 页返回状态码 %d", page, response.StatusCode)
	}
	content, err := extractRightColumn(response.Body)
	if err != nil {
		return dailybesttypes.SourcePage{}, fmt.Errorf("解析 DailyBest 第 %d 页容器失败: %w", page, err)
	}
	if isBlankContent(content) {
		return dailybesttypes.SourcePage{}, dailybesttypes.ErrBlankPage
	}
	sum := sha256.Sum256([]byte(content))
	return dailybesttypes.SourcePage{
		PageNumber: page,
		SourceURL:  targetURL,
		Content:    content,
		Checksum:   hex.EncodeToString(sum[:]),
	}, nil
}

func extractRightColumn(body []byte) (string, error) {
	document, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	return document.Find("div[id=rightcolumn]").Html()
}

func isBlankContent(content string) bool {
	if !strings.Contains(content, "videothumblist") {
		return true
	}
	return strings.Contains(content, "搜寻没有结果") || strings.Contains(content, "空的列表")
}
