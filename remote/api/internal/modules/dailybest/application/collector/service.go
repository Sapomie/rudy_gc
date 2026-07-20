package dailybestcollector

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	contracts "rudy_gc/shared/dailybest"
)

type Service struct {
	repo   dailybest.Repository
	source dailybest.PageSource
	root   context.Context
	loc    *time.Location

	runMu sync.Mutex
	wg    sync.WaitGroup
}

func NewService(root context.Context, repo dailybest.Repository, source dailybest.PageSource, loc *time.Location) *Service {
	if root == nil {
		root = context.Background()
	}
	if loc == nil {
		loc = time.Local
	}
	return &Service{repo: repo, source: source, root: root, loc: loc}
}

func (s *Service) CurrentBusinessDate(now time.Time) string {
	return now.In(s.loc).Add(-8 * time.Hour).Format(time.DateOnly)
}

func (s *Service) Enqueue(ctx context.Context, req contracts.DailyBestFetchRequest, triggerType string) (contracts.DailyBestFetchRun, error) {
	date := strings.TrimSpace(req.BusinessDate)
	if date == "" {
		date = s.CurrentBusinessDate(time.Now())
	}
	if _, err := time.ParseInLocation(time.DateOnly, date, s.loc); err != nil {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("business_date 格式错误: %w", err)
	}
	currentDate := s.CurrentBusinessDate(time.Now())
	if date != currentDate {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("源站只提供当前榜单，business_date 必须为 %s", currentDate)
	}
	category := req.Category
	if category == 0 {
		category = contracts.DailyBestCategoryMonth
	}
	if category != contracts.DailyBestCategoryMonth {
		return contracts.DailyBestFetchRun{}, fmt.Errorf("当前只支持 DailyBest 月榜 category=%d", contracts.DailyBestCategoryMonth)
	}
	if triggerType == "" {
		triggerType = dailybest.TriggerManual
	}
	run, err := s.repo.CreateFetchRun(ctx, date, category, triggerType)
	if err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	s.wg.Add(1)
	go func(runID int64, businessDate string, category int64) {
		defer s.wg.Done()
		s.execute(s.root, runID, businessDate, category)
	}(run.ID, date, category)
	return run, nil
}

func (s *Service) execute(ctx context.Context, runID int64, businessDate string, category int64) {
	s.runMu.Lock()
	defer s.runMu.Unlock()
	startedAt := time.Now().Unix()
	if err := s.repo.MarkFetchRunRunning(ctx, runID, startedAt); err != nil {
		return
	}
	if err := s.collectAndPublish(ctx, runID, businessDate, category); err != nil {
		_ = s.repo.MarkFetchRunFailed(context.WithoutCancel(ctx), runID, time.Now().Unix(), err.Error())
	}
}

func (s *Service) collectAndPublish(ctx context.Context, runID int64, businessDate string, category int64) error {
	pages := make([]contracts.DailyBestSnapshotPage, 0)
	seenPageChecksums := make(map[string]struct{})
	for pageNumber := int64(1); ; pageNumber++ {
		page, err := s.source.FetchPage(ctx, pageNumber)
		if errors.Is(err, dailybest.ErrBlankPage) {
			if pageNumber == 1 {
				return fmt.Errorf("DailyBest 第一页为空")
			}
			break
		}
		if err != nil {
			return err
		}
		if _, exists := seenPageChecksums[page.Checksum]; exists {
			return fmt.Errorf("DailyBest 第 %d 页与已有页面内容重复，停止发布", pageNumber)
		}
		seenPageChecksums[page.Checksum] = struct{}{}
		pages = append(pages, page)
	}
	items, err := ParsePages(pages)
	if err != nil {
		return err
	}
	draftResponse := contracts.DailyBestSnapshotResponse{
		DailyBestSnapshotManifest: contracts.DailyBestSnapshotManifest{
			BusinessDate: businessDate,
			Category:     category,
			PageCount:    int64(len(pages)),
			ItemCount:    int64(len(items)),
		},
		Pages: pages,
		Items: items,
	}
	checksum := contracts.ComputeDailyBestSnapshotChecksum(draftResponse)
	_, err = s.repo.PublishSnapshot(ctx, runID, dailybest.SnapshotDraft{
		BusinessDate: businessDate,
		Category:     category,
		Pages:        pages,
		Items:        items,
		Checksum:     checksum,
	}, time.Now().Unix())
	return err
}

func (s *Service) Wait() {
	s.wg.Wait()
}
