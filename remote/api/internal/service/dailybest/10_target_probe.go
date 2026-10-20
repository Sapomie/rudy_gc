package dailybest

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (s *Service) StartTargetProbeScheduler(ctx context.Context, interval time.Duration) error {
	if s.runtimeRepository == nil {
		return fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	if interval <= 0 {
		return fmt.Errorf("DailyBest 目标站探测间隔必须为正数")
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.executeScheduledTargetProbe(ctx)
		for {
			next := time.Now().Add(interval)
			s.runtimeMu.Lock()
			s.nextProbeAt = next.Unix()
			s.runtimeMu.Unlock()
			timer := time.NewTimer(time.Until(next))
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			case <-timer.C:
				s.executeScheduledTargetProbe(ctx)
			}
		}
	}()
	return nil
}

func (s *Service) executeScheduledTargetProbe(ctx context.Context) {
	if _, err := s.TriggerTargetProbe(ctx); err != nil {
		log.Printf("DailyBest 目标站探测执行失败: %v", err)
	}
}

func (s *Service) TriggerTargetProbe(ctx context.Context) (dailybesttypes.TargetProbe, error) {
	if !s.probeMu.TryLock() {
		return dailybesttypes.TargetProbe{}, ErrActiveTargetProbe
	}
	defer s.probeMu.Unlock()
	return s.executeTargetProbe(ctx)
}

func (s *Service) executeTargetProbe(ctx context.Context) (dailybesttypes.TargetProbe, error) {
	if s.agent == nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测未执行: FlareSolverr 客户端为空")
	}
	if s.runtimeRepository == nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 运行状态仓库为空")
	}
	config, err := s.GetSourceConfig(ctx)
	if err != nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测读取站点配置失败: %w", err)
	}
	address, err := normalizeJavAddress(config.JavAddress)
	if err != nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测读取域名失败: %w", err)
	}
	targetURL := dailybesttypes.SourcePageURL(address, 1)
	requestID, err := targetProbeRequestID()
	if err != nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测生成请求 ID 失败: %w", err)
	}
	probe, err := s.runtimeRepository.CreateTargetProbe(ctx, dailybesttypes.TargetProbeDraft{RequestID: requestID, TargetURL: targetURL})
	if err != nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测创建数据库记录失败，已阻止目标站请求: %w", err)
	}
	result, agentHTTPStatus, requestErr := s.agent.Request(ctx, requestID, targetURL, browseragenttypes.PurposeProbe)
	probeResult := evaluateTargetProbe(targetURL, agentHTTPStatus, result, requestErr)
	if err := s.runtimeRepository.CompleteTargetProbe(context.WithoutCancel(ctx), probe.ID, probeResult); err != nil {
		return dailybesttypes.TargetProbe{}, fmt.Errorf("DailyBest 目标站探测 %d 更新结果失败: %w", probe.ID, err)
	}
	completed := dailybesttypes.TargetProbe{
		ID: probe.ID, RequestID: requestID, Status: probeResult.Status, TargetURL: targetURL,
		FinalURL: probeResult.FinalURL, HTTPStatus: probeResult.HTTPStatus, DurationMS: probeResult.DurationMS,
		BrowserStatus: probeResult.BrowserStatus, ErrorType: probeResult.ErrorType, ErrorMessage: probeResult.ErrorMessage,
		StartedAt: probeResult.StartedAt, FinishedAt: probeResult.FinishedAt, CreatedOn: probe.CreatedOn, UpdatedOn: time.Now().Unix(),
	}
	s.updateTargetState(context.WithoutCancel(ctx), completed)
	return completed, nil
}

