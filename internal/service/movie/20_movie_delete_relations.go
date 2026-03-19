package movie

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/data/modelx/spiderx"
	"rudy_gc/internal/consts"
)

type movieDeleteContext struct {
	JavID      string
	MovieName  string
	Movie      *moviex.AMovie
	Minfo      *moviex.BmMinfo
	Murl       *moviex.BmMurl
	Item       *moviex.EItem
	Detail     *spiderx.DDetail
	Film       *moviex.VFilm
	RankRows   []*moviex.CRank
	GListRows  []*moviex.GList
	ComeScRows []*moviex.GSc
	CastIDs    []int64
	GenreIDs   []int64
	DirectorID int64
	LabelID    int64
	MakerID    int64
	PrefixID   int64
	Now        int64
}

type deletedMovieSnapshot struct {
	JavID       string            `json:"jav_id"`
	MovieName   string            `json:"movie_name"`
	Movie       *moviex.AMovie    `json:"movie,omitempty"`
	Minfo       *moviex.BmMinfo   `json:"minfo,omitempty"`
	Murl        *moviex.BmMurl    `json:"murl,omitempty"`
	Item        *moviex.EItem     `json:"item,omitempty"`
	Detail      *spiderx.DDetail  `json:"detail,omitempty"`
	Film        *moviex.VFilm     `json:"film,omitempty"`
	CastIDs     []int64           `json:"cast_ids,omitempty"`
	GenreIDs    []int64           `json:"genre_ids,omitempty"`
	RankIDs     []int64           `json:"rank_ids,omitempty"`
	GListNames  []string          `json:"g_list_names,omitempty"`
	ComeScNames []string          `json:"come_sc_names,omitempty"`
	Extra       map[string]string `json:"extra,omitempty"`
}

type castScInfo struct {
	ScTimes    int64
	ComeTimes  int64
	LastScTime int64
}

type castRankInfo struct {
	Rank500MovieNumber int64
	Rank20MovieNumber  int64
	Rank1MovieNumber   int64
	HighestRank        int64
	RankTimes          int64
}

func (s *Service) loadMovieDeleteContext(ctx context.Context, javID, fallbackName string) (*movieDeleteContext, error) {
	now := time.Now().Unix()
	out := &movieDeleteContext{
		JavID: javID,
		Now:   now,
	}

	movieRow, err := s.deps.MovieModel.FindOneByJavId(ctx, javID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("find a_movie by jav_id failed: %w", err)
	}
	out.Movie = movieRow
	if movieRow != nil {
		out.MovieName = strings.TrimSpace(movieRow.Name)
		out.DirectorID = movieRow.DirectorId
		out.LabelID = movieRow.LabelId
		out.MakerID = movieRow.MakerId
		out.PrefixID = movieRow.PrefixId
	}

	minfoRow, err := s.deps.MinfoModel.FindOneByJavId(ctx, javID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("find bm_minfo by jav_id failed: %w", err)
	}
	out.Minfo = minfoRow

	murlRow, err := s.deps.MurlModel.FindOneByJavId(ctx, javID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("find bm_murl by jav_id failed: %w", err)
	}
	out.Murl = murlRow

	itemRow, err := s.deps.ItemModel.FindOneByJavId(ctx, javID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("find e_item by jav_id failed: %w", err)
	}
	out.Item = itemRow

	detailRow, err := s.deps.DetailModel.FindOneByJavId(ctx, javID)
	if err != nil && !errors.Is(err, spiderx.ErrNotFound) {
		return nil, fmt.Errorf("find d_detail by jav_id failed: %w", err)
	}
	out.Detail = detailRow

	filmRow, err := s.deps.FilmModel.FindOneByMovieJavId(ctx, javID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return nil, fmt.Errorf("find v_film by movie_jav_id failed: %w", err)
	}
	out.Film = filmRow

	out.MovieName = firstNonEmpty(
		out.MovieName,
		fallbackName,
		nameFromItem(out.Item),
		nameFromDetail(out.Detail),
		nameFromFilm(out.Film),
		nameFromMinfo(out.Minfo),
		nameFromMurl(out.Murl),
	)

	if out.Film == nil && out.MovieName != "" {
		filmByName, ferr := s.deps.FilmModel.FindOneByMovieName(ctx, out.MovieName)
		if ferr != nil && !errors.Is(ferr, moviex.ErrNotFound) {
			return nil, fmt.Errorf("find v_film by movie_name failed: %w", ferr)
		}
		if filmByName != nil && strings.TrimSpace(filmByName.MovieJavId) == javID {
			out.Film = filmByName
		}
	}

	castIDs, err := s.deps.MovieCastModel.ListCastIDsByMovieJavId(ctx, javID)
	if err != nil {
		return nil, fmt.Errorf("list cast ids by movie_jav_id failed: %w", err)
	}
	out.CastIDs = uniqueInt64s(castIDs)

	genreIDs, err := s.deps.MovieGenreModel.ListGenreIDsByMovieJavId(ctx, javID)
	if err != nil {
		return nil, fmt.Errorf("list genre ids by movie_jav_id failed: %w", err)
	}
	out.GenreIDs = uniqueInt64s(genreIDs)

	rankRows, err := s.deps.RankModel.ListByMovieJavId(ctx, javID)
	if err != nil {
		return nil, fmt.Errorf("list c_rank by movie_jav_id failed: %w", err)
	}
	out.RankRows = rankRows

	gListRows, err := s.deps.GListModel.ListByMovieJavId(ctx, javID)
	if err != nil {
		return nil, fmt.Errorf("list g_list by movie_jav_id failed: %w", err)
	}
	out.GListRows = gListRows

	if out.MovieName != "" {
		comeScRows, err := s.deps.ScModel.ListByComeMovieName(ctx, out.MovieName)
		if err != nil {
			return nil, fmt.Errorf("list g_sc by come_movie_name failed: %w", err)
		}
		out.ComeScRows = comeScRows
	}

	return out, nil
}

