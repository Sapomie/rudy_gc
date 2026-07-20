package dailybestremote

import (
	"rudy_gc/remote/api/internal/modules/dailybest/adapter/outbound/mysql"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	return db.AutoMigrate(
		new(dailybestremotex.FetchRun),
		new(dailybestremotex.Snapshot),
		new(dailybestremotex.SnapshotItem),
		new(dailybestremotex.SnapshotPage),
	)
}
