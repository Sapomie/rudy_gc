package spider

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/dep"
	"rudy_gc/local/api/internal/model/modelx/spiderflowx"
	"rudy_gc/local/api/internal/model/modelx/spiderx"
	fetchsiterepo "rudy_gc/local/api/internal/repo/fetchsite"
	"rudy_gc/local/api/internal/service/spider/fetcher"
	fetchsitetypes "rudy_gc/local/api/internal/types/fetchsite"
)

type bestinvTestResult struct{}

func (bestinvTestResult) LastInsertId() (int64, error) { return 1, nil }
func (bestinvTestResult) RowsAffected() (int64, error) { return 1, nil }

type bestinvTestConn struct {
	sqlx.SqlConn
	row *spiderx.DBestinv
}

func (c *bestinvTestConn) QueryRowCtx(_ context.Context, value any, query string, args ...any) error {
	destination, ok := value.(*spiderx.DBestinv)
	if !ok {
		return fmt.Errorf("unexpected QueryRowCtx destination %T", value)
	}
	if strings.Contains(query, "where `name` = ?") {
		if c.row == nil || len(args) != 1 || c.row.Name != args[0] {
			return sqlx.ErrNotFound
		}
		*destination = *c.row
		return nil
	}
	if strings.Contains(query, "where `id` = ?") {
		if c.row == nil || len(args) != 1 || c.row.Id != args[0] {
			return sqlx.ErrNotFound
		}
		*destination = *c.row
		return nil
	}
	return fmt.Errorf("unexpected QueryRowCtx query: %s", query)
}

func (c *bestinvTestConn) QueryRowsCtx(_ context.Context, value any, query string, _ ...any) error {
	ids, ok := value.(*[]int64)
	if !ok || !strings.Contains(query, "`need_scan` = ?") {
		return fmt.Errorf("unexpected QueryRowsCtx query: %s", query)
	}
	if c.row != nil && c.row.NeedScan == consts.BestinvNeedScan {
		*ids = []int64{c.row.Id}
	}
	return nil
}

func (c *bestinvTestConn) ExecCtx(_ context.Context, query string, args ...any) (sql.Result, error) {
	switch {
	case strings.HasPrefix(query, "insert into `d_bestinv`"):
		if len(args) != 11 {
			return nil, fmt.Errorf("unexpected insert args: %d", len(args))
		}
		c.row = &spiderx.DBestinv{
			Id:            1,
			Name:          args[0].(string),
			NeedScan:      args[1].(int64),
			NeedRankCheck: args[2].(int64),
			Category:      args[3].(int64),
			Page:          args[4].(int64),
			DayNumber:     args[5].(int64),
			Content:       args[6].(string),
			LastQueryTime: args[7].(int64),
			Date:          args[8].(string),
			CreatedOn:     args[9].(int64),
			UpdatedOn:     args[10].(int64),
		}
		return bestinvTestResult{}, nil
	case strings.HasPrefix(query, "update `d_bestinv`"):
		if c.row == nil || len(args) != 12 {
			return nil, fmt.Errorf("unexpected update args: %d", len(args))
		}
		c.row.NeedScan = args[1].(int64)
		c.row.UpdatedOn = args[10].(int64)
		return bestinvTestResult{}, nil
	default:
		return nil, fmt.Errorf("unexpected ExecCtx query: %s", query)
	}
}

type bestinvRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn bestinvRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestFetchAndParseDailyBestinvStopsAtDuplicatePageAndContinues(t *testing.T) {
	content := `<html><body><div id="rightcolumn"><div class="videothumblist">same ranking page</div></div></body></html>`
	requestTimes := make([]time.Time, 0, 2)
	originalTransport := http.DefaultTransport
	http.DefaultTransport = bestinvRoundTripFunc(func(request *http.Request) (*http.Response, error) {
		requestTimes = append(requestTimes, time.Now())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(content)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	conn := &bestinvTestConn{}
	bestinvModel := spiderx.NewDBestinvModel(conn)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	service := &Service{deps: &runtimeDeps{
		Dep: &dep.Dep{
			Config: config.Config{Fetcher: config.FetcherConf{JavAddress: "dailybest.test"}},
			Log:    logger,
			FetchSiteRegistry: fetchsiterepo.NewRegistry(map[string]fetchsitetypes.Config{
				fetchsitetypes.CodeJavlib: {
					SiteCode: fetchsitetypes.CodeJavlib, BaseURL: "https://dailybest.test", Status: 1,
				},
			}),
			Fetcher: fetcher.NewFetcher(fetcher.Config{
				Timeout: 10 * time.Second,
			}),
		},
		Repositories: spiderflowx.NewRepositories(spiderflowx.Dependencies{
			BestinvModel: bestinvModel,
		}),
	}}

	if _, err := service.FetchAndParseDailyBestinv(context.Background(), false, nil); err != nil {
		t.Fatalf("FetchAndParseDailyBestinv: %v", err)
	}
	if len(requestTimes) != 2 {
		t.Fatalf("requests=%d, want 2", len(requestTimes))
	}
	if gap := requestTimes[1].Sub(requestTimes[0]); gap < 2*time.Second {
		t.Fatalf("request gap=%v, real random wait was skipped", gap)
	}
	if conn.row == nil || conn.row.Page != 1 {
		t.Fatalf("persisted row=%+v, duplicate page must not be stored", conn.row)
	}
	if conn.row.NeedScan != consts.BestinvNoNeedScan {
		t.Fatalf("need_scan=%d, ParseBestinv did not execute", conn.row.NeedScan)
	}
}