func (s *Service) upsertDeletedMovie(ctx context.Context, delCtx *movieDeleteContext, deleteSource string) error {
	snapshotJSON, err := buildDeletedMovieSnapshotJSON(delCtx)
	if err != nil {
		return fmt.Errorf("marshal deleted movie snapshot failed: %w", err)
	}

	existing, err := s.deps.DeletedMovieModel.FindOneByJavId(ctx, delCtx.JavID)
	if err != nil && !errors.Is(err, moviex.ErrNotFound) {
		return fmt.Errorf("find e_deleted_movie by jav_id failed: %w", err)
	}

	name := firstNonEmpty(delCtx.MovieName, delCtx.JavID)
	if existing != nil {
		existing.Name = name
		existing.DeleteSource = deleteSource
		existing.DeletedOn = delCtx.Now
		existing.SnapshotJson = snapshotJSON
		existing.UpdatedOn = delCtx.Now
		if err := s.deps.DeletedMovieModel.Update(ctx, existing); err != nil {
			return fmt.Errorf("update e_deleted_movie failed: %w", err)
		}
		return nil
	}

	_, err = s.deps.DeletedMovieModel.Insert(ctx, &moviex.EDeletedMovie{
		JavId:        delCtx.JavID,
		Name:         name,
		DeleteSource: deleteSource,
		DeletedOn:    delCtx.Now,
		SnapshotJson: snapshotJSON,
		CreatedOn:    delCtx.Now,
		UpdatedOn:    delCtx.Now,
	})
	if err != nil {
		return fmt.Errorf("insert e_deleted_movie failed: %w", err)
	}
	return nil
}

