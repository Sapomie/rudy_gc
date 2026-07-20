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
		new(DailyBestFetchRun),
		new(DailyBestMovie),
		new(Rank),
		new(RankDayStat),
		new(RankPeriod),
		new(RankPeriodItem),
		new(RankPeriodStat),
	); err != nil {
		return err
	}
	return migrateDailyBestSnapshots(db)
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
ON DUPLICATE KEY UPDATE
	movie_jav_id = VALUES(movie_jav_id),
	day_number = VALUES(day_number),
	rank_pos = VALUES(rank_pos),
	category = VALUES(category),
	updated_on = VALUES(updated_on)`).Error; err != nil {
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
FROM c_rank AS rank_row
JOIN (
	SELECT DATEDIFF(business_date, '2023-01-01') + 1 AS day_number, category
	FROM d_dailybest_snapshot
	GROUP BY business_date, category
) AS source_day ON source_day.day_number = rank_row.day_number AND source_day.category = rank_row.category`).Scan(&actual).Error; err != nil {
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
