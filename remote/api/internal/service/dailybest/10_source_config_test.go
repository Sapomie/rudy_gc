package dailybest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type sourceConfigMemory struct {
	mu      sync.Mutex
	config  *dailybesttypes.SourceConfig
	failure error
}

func (r *sourceConfigMemory) FindSourceConfig(context.Context) (dailybesttypes.SourceConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failure != nil {
		return dailybesttypes.SourceConfig{}, r.failure
	}
	if r.config == nil {
		return dailybesttypes.SourceConfig{}, ErrNotFound
	}
	return *r.config, nil
}
func (r *sourceConfigMemory) InsertSourceConfigIfMissing(_ context.Context, address string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failure != nil {
		return r.failure
	}
	if r.config == nil {
		r.config = &dailybesttypes.SourceConfig{JavAddress: address, CreatedOn: 10, UpdatedOn: 10}
	}
	return nil
}
func (r *sourceConfigMemory) SaveSourceConfig(_ context.Context, address string) (dailybesttypes.SourceConfig, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failure != nil {
		return dailybesttypes.SourceConfig{}, r.failure
	}
	created := int64(10)
	if r.config != nil {
		created = r.config.CreatedOn
	}
	r.config = &dailybesttypes.SourceConfig{JavAddress: address, CreatedOn: created, UpdatedOn: 20}
	return *r.config, nil
}

type sourceConfigTestRepository struct {
	Repository
	sourceConfigMemory
}

func TestSourceConfigImportsOnceAndSurvivesServiceRestart(t *testing.T) {
	ctx := context.Background()
	repository := &sourceConfigTestRepository{}
	service := New(ctx, repository, nil, time.UTC)
	if err := service.PrepareSourceConfig(ctx, " OLD.EXAMPLE.COM "); err != nil {
		t.Fatal(err)
	}
	config, err := service.GetSourceConfig(ctx)
	if err != nil || config.JavAddress != "old.example.com" {
		t.Fatalf("config=%+v err=%v", config, err)
	}
	saved, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: " NEW.EXAMPLE.COM "})
	if err != nil || saved.JavAddress != "new.example.com" || saved.CreatedOn != 10 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	restarted := New(ctx, repository, nil, time.UTC)
	if err := restarted.PrepareSourceConfig(ctx, "old.example.com"); err != nil {
		t.Fatal(err)
	}
	config, err = restarted.GetSourceConfig(ctx)
	if err != nil || config.JavAddress != saved.JavAddress || config.TargetProbeURL != dailybesttypes.SourcePageURL(saved.JavAddress, 1) {
		t.Fatalf("config=%+v err=%v", config, err)
	}
}

func TestSourceConfigAllowsInitialConfigurationAndRejectsInvalidDomain(t *testing.T) {
	ctx := context.Background()
	repository := &sourceConfigTestRepository{}
	service := New(ctx, repository, nil, time.UTC)
	if err := service.PrepareSourceConfig(ctx, ""); err != nil {
		t.Fatal(err)
	}
	config, err := service.GetSourceConfig(ctx)
	if err != nil || config.JavAddress != "" || config.TargetProbeURL != "" {
		t.Fatalf("config=%+v err=%v", config, err)
	}
	for _, invalid := range []string{"", "https://example.com", "example.com/cn", "example.com:443", "a..com", "-a.com", "a-.com", "a.com?x=y", "a.com#x", "u@a.com"} {
		if _, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: invalid}); !errors.Is(err, ErrInvalidSourceConfig) {
			t.Fatalf("address=%q err=%v", invalid, err)
		}
	}
	if repository.config != nil {
		t.Fatal("无效域名写入了配置")
	}
	if _, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: "example.com"}); err != nil {
		t.Fatal(err)
	}
	repository.failure = errors.New("database unavailable")
	if err := service.PrepareSourceConfig(ctx, "old.example.com"); !errors.Is(err, repository.failure) {
		t.Fatalf("err=%v", err)
	}
	if _, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: "new.example.com"}); !errors.Is(err, repository.failure) {
		t.Fatalf("err=%v", err)
	}
}

type configuredCollectRepository struct {
	collectPublishRepository
	sourceConfigMemory
	publishCount int
}

func (r *configuredCollectRepository) PublishRanks(ctx context.Context, id int64, draft dailybesttypes.RankDraft, stat contracts.DailyBestRankDayStat, finished int64) error {
	r.publishCount++
	return r.collectPublishRepository.PublishRanks(ctx, id, draft, stat, finished)
}

type configuredCollectSource struct {
	repeatedLastPageSource
	address    string
	urls       *[]string
	afterFirst func()
}

