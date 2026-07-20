package dailybestremotex

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rudy_gc/remote/api/internal/modules/dailybest/domain"
	contracts "rudy_gc/shared/dailybest"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ dailybest.Repository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateFetchRun(ctx context.Context, businessDate string, category int64, triggerType string) (contracts.DailyBestFetchRun, error) {
	now := time.Now().Unix()
	row := &FetchRun{
		BusinessDate: businessDate,
		Category:     category,
		TriggerType:  triggerType,
		Status:       dailybest.RunStatusQueued,
		ErrorMessage: "",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	return mapFetchRun(row), nil
}

func (r *Repository) MarkFetchRunRunning(ctx context.Context, runID int64, startedAt int64) error {
	return r.db.WithContext(ctx).Model(&FetchRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status":        dailybest.RunStatusRunning,
		"started_at":    startedAt,
		"error_message": "",
		"updated_at":    startedAt,
	}).Error
}

func (r *Repository) MarkFetchRunFailed(ctx context.Context, runID int64, finishedAt int64, message string) error {
	return r.db.WithContext(ctx).Model(&FetchRun{}).Where("id = ?", runID).Updates(map[string]any{
		"status":        dailybest.RunStatusFailed,
		"finished_at":   finishedAt,
		"error_message": message,
		"updated_at":    finishedAt,
	}).Error
}

func (r *Repository) PublishSnapshot(ctx context.Context, runID int64, draft dailybest.SnapshotDraft, finishedAt int64) (dailybest.PublishResult, error) {
	var (
		publishedID int64
		created     bool
	)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing Snapshot
		err := tx.Where(
			"business_date = ? AND category = ? AND checksum = ?",
			draft.BusinessDate,
			draft.Category,
			draft.Checksum,
		).First(&existing).Error
		if err == nil {
			publishedID = existing.ID
			return tx.Model(&FetchRun{}).Where("id = ?", runID).Updates(map[string]any{
				"status":        dailybest.RunStatusNoChange,
				"finished_at":   finishedAt,
				"page_count":    len(draft.Pages),
				"item_count":    len(draft.Items),
				"snapshot_id":   existing.ID,
				"error_message": "",
				"updated_at":    finishedAt,
			}).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		snapshot := &Snapshot{
			FetchRunID:   runID,
			BusinessDate: draft.BusinessDate,
			Category:     draft.Category,
			PageCount:    int64(len(draft.Pages)),
			ItemCount:    int64(len(draft.Items)),
			Checksum:     draft.Checksum,
			PublishedAt:  finishedAt,
			CreatedAt:    finishedAt,
		}
		if err := tx.Create(snapshot).Error; err != nil {
			return err
		}
		publishedID = snapshot.ID
		created = true

		for _, page := range draft.Pages {
			row := &SnapshotPage{
				SnapshotID: snapshot.ID,
				PageNumber: page.PageNumber,
				SourceURL:  page.SourceURL,
				Content:    page.Content,
				Checksum:   page.Checksum,
				CreatedAt:  finishedAt,
			}
			if err := tx.Create(row).Error; err != nil {
				return fmt.Errorf("保存 DailyBest 快照页 %d 失败: %w", page.PageNumber, err)
			}
		}
		for _, item := range draft.Items {
			row := &SnapshotItem{
				SnapshotID:  snapshot.ID,
				RankPos:     item.RankPos,
				SourcePage:  item.SourcePage,
				SourceIndex: item.SourceIndex,
				JavID:       item.JavID,
				MovieName:   item.MovieName,
				MovieCode:   item.MovieCode,
				Title:       item.Title,
				CoverURL:    item.CoverURL,
				CreatedAt:   finishedAt,
			}
			if err := tx.Create(row).Error; err != nil {
				return fmt.Errorf("保存 DailyBest 快照排名 %d 失败: %w", item.RankPos, err)
			}
		}
		return tx.Model(&FetchRun{}).Where("id = ?", runID).Updates(map[string]any{
			"status":        dailybest.RunStatusSucceeded,
			"finished_at":   finishedAt,
			"page_count":    len(draft.Pages),
			"item_count":    len(draft.Items),
			"snapshot_id":   snapshot.ID,
			"error_message": "",
			"updated_at":    finishedAt,
		}).Error
	})
	if err != nil {
		return dailybest.PublishResult{}, err
	}

	snapshot, err := r.FindSnapshot(ctx, publishedID, true)
	if err != nil {
		return dailybest.PublishResult{}, err
	}
	return dailybest.PublishResult{Snapshot: snapshot, Created: created}, nil
}

