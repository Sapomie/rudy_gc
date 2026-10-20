package modelg

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type gScOptionSplitSpec struct {
	optionType string
	tableName  string
	idColumn   string
}

func gScOptionSplitSpecs() []gScOptionSplitSpec {
	return []gScOptionSplitSpec{
		{optionType: "kind", tableName: GScKindOptionTableName, idColumn: "kind_option_id"},
		{optionType: "fg", tableName: GScFGOptionTableName, idColumn: "fg_option_id"},
		{optionType: "vessel", tableName: GScVesselOptionTableName, idColumn: "vessel_option_id"},
	}
}

func migrateSplitGScOptions(db *gorm.DB) error {
	specs := gScOptionSplitSpecs()
	for _, spec := range specs {
		if err := requireGScOptionTableColumns(db, spec.tableName, []string{
			"id", "name", "sort_order", "is_enabled", "sc_event_times", "last_sc_event_time", "created_on", "updated_on",
		}); err != nil {
			return err
		}
	}
	if err := requireGScOptionTableColumns(db, ScTableName, []string{
		"kind_option_id", "fg_option_id", "vessel_option_id",
	}); err != nil {
		return err
	}

	hasLegacyTable, err := hasTable(db, gScOptionLegacyTableName)
	if err != nil {
		return err
	}
	if hasLegacyTable {
		if err := requireGScOptionTableColumns(db, gScOptionLegacyTableName, []string{
			"id", "option_type", "option_name", "sort_order", "is_enabled", "created_on", "updated_on",
		}); err != nil {
			return err
		}
		if err := validateLegacyGScOptionTypes(db); err != nil {
			return err
		}
		for _, spec := range specs {
			if err := copyAndValidateGScOptions(db, spec); err != nil {
				return err
			}
		}
	}

	for _, spec := range specs {
		if err := validateGScOptionReferences(db, spec); err != nil {
			return err
		}
	}
	if !hasLegacyTable {
		return nil
	}
	if err := db.Migrator().DropTable(gScOptionLegacyTableName); err != nil {
		return fmt.Errorf("删除旧 SC 选项表失败: %w", err)
	}
	return nil
}

func requireGScOptionTableColumns(db *gorm.DB, tableName string, columns []string) error {
	hasTargetTable, err := hasTable(db, tableName)
	if err != nil {
		return err
	}
	if !hasTargetTable {
		return fmt.Errorf("缺少必要表 %s，停止 SC 选项拆分迁移", tableName)
	}
	for _, column := range columns {
		hasTargetColumn, err := hasColumn(db, tableName, column)
		if err != nil {
			return err
		}
		if !hasTargetColumn {
			return fmt.Errorf("%s 缺少必要字段 %s，停止 SC 选项拆分迁移", tableName, column)
		}
	}
	return nil
}

func validateLegacyGScOptionTypes(db *gorm.DB) error {
	var unknownTypeCount int64
	err := db.Raw(`
SELECT COUNT(*)
FROM g_sc_option
WHERE option_type IS NULL
   OR option_type NOT IN (?, ?, ?)`, "kind", "fg", "vessel").Scan(&unknownTypeCount).Error
	if err != nil {
		return fmt.Errorf("校验旧 SC 选项类型失败: %w", err)
	}
	if unknownTypeCount > 0 {
		return fmt.Errorf("旧 SC 选项表存在 %d 条未知类型数据，停止拆分迁移", unknownTypeCount)
	}
	return nil
}

func copyAndValidateGScOptions(db *gorm.DB, spec gScOptionSplitSpec) error {
	copySQL := fmt.Sprintf(`
INSERT IGNORE INTO %s
    (id, name, sort_order, is_enabled, created_on, updated_on)
SELECT id, option_name, sort_order, is_enabled, created_on, updated_on
FROM g_sc_option
WHERE option_type = ?`, quoteGScOptionIdentifier(spec.tableName))
	if err := db.Exec(copySQL, spec.optionType).Error; err != nil {
		return fmt.Errorf("复制 SC %s 选项失败: %w", spec.optionType, err)
	}

	validationSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM g_sc_option source_row
LEFT JOIN %s target_row ON target_row.id = source_row.id
WHERE source_row.option_type = ?
  AND (
    target_row.id IS NULL
    OR NOT (target_row.name <=> source_row.option_name)
    OR NOT (target_row.sort_order <=> source_row.sort_order)
    OR NOT (target_row.is_enabled <=> source_row.is_enabled)
    OR NOT (target_row.created_on <=> source_row.created_on)
    OR NOT (target_row.updated_on <=> source_row.updated_on)
  )`, quoteGScOptionIdentifier(spec.tableName))
	var mismatchCount int64
	if err := db.Raw(validationSQL, spec.optionType).Scan(&mismatchCount).Error; err != nil {
		return fmt.Errorf("校验 SC %s 选项拆分结果失败: %w", spec.optionType, err)
	}
	if mismatchCount > 0 {
		return fmt.Errorf("SC %s 选项源表与目标表存在 %d 条字段不一致数据，停止拆分迁移", spec.optionType, mismatchCount)
	}

	extraTargetSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s target_row
LEFT JOIN g_sc_option source_row
  ON source_row.id = target_row.id
 AND source_row.option_type = ?
WHERE source_row.id IS NULL`, quoteGScOptionIdentifier(spec.tableName))
	var extraTargetCount int64
	if err := db.Raw(extraTargetSQL, spec.optionType).Scan(&extraTargetCount).Error; err != nil {
		return fmt.Errorf("校验 SC %s 选项目标表额外数据失败: %w", spec.optionType, err)
	}
	if extraTargetCount > 0 {
		return fmt.Errorf("SC %s 选项目标表存在 %d 条源表中不存在的数据，停止拆分迁移", spec.optionType, extraTargetCount)
	}
	return nil
}

func validateGScOptionReferences(db *gorm.DB, spec gScOptionSplitSpec) error {
	validationSQL := fmt.Sprintf(`
SELECT COUNT(*)
FROM g_sc sc
LEFT JOIN %s option_row ON option_row.id = sc.%s
WHERE sc.%s IS NOT NULL
  AND option_row.id IS NULL`,
		quoteGScOptionIdentifier(spec.tableName),
		quoteGScOptionIdentifier(spec.idColumn),
		quoteGScOptionIdentifier(spec.idColumn),
	)
	var invalidReferenceCount int64
	if err := db.Raw(validationSQL).Scan(&invalidReferenceCount).Error; err != nil {
		return fmt.Errorf("校验 g_sc 的 %s 选项引用失败: %w", spec.optionType, err)
	}
	if invalidReferenceCount > 0 {
		return fmt.Errorf("g_sc 存在 %d 条无效的 %s 选项引用，停止删除旧 SC 选项表", invalidReferenceCount, spec.optionType)
	}
	return nil
}

func quoteGScOptionIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
