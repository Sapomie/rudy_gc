package modelg

import (
	"fmt"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if err := migrateDailyBestTableNames(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(
		new(DailyBestSourceConfig),
		new(DailyBestFetchRun),
		new(DailyBestFetchLog),
		new(DailyBestRuntimeEvent),
		new(DailyBestTargetProbe),
		new(DailyBestMovie),
		new(Rank),
		new(RankDayStat),
	); err != nil {
		return err
	}
	if err := ensureDailyBestActiveRunConstraint(db); err != nil {
		return err
	}
	if err := dropRankDayStatRedundantColumns(db); err != nil {
		return err
	}
	if err := migrateDailyBestSnapshots(db); err != nil {
		return err
	}
	if err := backfillRankDayStatTiming(db); err != nil {
		return err
	}
	return dropDailyBestPeriodTables(db)
}

func ensureDailyBestActiveRunConstraint(db *gorm.DB) error {
	const tableName = "d_dailybest_fetch_run"
	const columnName = "active_key"
	const indexName = "uidx_dailybest_active_run"
	if !db.Migrator().HasColumn(tableName, columnName) {
		if err := db.Exec(`
ALTER TABLE d_dailybest_fetch_run
ADD COLUMN active_key VARCHAR(32)
GENERATED ALWAYS AS (
	CASE
		WHEN status IN ('queued', 'running') THEN CONCAT(business_date, ':', category)
		ELSE NULL
	END
) STORED`).Error; err != nil {
			return fmt.Errorf("增加DailyBest活动任务约束字段失败: %w", err)
		}
	}
	if !db.Migrator().HasIndex(tableName, indexName) {
		if err := db.Exec(`CREATE UNIQUE INDEX uidx_dailybest_active_run ON d_dailybest_fetch_run (active_key)`).Error; err != nil {
			return fmt.Errorf("增加DailyBest活动任务唯一约束失败: %w", err)
		}
	}
	return nil
}

func dropRankDayStatRedundantColumns(db *gorm.DB) error {
	for _, column := range []string{"rise_movie_count", "fall_movie_count", "unchanged_movie_count"} {
		if err := dropColumnIfExists(db, RankDayStatTableName, column); err != nil {
			return fmt.Errorf("清理每日排行统计冗余字段失败: %w", err)
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
		business_date,
		MIN(NULLIF(finished_at, 0)) AS fetched_on
	FROM d_dailybest_fetch_run
	WHERE status = 'succeeded'
	GROUP BY business_date
) AS run_source ON run_source.business_date = stat.rank_date
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
			THEN COALESCE(run_source.fetched_on, rank_source.fetched_on, stat.created_on, 0)
		ELSE stat.fetched_on
	END,
	stat.synced_on = CASE
		WHEN stat.synced_on = 0
			THEN COALESCE(NULLIF(stat.created_on, 0), run_source.fetched_on, rank_source.fetched_on, 0)
		ELSE stat.synced_on
	END
WHERE stat.fetched_on = 0 OR stat.synced_on = 0`).Error; err != nil {
		return fmt.Errorf("回填 %s 抓取和同步时间失败: %w", RankDayStatTableName, err)
	}
	return nil
}

func dropDailyBestPeriodTables(db *gorm.DB) error {
	for _, table := range []string{"c_rank_period_stat", "c_rank_period_item", "c_rank_period"} {
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("删除已停用的数据表 %s 失败: %w", table, err)
		}
	}
	return nil
}

func migrateDailyBestTableNames(db *gorm.DB) error {
	tables := [][2]string{
		{"dailybest_fetch_run", "d_dailybest_fetch_run"},
		{"dailybest_snapshot", "d_dailybest_snapshot"},
		{"dailybest_snapshot_item", "d_dailybest_snapshot_item"},
		{"dailybest_snapshot_page", "d_dailybest_snapshot_page"},
	}
	for _, table := range tables {
		oldName, newName := table[0], table[1]
		if !db.Migrator().HasTable(oldName) {
			continue
		}
		if db.Migrator().HasTable(newName) {
			return fmt.Errorf("DailyBest 数据表命名迁移冲突: %s 与 %s 同时存在", oldName, newName)
		}
		if err := db.Migrator().RenameTable(oldName, newName); err != nil {
			return fmt.Errorf("DailyBest 数据表 %s 重命名为 %s 失败: %w", oldName, newName, err)
		}
	}
	if err := renameColumnIfNeeded(db, "d_dailybest_fetch_run", "created_at", "created_on"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "d_dailybest_fetch_run", "updated_at", "updated_on"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "d_dailybest_snapshot", "created_at", "created_on"); err != nil {
		return err
	}
	if err := renameColumnIfNeeded(db, "d_dailybest_snapshot_item", "created_at", "created_on"); err != nil {
		return err
	}
	return renameColumnIfNeeded(db, "d_dailybest_snapshot_page", "created_at", "created_on")
}

func migrateDailyBestSnapshots(db *gorm.DB) error {
	if !db.Migrator().HasTable("d_dailybest_snapshot") {
		return dropColumnIfExists(db, "d_dailybest_fetch_run", "snapshot_id")
	}
	if !db.Migrator().HasTable("d_dailybest_snapshot_item") {
		return fmt.Errorf("DailyBest 旧快照主表存在，但明细表 d_dailybest_snapshot_item 缺失")
	}
	if db.Migrator().HasColumn("d_dailybest_fetch_run", "snapshot_id") {
		if err := db.Exec(`
UPDATE d_dailybest_fetch_run AS run
JOIN d_dailybest_snapshot AS snapshot ON snapshot.id = run.snapshot_id
SET run.checksum = snapshot.checksum
WHERE TRIM(COALESCE(run.checksum, '')) = ''`).Error; err != nil {
			return fmt.Errorf("按 snapshot_id 回填抓取记录 checksum 失败: %w", err)
		}
	}

	if err := db.Exec(`
UPDATE d_dailybest_fetch_run AS run
JOIN d_dailybest_snapshot AS snapshot ON snapshot.fetch_run_id = run.id
SET run.checksum = snapshot.checksum
WHERE TRIM(COALESCE(run.checksum, '')) = ''`).Error; err != nil {
		return fmt.Errorf("回填抓取记录 checksum 失败: %w", err)
	}
	if err := migrateDailyBestMovies(db); err != nil {
		return err
	}

	if err := db.Exec(`
INSERT INTO c_rank (
	rank_key, movie_jav_id, day_number, rank_pos, category, created_on, updated_on
)
SELECT
	CONCAT(snapshot.business_date, '_', snapshot.category, '_', LPAD(item.rank_pos, 3, '0')),
	item.jav_id,
	DATEDIFF(snapshot.business_date, '2023-01-01') + 1,
	item.rank_pos,
	snapshot.category,
	snapshot.published_at,
	snapshot.published_at
FROM d_dailybest_snapshot_item AS item
JOIN d_dailybest_snapshot AS snapshot ON snapshot.id = item.snapshot_id
JOIN (
	SELECT business_date, category, MAX(id) AS snapshot_id
	FROM d_dailybest_snapshot
	GROUP BY business_date, category
) AS latest ON latest.snapshot_id = snapshot.id
WHERE NOT EXISTS (
	SELECT 1
	FROM c_rank AS existing
	WHERE existing.day_number = DATEDIFF(snapshot.business_date, '2023-01-01') + 1
	  AND existing.category = snapshot.category
)`).Error; err != nil {
		return fmt.Errorf("迁移旧快照排名失败: %w", err)
	}

	var expected, actual int64
	if err := db.Raw(`
SELECT COUNT(*)
FROM d_dailybest_snapshot_item AS item
JOIN (
	SELECT business_date, category, MAX(id) AS snapshot_id
	FROM d_dailybest_snapshot
	GROUP BY business_date, category
) AS latest ON latest.snapshot_id = item.snapshot_id`).Scan(&expected).Error; err != nil {
		return err
	}
	if err := db.Raw(`
SELECT COUNT(*)
FROM d_dailybest_snapshot_item AS item
JOIN d_dailybest_snapshot AS snapshot ON snapshot.id = item.snapshot_id
JOIN (
	SELECT business_date, category, MAX(id) AS snapshot_id
	FROM d_dailybest_snapshot
	GROUP BY business_date, category
) AS latest ON latest.snapshot_id = snapshot.id
JOIN c_rank AS rank_row
  ON rank_row.day_number = DATEDIFF(snapshot.business_date, '2023-01-01') + 1
 AND rank_row.category = snapshot.category
 AND rank_row.rank_pos = item.rank_pos
 AND rank_row.movie_jav_id = item.jav_id`).Scan(&actual).Error; err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("旧快照排名迁移校验失败: 源明细=%d, 排名=%d", expected, actual)
	}

	for _, table := range []string{"d_dailybest_snapshot_page", "d_dailybest_snapshot_item", "d_dailybest_snapshot"} {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				return fmt.Errorf("删除旧快照表 %s 失败: %w", table, err)
			}
		}
	}
	return dropColumnIfExists(db, "d_dailybest_fetch_run", "snapshot_id")
}

func migrateDailyBestMovies(db *gorm.DB) error {
	if err := db.Exec(`
INSERT INTO d_dailybest_movie (
	movie_jav_id, movie_name, movie_code, title, cover_url, created_on, updated_on
)
SELECT
	item.jav_id, item.movie_name, item.movie_code, item.title, item.cover_url,
	item.created_on, item.created_on
FROM d_dailybest_snapshot_item AS item
JOIN (
	SELECT jav_id, MAX(id) AS item_id
	FROM d_dailybest_snapshot_item
	WHERE TRIM(jav_id) <> ''
	GROUP BY jav_id
) AS latest ON latest.item_id = item.id
ON DUPLICATE KEY UPDATE
	movie_name = COALESCE(NULLIF(VALUES(movie_name), ''), movie_name),
	movie_code = COALESCE(NULLIF(VALUES(movie_code), ''), movie_code),
	title = COALESCE(NULLIF(VALUES(title), ''), title),
	cover_url = COALESCE(NULLIF(VALUES(cover_url), ''), cover_url),
	updated_on = GREATEST(updated_on, VALUES(updated_on))`).Error; err != nil {
		return fmt.Errorf("迁移旧快照影片基本信息失败: %w", err)
	}
	return nil
}

func renameColumnIfNeeded(db *gorm.DB, tableName, oldName, newName string) error {
	if !db.Migrator().HasTable(tableName) || !db.Migrator().HasColumn(tableName, oldName) {
		return nil
	}
	if db.Migrator().HasColumn(tableName, newName) {
		return fmt.Errorf("DailyBest 数据列命名迁移冲突: %s.%s 与 %s.%s 同时存在", tableName, oldName, tableName, newName)
	}
	if err := db.Migrator().RenameColumn(tableName, oldName, newName); err != nil {
		return fmt.Errorf("DailyBest 数据列 %s.%s 重命名为 %s 失败: %w", tableName, oldName, newName, err)
	}
	return nil
}

func dropColumnIfExists(db *gorm.DB, tableName, columnName string) error {
	if !db.Migrator().HasTable(tableName) || !db.Migrator().HasColumn(tableName, columnName) {
		return nil
	}
	if err := db.Migrator().DropColumn(tableName, columnName); err != nil {
		return fmt.Errorf("删除旧字段 %s.%s 失败: %w", tableName, columnName, err)
	}
	return nil
}