func (r *Repository) ListSnapshotsAfter(ctx context.Context, revision int64) ([]contracts.DailyBestSnapshotManifest, error) {
	var rows []Snapshot
	if err := r.db.WithContext(ctx).Where("id > ?", revision).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]contracts.DailyBestSnapshotManifest, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapManifest(row))
	}
	return out, nil
}

func (r *Repository) FindSnapshot(ctx context.Context, revision int64, includePages bool) (contracts.DailyBestSnapshotResponse, error) {
	var row Snapshot
	if err := r.db.WithContext(ctx).First(&row, revision).Error; err != nil {
		return contracts.DailyBestSnapshotResponse{}, err
	}
	return r.loadSnapshot(ctx, row, includePages)
}

func (r *Repository) FindLatestSnapshotByDate(ctx context.Context, businessDate string, category int64, includePages bool) (contracts.DailyBestSnapshotResponse, error) {
	var row Snapshot
	if err := r.db.WithContext(ctx).
		Where("business_date = ? AND category = ?", businessDate, category).
		Order("id DESC").
		First(&row).Error; err != nil {
		return contracts.DailyBestSnapshotResponse{}, err
	}
	return r.loadSnapshot(ctx, row, includePages)
}

func (r *Repository) loadSnapshot(ctx context.Context, row Snapshot, includePages bool) (contracts.DailyBestSnapshotResponse, error) {
	var itemRows []SnapshotItem
	if err := r.db.WithContext(ctx).Where("snapshot_id = ?", row.ID).Order("rank_pos ASC").Find(&itemRows).Error; err != nil {
		return contracts.DailyBestSnapshotResponse{}, err
	}
	resp := contracts.DailyBestSnapshotResponse{
		DailyBestSnapshotManifest: mapManifest(row),
		Items:                     make([]contracts.DailyBestSnapshotItem, 0, len(itemRows)),
	}
	for _, item := range itemRows {
		resp.Items = append(resp.Items, contracts.DailyBestSnapshotItem{
			RankPos:     item.RankPos,
			SourcePage:  item.SourcePage,
			SourceIndex: item.SourceIndex,
			JavID:       item.JavID,
			MovieName:   item.MovieName,
			MovieCode:   item.MovieCode,
			Title:       item.Title,
			CoverURL:    item.CoverURL,
		})
	}
	if includePages {
		var pageRows []SnapshotPage
		if err := r.db.WithContext(ctx).Where("snapshot_id = ?", row.ID).Order("page_number ASC").Find(&pageRows).Error; err != nil {
			return contracts.DailyBestSnapshotResponse{}, err
		}
		resp.Pages = make([]contracts.DailyBestSnapshotPage, 0, len(pageRows))
		for _, page := range pageRows {
			resp.Pages = append(resp.Pages, contracts.DailyBestSnapshotPage{
				PageNumber: page.PageNumber,
				SourceURL:  page.SourceURL,
				Content:    page.Content,
				Checksum:   page.Checksum,
			})
		}
	}
	return resp, nil
}

func (r *Repository) ListFetchRuns(ctx context.Context) ([]contracts.DailyBestFetchRun, error) {
	var rows []FetchRun
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]contracts.DailyBestFetchRun, 0, len(rows))
	for i := range rows {
		out = append(out, mapFetchRun(&rows[i]))
	}
	return out, nil
}

func (r *Repository) FindFetchRun(ctx context.Context, runID int64) (contracts.DailyBestFetchRun, error) {
	var row FetchRun
	if err := r.db.WithContext(ctx).First(&row, runID).Error; err != nil {
		return contracts.DailyBestFetchRun{}, err
	}
	return mapFetchRun(&row), nil
}

func (r *Repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func mapManifest(row Snapshot) contracts.DailyBestSnapshotManifest {
	return contracts.DailyBestSnapshotManifest{
		Revision:     row.ID,
		BusinessDate: row.BusinessDate,
		Category:     row.Category,
		PageCount:    row.PageCount,
		ItemCount:    row.ItemCount,
		Checksum:     row.Checksum,
		PublishedAt:  row.PublishedAt,
	}
}

func mapFetchRun(row *FetchRun) contracts.DailyBestFetchRun {
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
		SnapshotID:   row.SnapshotID,
		ErrorMessage: row.ErrorMessage,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
