package dailybest

import (
	"context"
	"sync"
	"time"

	"rudy_gc/remote/api/internal/notify"
	browseragenttypes "rudy_gc/remote/api/internal/types/browseragent"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

var (
	ErrNotFound       = dailybesttypes.ErrNotFound
	ErrActiveFetchRun = dailybesttypes.ErrActiveFetchRun
)

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
	PublishRanks(ctx context.Context, runID int64, draft dailybesttypes.RankDraft, stat contracts.DailyBestRankDayStat, finishedAt int64) error
	ListRankDayStatDayNumbers(ctx context.Context) ([]int64, error)
	InsertRankDayStatsIfMissing(ctx context.Context, stats []contracts.DailyBestRankDayStat) error
	ListAllRanks(ctx context.Context) ([]contracts.DailyBestRank, error)
	ListRankDays(ctx context.Context, filter dailybesttypes.RankDayFilter) ([]contracts.DailyBestRankDay, int64, error)
	ListRanks(ctx context.Context, filter dailybesttypes.RankFilter) ([]contracts.DailyBestRank, int64, error)
	ListRankDayStats(ctx context.Context, filter dailybesttypes.RankDayStatFilter) ([]contracts.DailyBestRankDayStat, int64, error)
	ListFetchRuns(ctx context.Context, page dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error)
	ListFetchLogs(ctx context.Context, filter dailybesttypes.FetchLogFilter) ([]dailybesttypes.FetchLog, int64, error)
	FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error)
	InterruptActiveFetches(ctx context.Context, finishedAt int64, message string) (int64, int64, error)
	Ping(ctx context.Context) error
}

type RuntimeRepository interface {
	CreateTargetProbe(ctx context.Context, draft dailybesttypes.TargetProbeDraft) (dailybesttypes.TargetProbe, error)
	CompleteTargetProbe(ctx context.Context, probeID int64, result dailybesttypes.TargetProbeResult) error
	InterruptRunningTargetProbes(ctx context.Context, finishedAt int64, message string) (int64, error)
	ListTargetProbes(ctx context.Context, filter dailybesttypes.TargetProbeFilter) ([]dailybesttypes.TargetProbe, int64, error)
	FindLatestTargetProbe(ctx context.Context) (dailybesttypes.TargetProbe, error)
	CreateRuntimeEvent(ctx context.Context, draft dailybesttypes.RuntimeEventDraft) (dailybesttypes.RuntimeEvent, error)
	ListRuntimeEvents(ctx context.Context, filter dailybesttypes.RuntimeEventFilter) ([]dailybesttypes.RuntimeEvent, int64, error)
}

type SchedulerRepository interface {
	FindAutomaticFetchState(ctx context.Context, businessDate string, category int64) (dailybesttypes.AutomaticFetchState, error)
}

type PageSource interface {
	FetchPage(ctx context.Context, fetchRunID, page int64) (dailybesttypes.SourcePage, error)
}

type AgentClient interface {
	Request(ctx context.Context, requestID, target, purpose string) (browseragenttypes.FetchResult, int, error)
	Status(ctx context.Context) (browseragenttypes.Status, error)
}

type Option func(*Service)

func WithSourceFactory(factory func(string) PageSource) Option {
	return func(service *Service) {
		service.sourceFactory = factory
	}
}

func WithAgent(agent AgentClient) Option {
	return func(service *Service) {
		service.agent = agent
	}
}

func WithNotifier(notifier notify.Notifier, detailBaseURL string) Option {
	return func(service *Service) {
		service.notifier = notifier
		service.detailBaseURL = detailBaseURL
	}
}

type Service struct {
	repository          Repository
	runtimeRepository   RuntimeRepository
	schedulerRepository SchedulerRepository
	source              PageSource
	sourceFactory       func(string) PageSource
	agent               AgentClient
	notifier            notify.Notifier
	detailBaseURL       string
	root                context.Context
	loc                 *time.Location

	runMu      sync.Mutex
	scheduleMu sync.Mutex
	runtimeMu  sync.Mutex
	wg         sync.WaitGroup

	targetStatus        string
	latestProbe         *dailybesttypes.TargetProbe
	nextProbeAt         int64
	lastProbeSuccess    int64
	lastProbeFailure    int64
	retryDelay          time.Duration
	retryWake           chan struct{}
	browserMonitorState string
}

func New(root context.Context, repository Repository, source PageSource, loc *time.Location, options ...Option) *Service {
	if root == nil {
		root = context.Background()
	}
	if loc == nil {
		loc = time.Local
	}
	service := &Service{
		repository: repository, source: source, root: root, loc: loc,
		targetStatus: dailybesttypes.TargetStatusUnknown, retryWake: make(chan struct{}, 1),
	}
	service.runtimeRepository, _ = repository.(RuntimeRepository)
	service.schedulerRepository, _ = repository.(SchedulerRepository)
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}
