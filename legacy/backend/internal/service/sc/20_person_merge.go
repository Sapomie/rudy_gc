package sc

import (
	"context"
	"errors"
	"sort"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

type personMergeState struct {
	keep                *moviex.CPerson
	sources             []*moviex.CPerson
	sourcePersonIDs     []int64
	allCasts            []*moviex.AmCast
	sourceCasts         []*moviex.AmCast
	castNamesByPersonID map[int64][]string
	moveCastNames       []string
	affectedMovieJavIDs []string
}

func (s *Service) loadPersonMergeState(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64) (*personMergeState, error) {
	if keepPersonID <= 0 {
		return nil, errors.New("keepPersonId 无效")
	}

	normalizedSourceIDs := normalizePersonMergeSourceIDs(keepPersonID, sourcePersonIDs)
	if len(normalizedSourceIDs) == 0 {
		return nil, errors.New("sourcePersonIds 不能为空")
	}

	keepRow, err := s.deps.PersonModel.FindOne(ctx, keepPersonID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}

	sourceRows := make([]*moviex.CPerson, 0, len(normalizedSourceIDs))
	for _, sourcePersonID := range normalizedSourceIDs {
		row, err := s.deps.PersonModel.FindOne(ctx, sourcePersonID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				return nil, types.ErrNotFound
			}
			return nil, err
		}
		sourceRows = append(sourceRows, row)
	}

	allPersonIDs := append([]int64{keepPersonID}, normalizedSourceIDs...)
	allCastRows, err := s.deps.CastModel.ListRowsByPersonIDs(ctx, allPersonIDs)
	if err != nil {
		return nil, err
	}

	castNamesByPersonID := buildPersonCastNames(allCastRows)
	ensurePersonNameInCastNames(castNamesByPersonID, keepRow)
	for _, sourceRow := range sourceRows {
		ensurePersonNameInCastNames(castNamesByPersonID, sourceRow)
	}

	sourceSet := make(map[int64]struct{}, len(normalizedSourceIDs))
	for _, id := range normalizedSourceIDs {
		sourceSet[id] = struct{}{}
	}

	sourceCastRows := make([]*moviex.AmCast, 0, len(allCastRows))
	for _, castRow := range allCastRows {
		if castRow == nil {
			continue
		}
		if _, ok := sourceSet[castRow.PersonId]; ok {
			sourceCastRows = append(sourceCastRows, castRow)
		}
	}

	moveCastNames := flattenCastNamesByPersonIDs(castNamesByPersonID, normalizedSourceIDs)
	affectedMovieJavIDs, err := s.collectMovieJavIDsByCastRows(ctx, allCastRows)
	if err != nil {
		return nil, err
	}

	return &personMergeState{
		keep:                keepRow,
		sources:             sourceRows,
		sourcePersonIDs:     normalizedSourceIDs,
		allCasts:            allCastRows,
		sourceCasts:         sourceCastRows,
		castNamesByPersonID: castNamesByPersonID,
		moveCastNames:       moveCastNames,
		affectedMovieJavIDs: affectedMovieJavIDs,
	}, nil
}