func (s *Service) deleteMovieRows(ctx context.Context, delCtx *movieDeleteContext) error {
	for _, rankRow := range delCtx.RankRows {
		if rankRow == nil {
			continue
		}
		if err := s.deps.RankModel.Delete(ctx, rankRow.Id); err != nil {
			return fmt.Errorf("delete c_rank(%d) failed: %w", rankRow.Id, err)
		}
	}

	for _, row := range delCtx.GListRows {
		if row == nil {
			continue
		}
		if err := s.deps.GListModel.Delete(ctx, row.Id); err != nil {
			return fmt.Errorf("delete g_list(%d) failed: %w", row.Id, err)
		}
	}

	for _, castID := range delCtx.CastIDs {
		row, err := s.deps.MovieCastModel.FindOneByMovieJavIdCastId(ctx, delCtx.JavID, castID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return fmt.Errorf("find amr_movie_cast(%s,%d) failed: %w", delCtx.JavID, castID, err)
		}
		if row == nil {
			continue
		}
		if err := s.deps.MovieCastModel.Delete(ctx, row.Id); err != nil {
			return fmt.Errorf("delete amr_movie_cast(%d) failed: %w", row.Id, err)
		}
	}

	for _, genreID := range delCtx.GenreIDs {
		row, err := s.deps.MovieGenreModel.FindOneByMovieJavIdGenreId(ctx, delCtx.JavID, genreID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return fmt.Errorf("find amr_movie_genre(%s,%d) failed: %w", delCtx.JavID, genreID, err)
		}
		if row == nil {
			continue
		}
		if err := s.deps.MovieGenreModel.Delete(ctx, row.Id); err != nil {
			return fmt.Errorf("delete amr_movie_genre(%d) failed: %w", row.Id, err)
		}
	}

	for _, row := range delCtx.ComeScRows {
		if row == nil || strings.TrimSpace(row.ComeMovieName) == "" {
			continue
		}
		row.ComeMovieName = ""
		row.UpdatedOn = delCtx.Now
		if err := s.deps.ScModel.Update(ctx, row); err != nil {
			return fmt.Errorf("clear g_sc come_movie_name(%d) failed: %w", row.Id, err)
		}
	}

	if delCtx.Film != nil {
		if err := s.deps.FilmModel.Delete(ctx, delCtx.Film.Id); err != nil {
			return fmt.Errorf("delete v_film(%d) failed: %w", delCtx.Film.Id, err)
		}
	}
	if delCtx.Detail != nil {
		if err := s.deps.DetailModel.Delete(ctx, delCtx.Detail.Id); err != nil {
			return fmt.Errorf("delete d_detail(%d) failed: %w", delCtx.Detail.Id, err)
		}
	}
	if delCtx.Item != nil {
		if err := s.deps.ItemModel.Delete(ctx, delCtx.Item.Id); err != nil {
			return fmt.Errorf("delete e_item(%d) failed: %w", delCtx.Item.Id, err)
		}
	}
	if delCtx.Murl != nil {
		if err := s.deps.MurlModel.Delete(ctx, delCtx.Murl.Id); err != nil {
			return fmt.Errorf("delete bm_murl(%d) failed: %w", delCtx.Murl.Id, err)
		}
	}
	if delCtx.Minfo != nil {
		if err := s.deps.MinfoModel.Delete(ctx, delCtx.Minfo.Id); err != nil {
			return fmt.Errorf("delete bm_minfo(%d) failed: %w", delCtx.Minfo.Id, err)
		}
	}
	if delCtx.Movie != nil {
		if err := s.deps.MovieModel.Delete(ctx, delCtx.Movie.Id); err != nil {
			return fmt.Errorf("delete a_movie(%d) failed: %w", delCtx.Movie.Id, err)
		}
	}

	return nil
}

func (s *Service) rebuildMovieDeleteAffectedStats(ctx context.Context, delCtx *movieDeleteContext) error {
	if err := s.rebuildScMovieNumbers(ctx, delCtx); err != nil {
		return err
	}

	for _, castID := range delCtx.CastIDs {
		if err := s.rebuildCastStatsByID(ctx, castID, delCtx.Now); err != nil {
			return err
		}
	}
	for _, genreID := range delCtx.GenreIDs {
		if err := s.updateGenreMovieNumbersByID(ctx, genreID, delCtx.Now); err != nil {
			return err
		}
	}
	if err := s.updateDirectorMovieNumbersByID(ctx, delCtx.DirectorID, delCtx.Now); err != nil {
		return err
	}
	if err := s.updateLabelMovieNumbersByID(ctx, delCtx.LabelID, delCtx.Now); err != nil {
		return err
	}
	if err := s.updateMakerMovieNumbersByID(ctx, delCtx.MakerID, delCtx.Now); err != nil {
		return err
	}
	if err := s.updatePrefixMovieNumbersByID(ctx, delCtx.PrefixID, delCtx.Now); err != nil {
		return err
	}

	return nil
}

