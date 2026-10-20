package dailybestsync

import (
	"context"
	"fmt"
	"sort"

	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) Sync(
	ctx context.Context,
	targets []dailybesttypes.SyncTarget,
	apply func(context.Context, Batch) error,
) (*Result, error) {
	targets, err := normalizeSyncTargets(targets)
	if err != nil {
		return nil, err
	}
	result := &Result{Batches: make([]Batch, 0, len(targets))}
	remoteDays := make(map[int64]contracts.DailyBestRankDay)
	type resolvedTarget struct {
		category contracts.DailyBestRankDayCategory
		manifest dailybesttypes.SyncManifest
		key      dailybesttypes.RankCategoryKey
	}
	resolved := make([]resolvedTarget, 0, len(targets))
	localQueries := make([]dailybesttypes.LocalRankCategoryQuery, 0, len(targets))
	for _, target := range targets {
		remoteDay, ok := remoteDays[target.DayNumber]
		if !ok {
			list, err := s.client.ListRankDays(ctx, dailybesttypes.RankDayQuery{DayNumber: target.DayNumber})
			if err != nil {
				return nil, err
			}
			remoteDay, err = findRemoteDay(list.Days, target.DayNumber)
			if err != nil {
				return nil, err
			}
			remoteDays[target.DayNumber] = remoteDay
		}
		category, err := findRemoteCategory(remoteDay, target.Category)
		if err != nil {
			return nil, err
		}
		manifest := dailybesttypes.SyncManifest{
			FetchRunID: category.FetchRunID, BusinessDate: remoteDay.BusinessDate, DayNumber: remoteDay.DayNumber,
			Category: category.Category, PageCount: category.PageCount, ItemCount: category.RankCount,
			Checksum: category.Checksum, FinishedAt: category.FinishedAt,
		}
		key := dailybesttypes.RankCategoryKey{DayNumber: remoteDay.DayNumber, Category: category.Category}
		resolved = append(resolved, resolvedTarget{category: category, manifest: manifest, key: key})
		localQueries = append(localQueries, dailybesttypes.LocalRankCategoryQuery{
			BusinessDate: remoteDay.BusinessDate, DayNumber: remoteDay.DayNumber, Category: category.Category,
		})
	}
	localStates, err := s.store.ListLocalRankCategoryStates(ctx, localQueries)
	if err != nil {
		return nil, fmt.Errorf("查询 DailyBest 本地日期状态失败: %w", err)
	}
	for _, target := range resolved {
		_, needsSync := rankCategoryStatus(target.category, localStates[target.key])
		if !needsSync {
			continue
		}
		manifest := target.manifest
		response, err := s.client.ListRanks(ctx, manifest.DayNumber, manifest.Category)
		if err != nil {
			return nil, err
		}
		batch := Batch{Manifest: manifest, Ranks: response.Ranks}
		if err := validateBatch(batch); err != nil {
			return nil, err
		}
		if err := apply(ctx, batch); err != nil {
			return nil, fmt.Errorf("导入 DailyBest fetch_run_id=%d 失败: %w", manifest.FetchRunID, err)
		}
		if err := s.store.SaveImported(ctx, manifest); err != nil {
			return nil, fmt.Errorf("保存 DailyBest fetch_run_id=%d 同步状态失败: %w", manifest.FetchRunID, err)
		}
		result.Batches = append(result.Batches, batch)
	}
	return result, nil
}

func (s *Service) MarkProcessed(ctx context.Context, result *Result) error {
	if result == nil {
		return nil
	}
	manifests := make([]dailybesttypes.SyncManifest, 0, len(result.Batches))
	for _, batch := range result.Batches {
		manifests = append(manifests, batch.Manifest)
	}
	return s.store.MarkProcessed(ctx, manifests)
}

func normalizeSyncTargets(targets []dailybesttypes.SyncTarget) ([]dailybesttypes.SyncTarget, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("请选择需要同步的 DailyBest 日期")
	}
	type key struct {
		DayNumber int64
		Category  int64
	}
	seen := make(map[key]struct{}, len(targets))
	out := make([]dailybesttypes.SyncTarget, 0, len(targets))
	for _, target := range targets {
		if target.DayNumber <= 0 || target.Category <= 0 {
			return nil, fmt.Errorf("DailyBest 同步日期和分类必须为正整数")
		}
		key := key{DayNumber: target.DayNumber, Category: target.Category}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, target)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].DayNumber != out[j].DayNumber {
			return out[i].DayNumber < out[j].DayNumber
		}
		return out[i].Category < out[j].Category
	})
	return out, nil
}

func findRemoteDay(days []contracts.DailyBestRankDay, dayNumber int64) (contracts.DailyBestRankDay, error) {
	for _, day := range days {
		if day.DayNumber == dayNumber {
			return day, nil
		}
	}
	return contracts.DailyBestRankDay{}, fmt.Errorf("Remote DailyBest day_number=%d 排名不存在", dayNumber)
}

func findRemoteCategory(day contracts.DailyBestRankDay, category int64) (contracts.DailyBestRankDayCategory, error) {
	for _, current := range day.Categories {
		if current.Category == category {
			return current, nil
		}
	}
	return contracts.DailyBestRankDayCategory{}, fmt.Errorf(
		"Remote DailyBest day_number=%d category=%d 排名不存在", day.DayNumber, category,
	)
}

func (s *Service) MarkFailed(ctx context.Context, result *Result, processErr error) {
	if result == nil || processErr == nil {
		return
	}
	for _, batch := range result.Batches {
		_ = s.store.MarkFailed(context.WithoutCancel(ctx), batch.Manifest, processErr.Error())
	}
}

func validateBatch(batch Batch) error {
	manifest := batch.Manifest
	dayNumber, err := contracts.RankDayNumber(manifest.BusinessDate)
	if err != nil || dayNumber != manifest.DayNumber {
		return fmt.Errorf("DailyBest fetch_run_id=%d 日期序号校验失败", manifest.FetchRunID)
	}
	if int64(len(batch.Ranks)) != manifest.ItemCount {
		return fmt.Errorf("DailyBest fetch_run_id=%d 排名数量校验失败", manifest.FetchRunID)
	}
	for _, rank := range batch.Ranks {
		if rank.DayNumber != manifest.DayNumber || rank.Category != manifest.Category || rank.RankPos <= 0 || rank.MovieJavID == "" {
			return fmt.Errorf("DailyBest fetch_run_id=%d 排名明细校验失败", manifest.FetchRunID)
		}
	}
	if checksum := contracts.ComputeDailyBestRankChecksum(manifest.BusinessDate, manifest.DayNumber, manifest.Category, batch.Ranks); checksum != manifest.Checksum {
		return fmt.Errorf("DailyBest fetch_run_id=%d checksum 校验失败", manifest.FetchRunID)
	}
	return nil
}
