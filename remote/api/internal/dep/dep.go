package dep

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"rudy_gc/remote/api/internal/config"
	"rudy_gc/remote/api/internal/model/modelg"
	dailybestremotex "rudy_gc/remote/api/internal/model/modelx/dailybest"
	dailybestrepo "rudy_gc/remote/api/internal/repo/dailybest"
)

type Dep struct {
	Config              config.Config
	DB                  *gorm.DB
	DailyBestRepository *dailybestremotex.Repository
	DailyBestSource     *dailybestrepo.Source
	Location            *time.Location
}

func BuildDep(cfg config.Config) (*Dep, error) {
	db, err := modelg.NewGormDBEngine(&modelg.Config{
		DSN:      strings.TrimSpace(cfg.DataSource),
		LogLevel: strings.TrimSpace(cfg.LogursLevel),
	})
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		closeDB(db)
		return nil, err
	}
	client := dailybestrepo.NewClient(dailybestrepo.Config{
		UserAgent: cfg.Source.UserAgent,
		Cookie:    cfg.Source.Cookie,
		Timeout:   time.Duration(cfg.Source.Timeout) * time.Second,
	})
	return &Dep{
		Config:              cfg,
		DB:                  db,
		DailyBestRepository: dailybestremotex.New(db),
		DailyBestSource:     dailybestrepo.NewSource(client, cfg.Source.JavAddress),
		Location:            loc,
	}, nil
}

func (d *Dep) Close() error {
	if d == nil {
		return nil
	}
	return closeDB(d.DB)
}

func closeDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
