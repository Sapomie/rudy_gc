package modelg

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"rudy_gc/internal/consts"

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
	if err := migrateDropLegacyAggDirtyTables(db); err != nil {
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
		new(PersonSc),
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
		new(Folder),
		new(Kv),
		new(Media),
		new(MediaBirthBucketStat),
		new(MediaBirthTopStat),
		new(MovieReleaseBucketStat),
		new(MovieReleaseTopStat),
		new(AggEvent),
		new(FetchSite),
		new(Album),
		new(AlbumItem),
		new(CMovieAlbum),
		new(CMovieAlbumItem),
		new(JavbusMagnet),
		new(SehuatangMagnet),
		new(JavbusMagnetFetch),
		new(SukebeiTorrent),
		new(SukebeiTorrentFetch),
	); err != nil {
		return err
	}
	if err := migrateAddAmCastOwnedWMediaNumber(db); err != nil {
		return err
	}
	if err := migrateAddCPersonOwnedWMediaNumber(db); err != nil {
		return err
	}
	if err := migrateAddCPersonOwnedWMediaRatio(db); err != nil {
		return err
	}
	if err := migrateWMediaSourceType(db); err != nil {
		return err
	}
	if err := migrateMovieReleaseAggMode(db); err != nil {
		return err
	}
	if err := migrateMovieNeedDownloadToAlbum(db); err != nil {
		return err
	}
	if err := migrateCMovieAlbumItemReleasingDate(db); err != nil {
		return err
	}
	if err := migrateWFolderSourceType(db); err != nil {
		return err
	}
	if err := migrateDSeedMovieStatsColumns(db); err != nil {
		return err
	}
	if err := ensureBaiduFanyiFetchSite(db); err != nil {
		return err
	}
	if err := backfillDSeedMovieStats(db); err != nil {
		return err
	}
	if err := dropDSeedLegacyMovieMaxNoColumns(db); err != nil {
		return err
	}
	if err := backfillGScStatRedundantColumns(db); err != nil {
		return err
	}
	if err := backfillCPersonStatsFromAmCast(db); err != nil {
		return err
	}

	return nil
}

func migrateAddAmCastOwnedWMediaNumber(db *gorm.DB) error {
	const tableName = "am_cast"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	hasColumnValue, err := hasColumn(db, tableName, "owned_w_media_number")
	if err != nil {
		return err
	}
	if !hasColumnValue {
		if err := db.Exec("ALTER TABLE `am_cast` ADD COLUMN `owned_w_media_number` bigint NOT NULL DEFAULT 0 AFTER `owned_movie_number`").Error; err != nil {
			return err
		}
	}

	return db.Exec(`
UPDATE am_cast ac
SET owned_w_media_number = (
	SELECT COUNT(DISTINCT amr.movie_jav_id)
	FROM amr_movie_cast amr
	JOIN w_media wm
	  ON wm.movie_jav_id = amr.movie_jav_id
	 AND wm.source_type = ?
	 AND wm.is_removed = ?
	WHERE amr.cast_id = ac.id
)`,
		consts.WMediaSourceNative,
		consts.FilmIsNotRemoved,
	).Error
}

func migrateDropLegacyAggDirtyTables(db *gorm.DB) error {
	if err := db.Exec("DROP TABLE IF EXISTS `w_media_agg_dirty`").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP TABLE IF EXISTS `movie_release_agg_dirty`").Error; err != nil {
		return err
	}
	return nil
}

