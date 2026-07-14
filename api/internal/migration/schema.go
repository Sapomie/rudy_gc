package migration

import (
	"rudy_gc/internal/schema"

	"gorm.io/gorm"
)

// Run executes schema migrations in their established order.
func Run(db *gorm.DB) error {
	return schema.AutoMigrate(db)
}
