package spider

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

func (l *Service) BackfillPersonData(ctx context.Context) error {
	now := time.Now().Unix()

	l.reportProgress(ctx, "person_backfill_prepare", "开始加载演员资料、人物资料和关联数据", 0, 0, 0, 0)

	cafoRows, err := l.deps.CafoModel.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list c_cafo failed: %w", err)
	}
	personRows, err := l.deps.PersonModel.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list c_person failed: %w", err)
	}
	cafoPersonLinks, err := l.deps.CafoModel.ListPersonLinks(ctx)
	if err != nil {
		return fmt.Errorf("list c_cafo person links failed: %w", err)
	}
	castRows, err := l.deps.CastModel.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list am_cast failed: %w", err)
	}

	lookup := newPersonBackfillLookup(personRows)

	createdProfiles := 0
	updatedProfiles := 0
	linkedProfiles := 0
	ambiguousProfiles := 0
	for i, cafoRow := range cafoRows {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		linkedPersonID := int64(0)
		if cafoRow != nil {
			linkedPersonID = cafoPersonLinks[cafoRow.Id]
		}
		created, updated, linked, ambiguous, err := l.backfillPersonFromCafo(ctx, lookup, cafoRow, linkedPersonID, now)
		if err != nil {
			return err
		}
		if created {
			createdProfiles++
		}
		if updated {
			updatedProfiles++
		}
		if linked {
			linkedProfiles++
		}
		if ambiguous {
			ambiguousProfiles++
		}
		l.reportPhaseProgress(
			ctx,
			"person_profile",
			"person_backfill_profile",
			fmt.Sprintf("已迁移 c_cafo -> c_person：%d/%d，新增=%d，更新=%d，新绑定=%d，待确认=%d", i+1, len(cafoRows), createdProfiles, updatedProfiles, linkedProfiles, ambiguousProfiles),
			i+1,
			len(cafoRows),
			i+1,
			0,
		)
	}

	boundCasts := 0
	createdByCast := 0
	for i, castRow := range castRows {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		created, changed, err := l.bindCastToPerson(ctx, lookup, castRow, now)
		if err != nil {
			return err
		}
		if created {
			createdByCast++
		}
		if changed {
			boundCasts++
		}
		l.reportPhaseProgress(
			ctx,
			"person_bind",
			"person_backfill_bind",
			fmt.Sprintf("已补 am_cast.person_id：%d/%d，补绑=%d，独立新建=%d", i+1, len(castRows), boundCasts, createdByCast),
			i+1,
			len(castRows),
			i+1,
			0,
		)
	}

	personIDs := lookup.ids()
	l.reportProgress(ctx, "person_backfill_sync", fmt.Sprintf("开始重算 c_person 统计，共 %d 人", len(personIDs)), 0, 0, 0, len(personIDs))
	if err := l.personStatSvc.SyncByIDs(ctx, personIDs, now); err != nil {
		return fmt.Errorf("sync c_person stats failed: %w", err)
	}

	movieJavIDs, err := l.collectMovieJavIDsByCastRows(ctx, castRows)
	if err != nil {
		return err
	}
	l.movieSvc.InvalidateMovieTypes(ctx, movieJavIDs...)
	l.reportProgress(ctx, "person_backfill_done", fmt.Sprintf("person 回填完成：person=%d，cast=%d，movie_cache=%d", len(personIDs), len(castRows), len(movieJavIDs)), len(castRows), len(castRows), 0, 0)

	return nil
}

