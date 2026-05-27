package modelx

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlc"

	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"
)

func (r *MovieReadRepo) listMovieIDsLegacy(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	if req == nil {
		req = &types.CardsListRequest{}
	}
	if isRandomView(req.View) {
		return r.listRandomCardIDsLegacy(ctx, req)
	}

	albumIDs, err := r.resolveCardAlbumIDs(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	if needCardAlbum(req) && len(albumIDs) == 0 {
		return nil, 0, nil
	}

	needAlbum := needCardAlbum(req)
	needA := needCardAMovie(req)
	needM := needCardMinfo(req)
	needW := needCardWMedia(req)
	needS := needCardGScStat(req)
	hasM2M := hasM2MCardFilters(req)

	onlyAlbumRelease := needAlbum && !hasM2M && !needA && !needM && !needW && !needS && req.OrderBy == consts.OrderByReleasingDate
	onlyA := !needAlbum && !hasM2M && needA && !needM && !needW && !needS
	onlyM := !needAlbum && !hasM2M && needM && !needA && !needW && !needS
	onlyW := !needAlbum && !hasM2M && needW && !needA && !needM && !needS
	onlyS := !needAlbum && !hasM2M && needS && !needA && !needM && !needW

	if onlyAlbumRelease {
		return r.listCardIDsFromAlbumOnly(ctx, req, albumIDs)
	}
	if onlyA {
		return r.listCardIDsFromAMovieOnly(ctx, req)
	}
	if onlyM {
		return r.listCardIDsFromMinfoOnly(ctx, req)
	}
	if onlyW {
		return r.listCardIDsFromWMediaOnly(ctx, req)
	}
	if onlyS {
		return r.listCardIDsFromGScStatOnly(ctx, req)
	}

	setAlbum, err := r.pickCardIDsFromAlbum(ctx, req, albumIDs)
	if err != nil {
		return nil, 0, err
	}
	setA, err := r.pickCardIDsFromAMovie(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setM, err := r.pickCardIDsFromMinfo(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setW, err := r.pickCardIDsFromWMedia(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setS, err := r.pickCardIDsFromGScStat(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	finalIDs := cardSetToSlice(intersectCardSets(setAlbum, setA, setM, setW, setS))
	total := int64(len(finalIDs))
	if total == 0 {
		return nil, 0, nil
	}
	page, size := legacyCardPage(req.Page, req.PageSize)
	offset := (page - 1) * size
	var ordered []string
	switch req.OrderBy {
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ordered, err = r.pageCardIDsOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank:
		ordered, err = r.pageCardIDsOnMinfo(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		ordered, err = r.pageCardIDsOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	default:
		ordered, err = r.pageCardIDsOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	}
	if err != nil {
		return nil, 0, err
	}
	return ordered, total, nil
}

func (r *MovieReadRepo) listRandomCardIDsLegacy(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	base := *req
	base.View = "cards"
	base.Page = 1
	base.PageSize = 1

	if ids, total, ok, err := r.listRandomCardIDsFromAMovie(ctx, &base, req.RandomN); ok || err != nil {
		return ids, total, err
	}

	_, total, err := r.listMovieIDsLegacy(ctx, &base)
	if err != nil || total == 0 {
		return nil, total, err
	}

	n := req.RandomN
	if n <= 0 {
		n = 6
	}
	if total <= n {
		target := *req
		target.View = "cards"
		target.Page = 1
		target.PageSize = n
		ids, _, err := r.listMovieIDsLegacy(ctx, &target)
		return ids, total, err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	positions := randomCardPositions(rng, total, n)

	ids := make([]string, 0, n)
	for _, pos := range positions {
		target := *req
		target.View = "cards"
		target.Page = pos
		target.PageSize = 1
		got, _, err := r.listMovieIDsLegacy(ctx, &target)
		if err != nil {
			return nil, 0, err
		}
		if len(got) > 0 {
			ids = append(ids, got[0])
		}
	}
	rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	if int64(len(ids)) > n {
		ids = ids[:n]
	}
	return ids, total, nil
}

func (r *MovieReadRepo) listRandomCardIDsFromAMovie(ctx context.Context, req *types.CardsListRequest, n int64) ([]string, int64, bool, error) {
	if !canUseAMovieIDFastPath(req) {
		return nil, 0, false, nil
	}
	where := r.cardAMovieWhere(req)
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("a_movie"), where)
	var total int64
	if err := queryOne(ctx, r, &total, countBuilder); err != nil {
		return nil, 0, true, err
	}
	if total == 0 {
		return nil, 0, true, nil
	}
	if n <= 0 {
		n = 6
	}
	if total <= n {
		listBuilder := applyWhere(squirrel.Select("jav_id").From("a_movie"), where).
			OrderBy(cardAMovieOrderClause(req.OrderBy, req.Order)).
			Limit(uint64(n))
		ids, err := queryStringList(ctx, r, listBuilder)
		return ids, total, true, err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	positions := randomCardPositions(rng, total, n)
	ids := make([]string, 0, n)
	for _, pos := range positions {
		listBuilder := applyWhere(squirrel.Select("jav_id").From("a_movie"), where).
			OrderBy(cardAMovieOrderClause(req.OrderBy, req.Order)).
			Offset(uint64(pos - 1)).
			Limit(1)
		got, err := queryStringList(ctx, r, listBuilder)
		if err != nil {
			return nil, 0, true, err
		}
		if len(got) > 0 {
			ids = append(ids, got[0])
		}
	}
	rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	return ids, total, true, nil
}

func randomCardPositions(rng *rand.Rand, total, n int64) []int64 {
	posSet := make(map[int64]struct{}, n)
	for int64(len(posSet)) < n {
		posSet[rng.Int63n(total)+1] = struct{}{}
	}
	positions := make([]int64, 0, len(posSet))
	for pos := range posSet {
		positions = append(positions, pos)
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })
	return positions
}

func (r *MovieReadRepo) listCardIDsFromAMovieOnly(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	where := r.cardAMovieWhere(req)
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("a_movie"), where)
	var total int64
	if err := queryOne(ctx, r, &total, countBuilder); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	page, size := legacyCardPage(req.Page, req.PageSize)
	listBuilder := applyWhere(squirrel.Select("jav_id").From("a_movie"), where).
		OrderBy(cardAMovieOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size))
	ids, err := queryStringList(ctx, r, listBuilder)
	return ids, total, err
}

func (r *MovieReadRepo) listCardIDsFromMinfoOnly(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	where := cardMinfoWhere(req)
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("bm_minfo"), where)
	var total int64
	if err := queryOne(ctx, r, &total, countBuilder); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	page, size := legacyCardPage(req.Page, req.PageSize)
	listBuilder := applyWhere(squirrel.Select("jav_id").From("bm_minfo"), where).
		OrderBy(cardMinfoOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size))
	ids, err := queryStringList(ctx, r, listBuilder)
	return ids, total, err
}

func (r *MovieReadRepo) listCardIDsFromAlbumOnly(ctx context.Context, req *types.CardsListRequest, albumIDs []int64) ([]string, int64, error) {
	if len(albumIDs) == 0 {
		return nil, 0, nil
	}
	where := cardAlbumWhere(albumIDs, req)
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("c_movie_album_item"), where)
	var total int64
	if err := queryOne(ctx, r, &total, countBuilder); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	page, size := legacyCardPage(req.Page, req.PageSize)
	listBuilder := applyWhere(squirrel.Select("movie_jav_id").From("c_movie_album_item"), where).
		OrderBy(cardAlbumOrderClause(req.Order)).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size))
	ids, err := queryStringList(ctx, r, listBuilder)
	return ids, total, err
}