func (s *Service) collectMovieJavIDsByCastRows(ctx context.Context, castRows []*moviex.AmCast) ([]string, error) {
	seen := make(map[string]struct{}, len(castRows))
	out := make([]string, 0, len(castRows))
	for _, castRow := range castRows {
		if castRow == nil || castRow.Id <= 0 {
			continue
		}
		movieJavIDs, err := s.movieCastListMovieJavIDsByCastID(ctx, castRow.Id)
		if err != nil {
			return nil, err
		}
		for _, movieJavID := range movieJavIDs {
			movieJavID = strings.TrimSpace(movieJavID)
			if movieJavID == "" {
				continue
			}
			if _, ok := seen[movieJavID]; ok {
				continue
			}
			seen[movieJavID] = struct{}{}
			out = append(out, movieJavID)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (s *Service) buildActorAliases(ctx context.Context, actor *types.Person) ([]string, error) {
	if actor == nil {
		return nil, nil
	}
	if actor.Id <= 0 {
		if strings.TrimSpace(actor.Name) == "" {
			return nil, nil
		}
		return []string{strings.TrimSpace(actor.Name)}, nil
	}

	castRows, err := s.deps.CastModel.ListRowsByPersonIDs(ctx, []int64{actor.Id})
	if err != nil {
		return nil, err
	}

	aliasMap := buildPersonCastNames(castRows)
	ensurePersonNameInCastNamesModel(aliasMap, actor.Id, actor.Name)
	return aliasMap[actor.Id], nil
}

func buildPersonMergeCandidates(persons []*types.Person, castNamesByPersonID map[int64][]string) []*types.PersonMergeCandidate {
	out := make([]*types.PersonMergeCandidate, 0, len(persons))
	for _, person := range persons {
		if person == nil {
			continue
		}
		out = append(out, buildPersonMergeCandidate(person, castNamesByPersonID[person.Id]))
	}
	return out
}

func buildPersonMergeCandidatesFromRows(rows []*moviex.CPerson, castNamesByPersonID map[int64][]string) []*types.PersonMergeCandidate {
	out := make([]*types.PersonMergeCandidate, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, buildPersonMergeCandidate(mapPersonModelToTypes(row), castNamesByPersonID[row.Id]))
	}
	return out
}

func buildPersonMergeCandidate(person *types.Person, castNames []string) *types.PersonMergeCandidate {
	if person == nil {
		return nil
	}
	names := append([]string{}, castNames...)
	if name := strings.TrimSpace(person.Name); name != "" && !containsString(names, name) {
		names = append([]string{name}, names...)
	}
	return &types.PersonMergeCandidate{
		Person:    person,
		CastNames: names,
	}
}

func buildPersonCastNames(castRows []*moviex.AmCast) map[int64][]string {
	agg := make(map[int64][]string)
	seen := make(map[int64]map[string]struct{})

	for _, castRow := range castRows {
		if castRow == nil || castRow.PersonId <= 0 {
			continue
		}
		name := strings.TrimSpace(castRow.Name)
		if name == "" {
			continue
		}
		if seen[castRow.PersonId] == nil {
			seen[castRow.PersonId] = map[string]struct{}{}
		}
		if _, ok := seen[castRow.PersonId][name]; ok {
			continue
		}
		seen[castRow.PersonId][name] = struct{}{}
		agg[castRow.PersonId] = append(agg[castRow.PersonId], name)
	}

	for personID := range agg {
		sort.Strings(agg[personID])
	}
	return agg
}

func ensurePersonNameInCastNames(castNamesByPersonID map[int64][]string, row *moviex.CPerson) {
	if row == nil {
		return
	}
	ensurePersonNameInCastNamesModel(castNamesByPersonID, row.Id, row.Name)
}

func ensurePersonNameInCastNamesModel(castNamesByPersonID map[int64][]string, personID int64, name string) {
	if personID <= 0 {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if containsString(castNamesByPersonID[personID], name) {
		return
	}
	castNamesByPersonID[personID] = append([]string{name}, castNamesByPersonID[personID]...)
}

func flattenCastNamesByPersonIDs(castNamesByPersonID map[int64][]string, personIDs []int64) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, personID := range personIDs {
		for _, name := range castNamesByPersonID[personID] {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func mergePersonProfiles(keep *moviex.CPerson, sources []*moviex.CPerson, castNamesByPersonID map[int64][]string) bool {
	if keep == nil || len(sources) == 0 {
		return false
	}

	changed := false
	for _, source := range sources {
		if source == nil {
			continue
		}
		if keep.Name == "" && strings.TrimSpace(source.Name) != "" {
			keep.Name = strings.TrimSpace(source.Name)
			changed = true
		}
		if keep.Chinese == "" && strings.TrimSpace(source.Chinese) != "" {
			keep.Chinese = strings.TrimSpace(source.Chinese)
			changed = true
		}
		if keep.BirthDay <= 0 && source.BirthDay > 0 {
			keep.BirthDay = source.BirthDay
			changed = true
		}
		if keep.Height <= 0 && source.Height > 0 {
			keep.Height = source.Height
			changed = true
		}
		if keep.Cup == "" && strings.TrimSpace(source.Cup) != "" {
			keep.Cup = strings.TrimSpace(source.Cup)
			changed = true
		}
		if keep.Bwh == "" && strings.TrimSpace(source.Bwh) != "" {
			keep.Bwh = strings.TrimSpace(source.Bwh)
			changed = true
		}
		if keep.Avatar == "" && strings.TrimSpace(source.Avatar) != "" {
			keep.Avatar = strings.TrimSpace(source.Avatar)
			changed = true
		}

		nextAlias := mergePersonAliasText(keep.Alias, keep.Name, source.Name, source.Alias, castNamesByPersonID[source.Id])
		if nextAlias != keep.Alias {
			keep.Alias = nextAlias
			changed = true
		}
	}
	return changed
}

func mergePersonAliasText(currentAlias, keepName, sourceName, sourceAlias string, sourceCastNames []string) string {
	ordered := make([]string, 0)
	seen := make(map[string]struct{})
	excluded := normalizeAliasValue(keepName)

	appendValue := func(raw string) {
		for _, part := range splitAliasParts(raw) {
			key := normalizeAliasValue(part)
			if key == "" || key == excluded {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			ordered = append(ordered, strings.TrimSpace(part))
		}
	}

	appendValue(currentAlias)
	appendValue(sourceAlias)
	appendValue(sourceName)
	for _, castName := range sourceCastNames {
		appendValue(castName)
	}

	return strings.Join(ordered, ", ")
}

func splitAliasParts(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	replacer := strings.NewReplacer("，", ",", "、", ",", "/", ",", "|", ",", "\n", ",", ";", ",", "；", ",")
	normalized := replacer.Replace(raw)
	parts := strings.Split(normalized, ",")
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

func normalizeAliasValue(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func normalizePersonMergeSourceIDs(keepPersonID int64, sourcePersonIDs []int64) []int64 {
	seen := make(map[int64]struct{}, len(sourcePersonIDs))
	out := make([]int64, 0, len(sourcePersonIDs))
	for _, sourcePersonID := range sourcePersonIDs {
		if sourcePersonID <= 0 || sourcePersonID == keepPersonID {
			continue
		}
		if _, ok := seen[sourcePersonID]; ok {
			continue
		}
		seen[sourcePersonID] = struct{}{}
		out = append(out, sourcePersonID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func containsString(values []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}
