package modelg

import (
	"fmt"

	"gorm.io/gorm"
)

// Existing daily buckets are additive. YEARWEEK(..., 3) matches Go's ISOWeek
// without depending on the database connection's timezone.
const releaseWeekBackfillSQL = `
INSERT INTO movie_release_bucket_stat
 (agg_mode, scope_key, level, year, quarter, month, day, count_all, count_owned,
  removed_count, size_bytes, latest_releasing_date, created_on, updated_on)
SELECT agg_mode,
 CONCAT('week:', FLOOR(iso_week / 100), '-W', LPAD(MOD(iso_week, 100), 2, '0')),
 'week', FLOOR(iso_week / 100), 0, 0, 0, SUM(count_all), SUM(count_owned),
 SUM(removed_count), SUM(size_bytes), MAX(latest_releasing_date), MIN(created_on), MAX(updated_on)
FROM (
 SELECT agg_mode, YEARWEEK(STR_TO_DATE(SUBSTRING(scope_key, 5), '%Y-%m-%d'), 3) AS iso_week,
  count_all, count_owned, removed_count, size_bytes, latest_releasing_date, created_on, updated_on
 FROM movie_release_bucket_stat WHERE level = 'day'
) daily
GROUP BY agg_mode, iso_week
ON DUPLICATE KEY UPDATE scope_key = VALUES(scope_key)`

const birthWeekBackfillSQL = `
INSERT INTO w_media_birth_bucket_stat
 (scope_key, level, year, quarter, month, day, media_count, removed_count, size_bytes,
  added_count, added_size_bytes, removed_size_bytes, net_size_bytes, has_sub_count,
  latest_birth_time, created_on, updated_on)
SELECT CONCAT('week:', FLOOR(iso_week / 100), '-W', LPAD(MOD(iso_week, 100), 2, '0')),
 'week', FLOOR(iso_week / 100), 0, 0, 0, SUM(media_count), SUM(removed_count), SUM(size_bytes),
 SUM(added_count), SUM(added_size_bytes), SUM(removed_size_bytes), SUM(net_size_bytes), SUM(has_sub_count),
 MAX(latest_birth_time), MIN(created_on), MAX(updated_on)
FROM (
 SELECT YEARWEEK(STR_TO_DATE(SUBSTRING(scope_key, 5), '%Y-%m-%d'), 3) AS iso_week,
  media_count, removed_count, size_bytes, added_count, added_size_bytes, removed_size_bytes,
  net_size_bytes, has_sub_count, latest_birth_time, created_on, updated_on
 FROM w_media_birth_bucket_stat WHERE level = 'day'
) daily
GROUP BY iso_week
ON DUPLICATE KEY UPDATE scope_key = VALUES(scope_key)`

func backfillCardPeriodWeeks(db *gorm.DB) error {
	// Preserve already rebuilt week rows; startup only fills missing periods.
	for _, query := range []string{releaseWeekBackfillSQL, birthWeekBackfillSQL} {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("回填卡片周周期失败: %w", err)
		}
	}
	return nil
}
