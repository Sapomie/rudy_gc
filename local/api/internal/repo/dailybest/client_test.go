package dailybestapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"rudy_gc/local/api/internal/config"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func TestListRankDaysUsesDirectConnectionWithoutAuthorization(t *testing.T) {
	var authorization string
	var requestURI string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		requestURI = request.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"days":[]}`))
	}))
	defer server.Close()

	client, err := New(config.DailyBestRemoteClientConf{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("创建 DailyBest 直连客户端失败: %v", err)
	}
	if _, err := client.ListRankDays(context.Background(), dailybesttypes.RankDayQuery{
		BusinessDateStart: "2026-07-01", Page: 2, PageSize: 24,
	}); err != nil {
		t.Fatalf("直连获取 DailyBest 排名日期失败: %v", err)
	}
	if authorization != "" {
		t.Fatalf("请求携带了 Authorization 头")
	}
	if requestURI != "/api/v1/dailybest/rank-days?business_date_start=2026-07-01&p=2&ps=24" {
		t.Fatalf("request uri=%s", requestURI)
	}
}

func TestListRankDaysUsesConfiguredProxyWithoutAuthorization(t *testing.T) {
	type observedRequest struct {
		authorization string
		target        string
	}
	directRequests := make(chan struct{}, 1)
	remoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		directRequests <- struct{}{}
		http.Error(w, "direct request rejected", http.StatusBadGateway)
	}))
	defer remoteServer.Close()

	proxiedRequests := make(chan observedRequest, 1)
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		proxiedRequests <- observedRequest{
			authorization: request.Header.Get("Authorization"),
			target:        request.URL.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"days":[]}`))
	}))
	defer proxyServer.Close()

	client, err := New(config.DailyBestRemoteClientConf{BaseURL: remoteServer.URL, Proxy: proxyServer.URL})
	if err != nil {
		t.Fatalf("创建 DailyBest 客户端失败: %v", err)
	}
	if _, err := client.ListRankDays(context.Background(), dailybesttypes.RankDayQuery{
		BusinessDateStart: "2026-07-01", Page: 2, PageSize: 24,
	}); err != nil {
		t.Fatalf("获取 DailyBest 排名日期失败: %v", err)
	}
	request := <-proxiedRequests
	if request.authorization != "" {
		t.Fatalf("请求携带了 Authorization 头")
	}
	expectedTarget := remoteServer.URL + "/api/v1/dailybest/rank-days?business_date_start=2026-07-01&p=2&ps=24"
	if request.target != expectedTarget {
		t.Fatalf("代理目标=%s，期望=%s", request.target, expectedTarget)
	}
	select {
	case <-directRequests:
		t.Fatal("DailyBest 请求绕过代理直连远端")
	default:
	}
}

func TestNewRejectsInvalidProxy(t *testing.T) {
	for _, proxy := range []string{"127.0.0.1:7897", "socks5://127.0.0.1:7897"} {
		if _, err := New(config.DailyBestRemoteClientConf{BaseURL: "http://remote.example.com", Proxy: proxy}); err == nil {
			t.Fatalf("Proxy=%q 时应拒绝创建客户端", proxy)
		}
	}
}
