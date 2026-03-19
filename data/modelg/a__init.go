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
	); err != nil {
		return err
	}

	return nil
}

func batchMigrate(db *gorm.DB, models ...interface{}) error {
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
