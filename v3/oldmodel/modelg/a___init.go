package modelg

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormDBEngine 初始化一个带有数据模型自动迁移的 GORM 数据库连接。
// dns - 用于连接 MySQL 数据库的数据源名称。
// logLevel - GORM 日志记录器的日志级别。
func NewGormDBEngine(dns string, logLevel string) (*gorm.DB, error) {
	db, err := initializeDatabase(dns, logLevel)
	if err != nil {
		return nil, err
	}

	if err := autoMigrateModels(db); err != nil {
		return nil, err
	}

	return db, nil
}

// NewGormDBEngineRemote 初始化一个不带自动迁移的 GORM 数据库连接。
func NewGormDBEngineRemote(dns string, logLevel string) (*gorm.DB, error) {
	return initializeDatabase(dns, logLevel)
}

// initializeDatabase 使用 GORM 打开一个 MySQL 数据库连接，
// 并根据提供的日志级别配置日志记录器。
func initializeDatabase(dns, logLevel string) (*gorm.DB, error) {
	lg := logger.Default.LogMode(getLogLevel(logLevel))

	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{Logger: lg})
	if err != nil {
		return nil, err // 如果连接失败，则返回错误
	}
	return db, nil
}

// autoMigrateModels 自动迁移所有指定模型的数据库架构。
func autoMigrateModels(db *gorm.DB) error {
	models := []interface{}{
		new(Inventory),
		new(Bestinv),
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}

// getLogLevel 将字符串日志级别转换为对应的 logger.LogLevel 常量。
func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "info":
		return logger.Info
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "silent":
		return logger.Silent
	default:
		return logger.Info // 如果没有提供有效的日志级别，则默认使用 info 级别
	}
}
