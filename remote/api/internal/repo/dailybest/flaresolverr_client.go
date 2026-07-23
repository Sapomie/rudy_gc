package dailybestsource

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
)

type FlareSolverrClientConfig struct {
	BaseURL string
	Session string
	Timeout time.Duration
}

type FlareSolverrClient struct {
	baseURL string
	session string
	timeout time.Duration
	client  *http.Client

	mu           sync.Mutex
	sessionReady bool
}

type flareSolverrResponse struct {
	Status    string               `json:"status"`
	Message   string               `json:"message"`
	Msg       string               `json:"msg"`
	Version   string               `json:"version"`
	UserAgent string               `json:"userAgent"`
	Solution  flareSolverrSolution `json:"solution"`
}

type flareSolverrSolution struct {
	Status    int    `json:"status"`
	Response  string `json:"response"`
	URL       string `json:"url"`
	UserAgent string `json:"userAgent"`
}

func NewFlareSolverrClient(cfg FlareSolverrClientConfig) (*FlareSolverrClient, error) {
	parsed, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("FlareSolverr BaseURL 无效")
	}
	if strings.TrimSpace(cfg.Session) == "" {
		return nil, fmt.Errorf("FlareSolverr Session 为空")
	}
	if cfg.Timeout <= 0 {
		return nil, fmt.Errorf("FlareSolverr 页面超时时间必须为正数")
	}
	return &FlareSolverrClient{
		baseURL: strings.TrimRight(parsed.String(), "/"),
		session: strings.TrimSpace(cfg.Session),
		timeout: cfg.Timeout,
		client:  &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (c *FlareSolverrClient) Get(ctx context.Context, target string) (Response, error) {
	requestID, err := newRequestID()
	if err != nil {
		return Response{}, err
	}
	return c.Fetch(ctx, requestID, target, browseragenttypes.PurposeDailyBest)
}

func (c *FlareSolverrClient) Fetch(ctx context.Context, requestID, target, purpose string) (Response, error) {
	result, statusCode, err := c.Request(ctx, requestID, target, purpose)
	response := Response{
		Body:         []byte(result.Body),
		StatusCode:   result.StatusCode,
		RequestID:    result.RequestID,
		FinalURL:     result.FinalURL,
		DurationMS:   result.DurationMS,
		StartedAt:    result.StartedAt,
		FinishedAt:   result.FinishedAt,
		BrowserState: result.BrowserState,
		ErrorType:    result.ErrorType,
	}
	if err != nil {
		return response, err
	}
	if statusCode != http.StatusOK {
		return response, fmt.Errorf("FlareSolverr 请求返回 HTTP %d", statusCode)
	}
	if result.ErrorMessage != "" {
		return response, fmt.Errorf("FlareSolverr 页面请求失败: %s", result.ErrorMessage)
	}
	return response, nil
}

func (c *FlareSolverrClient) Request(
	ctx context.Context,
	requestID, target, _ string,
) (browseragenttypes.FetchResult, int, error) {
	startedAt := time.Now()
	result := browseragenttypes.FetchResult{
		RequestID:    strings.TrimSpace(requestID),
		StartedAt:    startedAt.Unix(),
		BrowserState: browseragenttypes.StateFetching,
	}
	if result.RequestID == "" {
		return c.finishResult(result, startedAt, "invalid_request", "FlareSolverr request_id 为空"), 0,
			fmt.Errorf("FlareSolverr request_id 为空")
	}
	parsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return c.finishResult(result, startedAt, "invalid_request", "FlareSolverr 目标地址无效"), 0,
			fmt.Errorf("FlareSolverr 目标地址无效")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.ensureSession(ctx); err != nil {
		message := fmt.Sprintf("准备 FlareSolverr 会话失败: %v", err)
		return c.finishResult(result, startedAt, "session", message), 0, fmt.Errorf("%s", message)
	}

	response, statusCode, err := c.requestPage(ctx, parsed.String())
	if err != nil && missingFlareSolverrSession(response.Message) {
		c.sessionReady = false
		if sessionErr := c.ensureSession(ctx); sessionErr != nil {
			message := fmt.Sprintf("恢复 FlareSolverr 会话失败: %v", sessionErr)
			return c.finishResult(result, startedAt, "session", message), statusCode, fmt.Errorf("%s", message)
		}
		response, statusCode, err = c.requestPage(ctx, parsed.String())
	}
	result.StatusCode = response.Solution.Status
	result.Body = response.Solution.Response
	result.FinalURL = response.Solution.URL
	if err != nil {
		errorType := "solver_request"
		if strings.Contains(strings.ToLower(response.Message), "challenge") {
			errorType = "cloudflare"
		}
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = err.Error()
		}
		result = c.finishResult(result, startedAt, errorType, message)
		return result, statusCode, fmt.Errorf("FlareSolverr 页面请求失败: %s", message)
	}
	result = c.finishResult(result, startedAt, "", "")
	return result, statusCode, nil
}

