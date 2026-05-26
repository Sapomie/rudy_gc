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
	am  AMovieModel
	mi  BmMinfoModel
	wm  WMediaModel
	gss GScStatModel
	ca  CMovieAlbumModel
	cai CMovieAlbumItemModel

	lb AmLabelModel
	mk AmMakerModel
	dr AmDirectorModel
	px AmPrefixModel

	cs AmCastModel
	gr AmGenreModel
	rc AmrMovieCastModel
	rg AmrMovieGenreModel

	wf WFolderModel
}

func NewMovieListRepoSqlx(
	am AMovieModel,
	mi BmMinfoModel,
	wm WMediaModel,
	gss GScStatModel,
	ca CMovieAlbumModel,
	cai CMovieAlbumItemModel,
	lb AmLabelModel,
	mk AmMakerModel,
	dr AmDirectorModel,
	px AmPrefixModel,
	cs AmCastModel,
	gr AmGenreModel,
	rc AmrMovieCastModel,
	rg AmrMovieGenreModel,
	wf WFolderModel,
) *MovieListRepoSqlx {
	return &MovieListRepoSqlx{
		am: am, mi: mi, wm: wm, gss: gss, ca: ca, cai: cai,
		lb: lb, mk: mk, dr: dr, px: px,
		cs: cs, gr: gr, rc: rc, rg: rg,
		wf: wf,
	}
}

/* ----------------------------- 顶层入口 ----------------------------- */

