package dailybestremotex

import (
	"context"
	"fmt"
	"time"

	"rudy_gc/remote/api/internal/model/modelg"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (r *Repository) CreateTargetProbe(ctx context.Context, draft dailybesttypes.TargetProbeDraft) (dailybesttypes.TargetProbe, error) {
	now := time.Now().Unix()
	row := &modelg.DailyBestTargetProbe{
		RequestID: draft.RequestID, Status: dailybesttypes.TargetProbeStatusRunning, TargetURL: draft.TargetURL,
		FinalURL: "", ErrorMessage: "", CreatedOn: now, UpdatedOn: now,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return dailybesttypes.TargetProbe{}, err
	}
	return mapTargetProbe(row), nil
}

func (r *Repository) CompleteTargetProbe(ctx context.Context, probeID int64, result dailybesttypes.TargetProbeResult) error {
	update := r.db.WithContext(ctx).Model(&modelg.DailyBestTargetProbe{}).
		Where("id = ? AND status = ?", probeID, dailybesttypes.TargetProbeStatusRunning).
		Updates(map[string]any{
			"status": result.Status, "final_url": result.FinalURL, "http_status": result.HTTPStatus,
			"duration_ms": result.DurationMS, "browser_status": result.BrowserStatus,
			"error_type": result.ErrorType, "error_message": result.ErrorMessage,
			"started_at": result.StartedAt, "finished_at": result.FinishedAt, "updated_on": time.Now().Unix(),
		})
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected != 1 {
		return fmt.Errorf("DailyBest 目标站探测记录 %d 不存在或已经完成", probeID)
	}
	return nil
}

func (r *Repository) InterruptRunningTargetProbes(ctx context.Context, finishedAt int64, message string) (int64, error) {
	update := r.db.WithContext(ctx).Model(&modelg.DailyBestTargetProbe{}).
		Where("status = ?", dailybesttypes.TargetProbeStatusRunning).
		Updates(map[string]any{
			"status": dailybesttypes.TargetProbeStatusInterrupted, "error_type": "api_restarted",
			"error_message": message, "finished_at": finishedAt, "updated_on": finishedAt,
		})
	return update.RowsAffected, update.Error
}

func (r *Repository) ListTargetProbes(ctx context.Context, filter dailybesttypes.TargetProbeFilter) ([]dailybesttypes.TargetProbe, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.DailyBestTargetProbe{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartedFrom > 0 {
		query = query.Where("started_at >= ?", filter.StartedFrom)
	}
	if filter.StartedTo > 0 {
		query = query.Where("started_at <= ?", filter.StartedTo)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []modelg.DailyBestTargetProbe
	if err := paginate(query, filter.PageQuery).Order("created_on DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	result := make([]dailybesttypes.TargetProbe, 0, len(rows))
	for index := range rows {
		result = append(result, mapTargetProbe(&rows[index]))
	}
	return result, total, nil
}

func (r *Repository) FindLatestTargetProbe(ctx context.Context) (dailybesttypes.TargetProbe, error) {
	var row modelg.DailyBestTargetProbe
	if err := r.db.WithContext(ctx).Order("created_on DESC, id DESC").First(&row).Error; err != nil {
		return dailybesttypes.TargetProbe{}, normalizeError(err)
	}
	return mapTargetProbe(&row), nil
}

func mapTargetProbe(row *modelg.DailyBestTargetProbe) dailybesttypes.TargetProbe {
	return dailybesttypes.TargetProbe{
		ID: row.ID, RequestID: row.RequestID, Status: row.Status, TargetURL: row.TargetURL,
		FinalURL: row.FinalURL, HTTPStatus: row.HTTPStatus, DurationMS: row.DurationMS,
		BrowserStatus: row.BrowserStatus, ErrorType: row.ErrorType, ErrorMessage: row.ErrorMessage,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
	}
}
