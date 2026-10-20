package moviex

import (
	"context"

	"github.com/Masterminds/squirrel"

	"rudy_gc/local/api/internal/consts"
	"rudy_gc/local/api/internal/types"
)

type cardPeriodSQLModel interface {
	QueryRowNoCacheCtx(context.Context, any, string, ...any) error
	QueryRowsNoCacheCtx(context.Context, any, string, ...any) error
}

// ListCardPeriod filters and pages in SQL, without materializing a period's IDs.
// The existing order builders keep card ordering identical to the other pages.
func (r *MovieListRepoSqlx) ListCardPeriod(ctx context.Context, req *types.ListMovieFullRequest) ([]string, int64, error) {
	from, id := r.am.TableName(), "jav_id"
	var model cardPeriodSQLModel = r.am
	order, guards := amovieOrdering(req.OrderBy, req.Order)
	mediaSource := false
	join := ""
	switch req.OrderBy {
	case consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank:
		from, model = r.mi.TableName(), r.mi
		order, guards = minfoOrdering(req.OrderBy, req.Order)
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		from, id, model = r.wm.TableName()+" wm", "wm.movie_jav_id", r.wm
		order, guards = wmediaOrdering(req.OrderBy, req.Order), nil
		mediaSource = true
	case consts.OrderByScTimes, consts.OrderByComeTimes, consts.OrderByLastSelectedScTime:
		from, id, model = r.wm.TableName()+" wm", "wm.movie_jav_id", r.wm
		join = r.gss.TableName() + " gss ON gss.movie_jav_id = wm.movie_jav_id"
		order, guards = gscStatOrdering(req.OrderBy, req.Order)
		mediaSource = true
	}
	// Native-media sorting and absence of native media are mutually exclusive.
	if mediaSource && req.MediaOwned == consts.OwnedNotOwned {
		return []string{}, 0, nil
	}
	outerMovieID := id
	if !mediaSource {
		outerMovieID = from + ".jav_id"
	}
	where := squirrel.And{}
	where = append(where, guards...)
	if req.ReleasingDateStart != "" || req.ReleasingDateEnd != "" {
		if from == r.am.TableName() {
			where = append(where, amovieBaseFilters(req)...)
		} else {
			// Release periods always use the movie's release date, regardless of sort.
			conditions := squirrel.And{squirrel.Expr("period_movie.jav_id = " + outerMovieID)}
			if start, ok := parseYMD(req.ReleasingDateStart); ok {
				conditions = append(conditions, squirrel.GtOrEq{"period_movie.releasing_date": start})
			}
			if end, ok := parseYMD(req.ReleasingDateEnd); ok {
				conditions = append(conditions, squirrel.LtOrEq{"period_movie.releasing_date": end})
			}
			releaseSQL, args, err := squirrel.Select("1").From(r.am.TableName() + " period_movie").Where(conditions).ToSql()
			if err != nil {
				return nil, 0, err
			}
			where = append(where, squirrel.Expr("EXISTS ("+releaseSQL+")", args...))
		}
	}
	if req.MediaOwned == consts.OwnedNotOwned {
		where = append(where, squirrel.Expr(buildNativeWMediaNotExists(r.wm.TableName(), "period_unowned", outerMovieID)))
	}
	ownedMedia := req.MediaOwned > consts.MovieAll && req.MediaOwned != consts.OwnedNotOwned
	if mediaSource || ownedMedia || req.MediaBirthTimeStart != "" || req.MediaBirthTimeEnd != "" {
		mediaFilter := *req
		mediaFilter.ReleasingDateStart, mediaFilter.ReleasingDateEnd = "", ""
		if mediaSource {
			condition, args := r.buildNativeMediaMatchCondition(&mediaFilter, "wm")
			where = append(where, squirrel.Expr(condition, args...))
		} else {
			condition, args := r.buildNativeMediaMatchCondition(&mediaFilter, "period_media")
			mediaSQL := "EXISTS (SELECT 1 FROM " + r.wm.TableName() + " period_media WHERE period_media.movie_jav_id = " + outerMovieID + " AND " + condition + ")"
			where = append(where, squirrel.Expr(mediaSQL, args...))
		}
	}
	base := squirrel.Select().From(from).Where(where)
	if join != "" {
		base = base.LeftJoin(join)
	}
	countSQL, args, err := base.Columns("COUNT(*)").ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := model.QueryRowNoCacheCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}
	ids := []string{}
	if total == 0 {
		return ids, 0, nil
	}
	page, size := normalizePage(req.Page, req.PageSize)
	pageSQL, args, err := base.Columns(id).OrderBy(order).Limit(uint64(size)).Offset(uint64((page - 1) * size)).ToSql()
	if err != nil {
		return nil, 0, err
	}
	if err := model.QueryRowsNoCacheCtx(ctx, &ids, pageSQL, args...); err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}