func (l *Service) backfillPersonFromCafo(ctx context.Context, lookup *personBackfillLookup, row *moviex.CCafo, linkedPersonID int64, now int64) (created bool, updated bool, linked bool, ambiguous bool, err error) {
	if row == nil {
		return false, false, false, false, nil
	}

	var target *moviex.CPerson
	if linkedPersonID > 0 {
		target = lookup.byID[linkedPersonID]
		if target == nil {
			if err := l.deps.CafoModel.SetPersonID(ctx, row.Id, 0); err != nil {
				return false, false, false, false, fmt.Errorf("clear invalid c_cafo person link(%d -> %d) failed: %w", row.Id, linkedPersonID, err)
			}
			linkedPersonID = 0
		}
	}
	if target == nil {
		target, ambiguous = lookup.matchCafo(row)
		if ambiguous {
			return false, false, false, true, nil
		}
	}
	if target == nil {
		target = newPersonRowFromCafo(row, now)
		res, err := l.deps.PersonModel.Insert(ctx, target)
		if err != nil {
			return false, false, false, false, fmt.Errorf("insert c_person from c_cafo(%d) failed: %w", row.Id, err)
		}
		target.Id, _ = res.LastInsertId()
		lookup.upsert(target)
		created = true
	}

	if !created && mergePersonProfileFromCafo(target, row) {
		target.UpdatedOn = now
		if err := l.deps.PersonModel.Update(ctx, target); err != nil {
			return false, false, false, false, fmt.Errorf("update c_person(%d) from c_cafo(%d) failed: %w", target.Id, row.Id, err)
		}
		updated = true
	}

	if linkedPersonID != target.Id {
		if err := l.deps.CafoModel.SetPersonID(ctx, row.Id, target.Id); err != nil {
			return false, false, false, false, fmt.Errorf("link c_cafo(%d) to c_person(%d) failed: %w", row.Id, target.Id, err)
		}
		linked = true
	}
	return created, updated, linked, false, nil
}

func (l *Service) bindCastToPerson(ctx context.Context, lookup *personBackfillLookup, row *moviex.AmCast, now int64) (created bool, changed bool, err error) {
	if row == nil {
		return false, false, nil
	}

	if row.PersonId > 0 {
		if personRow := lookup.byID[row.PersonId]; personRow != nil {
			return false, false, nil
		}
	}

	target, ambiguous := lookup.matchCast(row)
	if ambiguous {
		return false, false, nil
	}
	if target == nil {
		target = newPersonRowFromCast(row, now)
		res, err := l.deps.PersonModel.Insert(ctx, target)
		if err != nil {
			return false, false, fmt.Errorf("insert c_person from am_cast(%d) failed: %w", row.Id, err)
		}
		target.Id, _ = res.LastInsertId()
		lookup.upsert(target)
		created = true
	}

	if row.PersonId == target.Id {
		return created, false, nil
	}

	row.PersonId = target.Id
	row.UpdatedOn = now
	if err := l.deps.CastModel.Update(ctx, row); err != nil {
		return created, false, fmt.Errorf("update am_cast(%d) person_id failed: %w", row.Id, err)
	}
	return created, true, nil
}

func (l *Service) collectMovieJavIDsByCastRows(ctx context.Context, castRows []*moviex.AmCast) ([]string, error) {
	if len(castRows) == 0 {
		return nil, nil
	}

	seen := make(map[string]struct{})
	out := make([]string, 0, len(castRows))
	for i, castRow := range castRows {
		if err := l.waitIfPaused(ctx); err != nil {
			return nil, err
		}
		if castRow == nil || castRow.Id <= 0 {
			continue
		}
		movieJavIDs, err := l.deps.MovieCastRepo.ListMovieJavIDsByCastID(ctx, castRow.Id)
		if err != nil {
			return nil, fmt.Errorf("list movie_jav_ids by cast_id(%d) failed: %w", castRow.Id, err)
		}
		for _, movieJavID := range movieJavIDs {
			if movieJavID == "" {
				continue
			}
			if _, ok := seen[movieJavID]; ok {
				continue
			}
			seen[movieJavID] = struct{}{}
			out = append(out, movieJavID)
		}
		l.reportPhaseProgress(
			ctx,
			"person_cache",
			"person_backfill_cache",
			fmt.Sprintf("已清理影片缓存准备集：%d/%d", i+1, len(castRows)),
			i+1,
			len(castRows),
			i+1,
			0,
		)
	}
	return out, nil
}