func (c *FlareSolverrClient) Status(ctx context.Context) (browseragenttypes.Status, error) {
	var response flareSolverrResponse
	statusCode, err := c.doJSON(ctx, http.MethodGet, "/", nil, &response)
	if err != nil {
		return browseragenttypes.Status{}, err
	}
	if statusCode != http.StatusOK || !strings.Contains(strings.ToLower(response.Msg), "ready") {
		return browseragenttypes.Status{}, fmt.Errorf("FlareSolverr 状态接口返回 HTTP %d: %s", statusCode, response.Msg)
	}
	return browseragenttypes.Status{
		State:           browseragenttypes.StateReady,
		LastHeartbeatAt: time.Now().Unix(),
		Version:         response.Version,
		UserAgent:       response.UserAgent,
	}, nil
}

func (c *FlareSolverrClient) ensureSession(ctx context.Context) error {
	if c.sessionReady {
		return nil
	}
	var response flareSolverrResponse
	statusCode, err := c.doJSON(ctx, http.MethodPost, "/v1", map[string]any{
		"cmd": "sessions.create", "session": c.session,
	}, &response)
	if err == nil && statusCode == http.StatusOK && response.Status == "ok" {
		c.sessionReady = true
		return nil
	}
	message := strings.ToLower(response.Message)
	if strings.Contains(message, "session") && (strings.Contains(message, "already") || strings.Contains(message, "exists")) {
		c.sessionReady = true
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("HTTP %d: %s", statusCode, response.Message)
}

func (c *FlareSolverrClient) requestPage(ctx context.Context, target string) (flareSolverrResponse, int, error) {
	var response flareSolverrResponse
	statusCode, err := c.doJSON(ctx, http.MethodPost, "/v1", map[string]any{
		"cmd":        "request.get",
		"url":        target,
		"session":    c.session,
		"maxTimeout": c.timeout.Milliseconds(),
	}, &response)
	if err != nil {
		return response, statusCode, err
	}
	if statusCode != http.StatusOK || response.Status != "ok" {
		return response, statusCode, fmt.Errorf("FlareSolverr 返回 HTTP %d: %s", statusCode, response.Message)
	}
	return response, statusCode, nil
}

func (c *FlareSolverrClient) doJSON(ctx context.Context, method, path string, input, output any) (int, error) {
	var body io.Reader
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return 0, err
	}
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("连接 FlareSolverr 失败: %w", err)
	}
	defer response.Body.Close()
	if output != nil {
		if err := json.NewDecoder(response.Body).Decode(output); err != nil {
			return response.StatusCode, fmt.Errorf("解析 FlareSolverr 响应失败: %w", err)
		}
	}
	return response.StatusCode, nil
}

func (c *FlareSolverrClient) finishResult(
	result browseragenttypes.FetchResult,
	startedAt time.Time,
	errorType, errorMessage string,
) browseragenttypes.FetchResult {
	result.DurationMS = time.Since(startedAt).Milliseconds()
	result.FinishedAt = time.Now().Unix()
	result.BrowserState = browseragenttypes.StateReady
	result.ErrorType = errorType
	result.ErrorMessage = errorMessage
	return result
}

func missingFlareSolverrSession(message string) bool {
	value := strings.ToLower(message)
	return strings.Contains(value, "session") &&
		(strings.Contains(value, "not exist") || strings.Contains(value, "does not exist") || strings.Contains(value, "not found"))
}

func newRequestID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("生成浏览器请求 ID 失败: %w", err)
	}
	return hex.EncodeToString(value), nil
}
