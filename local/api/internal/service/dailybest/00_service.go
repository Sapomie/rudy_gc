package dailybestsync

import (
	"context"

	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

type Batch struct {
	Manifest dailybesttypes.SyncManifest
	Ranks    []contracts.DailyBestRank
}

type Result struct {
	Batches []Batch
}

type RemoteClient interface {
	ListRankDays(ctx context.Context, filter dailybesttypes.RankDayQuery) (contracts.DailyBestRankDayListResponse, error)
	ListRanks(ctx context.Context, dayNumber, category int64) (contracts.DailyBestRankListResponse, error)
}

type RankStore interface {
	ListLocalRankCategoryStates(ctx context.Context, queries []dailybesttypes.LocalRankCategoryQuery) (map[dailybesttypes.RankCategoryKey]dailybesttypes.LocalRankCategoryState, error)
	SaveImported(ctx context.Context, manifest dailybesttypes.SyncManifest) error
	MarkProcessed(ctx context.Context, manifests []dailybesttypes.SyncManifest) error
	MarkFailed(ctx context.Context, manifest dailybesttypes.SyncManifest, message string) error
}

type Service struct {
	client RemoteClient
	store  RankStore
}

func New(client RemoteClient, store RankStore) *Service {
	return &Service{client: client, store: store}
}