type personBackfillLookup struct {
	byID             map[int64]*moviex.CPerson
	byName           map[string]*moviex.CPerson
	byAlias          map[string]*moviex.CPerson
	byChinese        map[string]*moviex.CPerson
	ambiguousName    map[string]struct{}
	ambiguousAlias   map[string]struct{}
	ambiguousChinese map[string]struct{}
}

func newPersonBackfillLookup(rows []*moviex.CPerson) *personBackfillLookup {
	out := &personBackfillLookup{
		byID:             make(map[int64]*moviex.CPerson, len(rows)),
		byName:           make(map[string]*moviex.CPerson, len(rows)),
		byAlias:          make(map[string]*moviex.CPerson, len(rows)),
		byChinese:        make(map[string]*moviex.CPerson, len(rows)),
		ambiguousName:    make(map[string]struct{}),
		ambiguousAlias:   make(map[string]struct{}),
		ambiguousChinese: make(map[string]struct{}),
	}
	for _, row := range rows {
		out.upsert(row)
	}
	return out
}

func (l *personBackfillLookup) upsert(row *moviex.CPerson) {
	if l == nil || row == nil || row.Id <= 0 {
		return
	}
	l.byID[row.Id] = row

	nameKey := normalizePersonLookupKey(row.Name)
	registerPersonLookup(l.byName, l.ambiguousName, nameKey, row)

	chineseKey := normalizePersonLookupKey(row.Chinese)
	registerPersonLookup(l.byChinese, l.ambiguousChinese, chineseKey, row)

	for _, alias := range splitPersonLookupAliases(row.Alias) {
		aliasKey := normalizePersonLookupKey(alias)
		registerPersonLookup(l.byAlias, l.ambiguousAlias, aliasKey, row)
	}
}

func registerPersonLookup(index map[string]*moviex.CPerson, ambiguous map[string]struct{}, key string, row *moviex.CPerson) {
	if key == "" || row == nil || row.Id <= 0 {
		return
	}
	if _, exists := ambiguous[key]; exists {
		return
	}
	if existing := index[key]; existing != nil && existing.Id != row.Id {
		delete(index, key)
		ambiguous[key] = struct{}{}
		return
	}
	index[key] = row
}

func (l *personBackfillLookup) matchCafo(row *moviex.CCafo) (*moviex.CPerson, bool) {
	if l == nil || row == nil {
		return nil, false
	}
	nameKey := normalizePersonLookupKey(row.Name)
	chineseKey := normalizePersonLookupKey(row.Chinese)
	if _, exists := l.ambiguousName[nameKey]; exists {
		return nil, true
	}
	if _, exists := l.ambiguousChinese[chineseKey]; exists {
		return nil, true
	}
	if _, exists := l.ambiguousAlias[nameKey]; exists {
		return nil, true
	}

	candidates := make(map[int64]*moviex.CPerson)
	for _, personRow := range []*moviex.CPerson{l.byName[nameKey], l.byChinese[chineseKey], l.byAlias[nameKey]} {
		if personRow != nil && personRow.Id > 0 {
			candidates[personRow.Id] = personRow
		}
	}
	if len(candidates) == 0 {
		return nil, false
	}
	if len(candidates) > 1 {
		return nil, true
	}
	for _, personRow := range candidates {
		return personRow, false
	}
	return nil, false
}

