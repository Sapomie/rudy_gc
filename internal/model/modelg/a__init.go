package modelg

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ------- 原有入口保持不变 -------

func MustAutoMigrate(db *gorm.DB) {
	if err := AutoMigrate(db); err != nil {
		panic(fmt.Sprintf("auto migrate err:%v", err))
	}
}

func AutoMigrate(db *gorm.DB) error {
	if err := migrateLegacySehuatangTable(db); err != nil {
		return err
	}
	if err := migrateLegacyMovieNameColumns(db); err != nil {
		return err
	}
	if err := migrateSehuatangInfoHashUniqueIndex(db); err != nil {
		return err
	}
	if err := migrateDropSehuatangRedundantColumns(db); err != nil {
		return err
	}
	if err := migrateAddSehuatangTagColumn(db); err != nil {
		return err
	}
	if err := backfillSehuatangTags(db); err != nil {
		return err
	}
	if err := migrateDropJavbusRedundantColumns(db); err != nil {
		return err
	}
	if err := migrateDropSukebeiRedundantColumns(db); err != nil {
		return err
	}
	if err := migrateAddGScStatRedundantColumns(db); err != nil {
		return err
	}
	if err := migrateDropWMediaAggRedundantColumns(db); err != nil {
		return err
	}

	// 1) 迁移表结构
	if err := batchMigrate(db,
		new(Seed),
		new(Inventory),
		new(Bestinv),
		new(Item),
		new(Detail),
		new(Rank),
		new(RankPeriod),
		new(RankPeriodItem),
		new(Cafo),
		new(Person),
		new(GScStat),

		new(Movie),
		new(Murl),
		new(Minfo),

		new(Cast),
		new(Genre),
		new(Director),
		new(Maker),
		new(Label),
		new(Prefix),
		new(MovieCast),
		new(MovieGenre),

		new(List),
		new(Sc),
		new(Record),
		new(DeletedMovie),

		new(Directory),
		new(Film),
		new(Folder),
		new(Kv),
		new(Media),
		new(MediaBirthBucketStat),
		new(MediaBirthTopStat),
		new(MediaAggDirty),
		new(MovieReleaseBucketStat),
		new(MovieReleaseTopStat),
		new(MovieReleaseAggDirty),
		new(AggEvent),
		new(FetchSite),
		new(Album),
		new(AlbumItem),
		new(JavbusMagnet),
		new(SehuatangMagnet),
		new(JavbusMagnetFetch),
		new(SukebeiTorrent),
		new(SukebeiTorrentFetch),
	); err != nil {
		return err
	}
	if err := ensureBaiduFanyiFetchSite(db); err != nil {
		return err
	}
	if err := backfillGScStatRedundantColumns(db); err != nil {
		return err
	}

	return nil
}

