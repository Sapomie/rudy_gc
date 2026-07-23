package dailybestsource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type Response struct {
	Body         []byte
	StatusCode   int
	RequestID    string
	FinalURL     string
	DurationMS   int64
	StartedAt    int64
	FinishedAt   int64
	BrowserState string
	ErrorType    string
}

type PageClient interface {
	Get(ctx context.Context, target string) (Response, error)
}

type RequestLogger interface {
	CreateFetchLog(ctx context.Context, draft dailybesttypes.FetchLogDraft) (int64, error)
	CompleteFetchLog(ctx context.Context, logID int64, result dailybesttypes.FetchLogResult) error
}

type Client struct {
	pageClient PageClient
	logger     RequestLogger
}

func NewClient(pageClient PageClient, logger RequestLogger) *Client {
	return &Client{pageClient: pageClient, logger: logger}
}

func (c *Client) Get(ctx context.Context, fetchRunID, pageNumber int64, target string) (Response, error) {
	if c.pageClient == nil {
		return Response{}, fmt.Errorf("DailyBest 页面客户端为空")
	}
	if c.logger == nil {
		return Response{}, fmt.Errorf("DailyBest 请求日志仓库为空")
	}
	logID, err := c.logger.CreateFetchLog(ctx, dailybesttypes.FetchLogDraft{
		FetchRunID:    fetchRunID,
		PageNumber:    pageNumber,
		AttemptNumber: 1,
		RequestURL:    target,
	})
	if err != nil {
		return Response{}, fmt.Errorf("创建 DailyBest 第 %d 页请求日志失败: %w", pageNumber, err)
	}

	startedAt := time.Now()
	response, err := c.pageClient.Get(ctx, target)
	if err != nil {
		requestErr := fmt.Errorf("执行浏览器页面请求失败: %w", err)
		if logErr := c.completeFetchLog(ctx, logID, dailybesttypes.FetchLogResult{
			Status:       dailybesttypes.FetchLogStatusFailed,
			HTTPStatus:   int64(response.StatusCode),
			DurationMS:   time.Since(startedAt).Milliseconds(),
			ErrorMessage: requestErr.Error(),
		}); logErr != nil {
			return Response{}, errors.Join(requestErr, logErr)
		}
		return Response{}, err
	}
	result := dailybesttypes.FetchLogResult{
		Status:     dailybesttypes.FetchLogStatusSucceeded,
		HTTPStatus: int64(response.StatusCode),
		DurationMS: time.Since(startedAt).Milliseconds(),
	}
	if response.StatusCode != http.StatusOK {
		result.Status = dailybesttypes.FetchLogStatusFailed
		result.ErrorMessage = fmt.Sprintf("HTTP 状态码 %d", response.StatusCode)
	}
	if err := c.completeFetchLog(ctx, logID, result); err != nil {
		return Response{}, err
	}
	return response, nil
}

func (c *Client) completeFetchLog(ctx context.Context, logID int64, result dailybesttypes.FetchLogResult) error {
	if err := c.logger.CompleteFetchLog(context.WithoutCancel(ctx), logID, result); err != nil {
		return fmt.Errorf("更新 DailyBest 请求日志 %d 失败: %w", logID, err)
	}
	return nil
}