func (r *MovieReadRepo) listCardIDsFromWMediaOnly(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	page, size := legacyCardPage(req.Page, req.PageSize)
	where := cardWMediaWhere(req)
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("w_media"), where)
	var total int64
	if err := queryOne(ctx, r, &total, countBuilder); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	listBuilder := applyWhere(squirrel.Select("movie_jav_id").From("w_media"), where).
		OrderBy(cardWMediaOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size))
	ids, err := queryStringList(ctx, r, listBuilder)
	return ids, total, err
}

func (r *MovieReadRepo) listCardIDsFromGScStatOnly(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	set, err := r.pickCardIDsFromGScStat(ctx, req)
	if err != nil || len(set) == 0 {
		return nil, 0, err
	}
	finalIDs := cardSetToSlice(set)
	page, size := legacyCardPage(req.Page, req.PageSize)
	ids, err := r.pageCardIDsOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, (page-1)*size, size)
	return ids, int64(len(finalIDs)), err
}

func (r *MovieReadRepo) pickCardIDsFromAMovie(ctx context.Context, req *types.CardsListRequest) (map[string]struct{}, error) {
	where := r.cardAMovieWhere(req)
	var base map[string]struct{}
	if len(where) > 0 {
		ids, err := queryStringList(ctx, r, applyWhere(squirrel.Select("jav_id").From("a_movie"), where))
		if err != nil {
			return nil, err
		}
		base = cardSliceToSet(ids)
	}
	for _, name := range parseList(req.CastNames) {
		ids, err := r.pickCardIDsByCastName(ctx, name)
		if err != nil {
			return nil, err
		}
		base = intersectTwoCardSets(base, ids)
		if len(base) == 0 {
			return base, nil
		}
	}
	for _, personID := range parseInt64List(req.PersonIds) {
		ids, err := r.pickCardIDsByPersonID(ctx, personID)
		if err != nil {
			return nil, err
		}
		base = intersectTwoCardSets(base, ids)
		if len(base) == 0 {
			return base, nil
		}
	}
	for _, name := range parseList(req.GenreNames) {
		ids, err := r.pickCardIDsByGenreName(ctx, name)
		if err != nil {
			return nil, err
		}
		base = intersectTwoCardSets(base, ids)
		if len(base) == 0 {
			return base, nil
		}
	}
	return base, nil
}