func migrateAddCPersonOwnedWMediaNumber(db *gorm.DB) error {
	const tableName = "c_person"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	hasColumnValue, err := hasColumn(db, tableName, "owned_w_media_number")
	if err != nil {
		return err
	}
	if !hasColumnValue {
		if err := db.Exec("ALTER TABLE `c_person` ADD COLUMN `owned_w_media_number` bigint NOT NULL DEFAULT 0 AFTER `owned_movie_number`").Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateAddCPersonOwnedWMediaRatio(db *gorm.DB) error {
	const tableName = "c_person"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	hasColumnValue, err := hasColumn(db, tableName, "owned_w_media_ratio")
	if err != nil {
		return err
	}
	if !hasColumnValue {
		if err := db.Exec("ALTER TABLE `c_person` ADD COLUMN `owned_w_media_ratio` bigint NOT NULL DEFAULT 0 AFTER `owned_w_media_number`").Error; err != nil {
			return err
		}
	}
	if err := db.Exec("ALTER TABLE `c_person` MODIFY COLUMN `owned_w_media_ratio` bigint NOT NULL DEFAULT 0 AFTER `owned_w_media_number`").Error; err != nil {
		return err
	}

	hasIndexValue, err := hasIndex(db, tableName, "idx_c_person_owned_w_media_ratio")
	if err != nil {
		return err
	}
	if !hasIndexValue {
		if err := db.Exec("ALTER TABLE `c_person` ADD INDEX `idx_c_person_owned_w_media_ratio` (`owned_w_media_ratio`)").Error; err != nil {
			return err
		}
	}

	return db.Exec(`
UPDATE c_person
SET owned_w_media_ratio = CASE
	WHEN owned_w_media_number <> 0 THEN 10000
	ELSE 0
END`).Error
}

func migrateDSeedMovieStatsColumns(db *gorm.DB) error {
	const tableName = "d_seed"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"movie_total",
		"",
		"ALTER TABLE `d_seed` ADD COLUMN `movie_total` bigint NOT NULL DEFAULT 0 AFTER `last_error`",
		"",
	); err != nil {
		return err
	}
	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"movie_latest_releasing_movie_jav_id",
		"movie_max_no_jav_id",
		"ALTER TABLE `d_seed` ADD COLUMN `movie_latest_releasing_movie_jav_id` varchar(64) NOT NULL DEFAULT '' AFTER `movie_total`",
		"ALTER TABLE `d_seed` CHANGE COLUMN `movie_max_no_jav_id` `movie_latest_releasing_movie_jav_id` varchar(64) NOT NULL DEFAULT ''",
	); err != nil {
		return err
	}
	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"movie_latest_releasing_movie_name",
		"movie_max_no_name",
		"ALTER TABLE `d_seed` ADD COLUMN `movie_latest_releasing_movie_name` varchar(191) NOT NULL DEFAULT '' AFTER `movie_latest_releasing_movie_jav_id`",
		"ALTER TABLE `d_seed` CHANGE COLUMN `movie_max_no_name` `movie_latest_releasing_movie_name` varchar(191) NOT NULL DEFAULT ''",
	); err != nil {
		return err
	}
	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"movie_last_added_time",
		"",
		"ALTER TABLE `d_seed` ADD COLUMN `movie_last_added_time` bigint NOT NULL DEFAULT 0 AFTER `movie_latest_releasing_movie_name`",
		"",
	); err != nil {
		return err
	}
	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"last_insert_count",
		"",
		"ALTER TABLE `d_seed` ADD COLUMN `last_insert_count` bigint NOT NULL DEFAULT 0 AFTER `movie_last_added_time`",
		"",
	); err != nil {
		return err
	}
	if err := ensureDSeedMovieStatColumn(
		db,
		tableName,
		"movie_latest_releasing_date",
		"",
		"ALTER TABLE `d_seed` ADD COLUMN `movie_latest_releasing_date` bigint NOT NULL DEFAULT 0 AFTER `last_insert_count`",
		"",
	); err != nil {
		return err
	}

	return nil
}

func ensureDSeedMovieStatColumn(db *gorm.DB, tableName string, newName string, oldName string, addDDL string, renameDDL string) error {
	hasNewColumn, err := hasColumn(db, tableName, newName)
	if err != nil {
		return err
	}
	if hasNewColumn {
		return nil
	}

	if oldName != "" {
		hasOldColumn, err := hasColumn(db, tableName, oldName)
		if err != nil {
			return err
		}
		if hasOldColumn {
			return db.Exec(renameDDL).Error
		}
	}

	return db.Exec(addDDL).Error
}

func backfillDSeedMovieStats(db *gorm.DB) error {
	const tableName = "d_seed"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	requiredColumns := []string{
		"movie_total",
		"movie_latest_releasing_movie_jav_id",
		"movie_latest_releasing_movie_name",
		"movie_latest_releasing_date",
	}
	for _, name := range requiredColumns {
		hasColumnValue, err := hasColumn(db, tableName, name)
		if err != nil {
			return err
		}
		if !hasColumnValue {
			return nil
		}
	}

	type seedRow struct {
		Id       int64  `gorm:"column:id"`
		Name     string `gorm:"column:name"`
		NameType int64  `gorm:"column:name_type"`
	}
	var rows []seedRow
	if err := db.Raw("SELECT id, name, name_type FROM `d_seed`").Scan(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		stats, err := calcDSeedMovieStatsForMigration(db, row.NameType, row.Name)
		if err != nil {
			return err
		}
		if err := db.Exec(
			"UPDATE `d_seed` SET `movie_total` = ?, `movie_latest_releasing_movie_jav_id` = ?, `movie_latest_releasing_movie_name` = ?, `movie_latest_releasing_date` = ? WHERE `id` = ?",
			stats.MovieTotal,
			stats.MovieLatestReleasingMovieJavId,
			stats.MovieLatestReleasingMovieName,
			stats.MovieLatestReleasingDate,
			row.Id,
		).Error; err != nil {
			return err
		}
	}

	return nil
}

func dropDSeedLegacyMovieMaxNoColumns(db *gorm.DB) error {
	const tableName = "d_seed"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	for _, name := range []string{"movie_max_no_jav_id", "movie_max_no_name"} {
		hasColumnValue, err := hasColumn(db, tableName, name)
		if err != nil {
			return err
		}
		if !hasColumnValue {
			continue
		}
		if err := db.Exec("ALTER TABLE `d_seed` DROP COLUMN `" + name + "`").Error; err != nil {
			return err
		}
	}

	return nil
}