func migrateDropWMediaAggRedundantColumns(db *gorm.DB) error {
	type target struct {
		table  string
		column string
	}

	targets := []target{
		{table: "w_media_birth_bucket_stat", column: "movie_count"},
		{table: "w_media_birth_top_stat", column: "movie_count"},
	}

	for _, item := range targets {
		hasTbl, err := hasTable(db, item.table)
		if err != nil {
			return err
		}
		if !hasTbl {
			continue
		}

		hasCol, err := hasColumn(db, item.table, item.column)
		if err != nil {
			return err
		}
		if !hasCol {
			continue
		}

		sql := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", item.table, item.column)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateAddGScStatRedundantColumns(db *gorm.DB) error {
	tableName := "g_sc_stat"

	hasTableFlag, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableFlag {
		return nil
	}

	type columnDef struct {
		name     string
		ddl      string
		indexDDL string
	}

	columns := []columnDef{
		{
			name:     "releasing_date",
			ddl:      "ALTER TABLE `g_sc_stat` ADD COLUMN `releasing_date` bigint NOT NULL DEFAULT 0 AFTER `last_sc_time`",
			indexDDL: "ALTER TABLE `g_sc_stat` ADD INDEX `idx_gss_release_name` (`releasing_date` DESC, `movie_name` DESC)",
		},
		{
			name:     "media_birth_time",
			ddl:      "ALTER TABLE `g_sc_stat` ADD COLUMN `media_birth_time` bigint NOT NULL DEFAULT 0 AFTER `releasing_date`",
			indexDDL: "ALTER TABLE `g_sc_stat` ADD INDEX `idx_gss_media_birth_name` (`media_birth_time` DESC, `movie_name` DESC)",
		},
	}

	for _, item := range columns {
		hasCol, err := hasColumn(db, tableName, item.name)
		if err != nil {
			return err
		}
		if !hasCol {
			if err := db.Exec(item.ddl).Error; err != nil {
				return err
			}
		}
	}

	if hasIdx, err := hasIndex(db, tableName, "idx_gss_release_name"); err != nil {
		return err
	} else if !hasIdx {
		if err := db.Exec(columns[0].indexDDL).Error; err != nil {
			return err
		}
	}

	if hasIdx, err := hasIndex(db, tableName, "idx_gss_media_birth_name"); err != nil {
		return err
	} else if !hasIdx {
		if err := db.Exec(columns[1].indexDDL).Error; err != nil {
			return err
		}
	}

	return nil
}

func backfillGScStatRedundantColumns(db *gorm.DB) error {
	tableName := "g_sc_stat"

	hasTableFlag, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableFlag {
		return nil
	}

	hasRelease, err := hasColumn(db, tableName, "releasing_date")
	if err != nil {
		return err
	}
	hasMediaBirth, err := hasColumn(db, tableName, "media_birth_time")
	if err != nil {
		return err
	}
	if !hasRelease || !hasMediaBirth {
		return nil
	}

	sql := `
UPDATE ` + "`g_sc_stat` gss" + `
LEFT JOIN ` + "`a_movie` am" + ` ON am.jav_id = gss.movie_jav_id
LEFT JOIN ` + "`w_media` wm" + ` ON wm.movie_jav_id = gss.movie_jav_id
SET
	gss.releasing_date = COALESCE(am.releasing_date, 0),
	gss.media_birth_time = COALESCE(wm.birth_time, 0)
`
	return db.Exec(sql).Error
}

func ensureBaiduFanyiFetchSite(db *gorm.DB) error {
	const (
		siteCode       = "baidu_fanyi"
		siteName       = "BaiduFanyi"
		baseURL        = "https://api.fanyi.baidu.com"
		defaultUA      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
		defaultProxy   = "http://127.0.0.1:7897/"
		timeoutSeconds = int64(60)
		requestSleepMs = int64(3000)
		retrySleepMs   = int64(3000)
		maxRetryTimes  = int64(45)
		statusEnabled  = int8(1)
	)

	nowUnix := time.Now().Unix()
	var count int64
	if err := db.Model(&FetchSite{}).Where("site_code = ?", siteCode).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return db.Model(&FetchSite{}).
			Where("site_code = ?", siteCode).
			Updates(map[string]any{
				"proxy":      defaultProxy,
				"updated_on": nowUnix,
			}).Error
	}

	row := &FetchSite{
		SiteCode:       siteCode,
		SiteName:       siteName,
		BaseURL:        baseURL,
		UserAgent:      defaultUA,
		Cookie:         "",
		Proxy:          defaultProxy,
		TimeoutSeconds: timeoutSeconds,
		RequestSleepMs: requestSleepMs,
		RetrySleepMs:   retrySleepMs,
		MaxRetryTimes:  maxRetryTimes,
		Status:         statusEnabled,
		CreatedOn:      nowUnix,
		UpdatedOn:      nowUnix,
	}
	return db.Create(row).Error
}