func (l *personBackfillLookup) matchCast(row *moviex.AmCast) (*moviex.CPerson, bool) {
	if l == nil || row == nil {
		return nil, false
	}
	if row.PersonId > 0 {
		if personRow := l.byID[row.PersonId]; personRow != nil {
			return personRow, false
		}
	}
	nameKey := normalizePersonLookupKey(row.Name)
	if _, exists := l.ambiguousName[nameKey]; exists {
		return nil, true
	}
	if _, exists := l.ambiguousAlias[nameKey]; exists {
		return nil, true
	}
	candidates := make(map[int64]*moviex.CPerson)
	for _, personRow := range []*moviex.CPerson{l.byName[nameKey], l.byAlias[nameKey]} {
		if personRow != nil && personRow.Id > 0 {
			candidates[personRow.Id] = personRow
		}
	}
	if len(candidates) > 1 {
		return nil, true
	}
	for _, personRow := range candidates {
		return personRow, false
	}
	return nil, false
}

func (l *personBackfillLookup) ids() []int64 {
	if l == nil || len(l.byID) == 0 {
		return nil
	}
	out := make([]int64, 0, len(l.byID))
	for id := range l.byID {
		if id > 0 {
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func newPersonRowFromCafo(row *moviex.CCafo, now int64) *moviex.CPerson {
	if row == nil {
		return &moviex.CPerson{CreatedOn: now, UpdatedOn: now}
	}
	return &moviex.CPerson{
		Name:      strings.TrimSpace(row.Name),
		Alias:     strings.TrimSpace(row.Alias),
		Chinese:   strings.TrimSpace(row.Chinese),
		BirthDay:  row.BirthDay,
		Height:    row.Height,
		Cup:       strings.TrimSpace(row.Cup),
		Bwh:       strings.TrimSpace(row.Bwh),
		Avatar:    strings.TrimSpace(row.Avtar),
		CreatedOn: now,
		UpdatedOn: now,
	}
}

func newPersonRowFromCast(row *moviex.AmCast, now int64) *moviex.CPerson {
	if row == nil {
		return &moviex.CPerson{CreatedOn: now, UpdatedOn: now}
	}
	return &moviex.CPerson{
		Name:               strings.TrimSpace(row.Name),
		MovieNumber:        row.MovieNumber,
		OwnedMovieNumber:   row.OwnedMovieNumber,
		OwnedWMediaNumber:  row.OwnedWMediaNumber,
		ScTimes:            row.ScTimes,
		ComeTimes:          row.ComeTimes,
		LastSelectedScTime: row.LastSelectedScTime,
		HighestRank:        row.HighestRank,
		RankTimes:          row.RankTimes,
		CreatedOn:          now,
		UpdatedOn:          now,
	}
}

func mergePersonProfileFromCafo(target *moviex.CPerson, source *moviex.CCafo) bool {
	if target == nil || source == nil {
		return false
	}
	changed := false

	if name := strings.TrimSpace(source.Name); name != "" && target.Name != name {
		target.Name = name
		changed = true
	}
	if alias := strings.TrimSpace(source.Alias); alias != "" && target.Alias != alias {
		target.Alias = alias
		changed = true
	}
	if chinese := strings.TrimSpace(source.Chinese); chinese != "" && target.Chinese != chinese {
		target.Chinese = chinese
		changed = true
	}
	if source.BirthDay > 0 && target.BirthDay != source.BirthDay {
		target.BirthDay = source.BirthDay
		changed = true
	}
	if source.Height > 0 && target.Height != source.Height {
		target.Height = source.Height
		changed = true
	}
	if cup := strings.TrimSpace(source.Cup); cup != "" && target.Cup != cup {
		target.Cup = cup
		changed = true
	}
	if bwh := strings.TrimSpace(source.Bwh); bwh != "" && target.Bwh != bwh {
		target.Bwh = bwh
		changed = true
	}
	if avatar := strings.TrimSpace(source.Avtar); avatar != "" && target.Avatar != avatar {
		target.Avatar = avatar
		changed = true
	}

	return changed
}

func normalizePersonLookupKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func splitPersonLookupAliases(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	replacer := strings.NewReplacer("，", ",", "、", ",", "/", ",", "|", ",", "\n", ",", ";", ",", "；", ",")
	parts := strings.Split(replacer.Replace(raw), ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
