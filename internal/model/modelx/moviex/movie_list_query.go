package moviex

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

/* ----------------------------- Repo ----------------------------- */

type MovieListRepoSqlx struct {
	am AMovieModel
	mi BmMinfoModel
	vf VFilmModel
	wm WMediaModel

	lb AmLabelModel
	mk AmMakerModel
	dr AmDirectorModel
	px AmPrefixModel

	cs AmCastModel
	gr AmGenreModel
	rc AmrMovieCastModel
	rg AmrMovieGenreModel

	vd VDirectoryModel
}

func NewMovieListRepoSqlx(
	am AMovieModel,
	mi BmMinfoModel,
	vf VFilmModel,
	wm WMediaModel,
	lb AmLabelModel,
	mk AmMakerModel,
	dr AmDirectorModel,
	px AmPrefixModel,
	cs AmCastModel,
	gr AmGenreModel,
	rc AmrMovieCastModel,
	rg AmrMovieGenreModel,
	vd VDirectoryModel,
) *MovieListRepoSqlx {
	return &MovieListRepoSqlx{
		am: am, mi: mi, vf: vf, wm: wm,
		lb: lb, mk: mk, dr: dr, px: px,
		cs: cs, gr: gr, rc: rc, rg: rg,
		vd: vd,
	}
}

/* ----------------------------- 顶层入口 ----------------------------- */