func (s *Service) rebuildScMovieNumbers(ctx context.Context, delCtx *movieDeleteContext) error {
	scNames := make([]string, 0, len(delCtx.GListRows)+len(delCtx.ComeScRows))
	for _, row := range delCtx.GListRows {
		if row != nil && strings.TrimSpace(row.ScName) != "" {
			scNames = append(scNames, strings.TrimSpace(row.ScName))
		}
	}
	for _, row := range delCtx.ComeScRows {
		if row != nil && strings.TrimSpace(row.Name) != "" {
			scNames = append(scNames, strings.TrimSpace(row.Name))
		}
	}

	for _, scName := range uniqueStrings(scNames) {
		scRow, err := s.deps.ScModel.FindOneByName(ctx, scName)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return fmt.Errorf("find g_sc by name failed: %w", err)
		}
		if scRow == nil {
			continue
		}

		gLists, err := s.deps.GListModel.ListByScName(ctx, scName)
		if err != nil {
			return fmt.Errorf("list g_list by sc_name failed: %w", err)
		}

		movieNumber := int64(len(gLists))
		if scRow.MovieNumber == movieNumber {
			continue
		}

		scRow.MovieNumber = movieNumber
		scRow.UpdatedOn = delCtx.Now
		if err := s.deps.ScModel.Update(ctx, scRow); err != nil {
			return fmt.Errorf("update g_sc movie_number(%s) failed: %w", scName, err)
		}
	}

	return nil
}

func (s *Service) rebuildCastStatsByID(ctx context.Context, castID, now int64) error {
	if castID <= 0 {
		return nil
	}

	castRow, err := s.deps.CastModel.FindOne(ctx, castID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_cast(%d) failed: %w", castID, err)
	}
	if castRow == nil {
		return nil
	}

	movieJavIDs, err := s.deps.MovieCastModel.ListMovieJavIDsByCastID(ctx, castID)
	if err != nil {
		return fmt.Errorf("list movie jav ids by cast id failed: %w", err)
	}
	movieJavIDs = uniqueStrings(movieJavIDs)

	movieNumber, ownedMovieNumber, err := s.deps.CastModel.GetMovieNumbersByID(ctx, castID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get cast movie numbers failed: %w", err)
	}

	scInfo, err := s.buildCastScInfo(ctx, movieJavIDs)
	if err != nil {
		return err
	}
	rankInfo, err := s.buildCastRankInfo(ctx, movieJavIDs)
	if err != nil {
		return err
	}

	if castRow.MovieNumber == movieNumber &&
		castRow.OwnedMovieNumber == ownedMovieNumber &&
		castRow.ScTimes == scInfo.ScTimes &&
		castRow.ComeTimes == scInfo.ComeTimes &&
		castRow.LastScTime == scInfo.LastScTime &&
		castRow.Rank500MovieNumber == rankInfo.Rank500MovieNumber &&
		castRow.Rank20MovieNumber == rankInfo.Rank20MovieNumber &&
		castRow.Rank1MovieNumber == rankInfo.Rank1MovieNumber &&
		castRow.HighestRank == rankInfo.HighestRank &&
		castRow.RankTimes == rankInfo.RankTimes {
		return nil
	}

	castRow.MovieNumber = movieNumber
	castRow.OwnedMovieNumber = ownedMovieNumber
	castRow.ScTimes = scInfo.ScTimes
	castRow.ComeTimes = scInfo.ComeTimes
	castRow.LastScTime = scInfo.LastScTime
	castRow.Rank500MovieNumber = rankInfo.Rank500MovieNumber
	castRow.Rank20MovieNumber = rankInfo.Rank20MovieNumber
	castRow.Rank1MovieNumber = rankInfo.Rank1MovieNumber
	castRow.HighestRank = rankInfo.HighestRank
	castRow.RankTimes = rankInfo.RankTimes
	castRow.UpdatedOn = now

	if err := s.deps.CastModel.Update(ctx, castRow); err != nil {
		return fmt.Errorf("update am_cast(%d) failed: %w", castID, err)
	}
	return nil
}

