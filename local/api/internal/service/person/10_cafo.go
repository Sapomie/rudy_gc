package person

import (
	"context"
	"errors"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

func (s *Service) ListCafoProfiles(ctx context.Context, personID int64) ([]*types.CafoProfile, error) {
	if personID <= 0 {
		return nil, types.ErrNotFound
	}
	if _, err := s.deps.PersonModel.FindOne(ctx, personID); err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	rows, err := s.deps.CafoModel.ListByPersonID(ctx, personID)
	if err != nil {
		return nil, err
	}
	out := make([]*types.CafoProfile, 0, len(rows))
	for _, row := range rows {
		if row != nil {
			out = append(out, mapCafoProfile(row))
		}
	}
	return out, nil
}

func (s *Service) UpdateCafoProfile(
	ctx context.Context,
	personID, cafoID int64,
	in *types.CafoProfileWriteForm,
) (*types.CafoProfile, *types.Person, error) {
	if personID <= 0 || cafoID <= 0 {
		return nil, nil, types.ErrNotFound
	}
	update, err := normalizeCafoProfileWrite(in)
	if err != nil {
		return nil, nil, err
	}
	update.UpdatedOn = time.Now().Unix()
	cafo, person, err := s.deps.CafoModel.UpdateProfileForPerson(ctx, personID, cafoID, update)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, nil, types.ErrNotFound
		}
		if strings.Contains(strings.ToLower(err.Error()), "duplicate entry") {
			return nil, nil, &types.BadRequestError{Message: "Name 或 Chinese 已被其他 c_cafo 使用"}
		}
		return nil, nil, err
	}
	return mapCafoProfile(cafo), mapPerson(person), nil
}

func normalizeCafoProfileWrite(in *types.CafoProfileWriteForm) (*moviex.CCafoProfileUpdate, error) {
	if in == nil {
		return nil, &types.BadRequestError{Message: "c_cafo 参数不能为空"}
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, &types.BadRequestError{Message: "Name 不能为空"}
	}
	return &moviex.CCafoProfileUpdate{
		Name:      name,
		Alias:     strings.TrimSpace(in.Alias),
		Chinese:   strings.TrimSpace(in.Chinese),
		BirthDay:  in.BirthDay,
		Height:    in.Height,
		Cup:       strings.TrimSpace(in.Cup),
		Bwh:       strings.TrimSpace(in.Bwh),
		Avatar:    strings.TrimSpace(in.Avatar),
		Biography: strings.TrimSpace(in.Biography),
	}, nil
}

func mapCafoProfile(row *moviex.CCafoProfile) *types.CafoProfile {
	if row == nil {
		return nil
	}
	return &types.CafoProfile{
		Id:        row.Id,
		PersonId:  row.PersonId,
		Name:      row.Name,
		Alias:     row.Alias,
		Chinese:   row.Chinese,
		BirthDay:  row.BirthDay,
		Height:    row.Height,
		Cup:       row.Cup,
		Bwh:       row.Bwh,
		Avatar:    row.Avatar,
		Biography: row.Biography,
		CreatedOn: row.CreatedOn,
		UpdatedOn: row.UpdatedOn,
	}
}

func mapPerson(row *moviex.CPerson) *types.Person {
	if row == nil {
		return nil
	}
	return &types.Person{
		Id:                     row.Id,
		Name:                   row.Name,
		Alias:                  row.Alias,
		Chinese:                row.Chinese,
		BirthDay:               row.BirthDay,
		Height:                 row.Height,
		Cup:                    row.Cup,
		Bwh:                    row.Bwh,
		Avatar:                 row.Avatar,
		MovieNumber:            row.MovieNumber,
		OwnedMovieNumber:       row.OwnedWMediaNumber,
		OwnedWMediaNumber:      row.OwnedWMediaNumber,
		OwnedWMediaRatio:       row.OwnedWMediaRatio,
		ScTimes:                row.ScTimes,
		ComeTimes:              row.ComeTimes,
		LastSelectedScTime:     row.LastSelectedScTime,
		LastParticipatedScTime: row.LastParticipatedScTime,
		HighestRank:            row.HighestRank,
		RankTimes:              row.RankTimes,
		CreatedOn:              row.CreatedOn,
		UpdatedOn:              row.UpdatedOn,
	}
}
