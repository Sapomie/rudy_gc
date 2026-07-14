package modelg

import (
	"regexp"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	moviedomain "rudy_gc/internal/domain/movie"

	"gorm.io/gorm"
)

var shMagnetInfoHashPattern = regexp.MustCompile(`(?i)btih:([0-9a-z]+)`)

type shMagnetBackfillResource struct {
	InfoHash  string
	MagnetURL string
}

func backfillSHMagnets(db *gorm.DB) error {
	type threadRow struct {
		ThreadTID     string `gorm:"column:thread_tid"`
		MovieJavID    string `gorm:"column:movie_jav_id"`
		MovieName     string `gorm:"column:movie_name"`
		MovieType     string `gorm:"column:movie_type"`
		SourceForumID string `gorm:"column:source_forum_id"`
		PostTime      int64  `gorm:"column:post_time"`
		MagnetText    string `gorm:"column:magnet_text"`
		InfoHashText  string `gorm:"column:info_hash_text"`
		FirstSeenTime int64  `gorm:"column:first_seen_time"`
		LastSeenTime  int64  `gorm:"column:last_seen_time"`
		CreatedOn     int64  `gorm:"column:created_on"`
		UpdatedOn     int64  `gorm:"column:updated_on"`
	}

	var rows []threadRow
	if err := db.Table("sh_thread").
		Select("thread_tid", "movie_jav_id", "movie_name", "movie_type", "source_forum_id", "post_time", "magnet_text", "info_hash_text", "first_seen_time", "last_seen_time", "created_on", "updated_on").
		Where("TRIM(COALESCE(info_hash_text, '')) <> '' OR TRIM(COALESCE(magnet_text, '')) <> ''").
		Find(&rows).Error; err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, row := range rows {
		createdOn := firstPositiveInt64(row.CreatedOn, row.FirstSeenTime, row.UpdatedOn, row.LastSeenTime, now)
		updatedOn := firstPositiveInt64(row.UpdatedOn, row.LastSeenTime, createdOn)
		for _, resource := range buildSHMagnetBackfillResources(row.InfoHashText, row.MagnetText) {
			if err := upsertSHMagnetMigrationRow(db, &ShMagnet{
				ThreadTID:     strings.TrimSpace(row.ThreadTID),
				MovieJavID:    strings.TrimSpace(row.MovieJavID),
				MovieName:     strings.TrimSpace(row.MovieName),
				MovieType:     strings.TrimSpace(row.MovieType),
				SourceForumID: strings.TrimSpace(row.SourceForumID),
				PostTime:      row.PostTime,
				MagnetURL:     resource.MagnetURL,
				InfoHash:      resource.InfoHash,
				CreatedOn:     createdOn,
				UpdatedOn:     updatedOn,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func upsertSHMagnetMigrationRow(db *gorm.DB, row *ShMagnet) error {
	return db.Exec(`
INSERT INTO sh_magnet (
	thread_tid, movie_jav_id, movie_name, movie_type, source_forum_id,
	post_time, magnet_url, info_hash, created_on, updated_on
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	thread_tid = CASE WHEN TRIM(VALUES(thread_tid)) <> '' THEN VALUES(thread_tid) ELSE thread_tid END,
	movie_jav_id = CASE WHEN TRIM(VALUES(movie_jav_id)) <> '' THEN VALUES(movie_jav_id) ELSE movie_jav_id END,
	movie_name = CASE WHEN TRIM(VALUES(movie_name)) <> '' THEN VALUES(movie_name) ELSE movie_name END,
	movie_type = CASE WHEN TRIM(VALUES(movie_type)) <> '' THEN VALUES(movie_type) ELSE movie_type END,
	source_forum_id = CASE WHEN TRIM(VALUES(source_forum_id)) <> '' THEN VALUES(source_forum_id) ELSE source_forum_id END,
	post_time = CASE WHEN VALUES(post_time) > 0 THEN VALUES(post_time) ELSE post_time END,
	magnet_url = CASE WHEN TRIM(VALUES(magnet_url)) <> '' THEN VALUES(magnet_url) ELSE magnet_url END,
	updated_on = GREATEST(updated_on, VALUES(updated_on))`,
		row.ThreadTID, row.MovieJavID, row.MovieName, row.MovieType, row.SourceForumID,
		row.PostTime, row.MagnetURL, row.InfoHash, row.CreatedOn, row.UpdatedOn,
	).Error
}

func migrateOptimizeSHMagnetColumns(db *gorm.DB) error {
	columns := []string{"tag", "thread_title", "thread_url", "post_date", "first_seen_time", "last_seen_time"}
	dropClauses := make([]string, 0, len(columns))
	for _, column := range columns {
		has, err := hasColumn(db, "sh_magnet", column)
		if err != nil {
			return err
		}
		if has {
			dropClauses = append(dropClauses, "DROP COLUMN `"+column+"`")
		}
	}
	if len(dropClauses) == 0 {
		return nil
	}
	return db.Exec("ALTER TABLE `sh_magnet` " + strings.Join(dropClauses, ", ")).Error
}

func migrateSHMagnetDownloadState(db *gorm.DB) error {
	const tableName = "sh_magnet"
	hasTableValue, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTableValue {
		return nil
	}

	hasColumnValue, err := hasColumn(db, tableName, "download_state")
	if err != nil {
		return err
	}
	if !hasColumnValue {
		if err := db.Exec("ALTER TABLE `sh_magnet` ADD COLUMN `download_state` tinyint NOT NULL DEFAULT 0 AFTER `info_hash`").Error; err != nil {
			return err
		}
	}

	hasIndexValue, err := hasIndex(db, tableName, "idx_sh_magnet_download_state")
	if err != nil {
		return err
	}
	if !hasIndexValue {
		if err := db.Exec("ALTER TABLE `sh_magnet` ADD INDEX `idx_sh_magnet_download_state` (`download_state`)").Error; err != nil {
			return err
		}
	}

	return db.Exec(`
UPDATE sh_magnet m
SET m.download_state = CASE
	WHEN LOWER(TRIM(m.movie_type)) = ? THEN CASE
		WHEN EXISTS (
			SELECT 1 FROM fc2_media fm_active
			WHERE fm_active.source_torrent_hash = m.info_hash
			  AND fm_active.is_removed = ?
		) THEN ?
		WHEN EXISTS (
			SELECT 1 FROM fc2_media fm_any
			WHERE fm_any.source_torrent_hash = m.info_hash
		) THEN ?
		ELSE ?
	END
	ELSE CASE
		WHEN EXISTS (
			SELECT 1 FROM w_media wm_active
			WHERE wm_active.source_type = ?
			  AND wm_active.source_torrent_hash = m.info_hash
			  AND wm_active.is_removed = ?
		) THEN ?
		WHEN EXISTS (
			SELECT 1 FROM w_media wm_any
			WHERE wm_any.source_type = ?
			  AND wm_any.source_torrent_hash = m.info_hash
		) THEN ?
		ELSE ?
	END
END`,
		moviedomain.SehuatangThreadMovieTypeFC2PPV,
		consts.FilmIsNotRemoved, consts.SHMagnetDownloadStateDownloaded,
		consts.SHMagnetDownloadStateRemoved, consts.SHMagnetDownloadStateMissing,
		consts.WMediaSourceNative, consts.FilmIsNotRemoved, consts.SHMagnetDownloadStateDownloaded,
		consts.WMediaSourceNative, consts.SHMagnetDownloadStateRemoved, consts.SHMagnetDownloadStateMissing,
	).Error
}

func buildSHMagnetBackfillResources(infoHashText, magnetText string) []shMagnetBackfillResource {
	urlsByHash := make(map[string]string)
	orderedHashes := make([]string, 0)
	seen := make(map[string]struct{})
	addHash := func(value string) {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		orderedHashes = append(orderedHashes, value)
	}

	for _, value := range splitSHMagnetLines(infoHashText) {
		addHash(value)
	}
	magnetLines := splitSHMagnetLines(magnetText)
	for _, magnetURL := range magnetLines {
		matches := shMagnetInfoHashPattern.FindStringSubmatch(magnetURL)
		if len(matches) < 2 {
			continue
		}
		infoHash := strings.ToUpper(strings.TrimSpace(matches[1]))
		addHash(infoHash)
		if _, ok := urlsByHash[infoHash]; !ok {
			urlsByHash[infoHash] = magnetURL
		}
	}

	out := make([]shMagnetBackfillResource, 0, len(orderedHashes))
	for index, infoHash := range orderedHashes {
		magnetURL := urlsByHash[infoHash]
		if magnetURL == "" && index < len(magnetLines) {
			magnetURL = magnetLines[index]
		}
		out = append(out, shMagnetBackfillResource{InfoHash: infoHash, MagnetURL: magnetURL})
	}
	return out
}

func splitSHMagnetLines(value string) []string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func firstPositiveInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
