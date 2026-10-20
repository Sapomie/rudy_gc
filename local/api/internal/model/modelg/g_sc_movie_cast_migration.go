package modelg

import (
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type legacyGScMovieCastRow struct {
	ID                int64         `gorm:"column:id"`
	MovieCast         string        `gorm:"column:movie_cast"`
	MovieCastPersonID sql.NullInt64 `gorm:"column:movie_cast_person_id"`
}

func (legacyGScMovieCastRow) TableName() string { return ScTableName }

type gScMovieCastResolver struct {
	castNames    map[string]int64
	chineseNames map[string]int64
	personNames  map[string]int64
	aliases      map[string]int64
}

func migrateGScMovieCastPersonID(db *gorm.DB) error {
	if !db.Migrator().HasTable(&Sc{}) || !db.Migrator().HasColumn(&Sc{}, "MovieCastPersonID") {
		return nil
	}
	if !db.Migrator().HasColumn(&legacyGScMovieCastRow{}, "MovieCast") {
		return validateGScMovieCastPersonReferences(db)
	}

	var rows []*legacyGScMovieCastRow
	if err := db.Table(ScTableName).Select("id", "movie_cast", "movie_cast_person_id").Order("id ASC").Scan(&rows).Error; err != nil {
		return fmt.Errorf("读取 g_sc 演员历史数据失败: %w", err)
	}
	resolver, err := loadGScMovieCastResolver(db)
	if err != nil {
		return err
	}

	resolved := make(map[int64]int64)
	unresolved := make([]string, 0)
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.MovieCastPersonID.Valid && row.MovieCastPersonID.Int64 > 0 {
			continue
		}
		legacyName := normalizeLegacyGScMovieCastName(row.MovieCast)
		if legacyName == "" {
			continue
		}
		personID := resolver.resolve(legacyName)
		if personID <= 0 {
			unresolved = append(unresolved, fmt.Sprintf("id=%d 演员=%q", row.ID, row.MovieCast))
			continue
		}
		resolved[row.ID] = personID
	}
	if len(unresolved) > 0 {
		return fmt.Errorf("g_sc 演员迁移存在无法唯一映射的历史数据: %s", strings.Join(unresolved, "; "))
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for eventID, personID := range resolved {
			if err := tx.Table(ScTableName).Where("id = ?", eventID).Update("movie_cast_person_id", personID).Error; err != nil {
				return fmt.Errorf("更新 g_sc(%d) 演员 person_id 失败: %w", eventID, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := validateGScMovieCastPersonReferences(db); err != nil {
		return err
	}
	if err := db.Migrator().DropColumn(&legacyGScMovieCastRow{}, "MovieCast"); err != nil {
		return fmt.Errorf("删除 g_sc.movie_cast 失败: %w", err)
	}
	return nil
}

func validateGScMovieCastPersonReferences(db *gorm.DB) error {
	var invalidCount int64
	if err := db.Table(ScTableName + " AS sc").
		Joins("LEFT JOIN c_person AS person ON person.id = sc.movie_cast_person_id").
		Where("sc.movie_cast_person_id IS NOT NULL AND (sc.movie_cast_person_id <= 0 OR person.id IS NULL)").
		Count(&invalidCount).Error; err != nil {
		return fmt.Errorf("校验 g_sc 演员 person_id 失败: %w", err)
	}
	if invalidCount > 0 {
		return fmt.Errorf("g_sc 存在 %d 条无效演员 person_id", invalidCount)
	}
	return nil
}

func loadGScMovieCastResolver(db *gorm.DB) (*gScMovieCastResolver, error) {
	type castNameRow struct {
		Name     string `gorm:"column:name"`
		PersonID int64  `gorm:"column:person_id"`
	}
	var casts []*castNameRow
	if err := db.Table("am_cast").Select("name", "person_id").Where("person_id > 0").Scan(&casts).Error; err != nil {
		return nil, fmt.Errorf("读取 am_cast 演员映射失败: %w", err)
	}
	var persons []*Person
	if err := db.Find(&persons).Error; err != nil {
		return nil, fmt.Errorf("读取 c_person 演员映射失败: %w", err)
	}

	resolver := &gScMovieCastResolver{
		castNames:    make(map[string]int64),
		chineseNames: make(map[string]int64),
		personNames:  make(map[string]int64),
		aliases:      make(map[string]int64),
	}
	for _, row := range casts {
		if row != nil {
			addGScMovieCastResolverName(resolver.castNames, row.Name, row.PersonID)
		}
	}
	for _, row := range persons {
		if row == nil {
			continue
		}
		addGScMovieCastResolverName(resolver.chineseNames, row.Chinese, row.Id)
		addGScMovieCastResolverName(resolver.personNames, row.Name, row.Id)
		for _, alias := range splitLegacyPersonAliases(row.Alias) {
			addGScMovieCastResolverName(resolver.aliases, alias, row.Id)
		}
	}
	return resolver, nil
}

func (r *gScMovieCastResolver) resolve(name string) int64 {
	key := normalizeGScMovieCastLookupName(name)
	if key == "" {
		return 0
	}
	for _, index := range []map[string]int64{r.castNames, r.chineseNames, r.personNames, r.aliases} {
		if personID := index[key]; personID > 0 {
			return personID
		}
		if _, ambiguous := index[key]; ambiguous {
			return 0
		}
	}
	return 0
}

func addGScMovieCastResolverName(index map[string]int64, name string, personID int64) {
	key := normalizeGScMovieCastLookupName(name)
	if key == "" || personID <= 0 {
		return
	}
	if existing, ok := index[key]; ok && existing != personID {
		index[key] = 0
		return
	}
	index[key] = personID
}

func normalizeGScMovieCastLookupName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func normalizeLegacyGScMovieCastName(raw string) string {
	raw = strings.TrimSpace(raw)
	if trimmed, ok := trimLegacyGScMovieCastAgeSuffix(raw, '(', ')'); ok {
		return trimmed
	}
	if trimmed, ok := trimLegacyGScMovieCastAgeSuffix(raw, '（', '）'); ok {
		return trimmed
	}
	return raw
}

func trimLegacyGScMovieCastAgeSuffix(raw string, left, right rune) (string, bool) {
	if raw == "" || !strings.HasSuffix(raw, string(right)) {
		return "", false
	}
	start := strings.LastIndex(raw, string(left))
	if start <= 0 {
		return "", false
	}
	inner := strings.TrimSpace(raw[start+len(string(left)) : len(raw)-len(string(right))])
	if inner == "" {
		return "", false
	}
	for _, ch := range inner {
		if (ch < '0' || ch > '9') && ch != '.' {
			return "", false
		}
	}
	return strings.TrimSpace(raw[:start]), true
}

func splitLegacyPersonAliases(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	replacer := strings.NewReplacer("，", ",", "、", ",", "/", ",", "|", ",", "\n", ",", ";", ",", "；", ",")
	parts := strings.Split(replacer.Replace(raw), ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