func (s *Service) buildCastScInfo(ctx context.Context, movieJavIDs []string) (castScInfo, error) {
	var out castScInfo
	if len(movieJavIDs) == 0 {
		return out, nil
	}

	pairs := make(map[string]int64)
	scNames := make(map[string]struct{})
	for _, movieJavID := range uniqueStrings(movieJavIDs) {
		rows, err := s.deps.GListModel.ListByMovieJavId(ctx, movieJavID)
		if err != nil {
			return out, fmt.Errorf("list g_list by movie_jav_id failed: %w", err)
		}
		for _, row := range rows {
			if row == nil || strings.TrimSpace(row.ScName) == "" {
				continue
			}
			key := movieJavID + "\x00" + strings.TrimSpace(row.ScName)
			if old, ok := pairs[key]; ok {
				if old == consts.GListIsCome || row.IsCome != consts.GListIsCome {
					continue
				}
			}
			pairs[key] = row.IsCome
			scNames[strings.TrimSpace(row.ScName)] = struct{}{}
		}
	}

	if len(pairs) == 0 {
		return out, nil
	}

	names := make([]string, 0, len(scNames))
	for scName := range scNames {
		names = append(names, scName)
	}
	scRows, err := s.deps.ScModel.ListByNames(ctx, names)
	if err != nil {
		return out, fmt.Errorf("list g_sc by names failed: %w", err)
	}
	scTimeMap := make(map[string]int64, len(scRows))
	for _, row := range scRows {
		if row == nil || strings.TrimSpace(row.Name) == "" {
			continue
		}
		scTimeMap[strings.TrimSpace(row.Name)] = row.ScTime
	}

	for key, isCome := range pairs {
		parts := strings.SplitN(key, "\x00", 2)
		if len(parts) != 2 {
			continue
		}
		out.ScTimes++
		if isCome == consts.GListIsCome {
			out.ComeTimes++
		}
		if scTimeMap[parts[1]] > out.LastScTime {
			out.LastScTime = scTimeMap[parts[1]]
		}
	}

	return out, nil
}

func (s *Service) buildCastRankInfo(ctx context.Context, movieJavIDs []string) (castRankInfo, error) {
	var out castRankInfo

	for _, movieJavID := range uniqueStrings(movieJavIDs) {
		minfo, err := s.deps.MinfoModel.FindOneByJavId(ctx, movieJavID)
		if err != nil {
			if errors.Is(err, moviex.ErrNotFound) {
				continue
			}
			return out, fmt.Errorf("find bm_minfo by jav_id failed: %w", err)
		}
		if minfo == nil {
			continue
		}
		if minfo.HighestRank <= 0 || minfo.HighestRank >= 1000 {
			continue
		}

		if minfo.DaysInRank > 0 {
			out.RankTimes += minfo.DaysInRank
		}
		if out.HighestRank == 0 || minfo.HighestRank < out.HighestRank {
			out.HighestRank = minfo.HighestRank
		}
		if minfo.HighestRank <= 500 {
			out.Rank500MovieNumber++
		}
		if minfo.HighestRank <= 20 {
			out.Rank20MovieNumber++
		}
		if minfo.HighestRank == 1 {
			out.Rank1MovieNumber++
		}
	}

	return out, nil
}

