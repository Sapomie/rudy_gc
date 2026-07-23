package dailybestremotex

import (
	"context"
	"fmt"
	"time"

	"rudy_gc/remote/api/internal/model/modelg"
	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

func (r *Repository) CreateFetchLog(ctx context.Context, draft dailybesttypes.FetchLogDraft) (int64, error) {
	now := time.Now().Unix()
	row := &modelg.DailyBestFetchLog{
		FetchRunID:    draft.FetchRunID,
		PageNumber:    draft.PageNumber,
		AttemptNumber: draft.AttemptNumber,
		RequestURL:    draft.RequestURL,
		Status:        dailybesttypes.FetchLogStatusRequesting,
		ErrorMessage:  "",
		CreatedOn:     now,
		UpdatedOn:     now,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *Repository) CompleteFetchLog(ctx context.Context, logID int64, result dailybesttypes.FetchLogResult) error {
	now := time.Now().Unix()
	update := r.db.WithContext(ctx).Model(&modelg.DailyBestFetchLog{}).Where("id = ?", logID).Updates(map[string]any{
		"status":        result.Status,
		"http_status":   result.HTTPStatus,
		"duration_ms":   result.DurationMS,
		"error_message": result.ErrorMessage,
		"updated_on":    now,
	})
	if update.Error != nil {
		return update.Error
	}
	if update.RowsAffected != 1 {
		return fmt.Errorf("DailyBest 抓取请求日志 %d 不存在", logID)
	}
	return nil
}

func (r *Repository) ListFetchLogs(ctx context.Context, filter dailybesttypes.FetchLogFilter) ([]dailybesttypes.FetchLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&modelg.DailyBestFetchLog{})
	if filter.FetchRunID > 0 {
		query = query.Where("fetch_run_id = ?", filter.FetchRunID)
	}
	if filter.PageNumber > 0 {
		query = query.Where("page_number = ?", filter.PageNumber)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []modelg.DailyBestFetchLog
	if err := paginate(query, filter.PageQuery).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	logs := make([]dailybesttypes.FetchLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, dailybesttypes.FetchLog{
			ID: row.ID, FetchRunID: row.FetchRunID, PageNumber: row.PageNumber,
			AttemptNumber: row.AttemptNumber, RequestURL: row.RequestURL, Status: row.Status,
			HTTPStatus: row.HTTPStatus, DurationMS: row.DurationMS, ErrorMessage: row.ErrorMessage,
			CreatedOn: row.CreatedOn, UpdatedOn: row.UpdatedOn,
		})
	}
	return logs, total, nil
}
