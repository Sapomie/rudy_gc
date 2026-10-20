package modelg

import (
	"fmt"

	"gorm.io/gorm"
)

func migrateRankKeysIncludeCategory(db *gorm.DB) error {
	if !db.Migrator().HasTable(RankTableName) {
		return nil
	}

	return db.Exec(`
UPDATE c_rank
SET rank_key = CONCAT(
    SUBSTRING_INDEX(rank_key, '_', 1),
    '_', category,
    '_', LPAD(rank_pos, 3, '0')
)
WHERE rank_key NOT REGEXP '^[^_]+_[0-9]+_[0-9]+$'
`).Error
}

func migrateResetLegacyRankPeriodStat(db *gorm.DB) error {
	if !db.Migrator().HasTable(RankPeriodStatTableName) {
		return nil
	}

	migrator := db.Migrator()
	legacyColumns := []string{"period_id", "prev_period_id"}
	for _, column := range legacyColumns {
		if migrator.HasColumn(RankPeriodStatTableName, column) {
			return migrator.DropTable(RankPeriodStatTableName)
		}
	}

	requiredColumns := []string{"period_type", "period_key", "start_day_number", "end_day_number"}
	for _, column := range requiredColumns {
		if !migrator.HasColumn(RankPeriodStatTableName, column) {
			return fmt.Errorf("%s 缺少必要字段 %s，停止自动迁移以保护现有数据", RankPeriodStatTableName, column)
		}
	}
	return nil
}

func migrateDropRankPeriodStatRedundantColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(RankPeriodStatTableName) {
		return nil
	}

	columns := []string{
		"continued_movie_count",
		"first_seen_movie_count",
		"returning_movie_count",
		"dropped_movie_count",
		"rise_movie_count",
		"fall_movie_count",
		"unchanged_movie_count",
		"peak_movie_count",
		"rank_one_movie_jav_id",
		"average_score",
		"average_days_in_rank",
	}
	for _, column := range columns {
		if !db.Migrator().HasColumn(RankPeriodStatTableName, column) {
			continue
		}
		if err := db.Migrator().DropColumn(RankPeriodStatTableName, column); err != nil {
			return fmt.Errorf("删除 %s.%s 失败: %w", RankPeriodStatTableName, column, err)
		}
	}
	return nil
}

func migrateDropRankDayStatRedundantColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(RankDayStatTableName) {
		return nil
	}

	columns := []string{"rise_movie_count", "fall_movie_count", "unchanged_movie_count"}
	for _, column := range columns {
		if !db.Migrator().HasColumn(RankDayStatTableName, column) {
			continue
		}
		if err := db.Migrator().DropColumn(RankDayStatTableName, column); err != nil {
			return fmt.Errorf("删除 %s.%s 失败: %w", RankDayStatTableName, column, err)
		}
	}
	return nil
}

func backfillRankDayStatTiming(db *gorm.DB) error {
	if !db.Migrator().HasTable(RankDayStatTableName) {
		return nil
	}

	if err := db.Exec(`
UPDATE c_rank_day_stat AS stat
LEFT JOIN (
	SELECT
		day_number,
		MIN(NULLIF(remote_finished_at, 0)) AS fetched_on,
		MIN(NULLIF(synced_at, 0)) AS synced_on
	FROM d_dailybest_sync_rank
	GROUP BY day_number
) AS sync_source ON sync_source.day_number = stat.day_number
LEFT JOIN (
	SELECT
		day_number,
		MIN(NULLIF(created_on, 0)) AS fetched_on
	FROM c_rank
	GROUP BY day_number
) AS rank_source ON rank_source.day_number = stat.day_number
SET
	stat.fetched_on = CASE
		WHEN stat.fetched_on = 0
			THEN COALESCE(sync_source.fetched_on, rank_source.fetched_on, 0)
		ELSE stat.fetched_on
	END,
	stat.synced_on = CASE
		WHEN stat.synced_on = 0
			THEN COALESCE(sync_source.synced_on, rank_source.fetched_on, 0)
		ELSE stat.synced_on
	END
WHERE stat.fetched_on = 0 OR stat.synced_on = 0`).Error; err != nil {
		return fmt.Errorf("回填 %s 抓取和同步时间失败: %w", RankDayStatTableName, err)
	}
	return nil
}