func (s *Service) updateGenreMovieNumbersByID(ctx context.Context, genreID, now int64) error {
	if genreID <= 0 {
		return nil
	}
	row, err := s.deps.GenreModel.FindOne(ctx, genreID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_genre(%d) failed: %w", genreID, err)
	}
	if row == nil {
		return nil
	}

	movieNumber, ownedMovieNumber, err := s.deps.GenreModel.GetMovieNumbersByID(ctx, genreID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get am_genre movie numbers failed: %w", err)
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now
	if err := s.deps.GenreModel.Update(ctx, row); err != nil {
		return fmt.Errorf("update am_genre(%d) failed: %w", genreID, err)
	}
	return nil
}

func (s *Service) updateDirectorMovieNumbersByID(ctx context.Context, directorID, now int64) error {
	if directorID <= 0 {
		return nil
	}
	row, err := s.deps.DirectorModel.FindOne(ctx, directorID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_director(%d) failed: %w", directorID, err)
	}
	if row == nil {
		return nil
	}

	movieNumber, ownedMovieNumber, err := s.deps.DirectorModel.GetMovieNumbersByID(ctx, directorID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get am_director movie numbers failed: %w", err)
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now
	if err := s.deps.DirectorModel.Update(ctx, row); err != nil {
		return fmt.Errorf("update am_director(%d) failed: %w", directorID, err)
	}
	return nil
}

func (s *Service) updateLabelMovieNumbersByID(ctx context.Context, labelID, now int64) error {
	if labelID <= 0 {
		return nil
	}
	row, err := s.deps.LabelModel.FindOne(ctx, labelID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_label(%d) failed: %w", labelID, err)
	}
	if row == nil {
		return nil
	}

	movieNumber, ownedMovieNumber, err := s.deps.LabelModel.GetMovieNumbersByID(ctx, labelID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get am_label movie numbers failed: %w", err)
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now
	if err := s.deps.LabelModel.Update(ctx, row); err != nil {
		return fmt.Errorf("update am_label(%d) failed: %w", labelID, err)
	}
	return nil
}

func (s *Service) updateMakerMovieNumbersByID(ctx context.Context, makerID, now int64) error {
	if makerID <= 0 {
		return nil
	}
	row, err := s.deps.MakerModel.FindOne(ctx, makerID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_maker(%d) failed: %w", makerID, err)
	}
	if row == nil {
		return nil
	}

	movieNumber, ownedMovieNumber, err := s.deps.MakerModel.GetMovieNumbersByID(ctx, makerID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get am_maker movie numbers failed: %w", err)
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now
	if err := s.deps.MakerModel.Update(ctx, row); err != nil {
		return fmt.Errorf("update am_maker(%d) failed: %w", makerID, err)
	}
	return nil
}

func (s *Service) updatePrefixMovieNumbersByID(ctx context.Context, prefixID, now int64) error {
	if prefixID <= 0 {
		return nil
	}
	row, err := s.deps.PrefixModel.FindOne(ctx, prefixID)
	if err != nil {
		if errors.Is(err, moviex.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("find am_prefix(%d) failed: %w", prefixID, err)
	}
	if row == nil {
		return nil
	}

	movieNumber, ownedMovieNumber, err := s.deps.PrefixModel.GetMovieNumbersByID(ctx, prefixID, consts.FilmIsNotRemoved)
	if err != nil {
		return fmt.Errorf("get am_prefix movie numbers failed: %w", err)
	}
	if row.MovieNumber == movieNumber && row.OwnedMovieNumber == ownedMovieNumber {
		return nil
	}

	row.MovieNumber = movieNumber
	row.OwnedMovieNumber = ownedMovieNumber
	row.UpdatedOn = now
	if err := s.deps.PrefixModel.Update(ctx, row); err != nil {
		return fmt.Errorf("update am_prefix(%d) failed: %w", prefixID, err)
	}
	return nil
}

func buildDeletedMovieSnapshotJSON(delCtx *movieDeleteContext) (string, error) {
	snapshot := deletedMovieSnapshot{
		JavID:       delCtx.JavID,
		MovieName:   delCtx.MovieName,
		Movie:       delCtx.Movie,
		Minfo:       delCtx.Minfo,
		Murl:        delCtx.Murl,
		Item:        delCtx.Item,
		Detail:      delCtx.Detail,
		Film:        delCtx.Film,
		CastIDs:     delCtx.CastIDs,
		GenreIDs:    delCtx.GenreIDs,
		RankIDs:     collectRankIDs(delCtx.RankRows),
		GListNames:  collectGListNames(delCtx.GListRows),
		ComeScNames: collectScNames(delCtx.ComeScRows),
		Extra: map[string]string{
			"generated_at": time.Unix(delCtx.Now, 0).Format(time.RFC3339),
		},
	}
	buf, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func collectRankIDs(rows []*moviex.CRank) []int64 {
	out := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row != nil && row.Id > 0 {
			out = append(out, row.Id)
		}
	}
	return out
}

func collectGListNames(rows []*moviex.GList) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row != nil && strings.TrimSpace(row.Name) != "" {
			out = append(out, strings.TrimSpace(row.Name))
		}
	}
	return uniqueStrings(out)
}

func collectScNames(rows []*moviex.GSc) []string {
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row != nil && strings.TrimSpace(row.Name) != "" {
			out = append(out, strings.TrimSpace(row.Name))
		}
	}
	return uniqueStrings(out)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func uniqueInt64s(items []int64) []int64 {
	seen := make(map[int64]struct{}, len(items))
	out := make([]int64, 0, len(items))
	for _, item := range items {
		if item <= 0 {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func nameFromItem(item *moviex.EItem) string {
	if item == nil {
		return ""
	}
	return item.Name
}

func nameFromDetail(detail *spiderx.DDetail) string {
	if detail == nil {
		return ""
	}
	return detail.Name
}

func nameFromFilm(film *moviex.VFilm) string {
	if film == nil {
		return ""
	}
	return film.MovieName
}

func nameFromMinfo(minfo *moviex.BmMinfo) string {
	if minfo == nil {
		return ""
	}
	return minfo.Name
}

func nameFromMurl(murl *moviex.BmMurl) string {
	if murl == nil {
		return ""
	}
	return murl.Name
}
