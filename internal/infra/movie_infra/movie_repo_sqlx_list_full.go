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
}

func NewMovieListRepoSqlx(am moviex.AMovieModel, mi moviex.BmMinfoModel, vf moviex.VFilmModel) *MovieListRepoSqlx {
	return &MovieListRepoSqlx{am: am, mi: mi, vf: vf}
}

// internal/infra/movie_infra/movie_repo_sqlx_list_full.go

func (r *MovieListRepoSqlx) ListFull(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	// 1) 先判定这次请求“命中”了哪些表（条件 or 排序）
	needA := needAMovie(req)
	needM := needMinfo(req)
	needF := needVFilm(req)

	onlyA := needA && !needM && !needF
	onlyM := needM && !needA && !needF
	onlyF := needF && !needA && !needM

	// 2) 单表直取（最快路径）
	if onlyA {
		return r.listFromAMovieOnly(ctx, req)
	}
	if onlyM {
		return r.listFromMinfoOnly(ctx, req)
	}
	if onlyF {
		return r.listFromVFilmOnly(ctx, req)
	}

	// 3) 命中多表：用你现有的交集逻辑
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

	// 4) 仍沿用“在排序所属表分页”的现有实现
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

/* ---------------- 单表直取：COUNT + ORDER/LIMIT（不做交集） ---------------- */

func (r *MovieListRepoSqlx) listFromAMovieOnly(ctx context.Context, req *types.ListMovieFullRequest) ([]*types.Movie, int64, error) {
	// WHERE
	w := whereAMovie(req)

	// COUNT(*)
	cntSql, cntArgs, _ := squirrel.Select("COUNT(*)").From(r.am.TableName()).Where(w).ToSql()
	var total int64
	if err := r.am.QueryRowNoCacheCtx(ctx, &total, cntSql, cntArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	// ORDER + PAGE
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
	w := whereMinfo(req)

	if req.OrderBy == consts.OrderByHighestRank {
		w = append(w, squirrel.Expr("highest_rank <> 0"))
	}
	if req.OrderBy == consts.OrderByRankDate {
		w = append(w, squirrel.Expr("days_in_rank <> 0"))
	}

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
	w := whereVFilm(req)

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

/* ---------------- 命中判断 & WHERE 复用 ---------------- */

func needAMovie(req *types.ListMovieFullRequest) bool {
	// 确认是否用了 a_movie 独有字段
	if req.CastAgeMin > 0 || req.CastAgeMax > 0 {
		return true
	}
	switch req.OrderBy {
	case consts.OrderByViewerWatched, consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc, consts.OrderByDetailUpdateTime:
		return true
	}

	// 仅发行日过滤/排序：如果本次请求还会命中 minfo 或 v_film，就让它们承接，不强制命中 a_movie
	releaseOnlyFilter := req.ReleasingDateStart != "" || req.ReleasingDateEnd != ""
	releaseOnlyOrder := req.OrderBy == consts.OrderByReleasingDate

	if releaseOnlyFilter || releaseOnlyOrder {
		// 若其它两表本就会被命中（例如 Owned/HasSub/Dir 或 NeedDownload/Word 等），则放弃 a_movie
		if needMinfo(req) || needVFilm(req) {
			return false
		}
		// 否则还是得靠 a_movie
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
	if req.Owned == 1 || req.HasSub > 0 || req.ComeTimesMin > 0 ||
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

func whereAMovie(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
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
	if req.CastAgeMin > 0 {
		w = append(w,
			squirrel.And{
				squirrel.GtOrEq{"cast_average_age": int64(req.CastAgeMin*10.0 + 0.5)},
				squirrel.NotEq{"cast_average_age": 0},
			},
		)
	}
	if req.CastAgeMax > 0 {
		w = append(w,
			squirrel.And{
				squirrel.LtOrEq{"cast_average_age": int64(req.CastAgeMax*10.0 + 0.5)},
				squirrel.NotEq{"cast_average_age": 0},
			},
		)
	}
	return w
}
func whereMinfo(req *types.ListMovieFullRequest) squirrel.And {
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
	// 发行日过滤（minfo 也有冗余）
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

func whereVFilm(req *types.ListMovieFullRequest) squirrel.And {
	w := squirrel.And{}
	if req.Owned == 1 {
		w = append(w, squirrel.Eq{"is_removed": 1})
	}
	if req.HasSub > 0 {
		w = append(w, squirrel.Eq{"has_sub": req.HasSub})
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
	// 发行日过滤（v_film 也有冗余）
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
	// 目录 LIKE
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

func normalizePage(page, size int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 18
	}
	return page, size
}

/* ------------------- A. 各表筛选（只查需要的表，无 JOIN） ------------------- */

func (r *MovieListRepoSqlx) pickFromAMovie(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := squirrel.And{}

	// CastNames / GenreNames / DirectorName / Prefix / Maker / Label:
	// 无 JOIN 的前提下，建议你有对应的“反查表” 或 在 a_movie 冗余名做 LIKE。
	// 这里给出“最保守”的做法：只用 a_movie 自身可用字段（name/title/…）与时间、年龄筛选。
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

	// 演员平均年龄（*10 存储）
	if req.CastAgeMin > 0 {
		w = append(w, squirrel.GtOrEq{"cast_average_age": int64(req.CastAgeMin*10.0 + 0.5)})
	}
	if req.CastAgeMax > 0 {
		w = append(w, squirrel.LtOrEq{"cast_average_age": int64(req.CastAgeMax*10.0 + 0.5)})
	}

	// 如果用户没有在 a_movie 维度提供任何条件，也可以不查 a_movie（返回 nil 表示“不限制”）
	if len(w) == 0 && req.OrderBy != consts.OrderByReleasingDate &&
		req.OrderBy != consts.OrderByViewerWatched && req.OrderBy != consts.OrderByCastAgeAsc &&
		req.OrderBy != consts.OrderByCastAgeDesc {
		return nil, nil
	}

	sb := squirrel.Select("jav_id").From(r.am.TableName()).Where(w)
	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}
	var ids []string
	if err := r.am.QueryRowsNoCacheCtx(ctx, &ids, sqlStr, args...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}

	return sliceToSet(ids), nil
}

func (r *MovieListRepoSqlx) pickFromMinfo(ctx context.Context, req *types.ListMovieFullRequest) (map[string]struct{}, error) {
	w := squirrel.And{}

	if req.StartRankingDate != "" {
		if ts, ok := parseYMD(req.StartRankingDate); ok {
			w = append(w, squirrel.GtOrEq{"first_rank_day_number": ts})
		}
	}
	if req.NeedDownload > 0 {
		w = append(w, squirrel.Eq{"need_download": req.NeedDownload})
	}
	// 关键词额外在 m.chinese 命中
	if req.Word != "" {
		w = append(w, squirrel.Like{"chinese": "%" + req.Word + "%"})
	}

	if req.OrderBy == consts.OrderByHighestRank {
		w = append(w, squirrel.Expr("highest_rank <> 0"))
	}

	// 如果该表没有被任何条件/排序命中，可以返回 nil 表示“不限制”
	if len(w) == 0 && req.OrderBy != consts.OrderByRankDate && req.OrderBy != consts.OrderByHighestRank {
		return nil, nil
	}

	sb := squirrel.Select("jav_id").From(r.mi.TableName()).Where(w)
	sqlStr, args, err := sb.ToSql()
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
	w := squirrel.And{}

	// Owned：1=本地有文件；0=没有。无 JOIN 时，Owned=1 => v_film 有记录且未标记删除；Owned=0 => v_film 无记录
	// 由于我们只查 v_film，一旦用户要求 Owned=0，这里无法“直接”给出集合（需要全体 javId 再差集）。
	// 为保持“只查一表”的约束，这里：
	// - Owned=1：查出 f.movie_jav_id 作为集合
	// - Owned=0：返回 nil（上层在排序分页时不依赖 v_film 集合；Owned=0 用 Service 侧兜底或另行处理）
	if req.Owned == 1 {
		w = append(w, squirrel.Eq{"is_removed": 1}) // 按你的常量替换
	}
	if req.HasSub > 0 {
		w = append(w, squirrel.Eq{"has_sub": req.HasSub})
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

	// 目录名筛选：走 full_dir LIKE，避免 JOIN
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

	// 若 v_film 完全未命中任何条件且排序也不依赖 v_film 字段，则返回 nil 表示“不限制”
	if len(w) == 0 && !orderBelongsToVFilm(req.OrderBy) {
		return nil, nil
	}

	sb := squirrel.Select("movie_jav_id").
		From(r.vf.TableName()).
		Where(w)
	sqlStr, args, err := sb.ToSql()
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

/* ------------------- B. 分表排序分页（WHERE jav_id IN (...)） ------------------- */

func (r *MovieListRepoSqlx) pageOnAMovie(ctx context.Context, finalIDs []string, od string, offset, limit int64) ([]string, error) {
	if len(finalIDs) == 0 {
		return nil, nil
	}
	order := "releasing_date DESC"
	switch od {
	case consts.OrderByReleasingDate:
		order = "releasing_date DESC"
	case consts.OrderByViewerWatched:
		order = "viewers_number_watched DESC"
	case consts.OrderByCastAgeAsc:
		order = "cast_average_age ASC"
	case consts.OrderByCastAgeDesc:
		order = "cast_average_age DESC"
	}
	sb := squirrel.Select("jav_id").
		From(r.am.TableName()).
		Where(squirrel.Eq{"jav_id": finalIDs}).
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
		// 若传错，退回按 amovie 排序
		return r.pageOnAMovie(ctx, finalIDs, consts.OrderByReleasingDate, offset, limit)
	}
	sb := squirrel.Select("jav_id").
		From(r.mi.TableName()).
		Where(squirrel.Eq{"jav_id": finalIDs}).
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
	sb := squirrel.Select("movie_jav_id").
		From(r.vf.TableName()).
		Where(squirrel.Eq{"movie_jav_id": finalIDs}).
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

/* ------------------- C. 辅助 ------------------- */

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

func setToSlice(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func intersectNonEmpty(sets ...map[string]struct{}) []string {
	// 跳过 nil（表示“该表不限制”），对非空集合取交集。
	nonNil := make([]map[string]struct{}, 0, len(sets))
	for _, s := range sets {
		if s != nil {
			nonNil = append(nonNil, s)
		}
	}
	if len(nonNil) == 0 {
		// 三表都不限制 -> 全表；但我们不能全表扫描。
		// 这里退化为：返回空（交给上层去处理 CountAll+按 OrderBy 的表分页拉取）
		return []string{}
	}
	// 以最小集合为基准交集
	sort.Slice(nonNil, func(i, j int) bool { return len(nonNil[i]) < len(nonNil[j]) })
	base := nonNil[0]
	res := make([]string, 0, len(base))
	for id := range base {
		ok := true
		for k := 1; k < len(nonNil); k++ {
			if _, hit := nonNil[k][id]; !hit {
				ok = false
				break
			}
		}
		if ok {
			res = append(res, id)
		}
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
	//}
}
