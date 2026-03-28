package spider

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"rudy_gc/data/modelx/moviex"
)

func (l *CrawlLogic) BackfillPersonData(ctx context.Context) error {
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
	castRows, err := l.deps.CastModel.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("list am_cast failed: %w", err)
	}

	lookup := newPersonBackfillLookup(personRows)

	createdProfiles := 0
	updatedProfiles := 0
	for i, cafoRow := range cafoRows {
		if err := l.waitIfPaused(ctx); err != nil {
			return err
		}
		created, updated, err := l.backfillPersonFromCafo(ctx, lookup, cafoRow, now)
		if err != nil {
			return err
		}
		if created {
			createdProfiles++
		}
		if updated {
			updatedProfiles++
		}
		l.reportPhaseProgress(
			ctx,
			"person_profile",
			"person_backfill_profile",
			fmt.Sprintf("已迁移 c_cafo -> c_person：%d/%d，新增=%d，更新=%d", i+1, len(cafoRows), createdProfiles, updatedProfiles),
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
	if err := l.deps.SyncPersonStatsByIDs(ctx, personIDs, now); err != nil {
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

func (l *CrawlLogic) backfillPersonFromCafo(ctx context.Context, lookup *personBackfillLookup, row *moviex.CCafo, now int64) (created bool, updated bool, err error) {
	if row == nil {
		return false, false, nil
	}

	target := lookup.matchCafo(row)
	if target == nil {
		target = newPersonRowFromCafo(row, now)
		res, err := l.deps.PersonModel.Insert(ctx, target)
		if err != nil {
			return false, false, fmt.Errorf("insert c_person from c_cafo(%d) failed: %w", row.Id, err)
		}
		target.Id, _ = res.LastInsertId()
		lookup.upsert(target)
		return true, false, nil
	}

	if !mergePersonProfileFromCafo(target, row) {
		lookup.upsert(target)
		return false, false, nil
	}
	target.UpdatedOn = now
	if err := l.deps.PersonModel.Update(ctx, target); err != nil {
		return false, false, fmt.Errorf("update c_person(%d) from c_cafo(%d) failed: %w", target.Id, row.Id, err)
	}
	lookup.upsert(target)
	return false, true, nil
}

func (l *CrawlLogic) bindCastToPerson(ctx context.Context, lookup *personBackfillLookup, row *moviex.AmCast, now int64) (created bool, changed bool, err error) {
	if row == nil {
		return false, false, nil
	}

	if row.PersonId > 0 {
		if personRow := lookup.byID[row.PersonId]; personRow != nil {
			return false, false, nil
		}
	}

	target := lookup.matchCast(row)
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

func (l *CrawlLogic) collectMovieJavIDsByCastRows(ctx context.Context, castRows []*moviex.AmCast) ([]string, error) {
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
	byID      map[int64]*moviex.CPerson
	byName    map[string]*moviex.CPerson
	byAlias   map[string]*moviex.CPerson
	byChinese map[string]*moviex.CPerson
}

func newPersonBackfillLookup(rows []*moviex.CPerson) *personBackfillLookup {
	out := &personBackfillLookup{
		byID:      make(map[int64]*moviex.CPerson, len(rows)),
		byName:    make(map[string]*moviex.CPerson, len(rows)),
		byAlias:   make(map[string]*moviex.CPerson, len(rows)),
		byChinese: make(map[string]*moviex.CPerson, len(rows)),
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
	if nameKey != "" {
		if exist := l.byName[nameKey]; exist == nil || exist.Id == row.Id {
			l.byName[nameKey] = row
		}
	}

	chineseKey := normalizePersonLookupKey(row.Chinese)
	if chineseKey != "" {
		if exist := l.byChinese[chineseKey]; exist == nil || exist.Id == row.Id {
			l.byChinese[chineseKey] = row
		}
	}

	for _, alias := range splitPersonLookupAliases(row.Alias) {
		aliasKey := normalizePersonLookupKey(alias)
		if aliasKey == "" {
			continue
		}
		if exist := l.byAlias[aliasKey]; exist == nil || exist.Id == row.Id {
			l.byAlias[aliasKey] = row
		}
	}
}

func (l *personBackfillLookup) matchCafo(row *moviex.CCafo) *moviex.CPerson {
	if l == nil || row == nil {
		return nil
	}
	if personRow := l.byName[normalizePersonLookupKey(row.Name)]; personRow != nil {
		return personRow
	}
	if personRow := l.byChinese[normalizePersonLookupKey(row.Chinese)]; personRow != nil {
		return personRow
	}
	if personRow := l.byAlias[normalizePersonLookupKey(row.Name)]; personRow != nil {
		return personRow
	}
	return nil
}

func (l *personBackfillLookup) matchCast(row *moviex.AmCast) *moviex.CPerson {
	if l == nil || row == nil {
		return nil
	}
	if row.PersonId > 0 {
		if personRow := l.byID[row.PersonId]; personRow != nil {
			return personRow
		}
	}
	if personRow := l.byName[normalizePersonLookupKey(row.Name)]; personRow != nil {
		return personRow
	}
	return l.byAlias[normalizePersonLookupKey(row.Name)]
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
		Name:             strings.TrimSpace(row.Name),
		MovieNumber:      row.MovieNumber,
		OwnedMovieNumber: row.OwnedMovieNumber,
		ScTimes:          row.ScTimes,
		ComeTimes:        row.ComeTimes,
		LastScTime:       row.LastScTime,
		HighestRank:      row.HighestRank,
		RankTimes:        row.RankTimes,
		CreatedOn:        now,
		UpdatedOn:        now,
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
	return strings.TrimSpace(v)
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