func migrateDropSehuatangRedundantColumns(db *gorm.DB) error {
	tableName := "t_sehuatang_magnet"

	hasSehuatangTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasSehuatangTable {
		return nil
	}

	hasCol, err := hasColumn(db, tableName, "magnet_url")
	if err != nil {
		return err
	}
	if !hasCol {
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `magnet_url`", tableName)
	return db.Exec(sql).Error
}

func migrateAddSehuatangTagColumn(db *gorm.DB) error {
	tableName := "t_sehuatang_magnet"

	hasSehuatangTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasSehuatangTable {
		return nil
	}

	hasCol, err := hasColumn(db, tableName, "tag")
	if err != nil {
		return err
	}
	if hasCol {
		return nil
	}

	sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `tag` varchar(100) NOT NULL DEFAULT '' AFTER `movie_name`", tableName)
	return db.Exec(sql).Error
}

func backfillSehuatangTags(db *gorm.DB) error {
	tableName := "t_sehuatang_magnet"

	hasSehuatangTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasSehuatangTable {
		return nil
	}

	hasCol, err := hasColumn(db, tableName, "tag")
	if err != nil {
		return err
	}
	if !hasCol {
		return nil
	}

	if err := db.Exec(
		fmt.Sprintf(
			"UPDATE `%s` SET `tag` = 'FC2PPV' WHERE (`tag` = '' OR `tag` IS NULL) AND UPPER(`thread_title`) LIKE '%%FC2PPV%%'",
			tableName,
		),
	).Error; err != nil {
		return err
	}

	if err := db.Exec(
		fmt.Sprintf(
			"UPDATE `%s` SET `tag` = '自提征用' WHERE (`tag` = '' OR `tag` IS NULL) AND (`thread_title` LIKE '%%[自译征用]%%' OR `thread_title` LIKE '%%[自提征用]%%')",
			tableName,
		),
	).Error; err != nil {
		return err
	}

	return nil
}

func migrateDropJavbusRedundantColumns(db *gorm.DB) error {
	tableName := "t_javbus_magnet"

	hasJavbusTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasJavbusTable {
		return nil
	}

	columns := []string{
		"page_url",
		"magnet_url",
		"dn",
		"size_text",
	}
	for _, columnName := range columns {
		hasCol, err := hasColumn(db, tableName, columnName)
		if err != nil {
			return err
		}
		if !hasCol {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, columnName)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateDropSukebeiRedundantColumns(db *gorm.DB) error {
	tableName := "t_sukebei_torrent"

	hasSukebeiTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasSukebeiTable {
		return nil
	}

	columns := []string{
		"search_url",
		"view_url",
		"torrent_url",
		"magnet_url",
		"dn",
		"size_text",
	}
	for _, columnName := range columns {
		hasCol, err := hasColumn(db, tableName, columnName)
		if err != nil {
			return err
		}
		if !hasCol {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, columnName)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateSehuatangInfoHashUniqueIndex(db *gorm.DB) error {
	tableName := "t_sehuatang_magnet"

	hasSehuatangTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasSehuatangTable {
		return nil
	}

	type duplicateInfoHash struct {
		InfoHash string `gorm:"column:info_hash"`
	}

	var duplicates []duplicateInfoHash
	if err := db.Raw(
		"SELECT info_hash FROM `t_sehuatang_magnet` GROUP BY info_hash HAVING COUNT(1) > 1",
	).Scan(&duplicates).Error; err != nil {
		return err
	}

	for _, item := range duplicates {
		var keepID int64
		if err := db.Raw(
			"SELECT id FROM `t_sehuatang_magnet` WHERE info_hash = ? ORDER BY last_seen_time DESC, updated_on DESC, id DESC LIMIT 1",
			item.InfoHash,
		).Scan(&keepID).Error; err != nil {
			return err
		}
		if keepID <= 0 {
			continue
		}
		if err := db.Exec(
			"DELETE FROM `t_sehuatang_magnet` WHERE info_hash = ? AND id <> ?",
			item.InfoHash,
			keepID,
		).Error; err != nil {
			return err
		}
	}

	hasOldUnique, err := hasIndex(db, tableName, "uk_movie_hash")
	if err != nil {
		return err
	}
	if hasOldUnique {
		if err := db.Exec("ALTER TABLE `t_sehuatang_magnet` DROP INDEX `uk_movie_hash`").Error; err != nil {
			return err
		}
	}

	hasInfoHashUnique, err := hasIndex(db, tableName, "uk_info_hash")
	if err != nil {
		return err
	}
	if !hasInfoHashUnique {
		if err := db.Exec("ALTER TABLE `t_sehuatang_magnet` ADD UNIQUE INDEX `uk_info_hash` (`info_hash`)").Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateLegacySehuatangTable(db *gorm.DB) error {
	oldTable := "t_98tang_magnet"
	newTable := "t_sehuatang_magnet"

	hasOld, err := hasTable(db, oldTable)
	if err != nil {
		return err
	}
	if !hasOld {
		return nil
	}
	hasNew, err := hasTable(db, newTable)
	if err != nil {
		return err
	}
	if hasNew {
		return nil
	}

	sql := fmt.Sprintf("RENAME TABLE `%s` TO `%s`", oldTable, newTable)
	return db.Exec(sql).Error
}

func migrateLegacyMovieNameColumns(db *gorm.DB) error {
	legacyColumn := "movie" + "_code"
	targetColumn := "movie_name"

	renames := []struct {
		table string
		typ   string
	}{
		{table: "t_javbus_magnet_fetch", typ: "varchar(191)"},
		{table: "t_sukebei_torrent_fetch", typ: "varchar(191)"},
		{table: "tm_album_item", typ: "varchar(191)"},
	}

	for _, item := range renames {
		hasOld, err := hasColumn(db, item.table, legacyColumn)
		if err != nil {
			return err
		}
		if !hasOld {
			continue
		}
		hasNew, err := hasColumn(db, item.table, targetColumn)
		if err != nil {
			return err
		}
		if hasNew {
			continue
		}

		sql := fmt.Sprintf("ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` %s NOT NULL", item.table, legacyColumn, targetColumn, item.typ)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func hasTable(db *gorm.DB, tableName string) (bool, error) {
	var count int64
	err := db.Raw(
		"SELECT COUNT(1) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		tableName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func hasColumn(db *gorm.DB, tableName string, columnName string) (bool, error) {
	var count int64
	err := db.Raw(
		"SELECT COUNT(1) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName,
		columnName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func hasIndex(db *gorm.DB, tableName string, indexName string) (bool, error) {
	var count int64
	err := db.Raw(
		"SELECT COUNT(1) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?",
		tableName,
		indexName,
	).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func batchMigrate(db *gorm.DB, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
