package dailybestsync

import (
	"context"
	"fmt"

	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
)

func (s *Service) ListRankDays(
	ctx context.Context,
	query dailybesttypes.RankDayQuery,
) (dailybesttypes.RankDayListResponse, error) {
	remote, err := s.client.ListRankDays(ctx, query)
	if err != nil {
		return dailybesttypes.RankDayListResponse{}, err
	}
	response := dailybesttypes.RankDayListResponse{
		Days:     make([]dailybesttypes.RankDayRow, 0, len(remote.Days)),
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	if remote.PageInfo != nil {
		response.Total = remote.PageInfo.Total
		response.Page = remote.PageInfo.Page
		response.PageSize = remote.PageInfo.PageSize
	} else {
		response.Total = int64(len(remote.Days))
	}
	localQueries := make([]dailybesttypes.LocalRankCategoryQuery, 0)
	for _, remoteDay := range remote.Days {
		for _, remoteCategory := range remoteDay.Categories {
			localQueries = append(localQueries, dailybesttypes.LocalRankCategoryQuery{
				BusinessDate: remoteDay.BusinessDate,
				DayNumber:    remoteDay.DayNumber,
				Category:     remoteCategory.Category,
			})
		}
	}
	localStates, err := s.store.ListLocalRankCategoryStates(ctx, localQueries)
	if err != nil {
		return dailybesttypes.RankDayListResponse{}, fmt.Errorf("查询 DailyBest 本地日期状态失败: %w", err)
	}
	for _, remoteDay := range remote.Days {
		row := dailybesttypes.RankDayRow{
			BusinessDate:    remoteDay.BusinessDate,
			DayNumber:       remoteDay.DayNumber,
			PageCount:       remoteDay.PageCount,
			ItemCount:       remoteDay.ItemCount,
			RemoteRankCount: remoteDay.RankCount,
			FinishedAt:      remoteDay.FinishedAt,
			Status:          dailybesttypes.RankDayStatusCompleted,
			Categories:      make([]dailybesttypes.RankDayCategoryRow, 0, len(remoteDay.Categories)),
		}
		for _, remoteCategory := range remoteDay.Categories {
			key := dailybesttypes.RankCategoryKey{DayNumber: remoteDay.DayNumber, Category: remoteCategory.Category}
			localState := localStates[key]
			status, needsSync := rankCategoryStatus(remoteCategory, localState)
			row.LocalRankCount += localState.RankCount
			row.NeedsSync = row.NeedsSync || needsSync
			row.Status = higherPriorityRankDayStatus(row.Status, status)
			row.Categories = append(row.Categories, dailybesttypes.RankDayCategoryRow{
				Remote: remoteCategory, Local: localState, Status: status, NeedsSync: needsSync,
			})
		}
		response.Days = append(response.Days, row)
	}
	return response, nil
}
