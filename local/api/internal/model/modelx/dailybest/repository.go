package dailybestlocalx

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	dailybesttypes "rudy_gc/local/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

const (
	SyncStatusImported     = "imported"
	SyncStatusFailed       = "failed"
	ProcessStatusPending   = "pending"
	ProcessStatusCompleted = "completed"
	ProcessStatusFailed    = "failed"
)

type Repository struct {
	conn sqlx.SqlConn
}

func New(conn sqlx.SqlConn) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) ListLocalRankCategoryStates(
	ctx context.Context,
	queries []dailybesttypes.LocalRankCategoryQuery,
) (map[dailybesttypes.RankCategoryKey]dailybesttypes.LocalRankCategoryState, error) {
	states := make(map[dailybesttypes.RankCategoryKey]dailybesttypes.LocalRankCategoryState, len(queries))
	if len(queries) == 0 {
		return states, nil
	}
	conditions := make(squirrel.Or, 0, len(queries))
	businessDates := make(map[dailybesttypes.RankCategoryKey]string, len(queries))
	for _, query := range queries {
		key := dailybesttypes.RankCategoryKey{DayNumber: query.DayNumber, Category: query.Category}
		businessDates[key] = query.BusinessDate
		conditions = append(conditions, squirrel.And{
			squirrel.Eq{"day_number": query.DayNumber},
			squirrel.Eq{"category": query.Category},
		})
	}
	type rankRow struct {
		MovieJavID string `db:"movie_jav_id"`
		DayNumber  int64  `db:"day_number"`
		RankPos    int64  `db:"rank_pos"`
		Category   int64  `db:"category"`
	}
	rankSQL, rankArgs, err := squirrel.Select("movie_jav_id", "day_number", "rank_pos", "category").
		From("c_rank").
		Where(conditions).
		OrderBy("day_number ASC", "category ASC", "rank_pos ASC", "movie_jav_id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	var ranks []rankRow
	err = r.conn.QueryRowsCtx(ctx, &ranks, rankSQL, rankArgs...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, err
	}
	ranksByKey := make(map[dailybesttypes.RankCategoryKey][]contracts.DailyBestRank, len(queries))
	for _, rank := range ranks {
		key := dailybesttypes.RankCategoryKey{DayNumber: rank.DayNumber, Category: rank.Category}
		ranksByKey[key] = append(ranksByKey[key], contracts.DailyBestRank{
			MovieJavID: rank.MovieJavID, DayNumber: rank.DayNumber,
			RankPos: rank.RankPos, Category: rank.Category,
		})
	}
	for key, checksumRanks := range ranksByKey {
		states[key] = dailybesttypes.LocalRankCategoryState{
			RankCount: int64(len(checksumRanks)),
			Checksum: contracts.ComputeDailyBestRankChecksum(
				businessDates[key], key.DayNumber, key.Category, checksumRanks,
			),
		}
	}
	type syncRow struct {
		DayNumber        int64  `db:"day_number"`
		Category         int64  `db:"category"`
		RemoteFetchRunID int64  `db:"remote_fetch_run_id"`
		SyncStatus       string `db:"sync_status"`
		ProcessStatus    string `db:"process_status"`
		SyncedAt         int64  `db:"synced_at"`
		ProcessedAt      int64  `db:"processed_at"`
		LastError        string `db:"last_error"`
	}
	syncSQL, syncArgs, err := squirrel.Select(
		"day_number", "category", "remote_fetch_run_id", "sync_status", "process_status",
		"synced_at", "processed_at", "last_error",
	).From("d_dailybest_sync_rank").Where(conditions).ToSql()
	if err != nil {
		return nil, err
	}
	var syncRows []syncRow
	if err := r.conn.QueryRowsCtx(ctx, &syncRows, syncSQL, syncArgs...); err != nil &&
		!errors.Is(err, sql.ErrNoRows) && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, err
	}
	for _, syncState := range syncRows {
		key := dailybesttypes.RankCategoryKey{DayNumber: syncState.DayNumber, Category: syncState.Category}
		state := states[key]
		state.RemoteFetchRunID = syncState.RemoteFetchRunID
		state.SyncStatus = syncState.SyncStatus
		state.ProcessStatus = syncState.ProcessStatus
		state.SyncedAt = syncState.SyncedAt
		state.ProcessedAt = syncState.ProcessedAt
		state.LastError = syncState.LastError
		states[key] = state
	}
	return states, nil
}

func (r *Repository) SaveImported(ctx context.Context, manifest dailybesttypes.SyncManifest) error {
	now := time.Now().Unix()
	_, err := r.conn.ExecCtx(ctx, `
INSERT INTO d_dailybest_sync_rank (
	remote_fetch_run_id, business_date, day_number, category, page_count, item_count, checksum,
	sync_status, process_status, remote_finished_at, synced_at, processed_at, last_error, created_on, updated_on
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, '', ?, ?)
ON DUPLICATE KEY UPDATE
	remote_fetch_run_id = VALUES(remote_fetch_run_id),
	business_date = VALUES(business_date),
	page_count = VALUES(page_count),
	item_count = VALUES(item_count),
	checksum = VALUES(checksum),
	sync_status = VALUES(sync_status),
	process_status = VALUES(process_status),
	remote_finished_at = VALUES(remote_finished_at),
	synced_at = VALUES(synced_at),
	processed_at = 0,
	last_error = '',
	updated_on = VALUES(updated_on)`,
		manifest.FetchRunID, manifest.BusinessDate, manifest.DayNumber, manifest.Category,
		manifest.PageCount, manifest.ItemCount, manifest.Checksum,
		SyncStatusImported, ProcessStatusPending, manifest.FinishedAt, now, now, now)
	return err
}

func (r *Repository) MarkProcessed(ctx context.Context, manifests []dailybesttypes.SyncManifest) error {
	if len(manifests) == 0 {
		return nil
	}
	now := time.Now().Unix()
	return r.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		for _, manifest := range manifests {
			if _, err := session.ExecCtx(ctx, `
UPDATE d_dailybest_sync_rank
SET process_status = ?, processed_at = ?, last_error = '', updated_on = ?
WHERE remote_fetch_run_id = ?`, ProcessStatusCompleted, now, now, manifest.FetchRunID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) MarkFailed(ctx context.Context, manifest dailybesttypes.SyncManifest, message string) error {
	now := time.Now().Unix()
	_, err := r.conn.ExecCtx(ctx, `
UPDATE d_dailybest_sync_rank
SET sync_status = ?, process_status = ?, last_error = ?, updated_on = ?
WHERE remote_fetch_run_id = ?`, SyncStatusFailed, ProcessStatusFailed, message, now, manifest.FetchRunID)
	return err
}
