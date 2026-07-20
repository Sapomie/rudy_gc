package dailybestremotex

import (
	"context"
	"time"

	"rudy_gc/remote/api/internal/model/modelg"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (r *Repository) CreateFetchRun(ctx context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error) {
	now := time.Now().Unix()
	row := &modelg.DailyBestFetchRun{
		BusinessDate: businessDate,
		Category:     category,
		TriggerType:  triggerType,
		Status:       dailybesttypes.RunStatusQueued,
		ErrorMessage: "",
		CreatedOn:    now,
		UpdatedOn:    now,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	return mapFetchRun(row), nil
}

func (r *Repository) MarkFetchRunRunning(ctx context.Context, runID int64, startedAt int64) error {
	return r.db.WithContext(ctx).Model(&modelg.DailyBestFetchRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status":        dailybesttypes.RunStatusRunning,
		"started_at":    startedAt,
		"error_message": "",
		"updated_on":    startedAt,
	}).Error
}

func (r *Repository) MarkFetchRunFailed(ctx context.Context, runID int64, finishedAt int64, message string) error {
	return r.db.WithContext(ctx).Model(&modelg.DailyBestFetchRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status":        dailybesttypes.RunStatusFailed,
		"finished_at":   finishedAt,
		"error_message": message,
		"updated_on":    finishedAt,
	}).Error
}

func (r *Repository) ListFetchRuns(ctx context.Context, page dailybesttypes.PageQuery) ([]contracts.DailyBestFetchRun, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.DailyBestFetchRun{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page.Enabled {
		query = query.Offset(page.Offset()).Limit(page.Limit())
	}
	var rows []modelg.DailyBestFetchRun
	if err := query.Order("id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]contracts.DailyBestFetchRun, 0, len(rows))
	for i := range rows {
		out = append(out, mapFetchRun(&rows[i]))
	}
	return out, total, nil
}

func (r *Repository) FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error) {
	var row modelg.DailyBestFetchRun
	if err := r.db.WithContext(ctx).First(&row, runID).Error; err != nil {
		return contracts.DailyBestFetchRun{}, normalizeError(err)
	}
	return mapFetchRun(&row), nil
}

func mapFetchRun(row *modelg.DailyBestFetchRun) contracts.DailyBestFetchRun {
	if row == nil {
		return contracts.DailyBestFetchRun{}
	}
	return contracts.DailyBestFetchRun{
		ID:           row.ID,
		BusinessDate: row.BusinessDate,
		Category:     row.Category,
		TriggerType:  row.TriggerType,
		Status:       row.Status,
		StartedAt:    row.StartedAt,
		FinishedAt:   row.FinishedAt,
		PageCount:    row.PageCount,
		ItemCount:    row.ItemCount,
		Checksum:     row.Checksum,
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedOn,
		UpdatedAt:    row.UpdatedOn,
	}
}
