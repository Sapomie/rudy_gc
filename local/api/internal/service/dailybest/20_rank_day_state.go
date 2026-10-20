package dailybestsync

import (
	dailybestlocalx "rudy_gc/local/api/internal/model/modelx/dailybest"
	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func rankCategoryStatus(
	remote contracts.DailyBestRankDayCategory,
	local dailybesttypes.LocalRankCategoryState,
) (string, bool) {
	if local.RankCount > 0 {
		if local.RankCount != remote.RankCount {
			return dailybesttypes.RankDayStatusIncomplete, false
		}
		if local.Checksum != remote.Checksum {
			return dailybesttypes.RankDayStatusVersionMismatch, false
		}
		if local.SyncStatus == dailybestlocalx.SyncStatusFailed || local.ProcessStatus == dailybestlocalx.ProcessStatusFailed {
			return dailybesttypes.RankDayStatusFailed, false
		}
		if local.SyncStatus == dailybestlocalx.SyncStatusImported && local.ProcessStatus != dailybestlocalx.ProcessStatusCompleted {
			return dailybesttypes.RankDayStatusPending, false
		}
		return dailybesttypes.RankDayStatusCompleted, false
	}
	if local.SyncStatus == dailybestlocalx.SyncStatusFailed || local.ProcessStatus == dailybestlocalx.ProcessStatusFailed {
		return dailybesttypes.RankDayStatusFailed, true
	}
	return dailybesttypes.RankDayStatusMissing, true
}

func higherPriorityRankDayStatus(current, candidate string) string {
	priority := map[string]int{
		dailybesttypes.RankDayStatusCompleted:       1,
		dailybesttypes.RankDayStatusPending:         2,
		dailybesttypes.RankDayStatusVersionMismatch: 3,
		dailybesttypes.RankDayStatusIncomplete:      4,
		dailybesttypes.RankDayStatusMissing:         5,
		dailybesttypes.RankDayStatusFailed:          6,
	}
	if priority[candidate] > priority[current] {
		return candidate
	}
	return current
}
