package dep

import (
	"strings"
	"time"

	"gorm.io/gorm"

	"rudy_gc/remote/api/internal/config"
	"rudy_gc/remote/api/internal/model/modelg"
	dailybestremotex "rudy_gc/remote/api/internal/model/modelx/dailybest"
	"rudy_gc/remote/api/internal/notify"
	dailybestrepo "rudy_gc/remote/api/internal/repo/dailybest"
)

type Dep struct {
	Config              config.Config
	DB                  *gorm.DB
	DailyBestRepository *dailybestremotex.Repository
	DailyBestBrowser    *dailybestrepo.FlareSolverrClient
	DailyBestSource     *dailybestrepo.Source
	Notifier            notify.Notifier
	Location            *time.Location
}

func BuildDep(cfg config.Config) (*Dep, error) {
	if err := cfg.Mail.Validate(); err != nil {
		return nil, err
	}
	mailNotifier, err := notify.NewSMTP(cfg.Mail)
	if err != nil {
		return nil, err
	}
	db, err := modelg.NewGormDBEngine(&modelg.Config{
		DSN:                  strings.TrimSpace(cfg.DataSource),
		LogLevel:             strings.TrimSpace(cfg.LogursLevel),
		IgnoreRecordNotFound: true,
		DisableColor:         true,
	})
	if err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(strings.TrimSpace(cfg.Timezone))
	if err != nil {
		closeDB(db)
		return nil, err
	}
	repository := dailybestremotex.New(db)
	browser, err := dailybestrepo.NewFlareSolverrClient(dailybestrepo.FlareSolverrClientConfig{
		BaseURL: cfg.Source.Browser.BaseURL,
		Session: cfg.Source.Browser.Session,
		Timeout: time.Duration(cfg.Source.Timeout) * time.Second,
	})
	if err != nil {
		closeDB(db)
		return nil, err
	}
	client := dailybestrepo.NewClient(browser, repository)
	return &Dep{
		Config:              cfg,
		DB:                  db,
		DailyBestRepository: repository,
		DailyBestBrowser:    browser,
		DailyBestSource:     dailybestrepo.NewSource(client, ""),
		Notifier:            mailNotifier,
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
