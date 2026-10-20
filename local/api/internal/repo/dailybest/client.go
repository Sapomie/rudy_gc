package dailybestapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"rudy_gc/local/api/internal/config"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg config.DailyBestRemoteClientConf) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	proxyURL, err := cfg.ProxyURL()
	if err != nil {
		return nil, err
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &Client{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		http:    &http.Client{Transport: transport},
	}, nil
}

func (c *Client) ListRankDays(ctx context.Context, filter dailybesttypes.RankDayQuery) (contracts.DailyBestRankDayListResponse, error) {
	query := url.Values{}
	if filter.BusinessDateStart != "" {
		query.Set("business_date_start", filter.BusinessDateStart)
	}
	if filter.BusinessDateEnd != "" {
		query.Set("business_date_end", filter.BusinessDateEnd)
	}
	if filter.DayNumber > 0 {
		query.Set("day_number", strconv.FormatInt(filter.DayNumber, 10))
	}
	if filter.Page > 0 {
		query.Set("p", strconv.FormatInt(filter.Page, 10))
	}
	if filter.PageSize > 0 {
		query.Set("ps", strconv.FormatInt(filter.PageSize, 10))
	}
	endpoint := c.baseURL + "/api/v1/dailybest/rank-days"
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	var response contracts.DailyBestRankDayListResponse
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return contracts.DailyBestRankDayListResponse{}, err
	}
	return response, nil
}

func (c *Client) ListRanks(ctx context.Context, dayNumber, category int64) (contracts.DailyBestRankListResponse, error) {
	query := url.Values{}
	query.Set("day_number", strconv.FormatInt(dayNumber, 10))
	query.Set("category", strconv.FormatInt(category, 10))
	var response contracts.DailyBestRankListResponse
	if err := c.getJSON(ctx, c.baseURL+"/api/v1/dailybest/ranks?"+query.Encode(), &response); err != nil {
		return contracts.DailyBestRankListResponse{}, err
	}
	return response, nil
}

func (c *Client) getJSON(ctx context.Context, target string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("请求 DailyBest 远端 API 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiErr contracts.DailyBestAPIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Message != "" {
			return fmt.Errorf("DailyBest 远端 API 返回 %d: %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("DailyBest 远端 API 返回状态码 %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析 DailyBest 远端 API 响应失败: %w", err)
	}
	return nil
}
