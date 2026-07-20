package dailybestremotex

import (
	"context"
	"errors"

	"gorm.io/gorm"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func normalizeError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dailybesttypes.ErrNotFound
	}
	return err
}
