package sc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/model/modelx/moviex"
	"rudy_gc/local/api/internal/types"
	moviedomain "rudy_gc/local/api/internal/types/movie"
)

func (s *Service) UpdateEventMeta(ctx context.Context, in *types.ScEventEditForm) (result *types.GSc, err error) {
	if in == nil {
		return nil, errors.New("nil input")
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, types.ErrNotFound
	}
	if in.DurationMinutes < 0 {
		return nil, &types.BadRequestError{Message: "时长不能为负数"}
	}

	row, err := s.deps.ScModel.FindOneByName(ctx, name)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if row == nil {
		return nil, types.ErrNotFound
	}
	kindOption, err := s.resolveScOption(ctx, ScOptionTypeKind, in.KindOptionID, scOptionID(row.KindOptionId))
	if err != nil {
		return nil, err
	}
	fgOption, err := s.resolveScOption(ctx, ScOptionTypeFG, in.FGOptionID, scOptionID(row.FgOptionId))
	if err != nil {
		return nil, err
	}
	vesselOption, err := s.resolveScOption(ctx, ScOptionTypeVessel, in.VesselOptionID, scOptionID(row.VesselOptionId))
	if err != nil {
		return nil, err
	}

	glRows, err := s.deps.GListModel.ListByScName(ctx, name)
	if err != nil {
		return nil, err
	}

	comeGListID := in.ComeGListId
	comeMovieName, castOptions, err := s.resolveScEventComeMovie(ctx, glRows, comeGListID)
	if err != nil {
		return nil, err
	}
	if in.MovieCastPersonID > 0 {
		if comeGListID <= 0 {
			return nil, &types.BadRequestError{Message: "未选择 Come Movie 时，Lead Cast 必须为空"}
		}
		if !containsScEventCastPersonID(castOptions, in.MovieCastPersonID) {
			return nil, &types.BadRequestError{Message: "Lead Cast 必须从当前 Come Movie 的演员中选择"}
		}
	}

	now := time.Now().Unix()
	row.KindOptionId = scOptionNullID(kindOption.ID)
	row.Duration = in.DurationMinutes
	row.FgOptionId = scOptionNullID(fgOption.ID)
	row.VesselOptionId = scOptionNullID(vesselOption.ID)
	row.MovieCastPersonId = nullablePositiveID(in.MovieCastPersonID)
	row.ComeMovieName = comeMovieName
	remarks := strings.TrimSpace(in.Remarks)
	row.Remarks = remarks
	var imagePath *string
	imageCommitted := false
	if in.Image != nil {
		path, saveErr := s.saveScEventImage(ctx, row.Id, in.Image)
		if saveErr != nil {
			return nil, saveErr
		}
		imagePath = &path
		defer func() {
			if !imageCommitted {
				err = errors.Join(err, discardScEventImage(path))
			}
		}()
	}
	updatedRows, updatedEvent, err := s.deps.ScEventWriteModel.UpdateEventMeta(ctx, row, comeGListID, now, imagePath)
	imageCommitted = updatedEvent != nil
	if err != nil {
		if imageCommitted {
			return nil, fmt.Errorf("SC Event 信息已保存，缓存刷新失败: %w", err)
		}
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if err := s.refreshPersonScSnapshotsByScNames(ctx, now, name); err != nil {
		return nil, err
	}
	javIDs, fc2IDs := splitGListMovieIdentities(updatedRows)
	if err := s.AddMovieAndCastScInfo(ctx, javIDs); err != nil {
		return nil, err
	}
	if err := s.AddFC2MovieScInfo(ctx, fc2IDs); err != nil {
		return nil, err
	}
	return s.mapScModelToTypes(ctx, updatedEvent)
}

func (s *Service) UpdateScEventMovieStatus(ctx context.Context, scName string, gListID, isSc int64) (*types.GList, error) {
	scName = strings.TrimSpace(scName)
	if scName == "" || gListID <= 0 {
		return nil, types.ErrNotFound
	}
	if isSc != consts.GListIsSc && isSc != consts.GListIsNotSc {
		return nil, &types.BadRequestError{Message: "影片 SC 状态无效"}
	}

	now := time.Now().Unix()
	target, _, err := s.deps.ScEventWriteModel.UpdateMovieStatus(ctx, scName, gListID, isSc, now)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if err := s.refreshPersonScSnapshotsByScNames(ctx, now, scName); err != nil {
		return nil, err
	}
	javIDs, fc2IDs := splitGListMovieIdentities([]*moviex.GList{target})
	if err := s.AddMovieAndCastScInfo(ctx, javIDs); err != nil {
		return nil, err
	}
	if err := s.AddFC2MovieScInfo(ctx, fc2IDs); err != nil {
		return nil, err
	}
	return mapGListModelToTypes(target), nil
}

func (s *Service) resolveScEventComeMovie(ctx context.Context, glRows []*moviex.GList, comeGListID int64) (string, []*types.ScEventCastOption, error) {
	if comeGListID <= 0 {
		return "", nil, nil
	}

	var selectedRow *moviex.GList
	for _, row := range glRows {
		if row == nil {
			continue
		}
		if row.Id != comeGListID {
			continue
		}
		if row.IsSc != consts.GListIsSc {
			return "", nil, &types.BadRequestError{Message: "Come Movie 必须从 is_sc=2 的电影中选择"}
		}
		selectedRow = row
		break
	}
	if selectedRow == nil {
		return "", nil, &types.BadRequestError{Message: "Come Movie 必须从 is_sc=2 的电影中选择"}
	}

	switch strings.TrimSpace(selectedRow.MovieType) {
	case moviedomain.SehuatangThreadMovieTypeJAV:
		comeMovieName := ""
		movieRow, err := s.movieFindOneByJavID(ctx, selectedRow.MovieJavId)
		if err != nil && !errors.Is(err, moviex.ErrNotFound) {
			return "", nil, fmt.Errorf("load come movie failed: %w", err)
		}
		if movieRow != nil {
			comeMovieName = strings.TrimSpace(movieRow.Name)
		}
		if comeMovieName == "" {
			comeMovieName = strings.TrimSpace(parseGListMovieName(mapGListModelToTypes(selectedRow)))
		}
		castOptions, err := s.listMovieCastOptions(ctx, selectedRow.MovieJavId)
		if err != nil {
			return "", nil, fmt.Errorf("load come movie casts failed: %w", err)
		}
		return comeMovieName, castOptions, nil
	case moviedomain.SehuatangThreadMovieTypeFC2PPV:
		movieRow, err := s.deps.FC2MovieModel.FindOne(ctx, selectedRow.FC2MovieId)
		if err != nil {
			return "", nil, fmt.Errorf("load fc2 come movie failed: %w", err)
		}
		return strings.TrimSpace(movieRow.FC2Key), []*types.ScEventCastOption{}, nil
	default:
		return "", nil, &types.BadRequestError{Message: "Come Movie 影片类型无效"}
	}
}

func splitGListMovieIdentities(rows []*moviex.GList) (map[string]struct{}, map[int64]struct{}) {
	javIDs := make(map[string]struct{})
	fc2IDs := make(map[int64]struct{})
	for _, row := range rows {
		if row == nil {
			continue
		}
		switch strings.TrimSpace(row.MovieType) {
		case moviedomain.SehuatangThreadMovieTypeJAV:
			if javID := strings.TrimSpace(row.MovieJavId); javID != "" {
				javIDs[javID] = struct{}{}
			}
		case moviedomain.SehuatangThreadMovieTypeFC2PPV:
			if row.FC2MovieId > 0 {
				fc2IDs[row.FC2MovieId] = struct{}{}
			}
		}
	}
	return javIDs, fc2IDs
}

func containsScEventCastPersonID(items []*types.ScEventCastOption, personID int64) bool {
	if personID <= 0 {
		return false
	}
	for _, item := range items {
		if item != nil && item.PersonID == personID {
			return true
		}
	}
	return false
}
