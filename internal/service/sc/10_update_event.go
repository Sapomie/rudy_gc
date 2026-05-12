package sc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
)

func (s *Service) UpdateEventMeta(ctx context.Context, in *types.ScEventEditForm) (*types.GSc, error) {
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

	glRows, err := s.deps.GListModel.ListByScName(ctx, name)
	if err != nil {
		return nil, err
	}

	comeMovieJavID := strings.TrimSpace(in.ComeMovieJavId)
	movieCast := strings.TrimSpace(in.MovieCast)
	comeMovieName, castOptions, err := s.resolveScEventComeMovie(ctx, glRows, comeMovieJavID)
	if err != nil {
		return nil, err
	}
	if movieCast != "" {
		if comeMovieJavID == "" {
			return nil, &types.BadRequestError{Message: "未选择 Come Movie 时，Lead Cast 必须为空"}
		}
		if !containsExactString(castOptions, movieCast) {
			return nil, &types.BadRequestError{Message: "Lead Cast 必须从当前 Come Movie 的演员中选择"}
		}
	}

	now := time.Now().Unix()
	changed := false
	kind := strings.TrimSpace(in.Kind)
	if row.Kind != kind {
		row.Kind = kind
		changed = true
	}
	if row.Duration != in.DurationMinutes {
		row.Duration = in.DurationMinutes
		changed = true
	}
	fg := strings.TrimSpace(in.Fg)
	if row.Fg != fg {
		row.Fg = fg
		changed = true
	}
	vessel := strings.TrimSpace(in.Vessel)
	if row.Vessel != vessel {
		row.Vessel = vessel
		changed = true
	}
	if row.MovieCast != movieCast {
		row.MovieCast = movieCast
		changed = true
	}
	if row.ComeMovieName != comeMovieName {
		row.ComeMovieName = comeMovieName
		changed = true
	}
	remarks := strings.TrimSpace(in.Remarks)
	if row.Remarks != remarks {
		row.Remarks = remarks
		changed = true
	}

	comeChanged, affectedMovieJavIDs, err := s.updateScEventComeFlags(ctx, glRows, comeMovieJavID, now)
	if err != nil {
		return nil, err
	}

	if changed {
		row.UpdatedOn = now
		if err := s.deps.ScModel.Update(ctx, row); err != nil {
			return nil, err
		}
	}
	if comeChanged {
		if err := s.refreshPersonScSnapshotsByScNames(ctx, now, name); err != nil {
			return nil, err
		}
		if err := s.AddMovieAndCastScInfo(ctx, affectedMovieJavIDs); err != nil {
			return nil, err
		}
	}

	return mapScModelToTypes(row), nil
}

func (s *Service) resolveScEventComeMovie(ctx context.Context, glRows []*moviex.GList, comeMovieJavID string) (string, []string, error) {
	comeMovieJavID = strings.TrimSpace(comeMovieJavID)
	if comeMovieJavID == "" {
		return "", nil, nil
	}

	var selectedRow *moviex.GList
	for _, row := range glRows {
		if row == nil {
			continue
		}
		if strings.TrimSpace(row.MovieJavId) != comeMovieJavID {
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

	comeMovieName := ""
	movieRow, err := s.movieFindOneByJavID(ctx, comeMovieJavID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return "", nil, fmt.Errorf("load come movie failed: %w", err)
	}
	if movieRow != nil && strings.TrimSpace(movieRow.Name) != "" {
		comeMovieName = strings.TrimSpace(movieRow.Name)
	}
	if comeMovieName == "" {
		comeMovieName = strings.TrimSpace(parseGListMovieName(mapGListModelToTypes(selectedRow)))
	}

	castOptions, err := s.listMovieCastDisplayNames(ctx, comeMovieJavID)
	if err != nil {
		return "", nil, fmt.Errorf("load come movie casts failed: %w", err)
	}
	if comeMovieName == "" {
		comeMovieName = strings.TrimSpace(comeMovieJavID)
	}

	return comeMovieName, castOptions, nil
}

func (s *Service) updateScEventComeFlags(ctx context.Context, glRows []*moviex.GList, comeMovieJavID string, now int64) (bool, map[string]struct{}, error) {
	affectedMovieJavIDs := make(map[string]struct{})
	comeMovieJavID = strings.TrimSpace(comeMovieJavID)
	changed := false
	for _, row := range glRows {
		if row == nil || strings.TrimSpace(row.MovieJavId) == "" {
			continue
		}
		want := consts.GListIsNotCome
		if comeMovieJavID != "" && strings.TrimSpace(row.MovieJavId) == comeMovieJavID {
			want = consts.GListIsCome
		}
		if row.IsCome == want {
			continue
		}
		row.IsCome = want
		row.UpdatedOn = now
		if err := s.deps.GListModel.Update(ctx, row); err != nil {
			return false, nil, err
		}
		affectedMovieJavIDs[strings.TrimSpace(row.MovieJavId)] = struct{}{}
		changed = true
	}
	return changed, affectedMovieJavIDs, nil
}

func containsExactString(items []string, want string) bool {
	want = strings.TrimSpace(want)
	if want == "" {
		return false
	}
	for _, item := range items {
		if strings.TrimSpace(item) == want {
			return true
		}
	}
	return false
}
