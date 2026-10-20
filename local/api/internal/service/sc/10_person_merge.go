package sc

import (
	"context"
	"time"

	"rudy_gc/local/api/internal/types"
)

func (s *Service) SearchPersonMergeCandidates(ctx context.Context, keepPersonID int64, keyword string, limit int64) ([]*types.PersonMergeCandidate, error) {
	persons, err := s.deps.PersonModel.SearchMergeCandidates(ctx, keyword, []int64{keepPersonID}, limit)
	if err != nil {
		return nil, err
	}

	personIDs := make([]int64, 0, len(persons))
	for _, person := range persons {
		if person == nil || person.Id <= 0 {
			continue
		}
		personIDs = append(personIDs, person.Id)
	}

	castRows, err := s.deps.CastModel.ListRowsByPersonIDs(ctx, personIDs)
	if err != nil {
		return nil, err
	}
	aliasMap := buildPersonCastNames(castRows)

	return buildPersonMergeCandidates(persons, aliasMap), nil
}

func (s *Service) PreviewPersonMerge(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64) (*types.PersonMergePreview, error) {
	state, err := s.loadPersonMergeState(ctx, keepPersonID, sourcePersonIDs)
	if err != nil {
		return nil, err
	}

	return &types.PersonMergePreview{
		Keep:               buildPersonMergeCandidate(mapPersonModelToTypes(state.keep), state.castNamesByPersonID[state.keep.Id]),
		Sources:            buildPersonMergeCandidatesFromRows(state.sources, state.castNamesByPersonID),
		MoveCastNames:      state.moveCastNames,
		RemovePersonIds:    state.sourcePersonIDs,
		AffectedMovieCount: len(state.affectedMovieJavIDs),
	}, nil
}

func (s *Service) MergePerson(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64) (*types.PersonMergeResult, error) {
	state, err := s.loadPersonMergeState(ctx, keepPersonID, sourcePersonIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if mergePersonProfiles(state.keep, state.sources, state.castNamesByPersonID) {
		state.keep.UpdatedOn = now
		if err := s.deps.PersonModel.Update(ctx, state.keep); err != nil {
			return nil, err
		}
	}

	for _, castRow := range state.sourceCasts {
		if castRow == nil || castRow.Id <= 0 || castRow.PersonId == state.keep.Id {
			continue
		}
		castRow.PersonId = state.keep.Id
		castRow.UpdatedOn = now
		if err := s.deps.CastModel.Update(ctx, castRow); err != nil {
			return nil, err
		}
	}
	if err := s.reassignScEventMovieCastPersons(ctx, state.keep.Id, state.sourcePersonIDs, now); err != nil {
		return nil, err
	}
	if s.deps.CafoModel != nil {
		if err := s.deps.CafoModel.ReassignPersonIDs(ctx, state.keep.Id, state.sourcePersonIDs); err != nil {
			return nil, err
		}
	}

	if err := s.personStatSvc.SyncByIDs(ctx, []int64{state.keep.Id}, now); err != nil {
		return nil, err
	}

	for _, sourceRow := range state.sources {
		if sourceRow == nil || sourceRow.Id <= 0 {
			continue
		}
		if err := s.deps.PersonModel.Delete(ctx, sourceRow.Id); err != nil {
			return nil, err
		}
	}
	if s.deps.CPersonScModel != nil {
		if err := s.deps.CPersonScModel.DeleteByPersonIDs(ctx, state.sourcePersonIDs); err != nil {
			return nil, err
		}
	}

	s.movieSvc.InvalidateMovieTypes(ctx, state.affectedMovieJavIDs...)

	return &types.PersonMergeResult{
		KeepPersonId:       state.keep.Id,
		RemovedPersonIds:   state.sourcePersonIDs,
		MoveCastNames:      state.moveCastNames,
		AffectedMovieCount: len(state.affectedMovieJavIDs),
	}, nil
}

func (s *Service) reassignScEventMovieCastPersons(ctx context.Context, keepPersonID int64, sourcePersonIDs []int64, now int64) error {
	if keepPersonID <= 0 || len(sourcePersonIDs) == 0 {
		return nil
	}
	sourceIDs := make(map[int64]struct{}, len(sourcePersonIDs))
	for _, personID := range sourcePersonIDs {
		if personID > 0 && personID != keepPersonID {
			sourceIDs[personID] = struct{}{}
		}
	}
	if len(sourceIDs) == 0 {
		return nil
	}
	rows, err := s.deps.ScModel.FindAll(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil || !row.MovieCastPersonId.Valid {
			continue
		}
		if _, ok := sourceIDs[row.MovieCastPersonId.Int64]; !ok {
			continue
		}
		row.MovieCastPersonId = nullablePositiveID(keepPersonID)
		row.UpdatedOn = now
		if err := s.deps.ScModel.Update(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