type dSeedMovieStatsMigration struct {
	MovieTotal                     int64  `gorm:"column:movie_total"`
	MovieLatestReleasingMovieJavId string `gorm:"column:movie_latest_releasing_movie_jav_id"`
	MovieLatestReleasingMovieName  string `gorm:"column:movie_latest_releasing_movie_name"`
	MovieLatestReleasingDate       int64  `gorm:"column:movie_latest_releasing_date"`
}

func calcDSeedMovieStatsForMigration(db *gorm.DB, nameType int64, name string) (*dSeedMovieStatsMigration, error) {
	var (
		query string
		args  []any
	)
	switch nameType {
	case 1:
		query = `
SELECT
	COUNT(*) AS movie_total,
	COALESCE(MAX(am.releasing_date), 0) AS movie_latest_releasing_date,
	COALESCE((
		SELECT am2.jav_id
		FROM a_movie am2
		JOIN am_prefix pf2 ON pf2.id = am2.prefix_id
		WHERE pf2.name = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_prefix pf3 ON pf3.id = am3.prefix_id
			WHERE pf3.name = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_jav_id,
	COALESCE((
		SELECT am2.name
		FROM a_movie am2
		JOIN am_prefix pf2 ON pf2.id = am2.prefix_id
		WHERE pf2.name = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_prefix pf3 ON pf3.id = am3.prefix_id
			WHERE pf3.name = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_name
FROM a_movie am
JOIN am_prefix pf ON pf.id = am.prefix_id
WHERE pf.name = ?
`
		args = []any{name, name, name, name, name}
	case 2:
		query = `
SELECT
	COUNT(*) AS movie_total,
	COALESCE(MAX(am.releasing_date), 0) AS movie_latest_releasing_date,
	COALESCE((
		SELECT am2.jav_id
		FROM a_movie am2
		JOIN am_label lb2 ON lb2.id = am2.label_id
		WHERE lb2.jav_id = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_label lb3 ON lb3.id = am3.label_id
			WHERE lb3.jav_id = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_jav_id,
	COALESCE((
		SELECT am2.name
		FROM a_movie am2
		JOIN am_label lb2 ON lb2.id = am2.label_id
		WHERE lb2.jav_id = ?
		  AND am2.releasing_date = (
			SELECT MAX(am3.releasing_date)
			FROM a_movie am3
			JOIN am_label lb3 ON lb3.id = am3.label_id
			WHERE lb3.jav_id = ?
		  )
		ORDER BY
		  CASE WHEN SUBSTRING_INDEX(am2.name, '-', -1) REGEXP '^[0-9]+$' THEN CAST(SUBSTRING_INDEX(am2.name, '-', -1) AS UNSIGNED) ELSE 0 END DESC,
		  am2.name DESC,
		  am2.jav_id DESC
		LIMIT 1
	), '') AS movie_latest_releasing_movie_name
FROM a_movie am
JOIN am_label lb ON lb.id = am.label_id
WHERE lb.jav_id = ?
`
		args = []any{name, name, name, name, name}
	default:
		return &dSeedMovieStatsMigration{}, nil
	}

	var out dSeedMovieStatsMigration
	if err := db.Raw(query, args...).Scan(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

func backfillCPersonStatsFromAmCast(db *gorm.DB) error {
	const query = `
UPDATE c_person p
LEFT JOIN (
	SELECT
		pm.person_id AS person_id,
		COUNT(*) AS movie_number,
		COALESCE(SUM(CASE WHEN wm_owned.is_removed = ? THEN 1 ELSE 0 END), 0) AS owned_movie_number,
		COALESCE(SUM(CASE WHEN wm.is_removed = ? THEN 1 ELSE 0 END), 0) AS owned_w_media_number,
		COALESCE(SUM(COALESCE(gss.sc_times, 0)), 0) AS sc_times,
		COALESCE(SUM(COALESCE(gss.come_times, 0)), 0) AS come_times,
		COALESCE(MAX(COALESCE(gss.last_sc_time, 0)), 0) AS last_sc_time,
		COALESCE(MIN(CASE WHEN mi.highest_rank > 0 AND mi.highest_rank < 1000 THEN mi.highest_rank END), 0) AS highest_rank,
		COALESCE(SUM(CASE WHEN mi.days_in_rank > 0 THEN mi.days_in_rank ELSE 0 END), 0) AS rank_times
	FROM (
		SELECT DISTINCT ac.person_id AS person_id, amr.movie_jav_id AS movie_jav_id
		FROM am_cast ac
		JOIN amr_movie_cast amr ON amr.cast_id = ac.id
		WHERE ac.person_id > 0
	) pm
	LEFT JOIN w_media wm_owned ON wm_owned.movie_jav_id = pm.movie_jav_id AND wm_owned.source_type = ?
	LEFT JOIN w_media wm ON wm.movie_jav_id = pm.movie_jav_id AND wm.source_type = ?
	LEFT JOIN g_sc_stat gss ON gss.movie_jav_id = pm.movie_jav_id
	LEFT JOIN bm_minfo mi ON mi.jav_id = pm.movie_jav_id
	GROUP BY pm.person_id
) agg ON agg.person_id = p.id
SET
	p.movie_number = COALESCE(agg.movie_number, 0),
	p.owned_movie_number = COALESCE(agg.owned_w_media_number, 0),
	p.owned_w_media_number = COALESCE(agg.owned_w_media_number, 0),
	p.owned_w_media_ratio = CASE
		WHEN COALESCE(agg.owned_w_media_number, 0) <> 0 THEN 10000
		ELSE 0
	END,
	p.sc_times = COALESCE(agg.sc_times, 0),
	p.come_times = COALESCE(agg.come_times, 0),
	p.last_sc_time = COALESCE(agg.last_sc_time, 0),
	p.highest_rank = COALESCE(agg.highest_rank, 0),
	p.rank_times = COALESCE(agg.rank_times, 0)`

	return db.Exec(query,
		consts.FilmIsNotRemoved,
		consts.FilmIsNotRemoved,
		consts.WMediaSourceNative,
		consts.WMediaSourceNative,
	).Error
}

func migrateMovieNeedDownloadToAlbum(db *gorm.DB) error {
	const (
		albumTable      = "c_movie_album"
		albumItemTable  = "c_movie_album_item"
		minfoTable      = "bm_minfo"
		legacyIndexName = "idx_need_reldate_name"
		legacyColumn    = "need_download"
	)

	if ok, err := hasTable(db, albumTable); err != nil {
		return err
	} else if !ok {
		return nil
	}
	if ok, err := hasTable(db, albumItemTable); err != nil {
		return err
	} else if !ok {
		return nil
	}

	now := time.Now().Unix()
	insertAlbumSQL := `
INSERT INTO c_movie_album (name, remark, created_on, updated_on)
SELECT ?, ?, ?, ?
FROM DUAL
WHERE NOT EXISTS (
	SELECT 1 FROM c_movie_album WHERE name = ?
)`
	if err := db.Exec(insertAlbumSQL,
		consts.MovieNeedDownloadAlbumName,
		consts.MovieNeedDownloadAlbumRemark,
		now,
		now,
		consts.MovieNeedDownloadAlbumName,
	).Error; err != nil {
		return err
	}

	hasLegacyColumn, err := hasColumn(db, minfoTable, legacyColumn)
	if err != nil {
		return err
	}
	if !hasLegacyColumn {
		return nil
	}

	backfillSQL := `
INSERT INTO c_movie_album_item (album_id, movie_jav_id, movie_name, releasing_date, sort_no, created_on, updated_on)
SELECT
	ca.id,
	mi.jav_id,
	COALESCE(NULLIF(am.name, ''), mi.name, mi.jav_id) AS movie_name,
	COALESCE(am.releasing_date, 0) AS releasing_date,
	UNIX_TIMESTAMP(),
	?,
	?
FROM bm_minfo mi
JOIN c_movie_album ca
  ON ca.name = ?
LEFT JOIN a_movie am
  ON am.jav_id = mi.jav_id
LEFT JOIN c_movie_album_item cai
  ON cai.album_id = ca.id
 AND cai.movie_jav_id = mi.jav_id
WHERE mi.need_download = ?
  AND cai.id IS NULL`
	if err := db.Exec(backfillSQL,
		now,
		now,
		consts.MovieNeedDownloadAlbumName,
		consts.MovieNeedDownLoadOK,
	).Error; err != nil {
		return err
	}

	hasLegacyIndex, err := hasIndex(db, minfoTable, legacyIndexName)
	if err != nil {
		return err
	}
	if hasLegacyIndex {
		if err := db.Exec("ALTER TABLE `bm_minfo` DROP INDEX `idx_need_reldate_name`").Error; err != nil {
			return err
		}
	}

	if err := db.Exec("ALTER TABLE `bm_minfo` DROP COLUMN `need_download`").Error; err != nil {
		return err
	}
	return nil
}

func migrateCMovieAlbumItemReleasingDate(db *gorm.DB) error {
	const tableName = "c_movie_album_item"

	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	hasColumnValue, err := hasColumn(db, tableName, "releasing_date")
	if err != nil {
		return err
	}
	if !hasColumnValue {
		if err := db.Exec("ALTER TABLE `c_movie_album_item` ADD COLUMN `releasing_date` bigint NOT NULL DEFAULT 0 AFTER `movie_name`").Error; err != nil {
			return err
		}
	}
	if err := db.Exec("ALTER TABLE `c_movie_album_item` MODIFY COLUMN `releasing_date` bigint NOT NULL DEFAULT 0 AFTER `movie_name`").Error; err != nil {
		return err
	}

	indexMatches, err := hasIndexColumns(db, tableName, "idx_album_release_name", []string{"album_id", "releasing_date", "movie_name", "movie_jav_id"})
	if err != nil {
		return err
	}
	if !indexMatches {
		hasIndexValue, err := hasIndex(db, tableName, "idx_album_release_name")
		if err != nil {
			return err
		}
		if hasIndexValue {
			if err := db.Exec("ALTER TABLE `c_movie_album_item` DROP INDEX `idx_album_release_name`").Error; err != nil {
				return err
			}
		}
		if err := db.Exec("ALTER TABLE `c_movie_album_item` ADD INDEX `idx_album_release_name` (`album_id`, `releasing_date`, `movie_name`, `movie_jav_id`)").Error; err != nil {
			return err
		}
	}

	return db.Exec(`
UPDATE c_movie_album_item cai
LEFT JOIN a_movie am
  ON am.jav_id = cai.movie_jav_id
SET cai.releasing_date = COALESCE(am.releasing_date, 0)
WHERE cai.releasing_date <> COALESCE(am.releasing_date, 0)`).Error
}

func migrateWMediaSourceType(db *gorm.DB) error {
	const tableName = "w_media"

	hasWMediaTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasWMediaTable {
		return nil
	}

	hasSourceType, err := hasColumn(db, tableName, "source_type")
	if err != nil {
		return err
	}
	if !hasSourceType {
		if err := db.Exec("ALTER TABLE `w_media` ADD COLUMN `source_type` tinyint NOT NULL DEFAULT 2 AFTER `file_name`").Error; err != nil {
			return err
		}
	}

	if err := db.Exec("UPDATE `w_media` SET `source_type` = 2 WHERE `source_type` = 0").Error; err != nil {
		return err
	}

	oldUniqueIndexes := []string{
		"idx_w_media_movie_jav_id",
		"idx_w_media_movie_name",
		"idx_w_media_file_name",
	}
	for _, indexName := range oldUniqueIndexes {
		hasOldIndex, err := hasIndex(db, tableName, indexName)
		if err != nil {
			return err
		}
		if hasOldIndex {
			if err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", tableName, indexName)).Error; err != nil {
				return err
			}
		}
	}

	type uniqueIndexDef struct {
		name    string
		columns string
	}

	newIndexes := []uniqueIndexDef{
		{name: "idx_w_media_movie_jav_id_source_type", columns: "`movie_jav_id`, `source_type`"},
		{name: "idx_w_media_movie_name_source_type", columns: "`movie_name`, `source_type`"},
		{name: "idx_w_media_file_name_source_type", columns: "`file_name`, `source_type`"},
	}
	for _, indexDef := range newIndexes {
		hasNewIndex, err := hasIndex(db, tableName, indexDef.name)
		if err != nil {
			return err
		}
		if hasNewIndex {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE INDEX `%s` (%s)", tableName, indexDef.name, indexDef.columns)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateMovieReleaseAggMode(db *gorm.DB) error {
	type tableSpec struct {
		tableName     string
		columnSQL     string
		oldIndexes    []string
		createIndexes []string
	}

	specs := []tableSpec{
		{
			tableName: "movie_release_bucket_stat",
			columnSQL: "ALTER TABLE `movie_release_bucket_stat` ADD COLUMN `agg_mode` varchar(16) NOT NULL DEFAULT 'all' AFTER `id`",
			oldIndexes: []string{
				"uk_mrbs_scope_key",
				"idx_mrbs_level_sort",
				"idx_mrbs_latest_release",
			},
			createIndexes: []string{
				"ALTER TABLE `movie_release_bucket_stat` ADD UNIQUE INDEX `uk_mrbs_mode_scope_key` (`agg_mode`, `scope_key`)",
				"ALTER TABLE `movie_release_bucket_stat` ADD INDEX `idx_mrbs_level_sort` (`agg_mode`, `level`, `year` DESC, `quarter` DESC, `month` DESC, `day` DESC)",
				"ALTER TABLE `movie_release_bucket_stat` ADD INDEX `idx_mrbs_latest_release` (`agg_mode`, `latest_releasing_date` DESC)",
			},
		},
		{
			tableName: "movie_release_top_stat",
			columnSQL: "ALTER TABLE `movie_release_top_stat` ADD COLUMN `agg_mode` varchar(16) NOT NULL DEFAULT 'all' AFTER `id`",
			oldIndexes: []string{
				"uk_mrts_scope_type_key",
				"idx_mrts_scope_type_rank",
				"idx_mrts_level_sort",
			},
			createIndexes: []string{
				"ALTER TABLE `movie_release_top_stat` ADD UNIQUE INDEX `uk_mrts_mode_scope_type_key` (`agg_mode`, `scope_key`, `agg_type`, `agg_key`)",
				"ALTER TABLE `movie_release_top_stat` ADD INDEX `idx_mrts_scope_type_rank` (`agg_mode`, `scope_key`, `agg_type`, `rank_no`)",
				"ALTER TABLE `movie_release_top_stat` ADD INDEX `idx_mrts_level_sort` (`agg_mode`, `level`, `year` DESC, `quarter` DESC, `month` DESC)",
			},
		},
	}

	for _, spec := range specs {
		hasTableValue, err := hasTable(db, spec.tableName)
		if err != nil {
			return err
		}
		if !hasTableValue {
			continue
		}

		hasAggMode, err := hasColumn(db, spec.tableName, "agg_mode")
		if err != nil {
			return err
		}
		if !hasAggMode {
			if err := db.Exec(spec.columnSQL).Error; err != nil {
				return err
			}
		}
		if err := db.Exec(fmt.Sprintf("UPDATE `%s` SET `agg_mode` = 'all' WHERE TRIM(COALESCE(`agg_mode`, '')) = ''", spec.tableName)).Error; err != nil {
			return err
		}

		for _, indexName := range spec.oldIndexes {
			hasOldIndex, err := hasIndex(db, spec.tableName, indexName)
			if err != nil {
				return err
			}
			if hasOldIndex {
				if err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", spec.tableName, indexName)).Error; err != nil {
					return err
				}
			}
		}

		for _, sql := range spec.createIndexes {
			indexName := migrationIndexName(sql)
			hasNewIndex, err := hasIndex(db, spec.tableName, indexName)
			if err != nil {
				return err
			}
			if hasNewIndex {
				continue
			}
			if err := db.Exec(sql).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func migrationIndexName(sql string) string {
	parts := strings.Split(sql, "`")
	if len(parts) < 4 {
		return ""
	}
	return parts[3]
}

type folderMigrationRow struct {
	ID         int64  `gorm:"column:id"`
	ParentID   int64  `gorm:"column:parent_id"`
	Name       string `gorm:"column:name"`
	SourceType int64  `gorm:"column:source_type"`
	Depth      int64  `gorm:"column:depth"`
	Path       string `gorm:"column:path"`
	PathHash   []byte `gorm:"column:path_hash"`
	CreatedOn  int64  `gorm:"column:created_on"`
	UpdatedOn  int64  `gorm:"column:updated_on"`
}

type mediaFolderPathRow struct {
	SourceType int64  `gorm:"column:source_type"`
	RootDir    string `gorm:"column:root_dir"`
	FullDir    string `gorm:"column:full_dir"`
}

type desiredFolderNode struct {
	SourceType int64
	ParentPath string
	Name       string
	Depth      int64
	Path       string
	PathHash   []byte
}

func migrateWFolderSourceType(db *gorm.DB) error {
	const tableName = "w_folder"

	hasWFolderTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasWFolderTable {
		return nil
	}

	hasSourceType, err := hasColumn(db, tableName, "source_type")
	if err != nil {
		return err
	}
	if !hasSourceType {
		if err := db.Exec("ALTER TABLE `w_folder` ADD COLUMN `source_type` tinyint NOT NULL DEFAULT 2 AFTER `name`").Error; err != nil {
			return err
		}
	}

	if err := db.Exec("UPDATE `w_folder` SET `source_type` = 2 WHERE `source_type` = 0").Error; err != nil {
		return err
	}

	oldUniqueIndexes := []string{
		"uk_w_folder_parent_name",
		"uk_w_folder_path",
	}
	for _, indexName := range oldUniqueIndexes {
		hasOldIndex, err := hasIndex(db, tableName, indexName)
		if err != nil {
			return err
		}
		if hasOldIndex {
			if err := db.Exec(fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", tableName, indexName)).Error; err != nil {
				return err
			}
		}
	}

	type indexDef struct {
		name   string
		column string
		unique bool
	}

	indexes := []indexDef{
		{name: "uk_w_folder_parent_name_source_type", column: "`parent_id`, `name`, `source_type`", unique: true},
		{name: "uk_w_folder_path_source_type", column: "`path`, `source_type`", unique: true},
		{name: "idx_w_folder_source_type", column: "`source_type`", unique: false},
	}
	for _, indexDef := range indexes {
		hasNewIndex, err := hasIndex(db, tableName, indexDef.name)
		if err != nil {
			return err
		}
		if hasNewIndex {
			continue
		}

		keyword := "INDEX"
		if indexDef.unique {
			keyword = "UNIQUE INDEX"
		}
		sql := fmt.Sprintf("ALTER TABLE `%s` ADD %s `%s` (%s)", tableName, keyword, indexDef.name, indexDef.column)
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return rebuildWFolderTreesFromWMedia(db)
}

func rebuildWFolderTreesFromWMedia(db *gorm.DB) error {
	hasWMediaTable, err := hasTable(db, "w_media")
	if err != nil {
		return err
	}
	if !hasWMediaTable {
		return nil
	}

	var mediaRows []mediaFolderPathRow
	if err := db.Raw(`
SELECT DISTINCT source_type, root_dir, full_dir
FROM `+"`w_media`"+`
WHERE source_type = ?
  AND TRIM(full_dir) <> ''
ORDER BY source_type ASC, full_dir ASC
`, consts.WFolderSourceNative).Scan(&mediaRows).Error; err != nil {
		return err
	}

	desiredNodes := buildDesiredFolderNodes(filterManagedMediaFolderRows(mediaRows))
	desiredByKey := make(map[string]desiredFolderNode, len(desiredNodes))
	for _, node := range desiredNodes {
		desiredByKey[folderSourcePathKey(node.SourceType, node.Path)] = node
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var currentRows []folderMigrationRow
		if err := tx.Raw(`
SELECT id, parent_id, name, source_type, depth, path, path_hash, created_on, updated_on
FROM ` + "`w_folder`" + `
ORDER BY depth ASC, id ASC`).Scan(&currentRows).Error; err != nil {
			return err
		}

		activeByKey := make(map[string]*folderMigrationRow, len(currentRows)+len(desiredNodes))
		for i := range currentRows {
			row := &currentRows[i]
			activeByKey[folderSourcePathKey(row.SourceType, row.Path)] = row
		}

		nowUnix := time.Now().Unix()
		for _, node := range desiredNodes {
			key := folderSourcePathKey(node.SourceType, node.Path)
			parentID := int64(0)
			if node.ParentPath != "" {
				parentRow, ok := activeByKey[folderSourcePathKey(node.SourceType, node.ParentPath)]
				if !ok {
					return fmt.Errorf("rebuild w_folder: missing parent node source_type=%d path=%s", node.SourceType, node.ParentPath)
				}
				parentID = parentRow.ID
			}

			if row, ok := activeByKey[key]; ok {
				createdOn := row.CreatedOn
				if createdOn <= 0 {
					createdOn = nowUnix
				}
				if row.ParentID != parentID || row.Name != node.Name || row.Depth != node.Depth || row.Path != node.Path || row.SourceType != node.SourceType || !bytes.Equal(row.PathHash, node.PathHash) || row.CreatedOn != createdOn {
					if err := tx.Exec(`
UPDATE `+"`w_folder`"+`
SET parent_id = ?, name = ?, source_type = ?, depth = ?, path = ?, path_hash = ?, created_on = ?, updated_on = ?
WHERE id = ?
`, parentID, node.Name, node.SourceType, node.Depth, node.Path, node.PathHash, createdOn, nowUnix, row.ID).Error; err != nil {
						return err
					}
					row.ParentID = parentID
					row.Name = node.Name
					row.SourceType = node.SourceType
					row.Depth = node.Depth
					row.Path = node.Path
					row.PathHash = node.PathHash
					row.CreatedOn = createdOn
					row.UpdatedOn = nowUnix
				}
				continue
			}

			insert := &Folder{
				ParentID:   parentID,
				Name:       node.Name,
				SourceType: node.SourceType,
				Depth:      int8(node.Depth),
				Path:       node.Path,
				PathHash:   node.PathHash,
				CreatedOn:  nowUnix,
				UpdatedOn:  nowUnix,
			}
			if err := tx.Create(insert).Error; err != nil {
				return err
			}

			activeByKey[key] = &folderMigrationRow{
				ID:         insert.ID,
				ParentID:   insert.ParentID,
				Name:       insert.Name,
				SourceType: insert.SourceType,
				Depth:      int64(insert.Depth),
				Path:       insert.Path,
				PathHash:   insert.PathHash,
				CreatedOn:  insert.CreatedOn,
				UpdatedOn:  insert.UpdatedOn,
			}
		}

		var mappingRows []struct {
			ID          int64  `gorm:"column:id"`
			SourceType  int64  `gorm:"column:source_type"`
			RootDir     string `gorm:"column:root_dir"`
			FullDir     string `gorm:"column:full_dir"`
			DirectoryID int64  `gorm:"column:directory_id"`
		}
		if err := tx.Raw(`
SELECT id, source_type, root_dir, full_dir, directory_id
FROM `+"`w_media`"+`
WHERE source_type = ?
ORDER BY id ASC
`, consts.WFolderSourceNative).Scan(&mappingRows).Error; err != nil {
			return err
		}
		for _, row := range mappingRows {
			targetPath := normalizeManagedMediaFolderPath(row.SourceType, row.RootDir, row.FullDir)
			targetID := int64(0)
			if targetPath != "" {
				targetRow, ok := activeByKey[folderSourcePathKey(row.SourceType, targetPath)]
				if !ok {
					return fmt.Errorf("rebuild w_folder: missing mapped folder source_type=%d path=%s", row.SourceType, targetPath)
				}
				targetID = targetRow.ID
			}
			if row.DirectoryID == targetID {
				continue
			}
			if err := tx.Exec("UPDATE `w_media` SET `directory_id` = ? WHERE `id` = ?", targetID, row.ID).Error; err != nil {
				return err
			}
		}

		staleRows := make([]folderMigrationRow, 0)
		for _, row := range currentRows {
			if _, ok := desiredByKey[folderSourcePathKey(row.SourceType, row.Path)]; ok {
				continue
			}
			staleRows = append(staleRows, row)
		}
		sort.Slice(staleRows, func(i, j int) bool {
			if staleRows[i].Depth == staleRows[j].Depth {
				return staleRows[i].ID > staleRows[j].ID
			}
			return staleRows[i].Depth > staleRows[j].Depth
		})
		for _, row := range staleRows {
			if err := tx.Exec("DELETE FROM `w_folder` WHERE id = ?", row.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func buildDesiredFolderNodes(mediaRows []mediaFolderPathRow) []desiredFolderNode {
	nodesByKey := make(map[string]desiredFolderNode, len(mediaRows)*4)
	for _, row := range mediaRows {
		path := normalizeManagedMediaFolderPath(row.SourceType, row.RootDir, row.FullDir)
		if path == "" {
			continue
		}

		parts := splitFolderPathParts(path)
		if len(parts) == 0 {
			continue
		}

		for i := range parts {
			nodePath := joinFolderPath(parts[:i+1])
			parentPath := ""
			if i > 0 {
				parentPath = joinFolderPath(parts[:i])
			}
			node := desiredFolderNode{
				SourceType: row.SourceType,
				ParentPath: parentPath,
				Name:       parts[i],
				Depth:      int64(i + 1),
				Path:       nodePath,
				PathHash:   folderPathHashBytes(nodePath),
			}
			nodesByKey[folderSourcePathKey(row.SourceType, nodePath)] = node
		}
	}

	nodes := make([]desiredFolderNode, 0, len(nodesByKey))
	for _, node := range nodesByKey {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].SourceType == nodes[j].SourceType {
			if nodes[i].Depth == nodes[j].Depth {
				return nodes[i].Path < nodes[j].Path
			}
			return nodes[i].Depth < nodes[j].Depth
		}
		return nodes[i].SourceType < nodes[j].SourceType
	})
	return nodes
}

func filterManagedMediaFolderRows(rows []mediaFolderPathRow) []mediaFolderPathRow {
	out := make([]mediaFolderPathRow, 0, len(rows))
	for _, row := range rows {
		if normalizeManagedMediaFolderPath(row.SourceType, row.RootDir, row.FullDir) == "" {
			continue
		}
		out = append(out, row)
	}
	return out
}

func normalizeManagedMediaFolderPath(sourceType int64, rootDir, fullDir string) string {
	path := filepath.Clean(strings.TrimSpace(fullDir))
	if path == "" || path == "." || path == string(filepath.Separator) {
		return ""
	}
	if sourceType != consts.WFolderSourceNative {
		return normalizeFolderNodePath(path)
	}

	root := filepath.Clean(strings.TrimSpace(rootDir))
	if root == "" || root == "." || root == string(filepath.Separator) {
		return ""
	}

	mediaBase := filepath.Join(root, "media")
	watchedBase := filepath.Join(root, "watched")
	if isFolderPathWithin(path, mediaBase) || isFolderPathWithin(path, watchedBase) {
		return normalizeFolderNodePath(path)
	}
	return ""
}

func normalizeFolderNodePath(path string) string {
	parts := splitFolderPathParts(path)
	if len(parts) == 0 {
		return ""
	}
	return joinFolderPath(parts)
}

func isFolderPathWithin(pathValue, baseDir string) bool {
	pathValue = filepath.Clean(strings.TrimSpace(pathValue))
	baseDir = filepath.Clean(strings.TrimSpace(baseDir))
	if pathValue == "" || baseDir == "" || pathValue == "." || baseDir == "." {
		return false
	}
	rel, err := filepath.Rel(baseDir, pathValue)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func folderSourcePathKey(sourceType int64, path string) string {
	return fmt.Sprintf("%d:%s", sourceType, path)
}

func splitFolderPathParts(path string) []string {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if cleaned == "" || cleaned == "." || cleaned == string(filepath.Separator) {
		return nil
	}

	cleaned = strings.TrimPrefix(cleaned, string(filepath.Separator))
	if cleaned == "" {
		return nil
	}

	raw := strings.Split(cleaned, string(filepath.Separator))
	out := make([]string, 0, len(raw))
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func joinFolderPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return string(filepath.Separator) + filepath.Join(parts...)
}

func folderPathHashBytes(path string) []byte {
	sum := md5.Sum([]byte(path))
	return sum[:]
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
LEFT JOIN ` + "`w_media` wm" + ` ON wm.movie_jav_id = gss.movie_jav_id AND wm.source_type = ` + fmt.Sprintf("%d", consts.WMediaSourceNative) + `
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

func hasIndexColumns(db *gorm.DB, tableName string, indexName string, columns []string) (bool, error) {
	type indexRow struct {
		SeqInIndex int64  `gorm:"column:SEQ_IN_INDEX"`
		ColumnName string `gorm:"column:COLUMN_NAME"`
	}

	var rows []indexRow
	err := db.Raw(
		"SELECT SEQ_IN_INDEX, COLUMN_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ? ORDER BY SEQ_IN_INDEX ASC",
		tableName,
		indexName,
	).Scan(&rows).Error
	if err != nil {
		return false, err
	}
	if len(rows) != len(columns) {
		return false, nil
	}
	for idx, column := range columns {
		if strings.TrimSpace(rows[idx].ColumnName) != column {
			return false, nil
		}
	}
	return true, nil
}

func batchMigrate(db *gorm.DB, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
