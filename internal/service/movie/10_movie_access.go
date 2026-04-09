package movie

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/types"
	"sort"
	"time"
)

const (
	fieldTypeName int = iota
	fieldTypeEncode
	defaultMovieJavId      = "javli6a53m"
	rankDayDefaultPageSize = 18
	rankDayMaxPageSize     = 20000
)

func (s *Service) GetMovieDetailByName(ctx context.Context, movieName string) (*types.MovieDetail, error) {
	movie, err := s.findOrFallbackMovie(ctx, movieName)
	if err != nil {
		return nil, err
	}
	return s.buildMovieDetail(ctx, movie)
}

func (s *Service) AddToDownloadLater(ctx context.Context, javId string) (int64, error) {
	row, err := s.deps.MinfoModel.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, fmt.Errorf("add to download later failed: %w", err)
	}
	if row.NeedDownload != consts.MovieNeedDownLoadOK {
		row.NeedDownload = consts.MovieNeedDownLoadOK
		row.UpdatedOn = time.Now().Unix()
		if err := s.deps.MinfoModel.Update(ctx, row); err != nil {
			return 0, fmt.Errorf("update download later failed: %w", err)
		}
	}
	s.InvalidateMovieType(ctx, javId)
	got, err := s.deps.MinfoModel.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, fmt.Errorf("find to download later failed: %w", err)
	}
	return got.NeedDownload, nil
}

func (s *Service) RemoveFromDownloadLater(ctx context.Context, javId string) (int64, error) {
	row, err := s.deps.MinfoModel.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, fmt.Errorf("remove from download later failed: %w", err)
	}
	if row.NeedDownload != 1 {
		row.NeedDownload = 1
		row.UpdatedOn = time.Now().Unix()
		if err := s.deps.MinfoModel.Update(ctx, row); err != nil {
			return 0, fmt.Errorf("update remove download later failed: %w", err)
		}
	}
	s.InvalidateMovieType(ctx, javId)
	got, err := s.deps.MinfoModel.FindOneByJavId(ctx, javId)
	if err != nil {
		return 0, fmt.Errorf("find after remove download later failed: %w", err)
	}
	return got.NeedDownload, nil
}

func (s *Service) ListMovieFull(ctx context.Context, r *types.ListMovieFullRequest) (*types.ListMovieResponse, error) {
	rows, total, err := s.movieListQuery().ListFull(ctx, r)
	if err != nil {
		return nil, err
	}

	out := make([]*types.MovieType, 0, len(rows))
	javIds := make([]string, 0, len(rows))
	for _, mv := range rows {
		mt, err := s.GetMovieType(ctx, mv.JavId)
		if err != nil {
			return nil, err
		}
		out = append(out, mt)
		javIds = append(javIds, mv.JavId)
	}

	return &types.ListMovieResponse{
		List:   out,
		Total:  total,
		JavIds: javIds,
	}, nil
}

func (s *Service) ListMovieFullRandom(ctx context.Context, r *types.ListMovieFullRequest, n int64) (*types.ListMovieResponse, error) {
	probe := *r
	probe.Page, probe.PageSize = 1, 1

	_, total, err := s.movieListQuery().ListFull(ctx, &probe)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return &types.ListMovieResponse{List: []*types.MovieType{}, Total: 0, JavIds: []string{}}, nil
	}
	if total <= n {
		target := *r
		target.Page, target.PageSize = 1, n
		return s.ListMovieFull(ctx, &target)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	posSet := make(map[int64]struct{}, n)
	for int64(len(posSet)) < n {
		p := rng.Int63n(total) + 1
		posSet[p] = struct{}{}
	}
	positions := make([]int64, 0, n)
	for p := range posSet {
		positions = append(positions, p)
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })

	type seg struct{ start, length int64 }
	segs := make([]seg, 0, n)
	start := positions[0]
	prev := positions[0]
	for i := 1; i < len(positions); i++ {
		if positions[i] == prev+1 {
			prev = positions[i]
			continue
		}
		segs = append(segs, seg{start: start, length: prev - start + 1})
		start, prev = positions[i], positions[i]
	}
	segs = append(segs, seg{start: start, length: prev - start + 1})

	allItems := make([]*types.MovieType, 0, n)
	allIDs := make([]string, 0, n)
	collected := int64(0)
	for _, g := range segs {
		if collected >= n {
			break
		}
		target := *r
		target.Page = g.start
		target.PageSize = g.length

		resp, err := s.ListMovieFull(ctx, &target)
		if err != nil {
			return nil, err
		}
		if len(resp.List) == 0 {
			continue
		}
		for i := 0; i < len(resp.List) && collected < n; i++ {
			allItems = append(allItems, resp.List[i])
			allIDs = append(allIDs, resp.JavIds[i])
			collected++
		}
	}

	rng.Shuffle(len(allItems), func(i, j int) {
		allItems[i], allItems[j] = allItems[j], allItems[i]
		allIDs[i], allIDs[j] = allIDs[j], allIDs[i]
	})

	return &types.ListMovieResponse{
		List:   allItems,
		Total:  total,
		JavIds: allIDs,
	}, nil
}

