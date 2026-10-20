package modelg

import (
	"fmt"

	"gorm.io/gorm"
)

type gScOptionTableRenameSpec struct {
	oldTable       string
	newTable       string
	oldUniqueIndex string
	newUniqueIndex string
	oldListIndex   string
	newListIndex   string
}

func gScOptionTableRenameSpecs() []gScOptionTableRenameSpec {
	return []gScOptionTableRenameSpec{
		{
			oldTable:       "g_sc_kind_option",
			newTable:       GScKindOptionTableName,
			oldUniqueIndex: "uk_g_sc_kind_option_name",
			newUniqueIndex: "uk_g_sc_option_kind_name",
			oldListIndex:   "idx_g_sc_kind_option_list",
			newListIndex:   "idx_g_sc_option_kind_list",
		},
		{
			oldTable:       "g_sc_fg_option",
			newTable:       GScFGOptionTableName,
			oldUniqueIndex: "uk_g_sc_fg_option_name",
			newUniqueIndex: "uk_g_sc_option_fg_name",
			oldListIndex:   "idx_g_sc_fg_option_list",
			newListIndex:   "idx_g_sc_option_fg_list",
		},
		{
			oldTable:       "g_sc_vessel_option",
			newTable:       GScVesselOptionTableName,
			oldUniqueIndex: "uk_g_sc_vessel_option_name",
			newUniqueIndex: "uk_g_sc_option_vessl_name",
			oldListIndex:   "idx_g_sc_vessel_option_list",
			newListIndex:   "idx_g_sc_option_vessl_list",
		},
	}
}

func migrateRenameGScOptionTables(db *gorm.DB) error {
	for _, spec := range gScOptionTableRenameSpecs() {
		hasOld, err := hasTable(db, spec.oldTable)
		if err != nil {
			return err
		}
		hasNew, err := hasTable(db, spec.newTable)
		if err != nil {
			return err
		}
		if hasOld && hasNew {
			return fmt.Errorf("SC 选项旧表 %s 与新表 %s 同时存在，停止表名迁移", spec.oldTable, spec.newTable)
		}
		if hasOld {
			if err := db.Migrator().RenameTable(spec.oldTable, spec.newTable); err != nil {
				return fmt.Errorf("重命名 SC 选项表 %s 为 %s 失败: %w", spec.oldTable, spec.newTable, err)
			}
			hasNew = true
		}
		if !hasNew {
			continue
		}
		if err := migrateRenameIndexIfPresent(db, spec.newTable, spec.oldUniqueIndex, spec.newUniqueIndex); err != nil {
			return fmt.Errorf("重命名 %s 唯一索引失败: %w", spec.newTable, err)
		}
		if err := migrateRenameIndexIfPresent(db, spec.newTable, spec.oldListIndex, spec.newListIndex); err != nil {
			return fmt.Errorf("重命名 %s 列表索引失败: %w", spec.newTable, err)
		}
	}
	return nil
}

func backfillGScOptionAggregates(db *gorm.DB) error {
	for _, spec := range gScOptionSplitSpecs() {
		if err := requireGScOptionTableColumns(db, spec.tableName, []string{
			"id", "sc_event_times", "last_sc_event_time",
		}); err != nil {
			return err
		}

		backfillSQL := fmt.Sprintf(`
UPDATE %s option_row
LEFT JOIN (
    SELECT %s AS option_id,
           COUNT(*) AS sc_event_times,
           COALESCE(MAX(sc_time), 0) AS last_sc_event_time
    FROM g_sc
    WHERE %s IS NOT NULL
    GROUP BY %s
) aggregate_row ON aggregate_row.option_id = option_row.id
SET option_row.sc_event_times = COALESCE(aggregate_row.sc_event_times, 0),
    option_row.last_sc_event_time = COALESCE(aggregate_row.last_sc_event_time, 0)`,
			quoteGScOptionIdentifier(spec.tableName),
			quoteGScOptionIdentifier(spec.idColumn),
			quoteGScOptionIdentifier(spec.idColumn),
			quoteGScOptionIdentifier(spec.idColumn),
		)
		if err := db.Exec(backfillSQL).Error; err != nil {
			return fmt.Errorf("回填 SC %s 选项聚合数据失败: %w", spec.optionType, err)
		}

		validationSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s option_row
LEFT JOIN (
    SELECT %s AS option_id,
           COUNT(*) AS sc_event_times,
           COALESCE(MAX(sc_time), 0) AS last_sc_event_time
    FROM g_sc
    WHERE %s IS NOT NULL
    GROUP BY %s
) aggregate_row ON aggregate_row.option_id = option_row.id
WHERE option_row.sc_event_times <> COALESCE(aggregate_row.sc_event_times, 0)
   OR option_row.last_sc_event_time <> COALESCE(aggregate_row.last_sc_event_time, 0)`,
			quoteGScOptionIdentifier(spec.tableName),
			quoteGScOptionIdentifier(spec.idColumn),
			quoteGScOptionIdentifier(spec.idColumn),
			quoteGScOptionIdentifier(spec.idColumn),
		)
		var mismatchCount int64
		if err := db.Raw(validationSQL).Scan(&mismatchCount).Error; err != nil {
			return fmt.Errorf("校验 SC %s 选项聚合数据失败: %w", spec.optionType, err)
		}
		if mismatchCount > 0 {
			return fmt.Errorf("SC %s 选项存在 %d 条聚合数据不一致记录", spec.optionType, mismatchCount)
		}
	}
	return nil
}
