package orm

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 提示：使用 time.Duration，直观且避免单位误解
type Config struct {
	DSN      string
	LogLevel string // "silent" | "error" | "warn" | "info"

	// 连接池
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// 日志增强
	SlowThreshold        time.Duration // 慢查询阈值
	IgnoreRecordNotFound bool          // 忽略 RecordNotFound 记日志
	DisableColor         bool          // 关闭彩色日志
	LoggerOutput         *log.Logger   // 自定义 writer；默认使用 os.Stdout

	// 高阶：可传入自定义 gorm.Config（如 NamingStrategy 等）
	GormConfig *gorm.Config
}

func MustNewGormDBEngine(c *Config) *gorm.DB {
	db, err := NewGormDBEngine(c)
	if err != nil {
		panic(fmt.Sprintf("gorm start err: %v", err))
	}
	return db
}

func NewGormDBEngine(cfg *Config) (*gorm.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("empty DSN")
	}
	c := cfg.withDefaults()

	// 构造 logger（比默认更可控）
	gl := logger.New(
		c.LoggerOutput,
		logger.Config{
			SlowThreshold:             c.SlowThreshold,
			LogLevel:                  parseLogLevel(c.LogLevel),
			IgnoreRecordNotFoundError: c.IgnoreRecordNotFound,
			Colorful:                  !c.DisableColor,
		},
	)

	gormConf := &gorm.Config{
		Logger: gl,
	}
	// 允许外部注入更高级的 gorm.Config
	if c.GormConfig != nil {
		*gormConf = *c.GormConfig
		// 确保我们上面的 logger 能生效（除非外部也指定了）
		if gormConf.Logger == nil {
			gormConf.Logger = gl
		}
	}

	db, err := gorm.Open(mysql.Open(c.DSN), gormConf)
	if err != nil {
		return nil, err
	}

	sdb, err := db.DB()
	if err != nil {
		return nil, err
	}
	sdb.SetMaxIdleConns(c.MaxIdleConns)
	sdb.SetMaxOpenConns(c.MaxOpenConns)
	sdb.SetConnMaxLifetime(c.ConnMaxLifetime)
	// Go 1.17+ 可用：为空闲连接设置独立过期时间
	if setConnMaxIdleTime := setMaxIdleTimeFunc(sdb); setConnMaxIdleTime != nil {
		setConnMaxIdleTime(c.ConnMaxIdleTime)
	}

	// 启动时做一次 Ping（带超时），尽早暴露配置/网络问题
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sdb.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// 默认值统一放这里
func (c *Config) withDefaults() *Config {
	cp := *c
	if cp.MaxIdleConns == 0 {
		cp.MaxIdleConns = 10
	}
	if cp.MaxOpenConns == 0 {
		cp.MaxOpenConns = 100
	}
	if cp.ConnMaxLifetime == 0 {
		cp.ConnMaxLifetime = time.Hour
	}
	// 可选：默认 15 分钟空闲过期（Go 1.17+）
	if cp.ConnMaxIdleTime == 0 {
		cp.ConnMaxIdleTime = 15 * time.Minute
	}
	if cp.SlowThreshold == 0 {
		cp.SlowThreshold = 200 * time.Millisecond
	}
	if cp.LogLevel == "" {
		cp.LogLevel = "warn"
	}
	if cp.LoggerOutput == nil {
		cp.LoggerOutput = log.New(os.Stdout, "", log.LstdFlags)
	}
	return &cp
}

func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn", "warning":
		return logger.Warn
	case "silent":
		return logger.Silent
	default:
		return logger.Warn
	}
}

// 为了兼容旧 Go 版本，如果不存在该方法就不设置
func setMaxIdleTimeFunc(db *sql.DB) func(time.Duration) {
	type setIdle interface{ SetConnMaxIdleTime(time.Duration) }
	if v, ok := any(db).(setIdle); ok {
		return v.SetConnMaxIdleTime
	}
	return nil
}