func (s *Service) ListMovieOwned(ctx context.Context) ([]*types.MovieType, error) {
	allRows, err := s.deps.WMediaModel.FindAllLegacyFilms(ctx, consts.FilmIsNotRemoved)
	if err != nil {
		return nil, err
	}
	out := make([]*types.MovieType, 0, len(allRows))
	for _, filmRow := range allRows {
		mt, err := s.GetMovieType(ctx, filmRow.MovieJavId)
		if err != nil {
			return nil, err
		}
		out = append(out, mt)
	}
	return out, nil
}

func (s *Service) FindLatestRankDayNumber(ctx context.Context) (int64, error) {
	return s.deps.RankModel.FindLatestDayNumber(ctx)
}

func (s *Service) FindEarliestRankDayNumber(ctx context.Context) (int64, error) {
	return s.deps.RankModel.FindEarliestDayNumber(ctx)
}

func (s *Service) ListMovieTypesByRankDay(ctx context.Context, dayNumber, page, pageSize int64) ([]*types.MovieType, int64, error) {
	ranks, err := s.deps.RankModel.FindByDayNumber(ctx, dayNumber)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(ranks))
	if total == 0 {
		return []*types.MovieType{}, 0, nil
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > rankDayMaxPageSize {
		pageSize = rankDayDefaultPageSize
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []*types.MovieType{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	rankSlice := ranks[start:end]
	out := make([]*types.MovieType, 0, len(rankSlice))
	for _, rk := range rankSlice {
		if rk == nil {
			continue
		}
		mt, err := s.GetMovieType(ctx, rk.MovieJavId)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, mt)
	}
	return out, total, nil
}

func (s *Service) ListRecords(ctx context.Context, startFrom int64, typ string, limit int) ([]*types.Record, error) {
	list, err := s.deps.RecordModel.FindByStartTimeAndType(ctx, startFrom, typ, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Record, 0, len(list))
	for _, row := range list {
		if row == nil {
			continue
		}
		out = append(out, &types.Record{
			Id:           row.Id,
			Name:         row.Name,
			StartTime:    row.StartTime,
			EndTime:      row.EndTime,
			Type:         row.Type,
			DetailNumber: row.DetailNumber,
			CreatedOn:    row.CreatedOn,
			UpdatedOn:    row.UpdatedOn,
		})
	}
	return out, nil
}

func (s *Service) findOrFallbackMovie(ctx context.Context, movieName string) (*types.Movie, error) {
	return s.findOrFallbackMovieByField(ctx, fieldTypeName, movieName)
}

func (s *Service) findOrFallbackMovieByField(ctx context.Context, fieldType int, fieldValue string) (*types.Movie, error) {
	var (
		rows []*moviex.AMovie
		err  error
	)
	switch fieldType {
	case fieldTypeName:
		rows, err = s.deps.MovieModel.FindMoviesByName(ctx, fieldValue)
	case fieldTypeEncode:
		rows, err = s.deps.MovieModel.FindMoviesByEncode(ctx, fieldValue)
	default:
		return nil, fmt.Errorf("invalid fieldType: %v", fieldType)
	}
	if err != nil {
		return nil, errors.New("failed to find movie: " + err.Error())
	}
	if len(rows) == 1 {
		return mapAMovieToTypes(rows[0]), nil
	}
	if len(rows) < 1 {
		row, err := s.deps.MovieModel.FindOneByJavId(ctx, defaultMovieJavId)
		if err != nil {
			return nil, err
		}
		return mapAMovieToTypes(row), nil
	}
	movie := mapAMovieToTypes(rows[0])
	movie.Name += "(----1----)"
	movie.Title += "(----1----)"
	return movie, nil
}