func evaluateTargetProbe(targetURL string, agentHTTPStatus int, result browseragenttypes.FetchResult, requestErr error) dailybesttypes.TargetProbeResult {
	finishedAt := result.FinishedAt
	if finishedAt == 0 {
		finishedAt = time.Now().Unix()
	}
	probeResult := dailybesttypes.TargetProbeResult{
		Status: dailybesttypes.TargetProbeStatusFailed, FinalURL: result.FinalURL,
		HTTPStatus: int64(result.StatusCode), DurationMS: result.DurationMS, BrowserStatus: result.BrowserState,
		ErrorType: result.ErrorType, ErrorMessage: result.ErrorMessage,
		StartedAt: result.StartedAt, FinishedAt: finishedAt,
	}
	if requestErr != nil {
		probeResult.ErrorType = firstValue(probeResult.ErrorType, "browser_request")
		probeResult.ErrorMessage = requestErr.Error()
		return probeResult
	}
	if agentHTTPStatus != http.StatusOK {
		probeResult.ErrorType = firstValue(probeResult.ErrorType, "browser_agent_http")
		probeResult.ErrorMessage = firstValue(probeResult.ErrorMessage, fmt.Sprintf("FlareSolverr 返回 HTTP %d", agentHTTPStatus))
		return probeResult
	}
	if result.StatusCode != http.StatusOK {
		probeResult.ErrorType = firstValue(probeResult.ErrorType, "target_http")
		probeResult.ErrorMessage = firstValue(probeResult.ErrorMessage, fmt.Sprintf("目标页面返回 HTTP %d", result.StatusCode))
		return probeResult
	}
	if err := validateFinalURL(targetURL, result.FinalURL); err != nil {
		probeResult.ErrorType = "unexpected_final_url"
		probeResult.ErrorMessage = err.Error()
		return probeResult
	}
	document, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(result.Body)))
	if err != nil {
		probeResult.ErrorType = "invalid_html"
		probeResult.ErrorMessage = err.Error()
		return probeResult
	}
	if document.Find("#rightcolumn").Length() == 0 {
		probeResult.ErrorType = "missing_rightcolumn"
		probeResult.ErrorMessage = "目标页面缺少 #rightcolumn"
		return probeResult
	}
	if document.Find(".videothumblist, #videothumblist, [class*='videothumblist']").Length() == 0 {
		probeResult.ErrorType = "missing_rank_list"
		probeResult.ErrorMessage = "目标页面缺少 videothumblist 排行列表"
		return probeResult
	}
	pageText := strings.ToLower(document.Text())
	pageTitle := strings.ToLower(document.Find("title").Text())
	if strings.Contains(pageTitle, "just a moment") || strings.Contains(pageTitle, "attention required") || strings.Contains(pageText, "cloudflare ray id") {
		probeResult.ErrorType = "cloudflare"
		probeResult.ErrorMessage = "目标页面进入 Cloudflare 验证或访问拒绝页面"
		return probeResult
	}
	probeResult.Status = dailybesttypes.TargetProbeStatusSucceeded
	probeResult.ErrorType = ""
	probeResult.ErrorMessage = ""
	return probeResult
}

func validateFinalURL(targetURL, finalURL string) error {
	target, targetErr := url.Parse(targetURL)
	final, finalErr := url.Parse(finalURL)
	if targetErr != nil || finalErr != nil || final.Hostname() == "" {
		return fmt.Errorf("目标页面最终地址无效")
	}
	if !strings.EqualFold(target.Hostname(), final.Hostname()) {
		return fmt.Errorf("目标页面最终地址主机变为 %s", final.Hostname())
	}
	return nil
}

func (s *Service) updateTargetState(ctx context.Context, probe dailybesttypes.TargetProbe) {
	newStatus := dailybesttypes.TargetStatusUnavailable
	eventType := "target_unavailable"
	level := dailybesttypes.RuntimeLevelError
	message := probe.ErrorMessage
	if probe.Status == dailybesttypes.TargetProbeStatusSucceeded {
		newStatus = dailybesttypes.TargetStatusAvailable
		eventType = "target_available"
		level = dailybesttypes.RuntimeLevelInfo
		message = "DailyBest 目标页面探测成功"
	}
	s.runtimeMu.Lock()
	previous := s.targetStatus
	s.targetStatus = newStatus
	s.latestProbe = &probe
	if probe.Status == dailybesttypes.TargetProbeStatusSucceeded {
		s.lastProbeSuccess = probe.FinishedAt
	} else {
		s.lastProbeFailure = probe.FinishedAt
	}
	s.runtimeMu.Unlock()
	if previous == newStatus {
		return
	}
	if _, err := s.runtimeRepository.CreateRuntimeEvent(ctx, dailybesttypes.RuntimeEventDraft{
		Component: dailybesttypes.RuntimeComponentTarget, EventType: eventType, Level: level,
		Message: message, TargetProbeID: probe.ID, OccurredAt: probe.FinishedAt,
	}); err != nil {
		log.Printf("DailyBest 目标站状态事件写入失败: %v", err)
	}
}

func targetProbeRequestID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func firstValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
