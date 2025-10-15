// internal/infra/movie_infra/movie_repo_sqlx_list_full.go
package movie_infra

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"rudy_gc/data/modelx/moviex"
	"rudy_gc/internal/consts"
	"rudy_gc/internal/types"
)

type MovieListRepoSqlx struct {
	am moviex.AMovieModel
	mi moviex.BmMinfoModel
	vf moviex.VFilmModel

	lb moviex.AmLabelModel
	mk moviex.AmMakerModel
	dr moviex.AmDirectorModel
	px moviex.AmPrefixModel

	cs moviex.AmCastModel
	gr moviex.AmGenreModel
	rc moviex.AmrMovieCastModel
	rg moviex.AmrMovieGenreModel
}

func NewMovieListRepoSqlx(
	am moviex.AMovieModel,
	mi moviex.BmMinfoModel,
	vf moviex.VFilmModel,
	lb moviex.AmLabelModel,
	mk moviex.AmMakerModel,
	dr moviex.AmDirectorModel,
	px moviex.AmPrefixModel,
	cs moviex.AmCastModel,
	gr moviex.AmGenreModel,
	rc moviex.AmrMovieCastModel,
	rg moviex.AmrMovieGenreModel,
) *MovieListRepoSqlx {
	return &MovieListRepoSqlx{
		am: am, mi: mi, vf: vf,
		lb: lb, mk: mk, dr: dr, px: px,
		cs: cs, gr: gr, rc: rc, rg: rg,
	}
}

/* ------------------------------ 顶层入口 ------------------------------ */

