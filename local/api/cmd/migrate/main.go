package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/local/api/internal/config"
	"rudy_gc/local/api/internal/model/modelg"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	fetchsiterepo "rudy_gc/local/api/internal/repo/fetchsite"
)

var configFile = flag.String("f", "config/config.yaml", "the config file")
var javlibOnly = flag.Bool("javlib-only", false, "仅初始化 Javlib 站点配置，保留已有配置，不执行其他迁移")

func main() {
	flag.Parse()

	c := config.MustLoad(*configFile)
	dsn := strings.TrimSpace(c.DataSource)
	if dsn == "" {
		log.Fatal("数据库迁移失败：配置中的 DataSource 为空")
	}
	initJavlib := func() error {
		logx.Disable()
		model := moviex.NewTFetchSiteModel(sqlx.NewMysql(dsn), c.Cache)
		return fetchsiterepo.EnsureJavlibConfig(context.Background(), c, model)
	}
	if *javlibOnly {
		if err := initJavlib(); err != nil {
			log.Fatalf("Javlib 站点配置初始化失败：%v", err)
		}
		log.Println("Javlib 站点配置初始化完成")
		return
	}

	db, err := modelg.NewGormDBEngine(&modelg.Config{
		DSN:      dsn,
		LogLevel: strings.TrimSpace(c.LogursLevel),
	})
	if err != nil {
		log.Fatalf("数据库迁移失败：连接数据库失败：%v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("数据库迁移失败：获取数据库连接失败：%v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败：%v", err)
		}
	}()

	log.Printf("开始执行数据库迁移，配置文件：%s", *configFile)
	if err := modelg.AutoMigrate(db); err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：%w", err))
	}
	if err := initJavlib(); err != nil {
		log.Fatalf("Javlib 站点配置初始化失败：%v", err)
	}
	var shMagnetColumnCount int64
	if err := db.Raw(`
SELECT COUNT(*)
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sh_magnet'`).Scan(&shMagnetColumnCount).Error; err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：统计 sh_magnet 字段失败：%w", err))
	}
	if shMagnetColumnCount != 12 {
		log.Fatalf("数据库迁移失败：sh_magnet 字段数量=%d，预期=12", shMagnetColumnCount)
	}
	var sourceThreadCount int64
	if err := db.Table("sh_thread").Where("TRIM(COALESCE(info_hash_text, '')) <> '' OR TRIM(COALESCE(magnet_text, '')) <> ''").Count(&sourceThreadCount).Error; err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：统计 SH 磁力来源帖子失败：%w", err))
	}
	var sourceHashOccurrenceCount int64
	if err := db.Raw(`
SELECT COALESCE(SUM(
	CASE
		WHEN TRIM(COALESCE(info_hash_text, '')) = '' THEN 0
		ELSE 1 + LENGTH(TRIM(info_hash_text)) - LENGTH(REPLACE(TRIM(info_hash_text), '\n', ''))
	END
), 0)
FROM sh_thread`).Scan(&sourceHashOccurrenceCount).Error; err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：统计 SH hash 出现次数失败：%w", err))
	}
	type shMagnetStats struct {
		CountAll        int64 `gorm:"column:count_all"`
		CountWithURL    int64 `gorm:"column:count_with_url"`
		CountWithTID    int64 `gorm:"column:count_with_tid"`
		CountWithForum  int64 `gorm:"column:count_with_forum"`
		CountWithMovie  int64 `gorm:"column:count_with_movie"`
		CountWithType   int64 `gorm:"column:count_with_type"`
		CountWithPost   int64 `gorm:"column:count_with_post"`
		CountMissing    int64 `gorm:"column:count_missing"`
		CountDownloaded int64 `gorm:"column:count_downloaded"`
		CountRemoved    int64 `gorm:"column:count_removed"`
	}
	var magnetStats shMagnetStats
	if err := db.Raw(`
SELECT
	COUNT(*) AS count_all,
	COALESCE(SUM(CASE WHEN TRIM(COALESCE(magnet_url, '')) <> '' THEN 1 ELSE 0 END), 0) AS count_with_url,
	COALESCE(SUM(CASE WHEN TRIM(COALESCE(thread_tid, '')) <> '' THEN 1 ELSE 0 END), 0) AS count_with_tid,
	COALESCE(SUM(CASE WHEN TRIM(COALESCE(source_forum_id, '')) <> '' THEN 1 ELSE 0 END), 0) AS count_with_forum,
	COALESCE(SUM(CASE WHEN TRIM(COALESCE(movie_name, '')) <> '' THEN 1 ELSE 0 END), 0) AS count_with_movie,
	COALESCE(SUM(CASE WHEN TRIM(COALESCE(movie_type, '')) <> '' THEN 1 ELSE 0 END), 0) AS count_with_type,
	COALESCE(SUM(CASE WHEN post_time > 0 THEN 1 ELSE 0 END), 0) AS count_with_post,
	COALESCE(SUM(CASE WHEN download_state = 0 THEN 1 ELSE 0 END), 0) AS count_missing,
	COALESCE(SUM(CASE WHEN download_state = 1 THEN 1 ELSE 0 END), 0) AS count_downloaded,
	COALESCE(SUM(CASE WHEN download_state = 2 THEN 1 ELSE 0 END), 0) AS count_removed
FROM sh_magnet`).Scan(&magnetStats).Error; err != nil {
		log.Fatal(fmt.Errorf("数据库迁移失败：统计 sh_magnet 失败：%w", err))
	}
	log.Printf(
		"SH 磁力回填核对：来源帖子=%d，hash 出现次数=%d，sh_magnet 去重行数=%d，含 magnet_url=%d，含 thread_tid=%d，含论坛=%d，含影片名=%d，含影片类型=%d，含 post_time=%d",
		sourceThreadCount, sourceHashOccurrenceCount, magnetStats.CountAll, magnetStats.CountWithURL,
		magnetStats.CountWithTID, magnetStats.CountWithForum, magnetStats.CountWithMovie, magnetStats.CountWithType, magnetStats.CountWithPost,
	)
	log.Printf("SH 磁力下载状态回填核对：未下载=%d，已下载=%d，已移除=%d", magnetStats.CountMissing, magnetStats.CountDownloaded, magnetStats.CountRemoved)
	log.Printf("SH 磁力表结构核对：字段数量=%d", shMagnetColumnCount)
	log.Println("数据库迁移完成")
}