func (r *MovieReadRepo) pickCardIDsFromMinfo(ctx context.Context, req *types.CardsListRequest) (map[string]struct{}, error) {
	if !needCardMinfo(req) {
		return nil, nil
	}
	ids, err := queryStringList(ctx, r, applyWhere(squirrel.Select("jav_id").From("bm_minfo"), cardMinfoWhere(req)))
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pickCardIDsFromWMedia(ctx context.Context, req *types.CardsListRequest) (map[string]struct{}, error) {
	if !needCardWMedia(req) {
		return nil, nil
	}
	ids, err := queryStringList(ctx, r, applyWhere(squirrel.Select("movie_jav_id").Distinct().From("w_media"), cardWMediaWhere(req)))
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pickCardIDsFromGScStat(ctx context.Context, req *types.CardsListRequest) (map[string]struct{}, error) {
	if !needCardGScStat(req) {
		return nil, nil
	}
	builder := squirrel.Select("wm.movie_jav_id").
		From(fmt.Sprintf("(SELECT DISTINCT movie_jav_id FROM w_media WHERE source_type = %d) wm", consts.WMediaSourceNative)).
		LeftJoin("g_sc_stat gss ON gss.movie_jav_id = wm.movie_jav_id")
	builder = applyWhere(builder, cardGScStatWhere(req))
	ids, err := queryStringList(ctx, r, builder)
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pickCardIDsFromAlbum(ctx context.Context, req *types.CardsListRequest, albumIDs []int64) (map[string]struct{}, error) {
	if !needCardAlbum(req) {
		return nil, nil
	}
	ids, err := queryStringList(ctx, r, applyWhere(squirrel.Select("movie_jav_id").From("c_movie_album_item"), cardAlbumWhere(albumIDs, req)))
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pageCardIDsOnAMovie(ctx context.Context, ids []string, orderBy, order string, offset, limit int64) ([]string, error) {
	return queryStringList(ctx, r, squirrel.Select("jav_id").
		From("a_movie").
		Where(squirrel.Eq{"jav_id": ids}).
		OrderBy(cardAMovieOrderClause(orderBy, order)).
		Offset(uint64(offset)).
		Limit(uint64(limit)))
}

func (r *MovieReadRepo) pageCardIDsOnMinfo(ctx context.Context, ids []string, orderBy, order string, offset, limit int64) ([]string, error) {
	return queryStringList(ctx, r, squirrel.Select("jav_id").
		From("bm_minfo").
		Where(squirrel.Eq{"jav_id": ids}).
		OrderBy(cardMinfoOrderClause(orderBy, order)).
		Offset(uint64(offset)).
		Limit(uint64(limit)))
}

func (r *MovieReadRepo) pageCardIDsOnWMedia(ctx context.Context, ids []string, orderBy, order string, offset, limit int64) ([]string, error) {
	return queryStringList(ctx, r, squirrel.Select("movie_jav_id").
		From("w_media").
		Where(squirrel.Eq{"movie_jav_id": ids}).
		Where(squirrel.Eq{"source_type": consts.WMediaSourceNative}).
		OrderBy(cardWMediaOrderClause(orderBy, order)).
		Offset(uint64(offset)).
		Limit(uint64(limit)))
}

func (r *MovieReadRepo) pageCardIDsOnGScStat(ctx context.Context, ids []string, orderBy, order string, offset, limit int64) ([]string, error) {
	builder := squirrel.Select("wm.movie_jav_id").
		From(fmt.Sprintf("(SELECT DISTINCT movie_jav_id FROM w_media WHERE source_type = %d) wm", consts.WMediaSourceNative)).
		LeftJoin("g_sc_stat gss ON gss.movie_jav_id = wm.movie_jav_id").
		Where(squirrel.Eq{"wm.movie_jav_id": ids}).
		OrderBy(cardGScStatOrderClause(orderBy, order)).
		Offset(uint64(offset)).
		Limit(uint64(limit))
	return queryStringList(ctx, r, builder)
}

func (r *MovieReadRepo) cardAMovieWhere(req *types.CardsListRequest) squirrel.And {
	where := squirrel.And{}
	if req.View == "cardstoday" {
		where = append(where, squirrel.LtOrEq{"releasing_date": startOfDay(time.Now())})
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			where = append(where, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			where = append(where, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	if req.CastAgeMin != nil {
		where = append(where, squirrel.GtOrEq{"cast_average_age": int64(*req.CastAgeMin*10.0 + 0.5)})
	}
	if req.CastAgeMax != nil {
		where = append(where, squirrel.LtOrEq{"cast_average_age": int64(*req.CastAgeMax*10.0 + 0.5)})
	}
	if req.ScoreMin != nil {
		where = append(where, squirrel.GtOrEq{"score": int64(*req.ScoreMin*10.0 + 0.5)})
	}
	if req.ScoreMax != nil {
		where = append(where, squirrel.LtOrEq{"score": int64(*req.ScoreMax*10.0 + 0.5)})
	}
	if req.ViewWatchedMin != nil {
		where = append(where, squirrel.GtOrEq{"viewers_number_watched": *req.ViewWatchedMin})
	}
	if req.ViewWatchedMax != nil {
		where = append(where, squirrel.LtOrEq{"viewers_number_watched": *req.ViewWatchedMax})
	}
	if req.DirectorName != "" {
		where = append(where, squirrel.Expr("director_id = (SELECT id FROM am_director WHERE name = ? LIMIT 1)", req.DirectorName))
	}
	if req.PrefixName != "" {
		where = append(where, squirrel.Expr("prefix_id = (SELECT id FROM am_prefix WHERE name = ? LIMIT 1)", req.PrefixName))
	}
	if req.MakerName != "" {
		where = append(where, squirrel.Expr("maker_id = (SELECT id FROM am_maker WHERE name = ? LIMIT 1)", req.MakerName))
	}
	if req.LabelName != "" {
		where = append(where, squirrel.Expr("label_id = (SELECT id FROM am_label WHERE name = ? LIMIT 1)", req.LabelName))
	}
	if req.LabelJavID != "" {
		where = append(where, squirrel.Expr("label_id = (SELECT id FROM am_label WHERE jav_id = ? LIMIT 1)", req.LabelJavID))
	}
	if req.MediaOwned == consts.OwnedNotOwned {
		where = append(where, squirrel.Expr("NOT EXISTS (SELECT 1 FROM w_media wm WHERE wm.movie_jav_id = a_movie.jav_id AND wm.source_type = ?)", consts.WMediaSourceNative))
	}
	if req.NeedDownload == consts.MovieNeedDownloadNone {
		where = append(where, squirrel.Expr("NOT EXISTS (SELECT 1 FROM c_movie_album_item cai JOIN c_movie_album ca ON ca.id = cai.album_id WHERE ca.name = ? AND cai.movie_jav_id = a_movie.jav_id)", consts.MovieNeedDownloadAlbumName))
	}
	return where
}

func cardMinfoWhere(req *types.CardsListRequest) squirrel.And {
	where := squirrel.And{}
	if req.View == "cardshasrank" && req.DaysInRankMin == nil {
		where = append(where, squirrel.Gt{"days_in_rank": 0})
	}
	if req.DaysInRankMin != nil {
		where = append(where, squirrel.GtOrEq{"days_in_rank": *req.DaysInRankMin})
	}
	if req.StartRankingDateStart != "" {
		where = append(where, squirrel.GtOrEq{"first_rank_day_number": consts.GetRankDayNumber(req.StartRankingDateStart)})
	}
	if req.StartRankingDateEnd != "" {
		where = append(where, squirrel.LtOrEq{"first_rank_day_number": consts.GetRankDayNumber(req.StartRankingDateEnd)})
	}
	if req.Word != "" {
		where = append(where, squirrel.Like{"chinese": "%" + req.Word + "%"})
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			where = append(where, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			where = append(where, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	return where
}

func cardWMediaWhere(req *types.CardsListRequest) squirrel.And {
	where := squirrel.And{squirrel.Eq{"source_type": consts.WMediaSourceNative}}
	appendWMediaOwnedFilters(&where, req.MediaOwned)
	if req.View == "cardsmediamowned" && req.MediaOwned == 0 {
		where = append(where, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	}
	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseDate(req.MediaBirthTimeStart); ok {
			where = append(where, squirrel.GtOrEq{"birth_time": ts})
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseDate(req.MediaBirthTimeEnd); ok {
			where = append(where, squirrel.LtOrEq{"birth_time": ts})
		}
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			where = append(where, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			where = append(where, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	appendMediaDirFilters(&where, "", req)
	return where
}

func cardGScStatWhere(req *types.CardsListRequest) squirrel.And {
	where := squirrel.And{}
	if req.ScTimesMin != nil {
		where = append(where, squirrel.GtOrEq{"gss.sc_times": *req.ScTimesMin})
	}
	if req.ScTimesMax != nil {
		where = append(where, squirrel.LtOrEq{"gss.sc_times": *req.ScTimesMax})
	}
	if req.ComeTimesMin != nil {
		where = append(where, squirrel.GtOrEq{"gss.come_times": *req.ComeTimesMin})
	}
	if req.ComeTimesMax != nil {
		where = append(where, squirrel.LtOrEq{"gss.come_times": *req.ComeTimesMax})
	}
	if req.LastScTimeMin != "" {
		if ts, ok := parseDate(req.LastScTimeMin); ok {
			where = append(where, squirrel.GtOrEq{"gss.last_sc_time": ts})
		}
	}
	if req.LastScTimeMax != "" {
		if ts, ok := parseDate(req.LastScTimeMax); ok {
			where = append(where, squirrel.LtOrEq{"gss.last_sc_time": ts})
		}
	}
	return where
}

func cardAlbumWhere(albumIDs []int64, req *types.CardsListRequest) squirrel.And {
	where := squirrel.And{squirrel.Eq{"album_id": albumIDs}}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			where = append(where, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			where = append(where, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	return where
}

func (r *MovieReadRepo) resolveCardAlbumIDs(ctx context.Context, req *types.CardsListRequest) ([]int64, error) {
	names := parseList(req.AlbumName)
	if req.View == "cardsneeddownload" || req.NeedDownload == consts.MovieNeedDownloadOK {
		names = []string{consts.MovieNeedDownloadAlbumName}
	}
	if len(names) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(names))
	seen := make(map[int64]struct{}, len(names))
	for _, name := range names {
		var id int64
		err := r.conn.QueryRowCtx(ctx, &id, "SELECT id FROM c_movie_album WHERE name = ? LIMIT 1", name)
		if isCardNotFound(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *MovieReadRepo) pickCardIDsByCastName(ctx context.Context, name string) (map[string]struct{}, error) {
	ids, err := r.queryCardSet(ctx, `
SELECT rc.movie_jav_id
FROM amr_movie_cast rc
JOIN am_cast c ON c.id = rc.cast_id
WHERE c.name = ?`, name)
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pickCardIDsByPersonID(ctx context.Context, personID int64) (map[string]struct{}, error) {
	ids, err := r.queryCardSet(ctx, `
SELECT DISTINCT rc.movie_jav_id
FROM amr_movie_cast rc
JOIN am_cast c ON c.id = rc.cast_id
WHERE c.person_id = ?`, personID)
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) pickCardIDsByGenreName(ctx context.Context, name string) (map[string]struct{}, error) {
	ids, err := r.queryCardSet(ctx, `
SELECT rg.movie_jav_id
FROM amr_movie_genre rg
JOIN am_genre g ON g.id = rg.genre_id
WHERE g.name = ?`, name)
	return cardSliceToSet(ids), err
}

func (r *MovieReadRepo) queryCardSet(ctx context.Context, query string, args ...any) ([]string, error) {
	var ids []string
	if err := r.conn.QueryRowsCtx(ctx, &ids, query, args...); err != nil {
		if isCardNotFound(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return ids, nil
}

func queryOne(ctx context.Context, r *MovieReadRepo, dest any, builder squirrel.SelectBuilder) error {
	sqlText, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	return r.conn.QueryRowCtx(ctx, dest, sqlText, args...)
}

func queryStringList(ctx context.Context, r *MovieReadRepo, builder squirrel.SelectBuilder) ([]string, error) {
	sqlText, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.conn.QueryRowsCtx(ctx, &ids, sqlText, args...); err != nil {
		if isCardNotFound(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return ids, nil
}

func needCardAlbum(req *types.CardsListRequest) bool {
	return strings.TrimSpace(req.AlbumName) != "" || req.View == "cardsneeddownload" || req.NeedDownload == consts.MovieNeedDownloadOK
}

func needCardAMovie(req *types.CardsListRequest) bool {
	if req.View == "" || req.View == "cards" || req.View == "cardstoday" {
		return true
	}
	if req.NeedDownload == consts.MovieNeedDownloadNone || req.MediaOwned == consts.OwnedNotOwned {
		return true
	}
	if req.DirectorName != "" || req.PrefixName != "" || req.MakerName != "" || req.LabelName != "" || req.LabelJavID != "" {
		return true
	}
	if req.CastAgeMin != nil || req.CastAgeMax != nil || req.ScoreMin != nil || req.ScoreMax != nil || req.ViewWatchedMin != nil || req.ViewWatchedMax != nil {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByViewerWatched, consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc, consts.OrderByDetailUpdateTime, consts.OrderByReleasingDate:
		return true
	default:
		return false
	}
}

func needCardMinfo(req *types.CardsListRequest) bool {
	if req.View == "cardshasrank" || req.StartRankingDateStart != "" || req.StartRankingDateEnd != "" || req.DaysInRankMin != nil || req.Word != "" {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank:
		return true
	default:
		return false
	}
}

func needCardWMedia(req *types.CardsListRequest) bool {
	if req.View == "cardsmediamowned" {
		return true
	}
	if req.MediaOwned > consts.OwnedAll && req.MediaOwned != consts.OwnedNotOwned {
		return true
	}
	if req.MediaBirthTimeStart != "" || req.MediaBirthTimeEnd != "" || req.MediaDir1 != "" || req.MediaDir2 != "" || req.MediaDir3 != "" || req.MediaDir4 != "" {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		return !needCardGScSignals(req)
	default:
		return false
	}
}

func needCardGScStat(req *types.CardsListRequest) bool {
	return needCardGScSignals(req) || req.OrderBy == consts.OrderByScTimes || req.OrderBy == consts.OrderByComeTimes || req.OrderBy == consts.OrderByLastScTime
}

func needCardGScSignals(req *types.CardsListRequest) bool {
	return req.LastScTimeMin != "" || req.LastScTimeMax != "" || req.ScTimesMin != nil || req.ScTimesMax != nil || req.ComeTimesMin != nil || req.ComeTimesMax != nil
}

func cardGScStatOrderClause(orderBy, order string) string {
	dir := cardOrderDirection(order)
	switch orderBy {
	case consts.OrderByComeTimes:
		return "COALESCE(gss.come_times, 0) " + dir + ", COALESCE(gss.last_sc_time, 0) " + dir + ", wm.movie_jav_id DESC"
	case consts.OrderByLastScTime:
		return "COALESCE(gss.last_sc_time, 0) " + dir + ", wm.movie_jav_id DESC"
	default:
		return "COALESCE(gss.sc_times, 0) " + dir + ", COALESCE(gss.last_sc_time, 0) " + dir + ", wm.movie_jav_id DESC"
	}
}

func legacyCardPage(page, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 18
	}
	return page, pageSize
}

func cardSliceToSet(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func cardSetToSlice(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func intersectCardSets(sets ...map[string]struct{}) map[string]struct{} {
	var out map[string]struct{}
	for _, set := range sets {
		if set == nil {
			continue
		}
		if out == nil {
			out = make(map[string]struct{}, len(set))
			for id := range set {
				out[id] = struct{}{}
			}
			continue
		}
		for id := range out {
			if _, ok := set[id]; !ok {
				delete(out, id)
			}
		}
	}
	if out == nil {
		return map[string]struct{}{}
	}
	return out
}

func intersectTwoCardSets(base, next map[string]struct{}) map[string]struct{} {
	if base == nil {
		return next
	}
	for id := range base {
		if _, ok := next[id]; !ok {
			delete(base, id)
		}
	}
	return base
}

func isCardNotFound(err error) bool {
	return errors.Is(err, sqlc.ErrNotFound) || errors.Is(err, ErrNotFound)
}