func (s *configuredCollectSource) FetchPage(ctx context.Context, id, page int64) (dailybesttypes.SourcePage, error) {
	result, err := s.repeatedLastPageSource.FetchPage(ctx, id, page)
	result.SourceURL = dailybesttypes.SourcePageURL(s.address, page)
	*s.urls = append(*s.urls, result.SourceURL)
	if page == 1 && s.afterFirst != nil {
		s.afterFirst()
	}
	return result, err
}

func TestConfiguredFetchKeepsDomainThroughRepeatedPageAndPublishesThenUsesNewDomain(t *testing.T) {
	ctx := context.Background()
	repository := &configuredCollectRepository{}
	if err := repository.InsertSourceConfigIfMissing(ctx, "old.example.com"); err != nil {
		t.Fatal(err)
	}
	var addresses, urls []string
	var service *Service
	service = New(ctx, repository, nil, time.UTC, WithSourceFactory(func(address string) PageSource {
		addresses = append(addresses, address)
		return &configuredCollectSource{address: address, urls: &urls, afterFirst: func() {
			if _, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: "new.example.com"}); err != nil {
				t.Fatal(err)
			}
		}}
	}))
	for _, id := range []int64{1, 2} {
		if err := service.collectAndPublish(ctx, id, "2026-10-12", contracts.DailyBestCategoryMonth); err != nil {
			t.Fatal(err)
		}
	}
	if fmt.Sprint(addresses) != "[old.example.com new.example.com]" || repository.publishCount != 2 {
		t.Fatalf("addresses=%v published=%d", addresses, repository.publishCount)
	}
	if len(urls) != 6 {
		t.Fatalf("urls=%v", urls)
	}
	for index, address := range addresses {
		for page := 1; page <= 3; page++ {
			if urls[index*3+page-1] != dailybesttypes.SourcePageURL(address, int64(page)) {
				t.Fatalf("urls=%v", urls)
			}
		}
	}
	if repository.publishedDraft.PageCount != 2 || len(repository.publishedDraft.Ranks) != 2 || repository.publishedStat.RankRowCount != 2 {
		t.Fatalf("draft=%+v stat=%+v", repository.publishedDraft, repository.publishedStat)
	}
}

type configuredProbeRepository struct {
	targetProbeRepository
	sourceConfigMemory
}

func (r *configuredProbeRepository) FindSourceConfig(ctx context.Context) (dailybesttypes.SourceConfig, error) {
	return r.sourceConfigMemory.FindSourceConfig(ctx)
}

func (r *configuredProbeRepository) InsertSourceConfigIfMissing(ctx context.Context, address string) error {
	return r.sourceConfigMemory.InsertSourceConfigIfMissing(ctx, address)
}

func (r *configuredProbeRepository) SaveSourceConfig(ctx context.Context, address string) (dailybesttypes.SourceConfig, error) {
	return r.sourceConfigMemory.SaveSourceConfig(ctx, address)
}

type configuredProbeAgent struct{ urls []string }

func (a *configuredProbeAgent) Request(_ context.Context, _, target, _ string) (browseragenttypes.FetchResult, int, error) {
	a.urls = append(a.urls, target)
	return browseragenttypes.FetchResult{StatusCode: 200, FinalURL: target, Body: `<div id="rightcolumn"><div class="videothumblist"></div></div>`}, 200, nil
}
func (*configuredProbeAgent) Status(context.Context) (browseragenttypes.Status, error) {
	return browseragenttypes.Status{}, nil
}

func TestTargetProbeReadsSavedDomainOnEveryExecution(t *testing.T) {
	ctx := context.Background()
	repository := &configuredProbeRepository{targetProbeRepository: targetProbeRepository{completed: make(chan dailybesttypes.TargetProbeResult, 2)}}
	agent := &configuredProbeAgent{}
	service := New(ctx, repository, nil, time.UTC, WithAgent(agent))
	for _, address := range []string{"old.example.com", "new.example.com"} {
		if _, err := service.UpdateSourceConfig(ctx, dailybesttypes.UpdateSourceConfigRequest{JavAddress: address}); err != nil {
			t.Fatal(err)
		}
		probe, err := service.TriggerTargetProbe(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if probe.Status != dailybesttypes.TargetProbeStatusSucceeded {
			t.Fatalf("probe=%+v", probe)
		}
		select {
		case result := <-repository.completed:
			if result.Status != dailybesttypes.TargetProbeStatusSucceeded {
				t.Fatalf("result=%+v", result)
			}
		default:
			t.Fatal("探测未完成")
		}
	}
	if len(agent.urls) != 2 || agent.urls[0] != dailybesttypes.SourcePageURL("old.example.com", 1) || agent.urls[1] != dailybesttypes.SourcePageURL("new.example.com", 1) {
		t.Fatalf("urls=%v", agent.urls)
	}
}