func (r *MovieListRepoSqlx) ListFull(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	// 命中判定
	needA := needAMovie(req)
	needM := needMinfo(req)
	needV := needVFilm(req)
	needW := needWMedia(req)
	needF := needV || needW

	// 一旦出现多对多（Cast/Genre），严禁任何 onlyX 快速路径 —— 必须走交集
	hasM2M := req.CastNames != "" || req.PersonIds != "" || req.GenreNames != ""

	onlyA := !hasM2M && needA && !needM && !needF
	onlyM := !hasM2M && needM && !needA && !needF
	onlyF := !hasM2M && needF && !needA && !needM

	// 1) 单表直取
	if onlyA {
		return r.listFromAMovieOnly(ctx, req)
	}
	if onlyM {
		return r.listFromMinfoOnly(ctx, req)
	}
	if onlyF {
		if needV && !needW {
			return r.listFromVFilmOnly(ctx, req)
		}
		return r.listFromMediaOnly(ctx, req)
	}

	// 2) 多表各自取 ID → 交集
	setA, err := r.pickFromAMovie(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setM, err := r.pickFromMinfo(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setF, err := r.pickFromMedia(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	finalIDs := intersectNonEmpty(setA, setM, setF)
	total := int64(len(finalIDs))
	if total == 0 {
		return nil, 0, nil
	}

	// 3) 在排序所属表分页
	page, size := normalizePage(req.Page, req.PageSize)
	offset := (page - 1) * size

	var ordered []string
	switch req.OrderBy {
	case consts.OrderByBirthTime:
		ordered, err = r.pageOnVFilm(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByMediaBirthTime:
		ordered, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByReleasingDate:
		if needW && !needV {
			ordered, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ordered, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ordered, err = r.pageOnVFilm(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank:
		ordered, err = r.pageOnMinfo(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	default:
		ordered, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	}
	if err != nil {
		return nil, 0, err
	}

	out := make([]*types.Movie, 0, len(ordered))
	for _, id := range ordered {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, total, nil
}

/* ---------------- 单表直取：COUNT + ORDER/LIMIT ---------------- */

func (r *MovieListRepoSqlx) listFromAMovieOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	w := amovieBaseFilters(req)

	// onlyA 路径下需要把四个单值外键也下推到 a_movie
	if req.LabelName != "" {
		if id, ok := r.idOfLabel(ctx, req.LabelName); ok {
			w = append(w, squirrel.Eq{"label_id": id})
		} else {
			return nil, 0, nil
		}
	}
	if req.MakerName != "" {
		if id, ok := r.idOfMaker(ctx, req.MakerName); ok {
			w = append(w, squirrel.Eq{"maker_id": id})
		} else {
			return nil, 0, nil
		}
	}
	if req.DirectorName != "" {
		if id, ok := r.idOfDirector(ctx, req.DirectorName); ok {
			w = append(w, squirrel.Eq{"director_id": id})
		} else {
			return nil, 0, nil
		}
	}
	if req.PrefixName != "" {
		if id, ok := r.idOfPrefix(ctx, req.PrefixName); ok {
			w = append(w, squirrel.Eq{"prefix_id": id})
		} else {
			return nil, 0, nil
		}
	}

	w = append(w, r.amovieOwnershipGuards(req)...)

	order, guards := amovieOrdering(req.OrderBy, req.Order)
	w = append(w, guards...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.am.TableName()).Where(w).ToSql()
	var total int64
	if err := r.am.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	page, size := normalizePage(req.Page, req.PageSize)
	sqlStr, args, _ := squirrel.
		Select("jav_id").
		From(r.am.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size)).
		ToSql()

	var ids []string
	if err := r.am.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, 0, err
	}
	out := make([]*types.Movie, 0, len(ids))
	for _, id := range ids {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, total, nil
}

func (r *MovieListRepoSqlx) listFromMinfoOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	w := minfoBaseFilters(req)
	order, guards := minfoOrdering(req.OrderBy, req.Order)
	w = append(w, guards...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.mi.TableName()).Where(w).ToSql()
	var total int64
	if err := r.mi.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	page, size := normalizePage(req.Page, req.PageSize)
	sqlStr, args, _ := squirrel.
		Select("jav_id").
		From(r.mi.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size)).
		ToSql()

	var ids []string
	if err := r.mi.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, 0, err
	}
	out := make([]*types.Movie, 0, len(ids))
	for _, id := range ids {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, total, nil
}

func (r *MovieListRepoSqlx) listFromVFilmOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	w := vfilmBaseFilters(ctx, r, req)
	order, guards := vfilmOrdering(req.OrderBy, req.Order)
	w = append(w, guards...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.vf.TableName()).Where(w).ToSql()
	var total int64
	if err := r.vf.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	page, size := normalizePage(req.Page, req.PageSize)
	sqlStr, args, _ := squirrel.
		Select("movie_jav_id").
		From(r.vf.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size)).
		ToSql()

	var ids []string
	if err := r.vf.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, 0, err
	}
	out := make([]*types.Movie, 0, len(ids))
	for _, id := range ids {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, total, nil
}

func (r *MovieListRepoSqlx) listFromMediaOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	set, err := r.pickFromMedia(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	if set == nil || len(set) == 0 {
		return nil, 0, nil
	}

	finalIDs := setToSlice(set)
	total := int64(len(finalIDs))

	page, size := normalizePage(req.Page, req.PageSize)
	offset := (page - 1) * size

	needV := needVFilm(req)
	needW := needWMedia(req)

	var ids []string
	switch req.OrderBy {
	case consts.OrderByBirthTime:
		ids, err = r.pageOnVFilm(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByMediaBirthTime:
		ids, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByReleasingDate:
		if needW && !needV {
			ids, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ids, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ids, err = r.pageOnVFilm(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	default:
		ids, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	}
	if err != nil {
		return nil, 0, err
	}

	out := make([]*types.Movie, 0, len(ids))
	for _, id := range ids {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, total, nil
}

/* ---------------- 命中判断 ---------------- */

func needAMovie(req *types.ListMovieFullRequest) bool {
	if req.Owned == consts.OwnedNotOwned || req.MediaOwned == consts.OwnedNotOwned {
		return true
	}
	// 只在 a_movie 特有字段/排序/单值外键时命中
	if req.DirectorName != "" || req.PrefixName != "" || req.MakerName != "" || req.LabelName != "" {
		return true
	}
	if req.CastAgeMin > 0 || req.CastAgeMax > 0 || req.ScoreMin > 0 || req.ScoreMax > 0 || req.ViewWatchedMin > 0 || req.ViewWatchedMax > 0 {
		return true
	}

	switch req.OrderBy {
	case consts.OrderByViewerWatched, consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc, consts.OrderByDetailUpdateTime:
		return true
	}

	// 仅发行日过滤/排序：若其他表会命中则不强制
	releaseOnlyFilter := req.ReleasingDateStart != "" || req.ReleasingDateEnd != ""
	releaseOnlyOrder := req.OrderBy == consts.OrderByReleasingDate
	if releaseOnlyFilter || releaseOnlyOrder {
		if needMinfo(req) || needVFilm(req) || needWMedia(req) {
			return false
		}
		return true
	}
	return false
}

func needMinfo(req *types.ListMovieFullRequest) bool {
	if req.StartRankingDateStart != "" || req.StartRankingDateEnd != "" || req.DaysInRankMin > 0 || req.NeedDownload > 0 || req.Word != "" {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank:
		return true
	}
	return false
}

func needVFilm(req *types.ListMovieFullRequest) bool {
	if hasVFilmFilters(req) {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByBirthTime:
		return true
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	}
	return false
}

func hasVFilmFilters(req *types.ListMovieFullRequest) bool {
	return (req.Owned > consts.MovieAll && req.Owned != consts.OwnedNotOwned) ||
		req.ComeTimesMin > 0 || req.ComeTimesMax != nil ||
		req.LastScTimeMin != "" || req.LastScTimeMax != "" ||
		req.ScTimesMin > 0 || req.ScTimesMax != nil ||
		req.FilmBirthTimeStart != "" || req.FilmBirthTimeEnd != "" ||
		req.Dir1 != "" || req.Dir2 != "" || req.Dir3 != "" || req.Dir4 != ""
}

func needWMedia(req *types.ListMovieFullRequest) bool {
	if req.OrderBy == consts.OrderByMediaBirthTime {
		return true
	}
	if (req.MediaOwned > consts.MovieAll && req.MediaOwned != consts.OwnedNotOwned) ||
		req.MediaBirthTimeStart != "" || req.MediaBirthTimeEnd != "" ||
		req.MediaDir1 != "" || req.MediaDir2 != "" || req.MediaDir3 != "" || req.MediaDir4 != "" {
		return true
	}
	return false
}

func needMedia(req *types.ListMovieFullRequest) bool {
	return needVFilm(req) || needWMedia(req)
}

/* ---------------- 共享的基础过滤器 ---------------- */

func amovieBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
	// 发行日
	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			w = append(w, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			w = append(w, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	// 年龄（排除 0）
	if req.CastAgeMin > 0 {
		w = append(w, squirrel.And{
			squirrel.GtOrEq{"cast_average_age": int64(req.CastAgeMin*10.0 + 0.5)},
			squirrel.NotEq{"cast_average_age": 0},
		})
	}
	if req.CastAgeMax > 0 {
		w = append(w, squirrel.And{
			squirrel.LtOrEq{"cast_average_age": int64(req.CastAgeMax*10.0 + 0.5)},
			squirrel.NotEq{"cast_average_age": 0},
		})
	}

	if req.ScoreMin > 0 {
		w = append(w, squirrel.And{
			squirrel.GtOrEq{"score": int64(req.ScoreMin*10.0 + 0.5)},
		})
	}
	if req.ScoreMax > 0 {
		w = append(w, squirrel.And{
			squirrel.LtOrEq{"score": int64(req.ScoreMax*10.0 + 0.5)},
		})
	}

	if req.ViewWatchedMin > 0 {
		w = append(w, squirrel.And{
			squirrel.GtOrEq{"viewers_number_watched": req.ViewWatchedMin},
		})
	}
	if req.ViewWatchedMax > 0 {
		w = append(w, squirrel.And{
			squirrel.LtOrEq{"viewers_number_watched": req.ViewWatchedMax},
		})
	}

	return w
}

func minfoBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
	if req.StartRankingDateStart != "" {
		ts := consts.GetRankDayNumber(req.StartRankingDateStart)
		w = append(w, squirrel.GtOrEq{"first_rank_day_number": ts})
		w = append(w, squirrel.NotEq{"days_in_rank": 0})
	}
	if req.StartRankingDateEnd != "" {
		ts := consts.GetRankDayNumber(req.StartRankingDateEnd)
		w = append(w, squirrel.LtOrEq{"first_rank_day_number": ts})
		w = append(w, squirrel.NotEq{"days_in_rank": 0})
	}

	if req.DaysInRankMin > 0 {
		w = append(w, squirrel.GtOrEq{"days_in_rank": req.DaysInRankMin})
	}
	if req.NeedDownload > 0 {
		w = append(w, squirrel.Eq{"need_download": req.NeedDownload})
	}
	if req.Word != "" {
		w = append(w, squirrel.Like{"chinese": "%" + req.Word + "%"})
	}
	// 发行日（冗余列）
	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			w = append(w, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			w = append(w, squirrel.LtOrEq{"releasing_date": ts})
		}
	}
	return w
}

func vfilmBaseFilters(ctx context.Context, r *MovieListRepoSqlx, req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}

	// Owned 映射
	switch req.Owned {
	case consts.OwnedHasSubNotRemoved:
		w = append(w, squirrel.Eq{"has_sub": consts.FilmHasSub}, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedNoSubNotRemoved:
		w = append(w, squirrel.Eq{"has_sub": consts.FilmNoSub}, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedAllNotRemoved:
		w = append(w, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedRemoved:
		w = append(w, squirrel.Eq{"is_removed": consts.FilmIsRemoved})
	case consts.OwnedAll:
		w = append(w, squirrel.Expr("1=1"))
	case consts.OwnedNotOwned:
		return squirrel.And{}
	case consts.MovieAll:
		return squirrel.And{}
	case 0:
	}

	if req.ComeTimesMin > 0 {
		w = append(w, squirrel.GtOrEq{"come_times": req.ComeTimesMin})
	}

	if req.LastScTimeMin != "" {
		if ts, ok := parseYMD(req.LastScTimeMin); ok {
			w = append(w, squirrel.GtOrEq{"last_sc_time": ts})
		}
	}
	if req.LastScTimeMax != "" {
		if ts, ok := parseYMD(req.LastScTimeMax); ok {
			w = append(w, squirrel.LtOrEq{"last_sc_time": ts}, squirrel.NotEq{"last_sc_time": 0})
		}
	}
	if req.ScTimesMin > 0 {
		w = append(w, squirrel.GtOrEq{"sc_times": req.ScTimesMin})
	}
	if req.ScTimesMax != nil {
		w = append(w, squirrel.LtOrEq{"sc_times": *req.ScTimesMax})
	}
	if req.ComeTimesMin > 0 {
		w = append(w, squirrel.GtOrEq{"come_times": req.ComeTimesMin})
	}
	if req.ComeTimesMax != nil {
		w = append(w, squirrel.LtOrEq{"come_times": *req.ComeTimesMax})
	}
	if req.FilmBirthTimeStart != "" {
		if ts, ok := parseYMD(req.FilmBirthTimeStart); ok {
			w = append(w, squirrel.GtOrEq{"birth_time": ts})
		}
	}
	if req.FilmBirthTimeEnd != "" {
		if ts, ok := parseYMD(req.FilmBirthTimeEnd); ok {
			w = append(w, squirrel.LtOrEq{"birth_time": ts})
		}
	}
	//发行日（v_film 冗余列）
	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			w = append(w, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			w = append(w, squirrel.LtOrEq{"releasing_date": ts})
		}
	}

	// ★★★ 目录过滤：Dir 名称 → v_directory.id → 用 dir*_id 精确过滤
	if req.Dir1 != "" {
		if id, ok := r.findDirIdByName(ctx, req.Dir1); ok {
			w = append(w, squirrel.Eq{"dir1_id": id})
		} else {
			return squirrel.And{squirrel.Expr("1=0")}
		}
	}
	if req.Dir2 != "" {
		if id, ok := r.findDirIdByName(ctx, req.Dir2); ok {
			w = append(w, squirrel.Eq{"dir2_id": id})
		} else {
			return squirrel.And{squirrel.Expr("1=0")}
		}
	}
	if req.Dir3 != "" {
		if id, ok := r.findDirIdByName(ctx, req.Dir3); ok {
			w = append(w, squirrel.Eq{"dir3_id": id})
		} else {
			return squirrel.And{squirrel.Expr("1=0")}
		}
	}
	if req.Dir4 != "" {
		if id, ok := r.findDirIdByName(ctx, req.Dir4); ok {
			w = append(w, squirrel.Eq{"dir4_id": id})
		} else {
			return squirrel.And{squirrel.Expr("1=0")}
		}
	}

	return w
}

func wmediaBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}

	switch req.MediaOwned {
	case consts.OwnedHasSubNotRemoved:
		w = append(w, squirrel.Eq{"has_sub": consts.FilmHasSub}, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedNoSubNotRemoved:
		w = append(w, squirrel.Eq{"has_sub": consts.FilmNoSub}, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedAllNotRemoved:
		w = append(w, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedRemoved:
		w = append(w, squirrel.Eq{"is_removed": consts.FilmIsRemoved})
	case consts.OwnedAll:
		w = append(w, squirrel.Expr("1=1"))
	case consts.OwnedNotOwned:
		return squirrel.And{}
	case consts.MovieAll:
		return squirrel.And{}
	case 0:
	}

	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseYMD(req.MediaBirthTimeStart); ok {
			w = append(w, squirrel.GtOrEq{"birth_time": ts})
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseYMD(req.MediaBirthTimeEnd); ok {
			w = append(w, squirrel.LtOrEq{"birth_time": ts})
		}
	}

	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			w = append(w, squirrel.GtOrEq{"releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			w = append(w, squirrel.LtOrEq{"releasing_date": ts})
		}
	}

	if req.MediaDir1 != "" {
		w = append(w, squirrel.Expr("CONCAT('/', full_dir, '/') LIKE ?", "%/"+req.MediaDir1+"/%"))
	}
	if req.MediaDir2 != "" {
		w = append(w, squirrel.Expr("CONCAT('/', full_dir, '/') LIKE ?", "%/"+req.MediaDir2+"/%"))
	}
	if req.MediaDir3 != "" {
		w = append(w, squirrel.Expr("CONCAT('/', full_dir, '/') LIKE ?", "%/"+req.MediaDir3+"/%"))
	}
	if req.MediaDir4 != "" {
		w = append(w, squirrel.Expr("CONCAT('/', full_dir, '/') LIKE ?", "%/"+req.MediaDir4+"/%"))
	}
	return w
}

/* ------------------- A. 各表筛选（取 ID 集） ------------------- */

func (r *MovieListRepoSqlx) pickFromAMovie(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := amovieBaseFilters(req)

	// 单值外键（名称→ID）
	if req.LabelName != "" {
		if id, ok := r.idOfLabel(ctx, req.LabelName); ok {
			w = append(w, squirrel.Eq{"label_id": id})
		} else {
			return map[string]struct{}{}, nil
		}
	}
	if req.MakerName != "" {
		if id, ok := r.idOfMaker(ctx, req.MakerName); ok {
			w = append(w, squirrel.Eq{"maker_id": id})
		} else {
			return map[string]struct{}{}, nil
		}
	}
	if req.DirectorName != "" {
		if id, ok := r.idOfDirector(ctx, req.DirectorName); ok {
			w = append(w, squirrel.Eq{"director_id": id})
		} else {
			return map[string]struct{}{}, nil
		}
	}
	if req.PrefixName != "" {
		if id, ok := r.idOfPrefix(ctx, req.PrefixName); ok {
			w = append(w, squirrel.Eq{"prefix_id": id})
		} else {
			return map[string]struct{}{}, nil
		}
	}

	// 排序护栏
	w = append(w, amovieOrderGuards(req.OrderBy)...)
	w = append(w, r.amovieOwnershipGuards(req)...)

	// 基集合：w 为空意味着“不限制 a_movie”（交由其它表或 M2M 决定）
	var base map[string]struct{} = nil
	if len(w) > 0 {
		sqlStr, args, err := squirrel.Select("jav_id").From(r.am.TableName()).Where(w).ToSql()
		if err != nil {
			return nil, err
		}
		var ids []string
		if err := r.am.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				base = map[string]struct{}{}
			} else {
				return nil, err
			}
		} else {
			base = sliceToSet(ids)
		}
	}

	// 多对多：演员
	if req.CastNames != "" {
		for _, name := range splitNames(req.CastNames) {
			sCast, err := r.pickFromCast(ctx, name)
			if err != nil {
				return nil, err
			}
			base = intersectTwo(base, sCast)
			if len(base) == 0 {
				return base, nil
			}
		}
	}
	if req.PersonIds != "" {
		for _, personID := range splitInt64Tokens(req.PersonIds) {
			sPerson, err := r.pickFromPerson(ctx, personID)
			if err != nil {
				return nil, err
			}
			base = intersectTwo(base, sPerson)
			if len(base) == 0 {
				return base, nil
			}
		}
	}
	// 多对多：类型
	if req.GenreNames != "" {
		for _, name := range splitNames(req.GenreNames) {
			sGenre, err := r.pickFromGenre(ctx, name)
			if err != nil {
				return nil, err
			}
			base = intersectTwo(base, sGenre)
			if len(base) == 0 {
				return base, nil
			}
		}
	}
	return base, nil
}

func (r *MovieListRepoSqlx) pickFromMinfo(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := minfoBaseFilters(req)
	w = append(w, minfoOrderGuards(req.OrderBy)...)

	// minfo 完全无条件且排序不依赖 minfo → “不限制”
	if len(w) == 0 && req.OrderBy != consts.OrderByRankDate && req.OrderBy != consts.OrderByHighestRank && req.OrderBy != consts.OrderByDaysInRank {
		return nil, nil
	}

	sqlStr, args, err := squirrel.Select("jav_id").From(r.mi.TableName()).Where(w).ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.mi.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromVFilm(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := vfilmBaseFilters(ctx, r, req)
	w = append(w, vfilmOrderGuards(req.OrderBy)...)

	if len(w) == 0 && !orderBelongsToVFilm(req.OrderBy) {
		return nil, nil
	}

	sqlStr, args, err := squirrel.Select("movie_jav_id").From(r.vf.TableName()).Where(w).ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.vf.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromWMedia(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := wmediaBaseFilters(req)
	w = append(w, wmediaOrderGuards(req.OrderBy)...)
	if len(w) == 0 && !orderBelongsToWMedia(req.OrderBy) {
		return nil, nil
	}

	sqlStr, args, err := squirrel.Select("movie_jav_id").From(r.wm.TableName()).Where(w).ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.wm.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromMedia(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	needV := needVFilm(req)
	needW := needWMedia(req)

	var (
		setV map[string]struct{}
		setW map[string]struct{}
		err  error
	)

	if needV {
		setV, err = r.pickFromVFilm(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	if needW {
		setW, err = r.pickFromWMedia(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	switch {
	case needV && needW:
		return intersectTwo(setV, setW), nil
	case needV:
		return setV, nil
	case needW:
		return setW, nil
	default:
		return nil, nil
	}
}

/* ---------------- 多对多关系查询（单名） ---------------- */

func (r *MovieListRepoSqlx) pickFromCast(ctx context.Context, castName string) (map[string]struct{}, error) {
	row, err := r.cs.FindOneByName(ctx, castName)
	if err != nil || row == nil {
		return map[string]struct{}{}, nil
	}
	sqlStr, args, _ := squirrel.
		Select("movie_jav_id").
		From(r.rc.TableName()).
		Where(squirrel.Eq{"cast_id": row.Id}).
		ToSql()
	var ids []string
	if err := r.rc.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromPerson(ctx context.Context, personID int64) (map[string]struct{}, error) {
	if personID <= 0 {
		return map[string]struct{}{}, nil
	}
	sqlStr, args, _ := squirrel.
		Select("DISTINCT amr.movie_jav_id").
		From(r.rc.TableName() + " amr").
		Join("`am_cast` ac ON ac.id = amr.cast_id").
		Where(squirrel.Eq{"ac.person_id": personID}).
		ToSql()
	var ids []string
	if err := r.rc.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromGenre(ctx context.Context, genreName string) (map[string]struct{}, error) {
	row, err := r.gr.FindOneByName(ctx, genreName)
	if err != nil || row == nil {
		return map[string]struct{}{}, nil
	}
	sqlStr, args, _ := squirrel.
		Select("movie_jav_id").
		From(r.rg.TableName()).
		Where(squirrel.Eq{"genre_id": row.Id}).
		ToSql()
	var ids []string
	if err := r.rg.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func splitInt64Tokens(raw string) []int64 {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '|' || r == ';' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]int64, 0, len(parts))
	seen := make(map[int64]struct{}, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		v, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || v <= 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

/* ------------------- B. 分表排序分页（WHERE ... IN） ------------------- */

func (r *MovieListRepoSqlx) pageOnAMovie(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order, guards := amovieOrdering(od, sortOrder)
	w := squirrel.And{squirrel.Eq{"jav_id": finalIDs}}
	w = append(w, guards...)

	sb := squirrel.Select("jav_id").
		From(r.am.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64(offset)).Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.am.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnMinfo(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order, guards := minfoOrdering(od, sortOrder)
	w := squirrel.And{squirrel.Eq{"jav_id": finalIDs}}
	w = append(w, guards...)

	sb := squirrel.Select("jav_id").
		From(r.mi.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64(offset)).Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.mi.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnVFilm(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order, guards := vfilmOrdering(od, sortOrder)
	w := squirrel.And{squirrel.Eq{"movie_jav_id": finalIDs}}
	w = append(w, guards...)

	sb := squirrel.Select("movie_jav_id").
		From(r.vf.TableName()).
		Where(w).
		OrderBy(order).
		Offset(uint64(offset)).Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.vf.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnMedia(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}

	order := mediaOrdering(od, sortOrder)
	sb := squirrel.Select("jav_id").
		From(r.am.TableName()).
		LeftJoin(r.vf.TableName() + " vf ON vf.movie_jav_id = jav_id").
		LeftJoin(r.wm.TableName() + " wm ON wm.movie_jav_id = jav_id").
		Where(squirrel.Eq{"jav_id": finalIDs}).
		OrderBy(order).
		Offset(uint64(offset)).Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.am.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnWMedia(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}

	order := wmediaOrdering(od, sortOrder)
	sb := squirrel.Select("movie_jav_id").
		From(r.wm.TableName()).
		Where(squirrel.Eq{"movie_jav_id": finalIDs}).
		OrderBy(order).
		Offset(uint64(offset)).Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.wm.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

/* ------------------- C. 名称→ID 查找（单值） ------------------- */

func (r *MovieListRepoSqlx) idOfLabel(ctx context.Context, name string) (int64, bool) {
	row, err := r.lb.FindOneByName(ctx, name)
	if err != nil || row == nil {
		return 0, false
	}
	return row.Id, true
}
func (r *MovieListRepoSqlx) idOfMaker(ctx context.Context, name string) (int64, bool) {
	row, err := r.mk.FindOneByName(ctx, name)
	if err != nil || row == nil {
		return 0, false
	}
	return row.Id, true
}
func (r *MovieListRepoSqlx) idOfDirector(ctx context.Context, name string) (int64, bool) {
	row, err := r.dr.FindOneByName(ctx, name)
	if err != nil || row == nil {
		return 0, false
	}
	return row.Id, true
}
func (r *MovieListRepoSqlx) idOfPrefix(ctx context.Context, name string) (int64, bool) {
	row, err := r.px.FindOneByName(ctx, name)
	if err != nil || row == nil {
		return 0, false
	}
	return row.Id, true
}

// 目录：按名称查 id（由你的 custom VDirectoryModel 提供）
func (r *MovieListRepoSqlx) findDirIdByName(ctx context.Context, name string) (int64, bool) {
	row, err := r.vd.FindOneByName(ctx, name)
	if err != nil || row == nil {
		return 0, false
	}
	return row.Id, true
}

/* ------------------- D. 排序护栏 & 排序字句 ------------------- */

func amovieOrderGuards(orderBy string) squirrel.And {
	w := squirrel.And{}
	switch orderBy {
	case consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc:
		w = append(w, squirrel.NotEq{"cast_average_age": 0})
	}
	return w
}

func (r *MovieListRepoSqlx) amovieOwnershipGuards(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
	if req.Owned == consts.OwnedNotOwned {
		w = append(w, squirrel.Expr("NOT EXISTS (SELECT 1 FROM "+r.vf.TableName()+" vf_abs WHERE vf_abs.movie_jav_id = jav_id)"))
	}
	if req.MediaOwned == consts.OwnedNotOwned {
		w = append(w, squirrel.Expr("NOT EXISTS (SELECT 1 FROM "+r.wm.TableName()+" wm_abs WHERE wm_abs.movie_jav_id = jav_id)"))
	}
	return w
}

func minfoOrderGuards(orderBy string) squirrel.And {
	w := squirrel.And{}
	switch orderBy {
	case consts.OrderByHighestRank:
		w = append(w, squirrel.Expr("highest_rank <> 0"))
	case consts.OrderByRankDate:
		w = append(w, squirrel.Expr("days_in_rank <> 0"))
	case consts.OrderByDaysInRank:
		w = append(w, squirrel.Expr("days_in_rank <> 0"))
	}
	return w
}

func vfilmOrderGuards(orderBy string) squirrel.And {
	w := squirrel.And{}
	switch orderBy {
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		// 固定护栏：这三种排序都要求 sc_times 非 0
		w = append(w, squirrel.Expr("sc_times <> 0"))
	}
	return w
}

func wmediaOrderGuards(orderBy string) squirrel.And {
	w := squirrel.And{}
	switch orderBy {
	case consts.OrderByMediaBirthTime:
		return w
	}
	return w
}

func amovieOrdering(orderBy string, sortOrder string) (order string, guards squirrel.And) {
	order = "releasing_date DESC,name DESC"
	switch orderBy {
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC,name DESC"
	case consts.OrderByDetailUpdateTime:
		order = "detail_update_time DESC,name DESC"
	case consts.OrderByViewerWatched:
		order = "viewers_number_watched DESC"
	case consts.OrderByCastAgeAsc:
		order = "cast_average_age ASC"
	case consts.OrderByCastAgeDesc:
		order = "cast_average_age DESC"
	}
	order = applyMovieListOrder(order, sortOrder)
	guards = amovieOrderGuards(orderBy)
	return
}

func minfoOrdering(orderBy string, sortOrder string) (order string, guards squirrel.And) {
	order = "first_rank_day_number desc,name desc"
	switch orderBy {
	case consts.OrderByHighestRank:
		order = "highest_rank ASC,name ASC"
	case consts.OrderByDaysInRank:
		order = "days_in_rank DESC"
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC,name DESC"
	case consts.OrderByRankDate:
		order = "first_rank_day_number DESC,name DESC"
	}
	order = applyMovieListOrder(order, sortOrder)
	guards = minfoOrderGuards(orderBy)
	return
}

func vfilmOrdering(orderBy string, sortOrder string) (order string, guards squirrel.And) {
	order = "birth_time DESC"
	switch orderBy {
	case consts.OrderByBirthTime:
		order = "birth_time DESC,movie_name DESC"
	case consts.OrderByScTimes:
		order = "sc_times DESC,last_sc_time DESC,movie_name DESC"
	case consts.OrderByComeTimes:
		order = "come_times DESC,last_sc_time DESC,movie_name DESC"
	case consts.OrderByLastScTime:
		order = "last_sc_time DESC,movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC,movie_name DESC"
	}
	order = applyMovieListOrder(order, sortOrder)
	guards = vfilmOrderGuards(orderBy)
	return
}

func mediaOrdering(orderBy string, sortOrder string) string {
	order := "vf.birth_time DESC,name DESC"
	switch orderBy {
	case consts.OrderByBirthTime:
		order = "vf.birth_time DESC,name DESC"
	case consts.OrderByMediaBirthTime:
		order = "wm.birth_time DESC,name DESC"
	case consts.OrderByReleasingDate:
		order = "COALESCE(NULLIF(vf.releasing_date, 0), NULLIF(wm.releasing_date, 0), releasing_date, 0) DESC,name DESC"
	}
	return applyMovieListOrder(order, sortOrder)
}

func wmediaOrdering(orderBy string, sortOrder string) string {
	order := "releasing_date DESC,movie_name DESC"
	switch orderBy {
	case consts.OrderByMediaBirthTime:
		order = "birth_time DESC,movie_name DESC"
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC,movie_name DESC"
	}
	return applyMovieListOrder(order, sortOrder)
}

func applyMovieListOrder(orderClause string, sortOrder string) string {
	currentOrder := normalizeMovieListOrder(sortOrder)
	if currentOrder == "" {
		return orderClause
	}

	target := strings.ToUpper(currentOrder)
	parts := strings.Split(orderClause, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		fields := strings.Fields(item)
		if len(fields) == 0 {
			continue
		}
		last := strings.ToUpper(fields[len(fields)-1])
		if last == "ASC" || last == "DESC" {
			fields[len(fields)-1] = target
			normalized = append(normalized, strings.Join(fields, " "))
			continue
		}
		normalized = append(normalized, item+" "+target)
	}
	if len(normalized) == 0 {
		return orderClause
	}
	return strings.Join(normalized, ", ")
}

func normalizeMovieListOrder(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "asc", "desc":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func orderBelongsToVFilm(od string) bool {
	switch od {
	case consts.OrderByBirthTime, consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	default:
		return false
	}
}

func orderBelongsToWMedia(od string) bool {
	switch od {
	case consts.OrderByMediaBirthTime, consts.OrderByReleasingDate:
		return true
	default:
		return false
	}
}

/* ------------------- E. 其它辅助 ------------------- */

func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 18
	}
	return page, size
}

func splitNames(s string) []string {
	seps := []string{",", "，", "、", "/", "|"}
	for _, sp := range seps {
		s = strings.ReplaceAll(s, sp, " ")
	}
	parts := strings.Fields(s)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func sliceToSet(ss []string) map[string]struct{} {
	if len(ss) == 0 {
		return map[string]struct{}{}
	}
	m := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		if s != "" {
			m[s] = struct{}{}
		}
	}
	return m
}

func intersectTwo(a, b map[string]struct{}) map[string]struct{} {
	if a == nil || b == nil {
		if a == nil {
			return b
		}
		return a
	}
	if len(a) > len(b) {
		a, b = b, a
	}
	res := make(map[string]struct{}, len(a))
	for k := range a {
		if _, ok := b[k]; ok {
			res[k] = struct{}{}
		}
	}
	return res
}

func setToSlice(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func unionNullableSets(a, b map[string]struct{}) map[string]struct{} {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	out := make(map[string]struct{}, len(a)+len(b))
	for k := range a {
		out[k] = struct{}{}
	}
	for k := range b {
		out[k] = struct{}{}
	}
	return out
}

func intersectNonEmpty(sets ...map[string]struct{}) []string {
	nonNil := make([]map[string]struct{}, 0, len(sets))
	for _, s := range sets {
		if s != nil {
			nonNil = append(nonNil, s)
		}
	}
	if len(nonNil) == 0 {
		return []string{}
	}
	sort.Slice(nonNil, func(i, j int) bool { return len(nonNil[i]) < len(nonNil[j]) })
	base := nonNil[0]
	res := make([]string, 0, len(base))
nextID:
	for id := range base {
		for k := 1; k < len(nonNil); k++ {
			if _, hit := nonNil[k][id]; !hit {
				continue nextID
			}
		}
		res = append(res, id)
	}
	return res
}

func parseYMD(s string) (int64, bool) {
	if len(strings.TrimSpace(s)) < 8 {
		return 0, false
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}
