package modelg

import (
	"fmt"

	"gorm.io/gorm"
)

// ------- 原有入口保持不变 -------

func MustAutoMigrate(db *gorm.DB) {
	if err := AutoMigrate(db); err != nil {
		panic(fmt.Sprintf("auto migrate err:%v", err))
	}
}

func AutoMigrate(db *gorm.DB) error {
	if err := migrateLegacyMovieNameColumns(db); err != nil {
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
		new(Media),
		new(FetchSite),
		new(Album),
		new(AlbumItem),
		new(JavbusMagnet),
		new(JavbusMagnetFetch),
		new(SukebeiTorrent),
		new(SukebeiTorrentFetch),
	); err != nil {
		return err
	}

	return nil
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

func batchMigrate(db *gorm.DB, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
