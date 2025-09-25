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

// ----------------- 新增：保障 Unknown & 自增起点 -----------------

func setAutoIncrementMin2(tx *gorm.DB, tables ...string) error {
	for _, t := range tables {
		// MySQL 在表已有更大 id 时，不会把自增指针调小，此操作是安全的
		sql := fmt.Sprintf(`ALTER TABLE %s AUTO_INCREMENT = 2`, t)
		if err := tx.Exec(sql).Error; err != nil {
			return fmt.Errorf("set AUTO_INCREMENT for %s failed: %w", t, err)
		}
	}
	return nil
}
