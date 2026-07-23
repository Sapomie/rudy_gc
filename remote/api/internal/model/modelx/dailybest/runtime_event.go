package dailybestremotex

import (
	"context"
	"time"

	"rudy_gc/remote/api/internal/model/modelg"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (r *Repository) CreateRuntimeEvent(ctx context.Context, draft dailybesttypes.RuntimeEventDraft) (dailybesttypes.RuntimeEvent, error) {
	now := time.Now().Unix()
	row := &modelg.DailyBestRuntimeEvent{
		Component: draft.Component, EventType: draft.EventType, Level: draft.Level, Message: draft.Message,
		FetchRunID: draft.FetchRunID, TargetProbeID: draft.TargetProbeID, OccurredAt: draft.OccurredAt, CreatedOn: now,
	}
	if row.OccurredAt == 0 {
		row.OccurredAt = now
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return dailybesttypes.RuntimeEvent{}, err
	}
	return mapRuntimeEvent(row), nil
}

func (r *Repository) ListRuntimeEvents(ctx context.Context, filter dailybesttypes.RuntimeEventFilter) ([]dailybesttypes.RuntimeEvent, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.DailyBestRuntimeEvent{})
	if filter.Component != "" {
		query = query.Where("component = ?", filter.Component)
	}
	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []modelg.DailyBestRuntimeEvent
	if err := paginate(query, filter.PageQuery).Order("occurred_at DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]dailybesttypes.RuntimeEvent, 0, len(rows))
	for index := range rows {
		result = append(result, mapRuntimeEvent(&rows[index]))
	}
	return result, total, nil
}

func mapRuntimeEvent(row *modelg.DailyBestRuntimeEvent) dailybesttypes.RuntimeEvent {
	return dailybesttypes.RuntimeEvent{
		ID: row.ID, Component: row.Component, EventType: row.EventType, Level: row.Level, Message: row.Message,
		FetchRunID: row.FetchRunID, TargetProbeID: row.TargetProbeID, OccurredAt: row.OccurredAt, CreatedOn: row.CreatedOn,
	}
}
