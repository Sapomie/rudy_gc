package sc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
)

const (
	ScOptionTypeKind   = "kind"
	ScOptionTypeFG     = "fg"
	ScOptionTypeVessel = "vessel"
)

type scOptionSelection struct {
	ID   int64
	Name string
}

func (s *Service) ListScOptions(ctx context.Context, includeDisabled bool) ([]*types.ScOption, error) {
	out := make([]*types.ScOption, 0)
	for _, optionType := range scOptionTypes() {
		model, err := s.scOptionModel(optionType)
		if err != nil {
			return nil, err
		}
		var rows []*moviex.GScOption
		if includeDisabled {
			rows, err = model.ListAll(ctx)
		} else {
			rows, err = model.ListEnabled(ctx)
		}
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if row == nil {
				continue
			}
			out = append(out, mapScOptionModelToTypes(optionType, row))
		}
	}
	return out, nil
}

func (s *Service) CreateScOption(ctx context.Context, in *types.ScOptionWriteForm) (*types.ScOption, error) {
	optionType, row, err := normalizeScOptionWrite(in)
	if err != nil {
		return nil, err
	}
	model, err := s.scOptionModel(optionType)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	row.CreatedOn = now
	row.UpdatedOn = now
	if err := model.Insert(ctx, row); err != nil {
		return nil, scOptionWriteError(err)
	}
	return mapScOptionModelToTypes(optionType, row), nil
}

func (s *Service) UpdateScOption(ctx context.Context, id int64, in *types.ScOptionWriteForm) (*types.ScOption, error) {
	if id <= 0 {
		return nil, types.ErrNotFound
	}
	optionType, row, err := normalizeScOptionWrite(in)
	if err != nil {
		return nil, err
	}
	model, err := s.scOptionModel(optionType)
	if err != nil {
		return nil, err
	}
	current, err := model.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	row.ID = current.ID
	row.ScEventTimes = current.ScEventTimes
	row.LastScEventTime = current.LastScEventTime
	row.CreatedOn = current.CreatedOn
	row.UpdatedOn = time.Now().Unix()
	if err := model.Update(ctx, row); err != nil {
		return nil, scOptionWriteError(err)
	}
	return mapScOptionModelToTypes(optionType, row), nil
}

func (s *Service) DisableScOption(ctx context.Context, optionType string, id int64) (*types.ScOption, error) {
	if id <= 0 {
		return nil, types.ErrNotFound
	}
	optionType = strings.TrimSpace(optionType)
	model, err := s.scOptionModel(optionType)
	if err != nil {
		return nil, err
	}
	row, err := model.FindOne(ctx, id)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if row.IsEnabled == 0 {
		return mapScOptionModelToTypes(optionType, row), nil
	}
	row.IsEnabled = 0
	row.UpdatedOn = time.Now().Unix()
	if err := model.Disable(ctx, row.ID, row.UpdatedOn); err != nil {
		return nil, err
	}
	return mapScOptionModelToTypes(optionType, row), nil
}

func normalizeScOptionWrite(in *types.ScOptionWriteForm) (string, *moviex.GScOption, error) {
	if in == nil {
		return "", nil, &types.BadRequestError{Message: "SC 选项参数不能为空"}
	}
	optionType := strings.TrimSpace(in.OptionType)
	if !validScOptionType(optionType) {
		return "", nil, &types.BadRequestError{Message: "选项类型必须是 kind、fg 或 vessel"}
	}
	optionName := strings.TrimSpace(in.OptionName)
	if optionName == "" {
		return "", nil, &types.BadRequestError{Message: "选项名称不能为空"}
	}
	if in.IsEnabled != 0 && in.IsEnabled != 1 {
		return "", nil, &types.BadRequestError{Message: "选项状态必须是启用或停用"}
	}
	return optionType, &moviex.GScOption{
		Name:      optionName,
		SortOrder: in.SortOrder,
		IsEnabled: in.IsEnabled,
	}, nil
}

func scOptionTypes() []string {
	return []string{ScOptionTypeKind, ScOptionTypeFG, ScOptionTypeVessel}
}

func (s *Service) scOptionModel(optionType string) (moviex.GScOptionModel, error) {
	if s == nil || s.deps == nil {
		return nil, errors.New("SC 选项模型未初始化")
	}
	switch strings.TrimSpace(optionType) {
	case ScOptionTypeKind:
		if s.deps.ScKindOptionModel != nil {
			return s.deps.ScKindOptionModel, nil
		}
	case ScOptionTypeFG:
		if s.deps.ScFGOptionModel != nil {
			return s.deps.ScFGOptionModel, nil
		}
	case ScOptionTypeVessel:
		if s.deps.ScVesselOptionModel != nil {
			return s.deps.ScVesselOptionModel, nil
		}
	default:
		return nil, &types.BadRequestError{Message: "选项类型必须是 kind、fg 或 vessel"}
	}
	return nil, errors.New("SC 选项模型未初始化")
}

func validScOptionType(optionType string) bool {
	switch optionType {
	case ScOptionTypeKind, ScOptionTypeFG, ScOptionTypeVessel:
		return true
	default:
		return false
	}
}

func scOptionWriteError(err error) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate entry") {
		return &types.BadRequestError{Message: "同类型的选项名称已存在"}
	}
	return err
}

func (s *Service) resolveScOption(ctx context.Context, optionType string, optionID, currentOptionID int64) (scOptionSelection, error) {
	optionType = strings.TrimSpace(optionType)
	if optionID <= 0 {
		return scOptionSelection{}, nil
	}
	model, err := s.scOptionModel(optionType)
	if err != nil {
		return scOptionSelection{}, err
	}
	row, err := model.FindOne(ctx, optionID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return scOptionSelection{}, &types.BadRequestError{Message: fmt.Sprintf("%s 选项不存在", scOptionLabel(optionType))}
		}
		return scOptionSelection{}, err
	}
	if row == nil {
		return scOptionSelection{}, &types.BadRequestError{Message: fmt.Sprintf("%s 选项不存在", scOptionLabel(optionType))}
	}
	if row.IsEnabled != 1 && row.ID != currentOptionID {
		return scOptionSelection{}, &types.BadRequestError{Message: fmt.Sprintf("%s 选项已停用", scOptionLabel(optionType))}
	}
	return scOptionSelection{ID: row.ID, Name: strings.TrimSpace(row.Name)}, nil
}

func scOptionLabel(optionType string) string {
	switch optionType {
	case ScOptionTypeKind:
		return "Kind"
	case ScOptionTypeFG:
		return "FG"
	case ScOptionTypeVessel:
		return "Vessel"
	default:
		return "SC"
	}
}

func scOptionNullID(id int64) sql.NullInt64 {
	return nullablePositiveID(id)
}

func nullablePositiveID(id int64) sql.NullInt64 {
	return sql.NullInt64{Int64: id, Valid: id > 0}
}

func scOptionID(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func mapScOptionModelToTypes(optionType string, row *moviex.GScOption) *types.ScOption {
	if row == nil {
		return nil
	}
	return &types.ScOption{
		ID:              row.ID,
		OptionType:      optionType,
		OptionName:      row.Name,
		SortOrder:       row.SortOrder,
		IsEnabled:       row.IsEnabled,
		ScEventTimes:    row.ScEventTimes,
		LastScEventTime: row.LastScEventTime,
		CreatedOn:       row.CreatedOn,
		UpdatedOn:       row.UpdatedOn,
	}
}