func (r *MovieListRepoSqlx) ListFull(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	// 1) 判定命中的表
	needA := needAMovie(req)
	needM := needMinfo(req)
	needF := needVFilm(req)

	// Cast/Genre 是多对多，一旦出现就不能走 onlyA（因为需要交集）
	hasM2M := req.CastNames != "" || req.GenreNames != ""

	onlyA := needA && !needM && !needF && !hasM2M
	onlyM := needM && !needA && !needF
	onlyF := needF && !needA && !needM

	// 2) 单表直取
	if onlyA {
		return r.listFromAMovieOnly(ctx, req)
	}
	if onlyM {
		return r.listFromMinfoOnly(ctx, req)
	}
	if onlyF {
		return r.listFromVFilmOnly(ctx, req)
	}

	// 3) 多表：各自挑出 jav_id 集合后做交集
	setA, err := r.pickFromAMovie(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setM, err := r.pickFromMinfo(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	setF, err := r.pickFromVFilm(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	finalIDs := intersectNonEmpty(setA, setM, setF)
	total := int64(len(finalIDs))
	if total == 0 {
		return nil, 0, nil
	}

	// 4) 在排序所属表分页
	page := req.Page
	size := req.PageSize
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 18
	}
	offset := (page - 1) * size

	var ordered []string
	switch req.OrderBy {
	case consts.OrderByBirthTime, consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		ordered, err = r.pageOnVFilm(ctx, finalIDs, req.OrderBy, offset, size)
	case consts.OrderByRankDate, consts.OrderByHighestRank:
		ordered, err = r.pageOnMinfo(ctx, finalIDs, req.OrderBy, offset, size)
	default:
		ordered, err = r.pageOnAMovie(ctx, finalIDs, req.OrderBy, offset, size)
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

	// ★ 在 onlyA 路径补充单值外键过滤（名称→ID）
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

	w = append(w, amovieOrderGuards(req.OrderBy)...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.am.TableName()).Where(w).ToSql()
	var total int64
	if err := r.am.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	order := "releasing_date DESC"
	switch req.OrderBy {
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC"
	case consts.OrderByDetailUpdateTime:
		order = "detail_update_time DESC"
	case consts.OrderByViewerWatched:
		order = "viewers_number_watched DESC"
	case consts.OrderByCastAgeAsc:
		order = "cast_average_age ASC"
	case consts.OrderByCastAgeDesc:
		order = "cast_average_age DESC"
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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, 0, nil
		}
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
	w = append(w, minfoOrderGuards(req.OrderBy)...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.mi.TableName()).Where(w).ToSql()
	var total int64
	if err := r.mi.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	order := "first_rank_day_number desc,name desc"
	switch req.OrderBy {
	case consts.OrderByHighestRank:
		order = "highest_rank ASC"
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC"
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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, 0, nil
		}
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
	w = append(w, vfilmOrderGuards(req.OrderBy)...)

	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.vf.TableName()).Where(w).ToSql()
	var total int64
	if err := r.vf.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	order := "birth_time DESC"
	switch req.OrderBy {
	case consts.OrderByBirthTime:
		order = "birth_time DESC"
	case consts.OrderByScTimes:
		order = "sc_times DESC"
	case consts.OrderByComeTimes:
		order = "come_times DESC"
	case consts.OrderByLastScTime:
		order = "last_sc_time DESC"
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC"
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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, 0, nil
		}
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
	// ★ 必修复：Owned==MovieAll 时必须命中 a_movie（走 onlyA 快速路径）
	if req.Owned == consts.MovieAll {
		return true
	}
	// ★ 一旦有 a_movie 单值外键过滤，必须命中 a_movie
	if req.DirectorName != "" || req.PrefixName != "" || req.MakerName != "" || req.LabelName != "" {
		return true
	}
	// a_movie 独有
	if req.CastAgeMin > 0 || req.CastAgeMax > 0 {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByViewerWatched, consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc, consts.OrderByDetailUpdateTime:
		return true
	}
	// 仅发行日过滤/排序：若其余表会命中则不强制
	releaseOnlyFilter := req.ReleasingDateStart != "" || req.ReleasingDateEnd != ""
	releaseOnlyOrder := req.OrderBy == consts.OrderByReleasingDate
	if releaseOnlyFilter || releaseOnlyOrder {
		if needMinfo(req) || needVFilm(req) {
			return false
		}
		return true
	}
	return false
}

func needMinfo(req *types.ListMovieFullRequest) bool {
	if req.StartRankingDate != "" || req.NeedDownload > 0 || req.Word != "" {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByRankDate, consts.OrderByHighestRank:
		return true
	}
	return false
}

func needVFilm(req *types.ListMovieFullRequest) bool {
	if req.Owned == consts.MovieAll {
		return false
	}
	if req.Owned > consts.MovieAll || // OwnedAll/OwnedAllNotRemoved/... 都需要 v_film
		req.ComeTimesMin > 0 ||
		req.LastScTimeMin != "" || req.ScTimesMin > 0 || req.ScTimesMax != nil ||
		req.FilmBirthTimeStart != "" || req.FilmBirthTimeEnd != "" ||
		req.Dir1 != "" || req.Dir2 != "" || req.Dir3 != "" || req.Dir4 != "" {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByBirthTime, consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	}
	return false
}

/* ---------------- 共享的“基础过滤构造器” ---------------- */

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
	return w
}

func minfoBaseFilters(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
	if req.StartRankingDate != "" {
		if ts, ok := parseYMD(req.StartRankingDate); ok {
			w = append(w, squirrel.GtOrEq{"first_rank_day_number": ts})
		}
	}
	if req.NeedDownload > 0 {
		w = append(w, squirrel.Eq{"need_download": req.NeedDownload})
	}
	if req.Word != "" {
		w = append(w, squirrel.Like{"chinese": "%" + req.Word + "%"})
	}
	// 发行日（minfo 冗余）
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
	case consts.MovieAll:
		return squirrel.And{}
	}

	if req.ComeTimesMin > 0 {
		w = append(w, squirrel.GtOrEq{"come_times": req.ComeTimesMin})
	}
	if req.LastScTimeMin != "" {
		if ts, ok := parseYMD(req.LastScTimeMin); ok {
			w = append(w, squirrel.GtOrEq{"last_sc_time": ts})
		}
	}
	if req.ScTimesMin > 0 {
		w = append(w, squirrel.GtOrEq{"sc_times": req.ScTimesMin})
	}
	if req.ScTimesMax != nil {
		w = append(w, squirrel.LtOrEq{"sc_times": *req.ScTimesMax})
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
	// 发行日（v_film 冗余）
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
	// full_dir LIKE
	if req.Dir1 != "" {
		w = append(w, squirrel.Like{"full_dir": "%/" + req.Dir1 + "/%"})
	}
	if req.Dir2 != "" {
		w = append(w, squirrel.Like{"full_dir": "%/" + req.Dir2 + "/%"})
	}
	if req.Dir3 != "" {
		w = append(w, squirrel.Like{"full_dir": "%/" + req.Dir3 + "/%"})
	}
	if req.Dir4 != "" {
		w = append(w, squirrel.Like{"full_dir": "%/" + req.Dir4 + "/%"})
	}
	return w
}

/* ---------------- whereXxx：仅返回“基础过滤”以供复用 ---------------- */

func whereAMovie(req *types.ListMovieFullRequest) squirrel.And { return amovieBaseFilters(req) }
func whereMinfo(req *types.ListMovieFullRequest) squirrel.And  { return minfoBaseFilters(req) }
func whereVFilm(ctx context.Context, r *MovieListRepoSqlx, req *types.ListMovieFullRequest) squirrel.And {
	return vfilmBaseFilters(ctx, r, req)
}

/* ------------------- A. 各表筛选（只查需要的表，无 JOIN） ------------------- */

func (r *MovieListRepoSqlx) pickFromAMovie(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := amovieBaseFilters(req)

	// 单值外键：精确命名 → 查 ID → 等值过滤
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

	// 基集合：只有当 w 非空时才查询 a_movie；否则置为 nil（表示“不限制”）
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

	// 无条件且排序不依赖 minfo：返回 nil “不限制”
	if len(w) == 0 && req.OrderBy != consts.OrderByRankDate && req.OrderBy != consts.OrderByHighestRank {
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

	// 若 v_film 完全未命中且排序也不依赖 v_film 字段，则返回 nil 表示“不限制”
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

/* ---------------- 多对多关系查询（单名；外层已拆分多名并逐名交集） ---------------- */

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

/* ------------------- B. 分表排序分页（WHERE ... IN） ------------------- */

func (r *MovieListRepoSqlx) pageOnAMovie(ctx context.Context, finalIDs []string, od string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order := "releasing_date DESC"
	switch od {
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC"
	case consts.OrderByDetailUpdateTime:
		order = "detail_update_time DESC"
	case consts.OrderByViewerWatched:
		order = "viewers_number_watched DESC"
	case consts.OrderByCastAgeAsc:
		order = "cast_average_age ASC"
	case consts.OrderByCastAgeDesc:
		order = "cast_average_age DESC"
	}

	w := squirrel.And{squirrel.Eq{"jav_id": finalIDs}}
	w = append(w, amovieOrderGuards(od)...)

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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnMinfo(ctx context.Context, finalIDs []string, od string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order := ""
	switch od {
	case consts.OrderByRankDate:
		order = "first_rank_day_number DESC"
	case consts.OrderByHighestRank:
		order = "highest_rank ASC"
	default:
		return r.pageOnAMovie(ctx, finalIDs, consts.OrderByReleasingDate, offset, limit)
	}

	w := squirrel.And{squirrel.Eq{"jav_id": finalIDs}}
	w = append(w, minfoOrderGuards(od)...)

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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}

func (r *MovieListRepoSqlx) pageOnVFilm(ctx context.Context, finalIDs []string, od string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order := "birth_time DESC"
	switch od {
	case consts.OrderByBirthTime:
		order = "birth_time DESC"
	case consts.OrderByScTimes:
		order = "sc_times DESC"
	case consts.OrderByComeTimes:
		order = "come_times DESC"
	case consts.OrderByLastScTime:
		order = "last_sc_time DESC"
	default:
		return r.pageOnAMovie(ctx, finalIDs, consts.OrderByReleasingDate, offset, limit)
	}

	w := squirrel.And{squirrel.Eq{"movie_jav_id": finalIDs}}
	w = append(w, vfilmOrderGuards(od)...)

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
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ids, nil
}

/* ------------------- C. ID 查找（单值，不 split） ------------------- */

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

/* ------------------- D. 辅助 ------------------- */

func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 18
	}
	return page, size
}

func splitNames(s string) []string {
	// 支持英文/中文逗号、空格、斜杠等常见分隔
	seps := []string{",", "，", "、", "/", "|"}
	for _, sp := range seps {
		s = strings.ReplaceAll(s, sp, " ")
	}
	parts := strings.Fields(s)
	// 去重
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
		// 任意一方“不限制”，返回另一方
		if a == nil {
			return b
		}
		return a
	}
	// 都非空，真正交集
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

func intersectNonEmpty(sets ...map[string]struct{}) []string {
	// 跳过 nil（“不限制”），对非空集合取交集
	nonNil := make([]map[string]struct{}, 0, len(sets))
	for _, s := range sets {
		if s != nil {
			nonNil = append(nonNil, s)
		}
	}
	if len(nonNil) == 0 {
		// 三表都不限制 → 交由上层做全表分页（此处返回空切片）
		return []string{}
	}
	// 以最小集合为基
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

func orderBelongsToVFilm(od string) bool {
	switch od {
	case consts.OrderByBirthTime, consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastScTime:
		return true
	default:
		return false
	}
}

/* ---------------- 排序护栏：保证不同路径一致 ---------------- */

func amovieOrderGuards(orderBy string) squirrel.And {
	w := squirrel.And{}
	switch orderBy {
	case consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc:
		w = append(w, squirrel.NotEq{"cast_average_age": 0})
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
	}
	return w
}

func vfilmOrderGuards(orderBy string) squirrel.And {
	// 目前无特殊护栏
	return squirrel.And{}
}