func (r *MovieListRepoSqlx) ListFull(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	// 命中判定
	needAlbum := needAlbum(req)
	needA := needAMovie(req)
	needM := needMinfo(req)
	needW := needWMedia(req)
	needS := needGScStat(req)
	needF := needW

	// 一旦出现多对多（Cast/Genre），严禁任何 onlyX 快速路径 —— 必须走交集
	hasM2M := req.CastNames != "" || req.PersonIds != "" || req.GenreNames != ""
	albumRows, err := r.resolveAlbums(ctx, req.AlbumName)
	if err != nil {
		return nil, 0, err
	}
	if needAlbum && len(albumRows) == 0 {
		return nil, 0, nil
	}
	albumIDs := movieAlbumRowsToIDs(albumRows)

	onlyA := !needAlbum && !hasM2M && needA && !needM && !needF && !needS
	onlyM := !needAlbum && !hasM2M && needM && !needA && !needF && !needS
	onlyF := !needAlbum && !hasM2M && needF && !needA && !needM && !needS
	onlyS := !needAlbum && !hasM2M && needS && !needA && !needM && !needF
	onlyAlbumRelease := needAlbum && !hasM2M && !needA && !needM && !needF && !needS && req.OrderBy == consts.OrderByReleasingDate

	// 1) 单表直取
	if onlyAlbumRelease {
		return r.listFromAlbumOnly(ctx, req, albumIDs)
	}
	if onlyA {
		return r.listFromAMovieOnly(ctx, req)
	}
	if onlyM {
		return r.listFromMinfoOnly(ctx, req)
	}
	if onlyF {
		return r.listFromMediaOnly(ctx, req)
	}
	if onlyS {
		return r.listFromGScStatOnly(ctx, req)
	}

	// 2) 多表各自取 ID → 交集
	setA, err := r.pickFromAMovie(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setAlbum, err := r.pickFromAlbum(ctx, req, albumIDs)
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
	setS, err := r.pickFromGScStat(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	finalIDs := intersectNonEmpty(setAlbum, setA, setM, setF, setS)
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
		ordered, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByMediaBirthTime:
		if needS && !needA && !needM && !needW {
			ordered, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ordered, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByReleasingDate:
		if needAlbum {
			ordered, err = r.pageOnAlbum(ctx, req, albumIDs, sliceToSet(finalIDs), offset, size)
		} else if needS && !needA && !needM && !needW {
			ordered, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else if needF {
			ordered, err = r.pageOnMedia(ctx, req, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ordered, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ordered, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
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
	if req.LabelJavID != "" {
		if id, ok := r.idOfLabelJavID(ctx, req.LabelJavID); ok {
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

func (r *MovieListRepoSqlx) listFromAlbumOnly(ctx context.Context, req *types.ListMovieFullRequest, albumIDs []int64) ([]*types.Movie, int64, error) {
	if len(albumIDs) == 0 {
		return nil, 0, nil
	}
	if len(albumIDs) == 1 {
		return r.listFromSingleAlbumOnly(ctx, req, albumIDs[0])
	}

	setAlbum, err := r.pickFromAlbum(ctx, req, albumIDs)
	if err != nil {
		return nil, 0, err
	}
	if len(setAlbum) == 0 {
		return nil, 0, nil
	}

	page, size := normalizePage(req.Page, req.PageSize)
	offset := (page - 1) * size
	ids, err := r.pageOnAlbum(ctx, req, albumIDs, setAlbum, offset, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*types.Movie, 0, len(ids))
	for _, id := range ids {
		out = append(out, &types.Movie{JavId: id})
	}
	return out, int64(len(setAlbum)), nil
}

func (r *MovieListRepoSqlx) listFromSingleAlbumOnly(ctx context.Context, req *types.ListMovieFullRequest, albumID int64) ([]*types.Movie, int64, error) {
	w := singleAlbumBaseFilters(albumID, req)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From("`c_movie_album_item`").Where(w).ToSql()
	var total int64
	if err := r.cai.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	page, size := normalizePage(req.Page, req.PageSize)
	order := applyMovieListOrder("releasing_date DESC,movie_name DESC,movie_jav_id DESC", req.Order)
	sqlStr, args, _ := squirrel.
		Select("movie_jav_id").
		From("`c_movie_album_item`").
		Where(w).
		OrderBy(order).
		Offset(uint64((page - 1) * size)).
		Limit(uint64(size)).
		ToSql()

	var ids []string
	if err := r.cai.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
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

func (r *MovieListRepoSqlx) listFromGScStatOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	set, err := r.pickFromGScStat(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	if set == nil || len(set) == 0 {
		return nil, 0, nil
	}

	finalIDs := setToSlice(set)
	total := int64(len(finalIDs))

	page, size := normalizePage(req.Page, req.PageSize)
	var ids []string
	offset := (page - 1) * size

	switch req.OrderBy {
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime, consts.OrderByReleasingDate, consts.OrderByMediaBirthTime:
		ids, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
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

	needW := needWMedia(req)
	needS := needGScStat(req)

	var ids []string
	switch req.OrderBy {
	case consts.OrderByBirthTime:
		ids, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
	case consts.OrderByMediaBirthTime:
		if needS && !needW {
			ids, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ids, err = r.pageOnWMedia(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByReleasingDate:
		if needS && !needW {
			ids, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
		} else {
			ids, err = r.pageOnMedia(ctx, req, finalIDs, req.OrderBy, req.Order, offset, size)
		}
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ids, err = r.pageOnGScStat(ctx, finalIDs, req.OrderBy, req.Order, offset, size)
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

func needAlbum(req *types.ListMovieFullRequest) bool {
	return strings.TrimSpace(req.AlbumName) != ""
}

func albumBaseFilters(albumIDs []int64, req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{squirrel.Eq{"album_id": albumIDs}}
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

func singleAlbumBaseFilters(albumID int64, req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{squirrel.Eq{"album_id": albumID}}
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

func needAMovie(req *types.ListMovieFullRequest) bool {
	if req.MediaOwned == consts.OwnedNotOwned {
		return true
	}
	// 只在 a_movie 特有字段/排序/单值外键时命中
	if req.DirectorName != "" || req.PrefixName != "" || req.MakerName != "" || req.LabelName != "" || req.LabelJavID != "" {
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
		if needAlbum(req) || needMinfo(req) || needWMedia(req) || needGScStat(req) {
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

func needDownloadAlbumFilter(req *types.ListMovieFullRequest) squirrel.Sqlizer {
	if req.NeedDownload == consts.MovieNeedDownLoadOK {
		return squirrel.Expr(
			"EXISTS (SELECT 1 FROM `c_movie_album_item` cai JOIN `c_movie_album` ca ON ca.`id` = cai.`album_id` WHERE ca.`name` = ? AND cai.`movie_jav_id` = `jav_id`)",
			consts.MovieNeedDownloadAlbumName,
		)
	}
	if req.NeedDownload == consts.MovieNeedDownLoadNone {
		return squirrel.Expr(
			"NOT EXISTS (SELECT 1 FROM `c_movie_album_item` cai JOIN `c_movie_album` ca ON ca.`id` = cai.`album_id` WHERE ca.`name` = ? AND cai.`movie_jav_id` = `jav_id`)",
			consts.MovieNeedDownloadAlbumName,
		)
	}
	return nil
}

func hasGScStatCoreSignals(req *types.ListMovieFullRequest) bool {
	if req.ComeTimesMin > 0 || req.ComeTimesMax != nil ||
		req.LastScTimeMin != "" || req.LastScTimeMax != "" ||
		req.ScTimesMin > 0 || req.ScTimesMax != nil {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	}
	return false
}

func needGScStat(req *types.ListMovieFullRequest) bool {
	return hasGScStatCoreSignals(req)
}

func needWMedia(req *types.ListMovieFullRequest) bool {
	if (req.OrderBy == consts.OrderByBirthTime || req.OrderBy == consts.OrderByMediaBirthTime) && !hasGScStatCoreSignals(req) {
		return true
	}
	if (req.MediaOwned > consts.MovieAll && req.MediaOwned != consts.OwnedNotOwned) ||
		((req.MediaBirthTimeStart != "" || req.MediaBirthTimeEnd != "") && !hasGScStatCoreSignals(req)) ||
		req.MediaDir1 != "" || req.MediaDir2 != "" || req.MediaDir3 != "" || req.MediaDir4 != "" ||
		req.MediaDirFull != "" {
		return true
	}
	return false
}

func needMedia(req *types.ListMovieFullRequest) bool {
	return needWMedia(req)
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
	if needDownloadFilter := needDownloadAlbumFilter(req); needDownloadFilter != nil {
		w = append(w, needDownloadFilter)
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

func gscStatBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}

	if req.LastScTimeMin != "" {
		if ts, ok := parseYMD(req.LastScTimeMin); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.last_sc_time, 0) >= ?", ts))
		}
	}
	if req.LastScTimeMax != "" {
		if ts, ok := parseYMD(req.LastScTimeMax); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.last_sc_time, 0) <= ?", ts))
		}
	}
	if req.ScTimesMin > 0 {
		w = append(w, squirrel.Expr("COALESCE(gss.sc_times, 0) >= ?", req.ScTimesMin))
	}
	if req.ScTimesMax != nil {
		w = append(w, squirrel.Expr("COALESCE(gss.sc_times, 0) <= ?", *req.ScTimesMax))
	}
	if req.ComeTimesMin > 0 {
		w = append(w, squirrel.Expr("COALESCE(gss.come_times, 0) >= ?", req.ComeTimesMin))
	}
	if req.ComeTimesMax != nil {
		w = append(w, squirrel.Expr("COALESCE(gss.come_times, 0) <= ?", *req.ComeTimesMax))
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.releasing_date, 0) >= ?", ts))
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.releasing_date, 0) <= ?", ts))
		}
	}
	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseYMD(req.MediaBirthTimeStart); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.media_birth_time, 0) >= ?", ts))
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseYMDEnd(req.MediaBirthTimeEnd); ok {
			w = append(w, squirrel.Expr("COALESCE(gss.media_birth_time, 0) <= ?", ts))
		}
	}
	return w
}

func wmediaBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{squirrel.Eq{"source_type": consts.WMediaSourceNative}}

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
		return w
	case consts.MovieAll:
		return w
	case 0:
	}

	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseYMD(req.MediaBirthTimeStart); ok {
			w = append(w, squirrel.GtOrEq{"birth_time": ts})
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseYMDEnd(req.MediaBirthTimeEnd); ok {
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
	if req.MediaDirFull != "" {
		if req.MediaDirSub {
			w = append(w, squirrel.Or{
				squirrel.Eq{"full_dir": req.MediaDirFull},
				squirrel.Like{"full_dir": req.MediaDirFull + "/%"},
			})
		} else {
			w = append(w, squirrel.Eq{"full_dir": req.MediaDirFull})
		}
	}
	return w
}

func (r *MovieListRepoSqlx) buildNativeMediaMatchCondition(req *types.ListMovieFullRequest, alias string) (string, []any) {
	conds := []string{alias + ".source_type = ?"}
	args := []any{consts.WMediaSourceNative}

	switch req.MediaOwned {
	case consts.OwnedHasSubNotRemoved:
		conds = append(conds, alias+".has_sub = ?", alias+".is_removed = ?")
		args = append(args, consts.FilmHasSub, consts.FilmIsNotRemoved)
	case consts.OwnedNoSubNotRemoved:
		conds = append(conds, alias+".has_sub = ?", alias+".is_removed = ?")
		args = append(args, consts.FilmNoSub, consts.FilmIsNotRemoved)
	case consts.OwnedAllNotRemoved:
		conds = append(conds, alias+".is_removed = ?")
		args = append(args, consts.FilmIsNotRemoved)
	case consts.OwnedRemoved:
		conds = append(conds, alias+".is_removed = ?")
		args = append(args, consts.FilmIsRemoved)
	case consts.OwnedAll, consts.OwnedNotOwned, consts.MovieAll, 0:
	}

	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseYMD(req.MediaBirthTimeStart); ok {
			conds = append(conds, alias+".birth_time >= ?")
			args = append(args, ts)
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseYMDEnd(req.MediaBirthTimeEnd); ok {
			conds = append(conds, alias+".birth_time <= ?")
			args = append(args, ts)
		}
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseYMD(req.ReleasingDateStart); ok {
			conds = append(conds, alias+".releasing_date >= ?")
			args = append(args, ts)
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseYMD(req.ReleasingDateEnd); ok {
			conds = append(conds, alias+".releasing_date <= ?")
			args = append(args, ts)
		}
	}

	if req.MediaDir1 != "" {
		conds = append(conds, "CONCAT('/', "+alias+".full_dir, '/') LIKE ?")
		args = append(args, "%/"+req.MediaDir1+"/%")
	}
	if req.MediaDir2 != "" {
		conds = append(conds, "CONCAT('/', "+alias+".full_dir, '/') LIKE ?")
		args = append(args, "%/"+req.MediaDir2+"/%")
	}
	if req.MediaDir3 != "" {
		conds = append(conds, "CONCAT('/', "+alias+".full_dir, '/') LIKE ?")
		args = append(args, "%/"+req.MediaDir3+"/%")
	}
	if req.MediaDir4 != "" {
		conds = append(conds, "CONCAT('/', "+alias+".full_dir, '/') LIKE ?")
		args = append(args, "%/"+req.MediaDir4+"/%")
	}
	if req.MediaDirFull != "" {
		if req.MediaDirSub {
			conds = append(conds, "("+alias+".full_dir = ? OR "+alias+".full_dir LIKE ?)")
			args = append(args, req.MediaDirFull, req.MediaDirFull+"/%")
		} else {
			conds = append(conds, alias+".full_dir = ?")
			args = append(args, req.MediaDirFull)
		}
	}

	return strings.Join(conds, " AND "), args
}

func (r *MovieListRepoSqlx) buildMediaMatchedIDSelect(req *types.ListMovieFullRequest, finalIDs []string) squirrel.SelectBuilder {
	nativeCond, nativeArgs := r.buildNativeMediaMatchCondition(req, "wn")
	sb := squirrel.
		Select("DISTINCT wn.movie_jav_id").
		From(r.wm.TableName()+" wn").
		Where(nativeCond, nativeArgs...)
	if len(finalIDs) > 0 {
		sb = sb.Where(squirrel.Eq{"wn.movie_jav_id": finalIDs})
	}
	return sb
}

func (r *MovieListRepoSqlx) buildMediaMatchedSortableSelect(req *types.ListMovieFullRequest, finalIDs []string) (squirrel.SelectBuilder, string) {
	nativeCond, nativeArgs := r.buildNativeMediaMatchCondition(req, "wn")
	sb := squirrel.
		Select(
			"wn.movie_jav_id AS movie_jav_id",
			"wn.birth_time AS native_birth_time",
			"wn.releasing_date AS native_releasing_date",
		).
		From(r.wm.TableName()+" wn").
		Where(nativeCond, nativeArgs...)
	if len(finalIDs) > 0 {
		sb = sb.Where(squirrel.Eq{"wn.movie_jav_id": finalIDs})
	}
	return sb, "q"
}

func (r *MovieListRepoSqlx) buildGScStatMediaBaseSelect(finalIDs []string) (squirrel.SelectBuilder, []any, error) {
	base := squirrel.
		Select("DISTINCT movie_jav_id").
		From(r.wm.TableName()).
		Where(squirrel.Eq{"source_type": consts.WMediaSourceNative})
	if len(finalIDs) > 0 {
		base = base.Where(squirrel.Eq{"movie_jav_id": finalIDs})
	}

	baseSQL, baseArgs, err := base.ToSql()
	if err != nil {
		return squirrel.SelectBuilder{}, nil, err
	}

	sb := squirrel.
		Select("wm.movie_jav_id").
		From("(" + baseSQL + ") wm").
		LeftJoin(r.gss.TableName() + " gss ON gss.movie_jav_id = wm.movie_jav_id")
	return sb, baseArgs, nil
}

func (r *MovieListRepoSqlx) resolveAlbums(ctx context.Context, albumNames string) ([]*CMovieAlbum, error) {
	names := splitAlbumNames(albumNames)
	if len(names) == 0 {
		return nil, nil
	}
	out := make([]*CMovieAlbum, 0, len(names))
	seen := make(map[int64]struct{}, len(names))
	for _, name := range names {
		row, err := r.ca.FindOneByName(ctx, name)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			return nil, err
		}
		if row == nil || row.Id <= 0 {
			continue
		}
		if _, ok := seen[row.Id]; ok {
			continue
		}
		seen[row.Id] = struct{}{}
		out = append(out, row)
	}
	return out, nil
}

func movieAlbumRowsToIDs(rows []*CMovieAlbum) []int64 {
	if len(rows) == 0 {
		return nil
	}
	out := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.Id <= 0 {
			continue
		}
		out = append(out, row.Id)
	}
	return out
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
	if req.LabelJavID != "" {
		if id, ok := r.idOfLabelJavID(ctx, req.LabelJavID); ok {
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

func (r *MovieListRepoSqlx) pickFromGScStat(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	if !shouldPickFromGScStat(req) {
		return nil, nil
	}

	w := gscStatBaseFilters(req)
	w = append(w, gscStatOrderGuards(req.OrderBy)...)

	if len(w) == 0 && !orderBelongsToGScStat(req.OrderBy) {
		return nil, nil
	}

	sb, baseArgs, err := r.buildGScStatMediaBaseSelect(nil)
	if err != nil {
		return nil, err
	}

	if len(w) > 0 {
		sb = sb.Where(w)
	}
	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	args = append(baseArgs, args...)
	var ids []string
	if err := r.wm.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
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

func (r *MovieListRepoSqlx) pickFromAlbum(ctx context.Context, req *types.ListMovieFullRequest, albumIDs []int64) (map[string]struct{}, error) {
	if len(albumIDs) == 0 {
		return nil, nil
	}
	w := albumBaseFilters(albumIDs, req)

	sqlStr, args, err := squirrel.
		Select("movie_jav_id").
		From("`c_movie_album_item`").
		Where(w).
		ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.cai.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromMedia(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	if !needMedia(req) {
		return nil, nil
	}

	sb := r.buildMediaMatchedIDSelect(req, nil)

	sqlStr, args, err := sb.ToSql()
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

type albumReleaseRow struct {
	MovieJavID    string `db:"movie_jav_id"`
	MovieName     string `db:"movie_name"`
	ReleasingDate int64  `db:"releasing_date"`
}

type albumReleaseCursor struct {
	albumID   int64
	offset    int64
	batchSize int64
	rows      []*albumReleaseRow
	index     int
	exhausted bool
}

func (r *MovieListRepoSqlx) pageOnAlbum(ctx context.Context, req *types.ListMovieFullRequest, albumIDs []int64, allowed map[string]struct{}, offset, limit int64) ([]string, error) {
	if len(albumIDs) == 0 || limit <= 0 {
		return nil, nil
	}
	if len(albumIDs) == 1 && len(allowed) == 0 {
		return r.pageOnSingleAlbum(ctx, req, albumIDs[0], offset, limit)
	}

	cursors := make([]*albumReleaseCursor, 0, len(albumIDs))
	for _, albumID := range albumIDs {
		cursor := &albumReleaseCursor{
			albumID:   albumID,
			batchSize: limit,
		}
		if err := r.loadAlbumReleaseBatch(ctx, req, cursor); err != nil {
			return nil, err
		}
		cursors = append(cursors, cursor)
	}

	seen := make(map[string]struct{}, int(offset+limit))
	skipped := int64(0)
	out := make([]string, 0, limit)
	for len(out) < int(limit) {
		best := -1
		for i, cursor := range cursors {
			row, err := r.peekAlbumReleaseRow(ctx, req, cursor)
			if err != nil {
				return nil, err
			}
			if row == nil {
				continue
			}
			if best == -1 {
				best = i
				continue
			}
			bestRow := cursors[best].rows[cursors[best].index]
			if albumReleaseRowBefore(row, bestRow, req.Order) {
				best = i
			}
		}
		if best == -1 {
			break
		}
		row := cursors[best].rows[cursors[best].index]
		cursors[best].index++
		if len(allowed) > 0 {
			if _, ok := allowed[row.MovieJavID]; !ok {
				continue
			}
		}
		if _, ok := seen[row.MovieJavID]; ok {
			continue
		}
		seen[row.MovieJavID] = struct{}{}
		if skipped < offset {
			skipped++
			continue
		}
		out = append(out, row.MovieJavID)
	}
	return out, nil
}

func (r *MovieListRepoSqlx) pageOnSingleAlbum(ctx context.Context, req *types.ListMovieFullRequest, albumID int64, offset, limit int64) ([]string, error) {
	if albumID <= 0 || limit <= 0 {
		return nil, nil
	}
	order := applyMovieListOrder("releasing_date DESC,movie_name DESC,movie_jav_id DESC", req.Order)
	sb := squirrel.Select("movie_jav_id").
		From("`c_movie_album_item`").
		Where(singleAlbumBaseFilters(albumID, req)).
		OrderBy(order).
		Offset(uint64(offset)).
		Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.cai.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) peekAlbumReleaseRow(ctx context.Context, req *types.ListMovieFullRequest, cursor *albumReleaseCursor) (*albumReleaseRow, error) {
	for cursor != nil {
		if cursor.index < len(cursor.rows) {
			return cursor.rows[cursor.index], nil
		}
		if cursor.exhausted {
			return nil, nil
		}
		if err := r.loadAlbumReleaseBatch(ctx, req, cursor); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *MovieListRepoSqlx) loadAlbumReleaseBatch(ctx context.Context, req *types.ListMovieFullRequest, cursor *albumReleaseCursor) error {
	if cursor == nil || cursor.exhausted {
		return nil
	}
	order := applyMovieListOrder("releasing_date DESC,movie_name DESC,movie_jav_id DESC", req.Order)
	sb := squirrel.
		Select("movie_jav_id", "movie_name", "releasing_date").
		From("`c_movie_album_item`").
		Where(singleAlbumBaseFilters(cursor.albumID, req)).
		OrderBy(order).
		Offset(uint64(cursor.offset)).
		Limit(uint64(cursor.batchSize))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return err
	}
	var rows []*albumReleaseRow
	if err := r.cai.QueryRowsNoCacheCtx(ctx, &rows, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			cursor.rows = nil
			cursor.index = 0
			cursor.exhausted = true
			return nil
		}
		return err
	}
	cursor.rows = rows
	cursor.index = 0
	cursor.offset += int64(len(rows))
	if len(rows) < int(cursor.batchSize) {
		cursor.exhausted = true
	}
	return nil
}

func albumReleaseRowBefore(a, b *albumReleaseRow, sortOrder string) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	desc := normalizeMovieListOrder(sortOrder) != "asc"
	if a.ReleasingDate != b.ReleasingDate {
		if desc {
			return a.ReleasingDate > b.ReleasingDate
		}
		return a.ReleasingDate < b.ReleasingDate
	}
	if a.MovieName != b.MovieName {
		if desc {
			return a.MovieName > b.MovieName
		}
		return a.MovieName < b.MovieName
	}
	if desc {
		return a.MovieJavID > b.MovieJavID
	}
	return a.MovieJavID < b.MovieJavID
}

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

func (r *MovieListRepoSqlx) pageOnGScStat(ctx context.Context, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order, guards := gscStatOrdering(od, sortOrder)
	sb, baseArgs, err := r.buildGScStatMediaBaseSelect(finalIDs)
	if err != nil {
		return nil, err
	}

	w := squirrel.And{}
	w = append(w, guards...)
	if len(w) > 0 {
		sb = sb.Where(w)
	}
	sb = sb.OrderBy(order).
		Offset(uint64(offset)).
		Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	args = append(baseArgs, args...)
	var ids []string
	if err := r.wm.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnMedia(ctx context.Context, req *types.ListMovieFullRequest, finalIDs []string, od string, sortOrder string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}

	inner, innerAlias := r.buildMediaMatchedSortableSelect(req, finalIDs)

	innerSQL, innerArgs, err := inner.ToSql()
	if err != nil {
		return nil, err
	}

	order := mediaMatchedOrdering(od, sortOrder, innerAlias)
	sb := squirrel.
		Select(innerAlias + ".movie_jav_id").
		From("(" + innerSQL + ") " + innerAlias).
		OrderBy(order).
		Offset(uint64(offset)).
		Limit(uint64(limit))

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	args = append(innerArgs, args...)
	var ids []string
	if err := r.wm.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
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
		Where(squirrel.Eq{"source_type": consts.WMediaSourceNative}).
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
func (r *MovieListRepoSqlx) idOfLabelJavID(ctx context.Context, javID string) (int64, bool) {
	row, err := r.lb.FindOneByJavId(ctx, javID)
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

// 目录：按名称查 id（由 WFolderModel 提供）
func (r *MovieListRepoSqlx) findDirIdByNameSourceType(ctx context.Context, name string, sourceType int64) (int64, bool) {
	row, err := r.wf.FindOneByNameSourceType(ctx, name, sourceType)
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
	if req.MediaOwned == consts.OwnedNotOwned {
		w = append(w, squirrel.Expr(buildNativeWMediaNotExists(r.wm.TableName(), "wm_abs", "jav_id")))
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

func gscStatOrderGuards(orderBy string) squirrel.And {
	return squirrel.And{}
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

func gscStatOrdering(orderBy string, sortOrder string) (order string, guards squirrel.And) {
	order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	switch orderBy {
	case consts.OrderByScTimes:
		order = "COALESCE(gss.sc_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	case consts.OrderByComeTimes:
		order = "COALESCE(gss.come_times, 0) DESC,COALESCE(gss.last_sc_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	case consts.OrderByLastScTime:
		order = "COALESCE(gss.last_sc_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	case consts.OrderByReleasingDate:
		order = "COALESCE(gss.releasing_date, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	case consts.OrderByMediaBirthTime:
		order = "COALESCE(gss.media_birth_time, 0) DESC,gss.movie_name DESC,wm.movie_jav_id DESC"
	}
	order = applyMovieListOrder(order, sortOrder)
	guards = gscStatOrderGuards(orderBy)
	return
}

func mediaMatchedOrdering(orderBy string, sortOrder string, alias string) string {
	primary := alias + ".native_releasing_date DESC"
	switch orderBy {
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		primary = alias + ".native_birth_time DESC"
	case consts.OrderByReleasingDate:
		primary = alias + ".native_releasing_date DESC"
	}
	order := primary + "," + alias + ".movie_jav_id DESC"
	return applyMovieListOrder(order, sortOrder)
}

func wmediaOrdering(orderBy string, sortOrder string) string {
	order := "releasing_date DESC,movie_name DESC"
	switch orderBy {
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
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

func orderBelongsToGScStat(od string) bool {
	switch od {
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	default:
		return false
	}
}

func shouldPickFromGScStat(req *types.ListMovieFullRequest) bool {
	return needGScStat(req) || orderBelongsToGScStat(req.OrderBy)
}

func orderBelongsToWMedia(od string) bool {
	switch od {
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime, consts.OrderByReleasingDate:
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

func splitAlbumNames(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case ',', '，', '、', '|', ';', '；', '\t', '\n', '\r':
			return true
		default:
			return false
		}
	})
	seen := map[string]struct{}{}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
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

func parseYMDEnd(s string) (int64, bool) {
	if len(strings.TrimSpace(s)) < 8 {
		return 0, false
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return 0, false
	}
	return t.Add(24*time.Hour - time.Second).Unix(), true
}
