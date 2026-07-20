package dailybest

import (
	"context"
	"sync"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

var ErrNotFound = dailybesttypes.ErrNotFound

type Repository interface {
	CreateFetchRun(ctx context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error)
	MarkFetchRunRunning(ctx context.Context, runID int64, startedAt int64) error
	MarkFetchRunFailed(ctx context.Context, runID int64, finishedAt int64, message string) error
	MarkFetchRunNoChange(
		ctx context.Context,
		runID, finishedAt, pageCount, itemCount int64,
		checksum string,
		movies []dailybesttypes.MovieDraft,
	) error
	PublishRanks(ctx context.Context, runID int64, draft dailybesttypes.RankDraft, projection dailybesttypes.RankProjection, finishedAt int64) error
	RankProjectionComplete(ctx context.Context) (bool, error)
	SyncRankProjection(ctx context.Context, projection dailybesttypes.RankProjection) error
	ListAllRanks(ctx context.Context) ([]contracts.DailyBestRank, error)
	ListRankSyncManifests(ctx context.Context) ([]contracts.DailyBestRankSyncManifest, error)
	ListRanks(ctx context.Context, filter dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error)
	ListRankDayStats(ctx context.Context, filter dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error)
	ListRankPeriods(ctx context.Context, filter dailybesttypes.RankPeriodFilter) ([]contracts.DailyBestRankPeriod, int64, error)
	ListRankPeriodItems(ctx context.Context, filter dailybesttypes.RankPeriodItemFilter) ([]contracts.DailyBestRankPeriodItem, int64, error)
	ListRankPeriodStats(ctx context.Context, filter dailybesttypes.RankPeriodStatFilter) ([]contracts.DailyBestRankPeriodStat, int64, error)
	ListFetchRuns(ctx context.Context, page dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error)
	FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error)
	Ping(ctx context.Context) error
}

type PageSource interface {
	FetchPage(ctx context.Context, page int64) (dailybesttypes.SourcePage, error)
}

type Service struct {
	repository Repository
	source     PageSource
	root       context.Context
	loc        *time.Location

	runMu sync.Mutex
	wg    sync.WaitGroup
}

func New(root context.Context, repository Repository, source PageSource, loc *time.Location) *Service {
	if root == nil {
		root = context.Background()
	}
	if loc == nil {
		loc = time.Local
	}
	return &Service{repository: repository, source: source, root: root, loc: loc}
}
