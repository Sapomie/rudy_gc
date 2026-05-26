package modelx

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rudy-gc-api/internal/config"
	"rudy-gc-api/internal/consts"
	"rudy-gc-api/internal/types"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MovieReadRepo struct {
	conn sqlx.SqlConn
	cfg  config.Config
}

type cardBaseRow struct {
	MovieJavID           string  `db:"movie_jav_id"`
	MovieName            string  `db:"movie_name"`
	Title                string  `db:"title"`
	EncodeName           string  `db:"encode_name"`
	ReleasingDate        int64   `db:"releasing_date"`
	DetailUpdateTime     int64   `db:"detail_update_time"`
	ScoreRaw             int64   `db:"score_raw"`
	ViewersNumberWatched int64   `db:"viewers_number_watched"`
	PrefixName           string  `db:"prefix_name"`
	MakerName            string  `db:"maker_name"`
	LabelName            string  `db:"label_name"`
	DirectorName         string  `db:"director_name"`
	ChineseTitle         string  `db:"chinese_title"`
	JacketImg            string  `db:"jacket_img"`
	JacketImgLocal       string  `db:"jacket_img_local"`
	HighestRank          int64   `db:"highest_rank"`
	DaysInRank           int64   `db:"days_in_rank"`
	FirstRankDayNumber   int64   `db:"first_rank_day_number"`
	ScTimes              int64   `db:"sc_times"`
	ComeTimes            int64   `db:"come_times"`
	LastScTime           int64   `db:"last_sc_time"`
	NeedDownload         int64   `db:"need_download"`
	WMediaID             int64   `db:"w_media_id"`
	WMediaBirthTime      int64   `db:"w_media_birth_time"`
	WMediaHasSub         int64   `db:"w_media_has_sub"`
	WMediaIsRemoved      int64   `db:"w_media_is_removed"`
	WMediaFullDir        string  `db:"w_media_full_dir"`
	WMediaFileName       string  `db:"w_media_file_name"`
	WMediaSourceHash     string  `db:"w_media_source_hash"`
	WMediaSize           int64   `db:"w_media_size"`
	WMediaHeight         int64   `db:"w_media_height"`
	WMediaBitRate        int64   `db:"w_media_bit_rate"`
	WMediaDuration       int64   `db:"w_media_duration"`
	WMediaFrameAverage   float64 `db:"w_media_frame_average"`
	WMediaSelfMake       int64   `db:"w_media_self_make"`
}

type movieScRow struct {
	Name          string `db:"name"`
	ScTime        int64  `db:"sc_time"`
	ComeMovieName string `db:"come_movie_name"`
	Cooldown      int64  `db:"cooldown"`
}

type movieRankRow struct {
	DayNumber int64 `db:"day_number"`
	RankPos   int64 `db:"rank_pos"`
}

type cardIDQueryPlan struct {
	source   string
	countSQL string
	countArg []any
	listSQL  string
	listArg  []any
}

func NewMovieReadRepo(conn sqlx.SqlConn, cfg config.Config) *MovieReadRepo {
	return &MovieReadRepo{conn: conn, cfg: cfg}
}

func (r *MovieReadRepo) ListMovieIDs(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	plan, ok, err := r.buildCardIDFastPathPlan(req)
	if err != nil {
		return nil, 0, err
	}
	if ok {
		return r.queryCardIDPlan(ctx, plan)
	}
	return r.listMovieIDsGeneric(ctx, req)
}

func (r *MovieReadRepo) queryCardIDPlan(ctx context.Context, plan *cardIDQueryPlan) ([]string, int64, error) {
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, plan.countSQL, plan.countArg...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	var ids []string
	if err := r.conn.QueryRowsCtx(ctx, &ids, plan.listSQL, plan.listArg...); err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}

func (r *MovieReadRepo) listMovieIDsGeneric(ctx context.Context, req *types.CardsListRequest) ([]string, int64, error) {
	builder := r.baseCardQuery(req)

	countSQL, countArgs, err := builder.
		Column("COUNT(DISTINCT m.jav_id) AS total").
		ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, countSQL, countArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	page := normalizePage(req.Page, 1)
	pageSize := normalizePageSize(req.PageSize, 18)
	offset := (page - 1) * pageSize

	listBuilder := r.baseCardQuery(req).
		Column("m.jav_id AS movie_jav_id").
		GroupBy("m.jav_id")

	if isRandomView(req.View) {
		listBuilder = listBuilder.OrderBy("RAND()").Limit(uint64(req.RandomN))
	} else {
		listBuilder = listBuilder.OrderBy(cardGroupedOrderClause(req.OrderBy, req.Order)).
			Offset(uint64(offset)).
			Limit(uint64(pageSize))
	}

	listSQL, listArgs, err := listBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var ids []string
	if err := r.conn.QueryRowsCtx(ctx, &ids, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}

func (r *MovieReadRepo) buildCardIDFastPathPlan(req *types.CardsListRequest) (*cardIDQueryPlan, bool, error) {
	if req == nil || isRandomView(req.View) {
		return nil, false, nil
	}
	if canUseAlbumIDFastPath(req) {
		plan, err := r.buildAlbumIDPlan(req)
		return plan, err == nil, err
	}
	if canUseWMediaIDFastPath(req) {
		plan, err := r.buildWMediaIDPlan(req)
		return plan, err == nil, err
	}
	if canUseMinfoIDFastPath(req) {
		plan, err := r.buildMinfoIDPlan(req)
		return plan, err == nil, err
	}
	if canUseAMovieIDFastPath(req) {
		plan, err := r.buildAMovieIDPlan(req)
		return plan, err == nil, err
	}
	return nil, false, nil
}

func (r *MovieReadRepo) buildAMovieIDPlan(req *types.CardsListRequest) (*cardIDQueryPlan, error) {
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

	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("a_movie"), where)
	listBuilder := applyWhere(squirrel.Select("jav_id").From("a_movie"), where).
		OrderBy(cardAMovieOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((normalizePage(req.Page, 1) - 1) * normalizePageSize(req.PageSize, 18))).
		Limit(uint64(normalizePageSize(req.PageSize, 18)))
	return buildCardIDPlan("a_movie", countBuilder, listBuilder)
}

func (r *MovieReadRepo) buildMinfoIDPlan(req *types.CardsListRequest) (*cardIDQueryPlan, error) {
	where := squirrel.And{}
	if req.View == "cardshasrank" {
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

	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("bm_minfo"), where)
	listBuilder := applyWhere(squirrel.Select("jav_id").From("bm_minfo"), where).
		OrderBy(cardMinfoOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((normalizePage(req.Page, 1) - 1) * normalizePageSize(req.PageSize, 18))).
		Limit(uint64(normalizePageSize(req.PageSize, 18)))
	return buildCardIDPlan("bm_minfo", countBuilder, listBuilder)
}

func (r *MovieReadRepo) buildWMediaIDPlan(req *types.CardsListRequest) (*cardIDQueryPlan, error) {
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

	countBuilder := applyWhere(squirrel.Select("COUNT(DISTINCT movie_jav_id)").From("w_media"), where)
	listBuilder := applyWhere(squirrel.Select("movie_jav_id").From("w_media"), where).
		GroupBy("movie_jav_id").
		OrderBy(cardWMediaOrderClause(req.OrderBy, req.Order)).
		Offset(uint64((normalizePage(req.Page, 1) - 1) * normalizePageSize(req.PageSize, 18))).
		Limit(uint64(normalizePageSize(req.PageSize, 18)))
	return buildCardIDPlan("w_media", countBuilder, listBuilder)
}

func (r *MovieReadRepo) buildAlbumIDPlan(req *types.CardsListRequest) (*cardIDQueryPlan, error) {
	albumName := strings.TrimSpace(req.AlbumName)
	if req.View == "cardsneeddownload" || req.NeedDownload == consts.MovieNeedDownloadOK {
		albumName = consts.MovieNeedDownloadAlbumName
	}
	where := squirrel.And{squirrel.Eq{"ca.name": albumName}}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			where = append(where, squirrel.GtOrEq{"cai.releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			where = append(where, squirrel.LtOrEq{"cai.releasing_date": ts})
		}
	}

	from := "c_movie_album_item cai"
	join := "c_movie_album ca ON ca.id = cai.album_id"
	countBuilder := applyWhere(squirrel.Select("COUNT(DISTINCT cai.movie_jav_id)").From(from).Join(join), where)
	listBuilder := applyWhere(squirrel.Select("cai.movie_jav_id").From(from).Join(join), where).
		GroupBy("cai.movie_jav_id").
		OrderBy(cardAlbumOrderClause(req.Order)).
		Offset(uint64((normalizePage(req.Page, 1) - 1) * normalizePageSize(req.PageSize, 18))).
		Limit(uint64(normalizePageSize(req.PageSize, 18)))
	return buildCardIDPlan("c_movie_album_item", countBuilder, listBuilder)
}

func buildCardIDPlan(source string, countBuilder, listBuilder squirrel.SelectBuilder) (*cardIDQueryPlan, error) {
	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	listSQL, listArgs, err := listBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	return &cardIDQueryPlan{
		source:   source,
		countSQL: countSQL,
		countArg: countArgs,
		listSQL:  listSQL,
		listArg:  listArgs,
	}, nil
}

func applyWhere(builder squirrel.SelectBuilder, where squirrel.And) squirrel.SelectBuilder {
	if len(where) == 0 {
		return builder
	}
	return builder.Where(where)
}

func canUseAMovieIDFastPath(req *types.CardsListRequest) bool {
	if req.View != "" && req.View != "cards" && req.View != "cardstoday" {
		return false
	}
	if hasM2MCardFilters(req) || hasMinfoCardFilters(req) || hasWMediaCardFilters(req) || hasGScCardFilters(req) {
		return false
	}
	if req.AlbumName != "" || req.NeedDownload != 0 {
		return false
	}
	switch req.OrderBy {
	case "", consts.OrderByReleasingDate, consts.OrderByDetailUpdateTime, consts.OrderByCastAgeAsc, consts.OrderByCastAgeDesc, consts.OrderByViewerWatched:
		return true
	default:
		return false
	}
}

func canUseMinfoIDFastPath(req *types.CardsListRequest) bool {
	if req.View != "cardshasrank" {
		return false
	}
	if hasM2MCardFilters(req) || hasAMovieOnlyCardFilters(req) || hasWMediaCardFilters(req) || hasGScCardFilters(req) {
		return false
	}
	if req.AlbumName != "" || req.NeedDownload != 0 {
		return false
	}
	switch req.OrderBy {
	case "", consts.OrderByRankDate, consts.OrderByHighestRank, consts.OrderByDaysInRank, consts.OrderByReleasingDate:
		return true
	default:
		return false
	}
}

func canUseWMediaIDFastPath(req *types.CardsListRequest) bool {
	if req.View != "cardsmediamowned" {
		return false
	}
	if hasM2MCardFilters(req) || hasAMovieOnlyCardFilters(req) || hasMinfoCardFilters(req) || hasGScCardFilters(req) {
		return false
	}
	if req.AlbumName != "" || req.NeedDownload != 0 {
		return false
	}
	switch req.OrderBy {
	case "", consts.OrderByMediaBirthTime, consts.OrderByBirthTime, consts.OrderByReleasingDate:
		return true
	default:
		return false
	}
}

func canUseAlbumIDFastPath(req *types.CardsListRequest) bool {
	if req.View != "cardsneeddownload" && strings.TrimSpace(req.AlbumName) == "" && req.NeedDownload != consts.MovieNeedDownloadOK {
		return false
	}
	if req.NeedDownload == consts.MovieNeedDownloadNone {
		return false
	}
	if strings.ContainsAny(strings.TrimSpace(req.AlbumName), ",|;") {
		return false
	}
	if hasM2MCardFilters(req) || hasAMovieOnlyCardFilters(req) || hasMinfoCardFilters(req) || hasWMediaCardFilters(req) || hasGScCardFilters(req) {
		return false
	}
	switch req.OrderBy {
	case "", consts.OrderByReleasingDate:
		return true
	default:
		return false
	}
}

func hasM2MCardFilters(req *types.CardsListRequest) bool {
	return req.CastNames != "" || req.PersonIds != "" || req.GenreNames != ""
}

func hasAMovieOnlyCardFilters(req *types.CardsListRequest) bool {
	return req.DirectorName != "" ||
		req.PrefixName != "" ||
		req.MakerName != "" ||
		req.LabelName != "" ||
		req.LabelJavID != "" ||
		req.CastAgeMin != nil ||
		req.CastAgeMax != nil ||
		req.ScoreMin != nil ||
		req.ScoreMax != nil ||
		req.ViewWatchedMin != nil ||
		req.ViewWatchedMax != nil
}

func hasMinfoCardFilters(req *types.CardsListRequest) bool {
	return req.StartRankingDateStart != "" ||
		req.StartRankingDateEnd != "" ||
		req.DaysInRankMin != nil ||
		req.Word != ""
}

func hasWMediaCardFilters(req *types.CardsListRequest) bool {
	return req.MediaOwned != 0 ||
		req.MediaBirthTimeStart != "" ||
		req.MediaBirthTimeEnd != "" ||
		req.MediaDir1 != "" ||
		req.MediaDir2 != "" ||
		req.MediaDir3 != "" ||
		req.MediaDir4 != ""
}

func hasGScCardFilters(req *types.CardsListRequest) bool {
	return req.LastScTimeMin != "" ||
		req.LastScTimeMax != "" ||
		req.ScTimesMin != nil ||
		req.ScTimesMax != nil ||
		req.ComeTimesMin != nil ||
		req.ComeTimesMax != nil
}

func cardAMovieOrderClause(orderBy, order string) string {
	dir := cardOrderDirection(order)
	tie := "name DESC, jav_id DESC"
	switch orderBy {
	case consts.OrderByDetailUpdateTime:
		return "detail_update_time " + dir + ", " + tie
	case consts.OrderByCastAgeAsc:
		return "COALESCE(NULLIF(cast_average_age, 0), 999999) ASC, " + tie
	case consts.OrderByCastAgeDesc:
		return "cast_average_age DESC, " + tie
	case consts.OrderByViewerWatched:
		return "viewers_number_watched " + dir + ", " + tie
	default:
		return "releasing_date " + dir + ", " + tie
	}
}

func cardMinfoOrderClause(orderBy, order string) string {
	dir := cardOrderDirection(order)
	tie := "jav_id DESC"
	switch orderBy {
	case consts.OrderByHighestRank:
		if dir == "ASC" {
			return "COALESCE(NULLIF(highest_rank, 0), 999999) ASC, " + tie
		}
		return "highest_rank DESC, " + tie
	case consts.OrderByDaysInRank:
		return "days_in_rank " + dir + ", " + tie
	case consts.OrderByReleasingDate:
		return "releasing_date " + dir + ", " + tie
	default:
		return "first_rank_day_number " + dir + ", " + tie
	}
}

func cardWMediaOrderClause(orderBy, order string) string {
	dir := cardOrderDirection(order)
	tie := "MAX(movie_name) DESC, movie_jav_id DESC"
	if orderBy == consts.OrderByReleasingDate {
		return "MAX(releasing_date) " + dir + ", " + tie
	}
	return "MAX(birth_time) " + dir + ", " + tie
}

func cardAlbumOrderClause(order string) string {
	return "MAX(cai.releasing_date) " + cardOrderDirection(order) + ", MAX(cai.movie_name) DESC, cai.movie_jav_id DESC"
}

func cardOrderDirection(order string) string {
	if strings.EqualFold(order, "asc") {
		return "ASC"
	}
	return "DESC"
}

func appendWMediaOwnedFilters(where *squirrel.And, mediaOwned int64) {
	switch mediaOwned {
	case consts.OwnedAll:
	case consts.OwnedAllNotRemoved:
		*where = append(*where, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved})
	case consts.OwnedHasSubNotRemoved:
		*where = append(*where, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved}, squirrel.Eq{"has_sub": consts.FilmHasSub})
	case consts.OwnedNoSubNotRemoved:
		*where = append(*where, squirrel.Eq{"is_removed": consts.FilmIsNotRemoved}, squirrel.Eq{"has_sub": consts.FilmNoSub})
	case consts.OwnedRemoved:
		*where = append(*where, squirrel.Eq{"is_removed": consts.FilmIsRemoved})
	}
}

func appendMediaDirFilters(where *squirrel.And, alias string, req *types.CardsListRequest) {
	column := "full_dir"
	if alias != "" {
		column = alias + ".full_dir"
	}
	if req.MediaDir1 != "" {
		*where = append(*where, squirrel.Expr("CONCAT('/', "+column+", '/') LIKE ?", "%/"+req.MediaDir1+"/%"))
	}
	if req.MediaDir2 != "" {
		*where = append(*where, squirrel.Expr("CONCAT('/', "+column+", '/') LIKE ?", "%/"+req.MediaDir2+"/%"))
	}
	if req.MediaDir3 != "" {
		*where = append(*where, squirrel.Expr("CONCAT('/', "+column+", '/') LIKE ?", "%/"+req.MediaDir3+"/%"))
	}
	if req.MediaDir4 != "" {
		*where = append(*where, squirrel.Expr("CONCAT('/', "+column+", '/') LIKE ?", "%/"+req.MediaDir4+"/%"))
	}
}

func (r *MovieReadRepo) LoadMovieCardsByIDs(ctx context.Context, ids []string) ([]*types.MovieCard, error) {
	if len(ids) == 0 {
		return []*types.MovieCard{}, nil
	}

	sqlText := `
SELECT
  m.jav_id AS movie_jav_id,
  m.name AS movie_name,
  m.title AS title,
  m.encode_name AS encode_name,
  m.releasing_date AS releasing_date,
  m.detail_update_time AS detail_update_time,
  m.score AS score_raw,
  m.viewers_number_watched AS viewers_number_watched,
  COALESCE(px.name, '') AS prefix_name,
  COALESCE(mk.name, '') AS maker_name,
  COALESCE(lb.name, '') AS label_name,
  COALESCE(dr.name, '') AS director_name,
  COALESCE(mi.chinese, '') AS chinese_title,
  COALESCE(mu.jacket_img, '') AS jacket_img,
  COALESCE(mu.jacket_img_local, '') AS jacket_img_local,
  COALESCE(mi.highest_rank, 0) AS highest_rank,
  COALESCE(mi.days_in_rank, 0) AS days_in_rank,
  COALESCE(mi.first_rank_day_number, 0) AS first_rank_day_number,
  COALESCE(gs.sc_times, 0) AS sc_times,
  COALESCE(gs.come_times, 0) AS come_times,
  COALESCE(gs.last_sc_time, 0) AS last_sc_time,
  CASE
    WHEN EXISTS (
      SELECT 1
      FROM c_movie_album_item cai
      JOIN c_movie_album ca ON ca.id = cai.album_id
      WHERE cai.movie_jav_id = m.jav_id AND ca.name = ?
    ) THEN 2 ELSE 1
  END AS need_download,
  COALESCE(wm.id, 0) AS w_media_id,
  COALESCE(wm.birth_time, 0) AS w_media_birth_time,
  COALESCE(wm.has_sub, 0) AS w_media_has_sub,
  COALESCE(wm.is_removed, 0) AS w_media_is_removed,
  COALESCE(wm.full_dir, '') AS w_media_full_dir,
  COALESCE(wm.file_name, '') AS w_media_file_name,
  COALESCE(wm.source_torrent_hash, '') AS w_media_source_hash,
  COALESCE(wm.size, 0) AS w_media_size,
  COALESCE(wm.height, 0) AS w_media_height,
  COALESCE(wm.bit_rate, 0) AS w_media_bit_rate,
  COALESCE(wm.duration, 0) AS w_media_duration,
  COALESCE(wm.frame_average, 0) AS w_media_frame_average,
  COALESCE(wm.self_make, 0) AS w_media_self_make
FROM a_movie m
LEFT JOIN bm_minfo mi ON mi.jav_id = m.jav_id
LEFT JOIN bm_murl mu ON mu.jav_id = m.jav_id
LEFT JOIN am_prefix px ON px.id = m.prefix_id
LEFT JOIN am_maker mk ON mk.id = m.maker_id
LEFT JOIN am_label lb ON lb.id = m.label_id
LEFT JOIN am_director dr ON dr.id = m.director_id
LEFT JOIN g_sc_stat gs ON gs.movie_jav_id = m.jav_id
LEFT JOIN w_media wm ON wm.movie_jav_id = m.jav_id AND wm.source_type = ?
WHERE m.jav_id IN (?)
`
	query := strings.Replace(sqlText, "(?)", "("+buildPlaceholders(len(ids))+")", 1)
	args := make([]any, 0, len(ids)+2)
	args = append(args, consts.MovieNeedDownloadAlbumName)
	args = append(args, consts.WMediaSourceNative)
	for _, id := range ids {
		args = append(args, id)
	}

	var rows []*cardBaseRow
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	cardMap := make(map[string]*types.MovieCard, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		cardMap[row.MovieJavID] = r.mapCardBaseRow(row)
	}
	if err := r.attachMovieGenres(ctx, cardMap); err != nil {
		return nil, err
	}
	if err := r.attachMovieCasts(ctx, cardMap); err != nil {
		return nil, err
	}

	items := make([]*types.MovieCard, 0, len(ids))
	for _, id := range ids {
		if card, ok := cardMap[id]; ok {
			items = append(items, card)
		}
	}
	return items, nil
}

func (r *MovieReadRepo) FindMovieCardByName(ctx context.Context, movieName string) (*types.MovieCard, error) {
	sqlText := `SELECT jav_id FROM a_movie WHERE name = ? LIMIT 1`
	var javID string
	if err := r.conn.QueryRowCtx(ctx, &javID, sqlText, movieName); err != nil {
		return nil, err
	}
	cards, err := r.LoadMovieCardsByIDs(ctx, []string{javID})
	if err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return nil, ErrNotFound
	}
	return cards[0], nil
}

func (r *MovieReadRepo) FindEarliestRankDayNumber(ctx context.Context) (int64, error) {
	var dayNumber int64
	if err := r.conn.QueryRowCtx(ctx, &dayNumber, `SELECT COALESCE(MIN(day_number), 0) FROM c_rank`); err != nil {
		return 0, err
	}
	return dayNumber, nil
}

func (r *MovieReadRepo) FindLatestRankDayNumber(ctx context.Context) (int64, error) {
	var dayNumber int64
	if err := r.conn.QueryRowCtx(ctx, &dayNumber, `SELECT COALESCE(MAX(day_number), 0) FROM c_rank`); err != nil {
		return 0, err
	}
	return dayNumber, nil
}

func (r *MovieReadRepo) ListRankDayMovieIDs(ctx context.Context, dayNumber, page, pageSize int64) ([]string, int64, error) {
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, `SELECT COUNT(*) FROM c_rank WHERE day_number = ?`, dayNumber); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	offset := (normalizePage(page, 1) - 1) * normalizePageSize(pageSize, 18)
	query := `SELECT movie_jav_id FROM c_rank WHERE day_number = ? ORDER BY rank_pos ASC LIMIT ? OFFSET ?`
	var ids []string
	if err := r.conn.QueryRowsCtx(ctx, &ids, query, dayNumber, normalizePageSize(pageSize, 18), offset); err != nil {
		return nil, 0, err
	}
	return ids, total, nil
}

func (r *MovieReadRepo) LoadMovieRanks(ctx context.Context, movieJavID string) ([]*types.MovieRankInfo, error) {
	query := `SELECT day_number, rank_pos FROM c_rank WHERE movie_jav_id = ? ORDER BY day_number DESC`
	var rows []*movieRankRow
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, movieJavID); err != nil {
		return nil, err
	}
	out := make([]*types.MovieRankInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		date := consts.GetDateStringByRankDayNumber(row.DayNumber)
		out = append(out, &types.MovieRankInfo{
			Date: date,
			Rank: row.RankPos,
			Href: "/moviecarddayrank?date=" + url.QueryEscape(date),
		})
	}
	return out, nil
}

func (r *MovieReadRepo) LoadMovieScEvents(ctx context.Context, movieName string) ([]*types.MovieScEvent, error) {
	query := `SELECT name, sc_time, come_movie_name, cooldown FROM g_sc WHERE movie_cast = ? OR come_movie_name = ? ORDER BY sc_time DESC LIMIT 32`
	var rows []*movieScRow
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, movieName, movieName); err != nil {
		return nil, err
	}
	out := make([]*types.MovieScEvent, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, &types.MovieScEvent{
			ScName:   row.Name,
			ScTime:   row.ScTime,
			Cooldown: row.Cooldown,
			IsCome:   row.ComeMovieName == movieName,
			Href:     "/sc-events/" + url.PathEscape(row.Name),
		})
	}
	return out, nil
}

func (r *MovieReadRepo) LoadRankPeriod(ctx context.Context, periodType, category int64, periodKey string) (*CRankPeriod, error) {
	query := `SELECT * FROM c_rank_period WHERE period_type = ? AND category = ?`
	args := []any{periodType, category}
	if strings.TrimSpace(periodKey) != "" {
		query += ` AND period_key = ?`
		args = append(args, periodKey)
	} else {
		query += ` ORDER BY start_day_number DESC LIMIT 1`
	}

	var row CRankPeriod
	if err := r.conn.QueryRowCtx(ctx, &row, query, args...); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *MovieReadRepo) LoadPrevRankPeriod(ctx context.Context, periodType, category, startDayNumber int64) (*CRankPeriod, error) {
	query := `SELECT * FROM c_rank_period WHERE period_type = ? AND category = ? AND start_day_number < ? ORDER BY start_day_number DESC LIMIT 1`
	var row CRankPeriod
	if err := r.conn.QueryRowCtx(ctx, &row, query, periodType, category, startDayNumber); err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *MovieReadRepo) LoadNextRankPeriod(ctx context.Context, periodType, category, startDayNumber int64) (*CRankPeriod, error) {
	query := `SELECT * FROM c_rank_period WHERE period_type = ? AND category = ? AND start_day_number > ? ORDER BY start_day_number ASC LIMIT 1`
	var row CRankPeriod
	if err := r.conn.QueryRowCtx(ctx, &row, query, periodType, category, startDayNumber); err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *MovieReadRepo) LoadRankPeriodItems(ctx context.Context, periodID, page, pageSize int64) ([]*CRankPeriodItem, int64, error) {
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, `SELECT COUNT(*) FROM c_rank_period_item WHERE period_id = ?`, periodID); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}
	offset := (normalizePage(page, 1) - 1) * normalizePageSize(pageSize, 18)
	query := `SELECT * FROM c_rank_period_item WHERE period_id = ? ORDER BY rank_pos ASC LIMIT ? OFFSET ?`
	var rows []*CRankPeriodItem
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, periodID, normalizePageSize(pageSize, 18), offset); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *MovieReadRepo) LoadMovieMedia(ctx context.Context, movieJavID string) (*types.MovieMedia, error) {
	query := `SELECT movie_jav_id, movie_name, file_name, full_dir, source_torrent_hash, birth_time, size, height, bit_rate, duration, frame_average, self_make, is_removed FROM w_media WHERE movie_jav_id = ? AND source_type = ? LIMIT 1`
	type mediaRow struct {
		MovieJavID        string  `db:"movie_jav_id"`
		MovieName         string  `db:"movie_name"`
		FileName          string  `db:"file_name"`
		FullDir           string  `db:"full_dir"`
		SourceTorrentHash string  `db:"source_torrent_hash"`
		BirthTime         int64   `db:"birth_time"`
		Size              int64   `db:"size"`
		Height            int64   `db:"height"`
		BitRate           int64   `db:"bit_rate"`
		Duration          int64   `db:"duration"`
		FrameAverage      float64 `db:"frame_average"`
		SelfMake          int64   `db:"self_make"`
		IsRemoved         int64   `db:"is_removed"`
	}
	var row mediaRow
	if err := r.conn.QueryRowCtx(ctx, &row, query, movieJavID, consts.WMediaSourceNative); err != nil {
		return nil, err
	}
	return &types.MovieMedia{
		MovieJavID:        row.MovieJavID,
		MovieName:         row.MovieName,
		FileName:          row.FileName,
		Directory:         row.FullDir,
		FilePath:          filepath.Join(row.FullDir, row.FileName),
		SourceTorrentHash: row.SourceTorrentHash,
		BirthTime:         formatDate(row.BirthTime),
		Size:              round1(float64(row.Size) / 1e9),
		Height:            row.Height,
		BitRate:           round1(float64(row.BitRate)),
		DurationMinutes:   round1(float64(row.Duration) / 60.0),
		Frame:             int64(math.Round(row.FrameAverage)),
		SelfMake:          row.SelfMake,
		IsRemoved:         row.IsRemoved,
	}, nil
}

func (r *MovieReadRepo) LoadJavbusFetch(ctx context.Context, movieJavID string) (*types.MovieFetchSiteStatus, error) {
	row, err := r.TJavbusMagnetFetchRow(ctx, movieJavID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return mapFetchStatus(row.MovieJavId, row.MovieName, row.ReleaseDate, row.FetchStatus, row.TryCount, row.LastFetchTime, row.LastError, row.TorrentHashCount, row.LatestPublishTime, row.SourceUrl), nil
}

func (r *MovieReadRepo) LoadSukebeiFetch(ctx context.Context, movieJavID string) (*types.MovieFetchSiteStatus, error) {
	row, err := r.TSukebeiFetchRow(ctx, movieJavID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return mapFetchStatus(row.MovieJavId, row.MovieName, row.ReleaseDate, row.FetchStatus, row.TryCount, row.LastFetchTime, row.LastError, row.TorrentHashCount, row.LatestPublishTime, row.SourceUrl), nil
}

func (r *MovieReadRepo) TJavbusMagnetRowList(ctx context.Context, movieJavID string) ([]*types.MovieJavbusMagnet, error) {
	type row struct {
		Id          int64  `db:"id"`
		MagnetName  string `db:"magnet_name"`
		InfoHash    string `db:"info_hash"`
		SizeBytes   int64  `db:"size_bytes"`
		ShareDate   int64  `db:"share_date"`
		HasHd       int64  `db:"has_hd"`
		HasSubtitle int64  `db:"has_subtitle"`
	}
	var rows []*row
	if err := r.conn.QueryRowsCtx(ctx, &rows, `SELECT id, magnet_name, info_hash, size_bytes, share_date, has_hd, has_subtitle FROM t_javbus_magnet WHERE movie_jav_id = ? ORDER BY share_date DESC`, movieJavID); err != nil {
		return nil, err
	}
	out := make([]*types.MovieJavbusMagnet, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		out = append(out, &types.MovieJavbusMagnet{
			RowID:       item.Id,
			MagnetName:  item.MagnetName,
			InfoHash:    item.InfoHash,
			SizeText:    formatResourceSizeBytes(item.SizeBytes),
			ShareDate:   formatDate(item.ShareDate),
			HasHD:       item.HasHd == 1,
			HasSubtitle: item.HasSubtitle == 1,
		})
	}
	return out, nil
}

func (r *MovieReadRepo) TSukebeiTorrentRowList(ctx context.Context, movieJavID string) ([]*types.MovieSukebeiTorrent, error) {
	type row struct {
		Id           int64  `db:"id"`
		TorrentTitle string `db:"torrent_title"`
		ViewId       int64  `db:"view_id"`
		InfoHash     string `db:"info_hash"`
		SizeBytes    int64  `db:"size_bytes"`
		PublishTime  int64  `db:"publish_time"`
		Seeders      int64  `db:"seeders"`
		Leechers     int64  `db:"leechers"`
		Completed    int64  `db:"completed"`
	}
	var rows []*row
	if err := r.conn.QueryRowsCtx(ctx, &rows, `SELECT id, torrent_title, view_id, info_hash, size_bytes, publish_time, seeders, leechers, completed FROM t_sukebei_torrent WHERE movie_jav_id = ? ORDER BY publish_time DESC`, movieJavID); err != nil {
		return nil, err
	}
	out := make([]*types.MovieSukebeiTorrent, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		out = append(out, &types.MovieSukebeiTorrent{
			RowID:        item.Id,
			TorrentTitle: item.TorrentTitle,
			ViewURL:      buildSukebeiViewURL(item.ViewId),
			InfoHash:     item.InfoHash,
			SizeText:     formatResourceSizeBytes(item.SizeBytes),
			PublishTime:  formatDate(item.PublishTime),
			Seeders:      item.Seeders,
			Leechers:     item.Leechers,
			Completed:    item.Completed,
		})
	}
	return out, nil
}

func (r *MovieReadRepo) TSehuatangMagnetRowList(ctx context.Context, movieJavID, movieName string) ([]*types.MovieSehuatangMagnet, error) {
	type row struct {
		Id          int64  `db:"id"`
		ThreadTitle string `db:"thread_title"`
		ThreadUrl   string `db:"thread_url"`
		InfoHash    string `db:"info_hash"`
		PostTime    int64  `db:"post_time"`
	}
	var rows []*row
	query := `SELECT id, thread_title, thread_url, info_hash, post_time FROM t_sehuatang_magnet WHERE movie_jav_id = ? OR movie_name = ? ORDER BY post_time DESC`
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, movieJavID, movieName); err != nil {
		return nil, err
	}
	out := make([]*types.MovieSehuatangMagnet, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		out = append(out, &types.MovieSehuatangMagnet{
			RowID:       item.Id,
			ThreadTitle: item.ThreadTitle,
			ThreadURL:   item.ThreadUrl,
			InfoHash:    item.InfoHash,
			PostTime:    formatDateTime(item.PostTime),
		})
	}
	return out, nil
}

func (r *MovieReadRepo) baseCardQuery(req *types.CardsListRequest) squirrel.SelectBuilder {
	q := squirrel.Select().
		From("a_movie m").
		LeftJoin("bm_minfo mi ON mi.jav_id = m.jav_id").
		LeftJoin(fmt.Sprintf("w_media wm ON wm.movie_jav_id = m.jav_id AND wm.source_type = %d", consts.WMediaSourceNative)).
		LeftJoin("g_sc_stat gs ON gs.movie_jav_id = m.jav_id").
		LeftJoin("am_label lb ON lb.id = m.label_id").
		LeftJoin("am_maker mk ON mk.id = m.maker_id").
		LeftJoin("am_director dr ON dr.id = m.director_id").
		LeftJoin("am_prefix px ON px.id = m.prefix_id")

	for _, filter := range r.cardFilters(req) {
		q = q.Where(filter)
	}
	return q
}

func (r *MovieReadRepo) cardFilters(req *types.CardsListRequest) []squirrel.Sqlizer {
	filters := make([]squirrel.Sqlizer, 0, 24)

	if req.View == "" || req.View == "cards" {
	} else if req.View == "cardstoday" {
		filters = append(filters, squirrel.LtOrEq{"m.releasing_date": startOfDay(time.Now())})
	} else if req.View == "cardshasrank" {
		filters = append(filters, squirrel.Gt{"mi.days_in_rank": 0})
	} else if req.View == "cardsmediamowned" {
		filters = append(filters, squirrel.NotEq{"wm.id": nil})
		filters = append(filters, squirrel.Eq{"wm.is_removed": 1})
	} else if req.View == "cardsneeddownload" {
		filters = append(filters, squirrel.Expr("EXISTS (SELECT 1 FROM c_movie_album_item cai JOIN c_movie_album ca ON ca.id = cai.album_id WHERE cai.movie_jav_id = m.jav_id AND ca.name = ?)", consts.MovieNeedDownloadAlbumName))
	}

	if req.PrefixName != "" {
		filters = append(filters, squirrel.Eq{"px.name": req.PrefixName})
	}
	if req.MakerName != "" {
		filters = append(filters, squirrel.Eq{"mk.name": req.MakerName})
	}
	if req.LabelName != "" {
		filters = append(filters, squirrel.Eq{"lb.name": req.LabelName})
	}
	if req.LabelJavID != "" {
		filters = append(filters, squirrel.Eq{"lb.jav_id": req.LabelJavID})
	}
	if req.DirectorName != "" {
		filters = append(filters, squirrel.Eq{"dr.name": req.DirectorName})
	}
	if req.Word != "" {
		filters = append(filters, squirrel.Like{"mi.chinese": "%" + req.Word + "%"})
	}
	if req.CastAgeMin != nil {
		filters = append(filters, squirrel.GtOrEq{"m.cast_average_age": int64(*req.CastAgeMin*10.0 + 0.5)})
	}
	if req.CastAgeMax != nil {
		filters = append(filters, squirrel.LtOrEq{"m.cast_average_age": int64(*req.CastAgeMax*10.0 + 0.5)})
	}
	if req.ReleasingDateStart != "" {
		if ts, ok := parseDate(req.ReleasingDateStart); ok {
			filters = append(filters, squirrel.GtOrEq{"m.releasing_date": ts})
		}
	}
	if req.ReleasingDateEnd != "" {
		if ts, ok := parseDate(req.ReleasingDateEnd); ok {
			filters = append(filters, squirrel.LtOrEq{"m.releasing_date": ts})
		}
	}
	if req.MediaBirthTimeStart != "" {
		if ts, ok := parseDate(req.MediaBirthTimeStart); ok {
			filters = append(filters, squirrel.GtOrEq{"wm.birth_time": ts})
		}
	}
	if req.MediaBirthTimeEnd != "" {
		if ts, ok := parseDate(req.MediaBirthTimeEnd); ok {
			filters = append(filters, squirrel.LtOrEq{"wm.birth_time": ts})
		}
	}
	if req.StartRankingDateStart != "" {
		filters = append(filters, squirrel.GtOrEq{"mi.first_rank_day_number": consts.GetRankDayNumber(req.StartRankingDateStart)})
	}
	if req.StartRankingDateEnd != "" {
		filters = append(filters, squirrel.LtOrEq{"mi.first_rank_day_number": consts.GetRankDayNumber(req.StartRankingDateEnd)})
	}
	if req.DaysInRankMin != nil {
		filters = append(filters, squirrel.GtOrEq{"mi.days_in_rank": *req.DaysInRankMin})
	}
	if req.ViewWatchedMin != nil {
		filters = append(filters, squirrel.GtOrEq{"m.viewers_number_watched": *req.ViewWatchedMin})
	}
	if req.ViewWatchedMax != nil {
		filters = append(filters, squirrel.LtOrEq{"m.viewers_number_watched": *req.ViewWatchedMax})
	}
	if req.ScoreMin != nil {
		filters = append(filters, squirrel.GtOrEq{"m.score": int64(*req.ScoreMin*10.0 + 0.5)})
	}
	if req.ScoreMax != nil {
		filters = append(filters, squirrel.LtOrEq{"m.score": int64(*req.ScoreMax*10.0 + 0.5)})
	}
	if req.ScTimesMin != nil {
		filters = append(filters, squirrel.GtOrEq{"gs.sc_times": *req.ScTimesMin})
	}
	if req.ScTimesMax != nil {
		filters = append(filters, squirrel.LtOrEq{"gs.sc_times": *req.ScTimesMax})
	}
	if req.ComeTimesMin != nil {
		filters = append(filters, squirrel.GtOrEq{"gs.come_times": *req.ComeTimesMin})
	}
	if req.ComeTimesMax != nil {
		filters = append(filters, squirrel.LtOrEq{"gs.come_times": *req.ComeTimesMax})
	}
	if req.LastScTimeMin != "" {
		if ts, ok := parseDate(req.LastScTimeMin); ok {
			filters = append(filters, squirrel.GtOrEq{"gs.last_sc_time": ts})
		}
	}
	if req.LastScTimeMax != "" {
		if ts, ok := parseDate(req.LastScTimeMax); ok {
			filters = append(filters, squirrel.LtOrEq{"gs.last_sc_time": ts})
		}
	}
	if req.MediaDir1 != "" {
		filters = append(filters, squirrel.Expr("CONCAT('/', wm.full_dir, '/') LIKE ?", "%/"+req.MediaDir1+"/%"))
	}
	if req.MediaDir2 != "" {
		filters = append(filters, squirrel.Expr("CONCAT('/', wm.full_dir, '/') LIKE ?", "%/"+req.MediaDir2+"/%"))
	}
	if req.MediaDir3 != "" {
		filters = append(filters, squirrel.Expr("CONCAT('/', wm.full_dir, '/') LIKE ?", "%/"+req.MediaDir3+"/%"))
	}
	if req.MediaDir4 != "" {
		filters = append(filters, squirrel.Expr("CONCAT('/', wm.full_dir, '/') LIKE ?", "%/"+req.MediaDir4+"/%"))
	}

	switch req.MediaOwned {
	case consts.OwnedAll:
		filters = append(filters, squirrel.NotEq{"wm.id": nil})
	case consts.OwnedAllNotRemoved:
		filters = append(filters, squirrel.NotEq{"wm.id": nil}, squirrel.Eq{"wm.is_removed": 1})
	case consts.OwnedHasSubNotRemoved:
		filters = append(filters, squirrel.NotEq{"wm.id": nil}, squirrel.Eq{"wm.is_removed": 1}, squirrel.Eq{"wm.has_sub": 2})
	case consts.OwnedNoSubNotRemoved:
		filters = append(filters, squirrel.NotEq{"wm.id": nil}, squirrel.Eq{"wm.is_removed": 1}, squirrel.Eq{"wm.has_sub": 1})
	case consts.OwnedRemoved:
		filters = append(filters, squirrel.NotEq{"wm.id": nil}, squirrel.Eq{"wm.is_removed": 2})
	case consts.OwnedNotOwned:
		filters = append(filters, squirrel.Expr("wm.id IS NULL"))
	}

	if req.NeedDownload == consts.MovieNeedDownloadOK {
		filters = append(filters, squirrel.Expr("EXISTS (SELECT 1 FROM c_movie_album_item cai JOIN c_movie_album ca ON ca.id = cai.album_id WHERE cai.movie_jav_id = m.jav_id AND ca.name = ?)", consts.MovieNeedDownloadAlbumName))
	}
	if req.NeedDownload == consts.MovieNeedDownloadNone {
		filters = append(filters, squirrel.Expr("NOT EXISTS (SELECT 1 FROM c_movie_album_item cai JOIN c_movie_album ca ON ca.id = cai.album_id WHERE cai.movie_jav_id = m.jav_id AND ca.name = ?)", consts.MovieNeedDownloadAlbumName))
	}

	if req.AlbumName != "" {
		filters = append(filters, squirrel.Expr("EXISTS (SELECT 1 FROM c_movie_album_item cai JOIN c_movie_album ca ON ca.id = cai.album_id WHERE cai.movie_jav_id = m.jav_id AND ca.name = ?)", req.AlbumName))
	}

	if castNames := parseList(req.CastNames); len(castNames) > 0 {
		filters = append(filters, squirrel.Expr(`
EXISTS (
  SELECT 1
  FROM amr_movie_cast rc
  JOIN am_cast c ON c.id = rc.cast_id
  WHERE rc.movie_jav_id = m.jav_id AND c.name = ?
)`, castNames[0]))
	}
	if personIDs := parseInt64List(req.PersonIds); len(personIDs) > 0 {
		filters = append(filters, squirrel.Expr(`
EXISTS (
  SELECT 1
  FROM amr_movie_cast rc
  JOIN am_cast c ON c.id = rc.cast_id
  WHERE rc.movie_jav_id = m.jav_id AND c.person_id = ?
)`, personIDs[0]))
	}
	if genreNames := parseList(req.GenreNames); len(genreNames) > 0 {
		filters = append(filters, squirrel.Expr(`
EXISTS (
  SELECT 1
  FROM amr_movie_genre rg
  JOIN am_genre g ON g.id = rg.genre_id
  WHERE rg.movie_jav_id = m.jav_id AND g.name = ?
)`, genreNames[0]))
	}
	return filters
}

func (r *MovieReadRepo) attachMovieGenres(ctx context.Context, cardMap map[string]*types.MovieCard) error {
	if len(cardMap) == 0 {
		return nil
	}
	ids := mapKeys(cardMap)
	query := `
SELECT rg.movie_jav_id AS movie_jav_id, g.name AS genre_name
FROM amr_movie_genre rg
JOIN am_genre g ON g.id = rg.genre_id
WHERE rg.movie_jav_id IN (?)
ORDER BY g.name ASC`
	query = strings.Replace(query, "(?)", "("+buildPlaceholders(len(ids))+")", 1)
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	type row struct {
		MovieJavID string `db:"movie_jav_id"`
		GenreName  string `db:"genre_name"`
	}
	var rows []*row
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return err
	}
	for _, item := range rows {
		if item == nil {
			continue
		}
		card := cardMap[item.MovieJavID]
		if card == nil {
			continue
		}
		card.Genre = append(card.Genre, item.GenreName)
	}
	return nil
}

func (r *MovieReadRepo) attachMovieCasts(ctx context.Context, cardMap map[string]*types.MovieCard) error {
	if len(cardMap) == 0 {
		return nil
	}
	ids := mapKeys(cardMap)
	query := `
SELECT
  rc.movie_jav_id AS movie_jav_id,
  c.person_id AS person_id,
  c.name AS cast_name,
  p.name AS person_name,
  p.chinese AS person_chinese,
  p.birth_day AS birth_day
FROM amr_movie_cast rc
JOIN am_cast c ON c.id = rc.cast_id
LEFT JOIN c_person p ON p.id = c.person_id
WHERE rc.movie_jav_id IN (?)
ORDER BY rc.movie_jav_id ASC, c.person_id ASC, c.name ASC`
	query = strings.Replace(query, "(?)", "("+buildPlaceholders(len(ids))+")", 1)
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	type row struct {
		MovieJavID    string `db:"movie_jav_id"`
		PersonID      int64  `db:"person_id"`
		CastName      string `db:"cast_name"`
		PersonName    string `db:"person_name"`
		PersonChinese string `db:"person_chinese"`
		BirthDay      int64  `db:"birth_day"`
	}
	var rows []*row
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		card := cardMap[item.MovieJavID]
		if card == nil || len(card.Cast) >= 9 {
			continue
		}
		displayName := firstNonEmpty(item.PersonChinese, item.PersonName, item.CastName)
		nameShow := displayName
		key := item.MovieJavID + ":" + strconv.FormatInt(item.PersonID, 10) + ":" + item.CastName
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		if item.BirthDay > 0 {
			nameShow = fmt.Sprintf("%s(%.1f)", displayName, round1(float64(cardReleaseTS(card.MovieName, card.ReleasingDate)-item.BirthDay)/(3600*24*365)))
		}
		card.Cast = append(card.Cast, &types.CastTag{
			PersonID:    item.PersonID,
			Name:        item.CastName,
			DisplayName: displayName,
			NameShow:    nameShow,
			Href:        buildCastHref(item.PersonID, item.CastName),
		})
	}
	return nil
}

func (r *MovieReadRepo) mapCardBaseRow(row *cardBaseRow) *types.MovieCard {
	title := firstNonEmpty(strings.TrimSpace(row.ChineseTitle), strings.TrimSpace(row.Title), row.MovieName)
	card := &types.MovieCard{
		MovieName:            row.MovieName,
		MovieJavID:           row.MovieJavID,
		MovieHref:            "/movie/" + url.PathEscape(row.MovieName),
		Title:                title,
		JacketImg:            firstNonEmpty(buildLocalJacketPath(r.cfg.Fetcher.LocalImageDir, row.JacketImgLocal), row.JacketImg),
		ComeTimes:            row.ComeTimes,
		ScTimes:              row.ScTimes,
		HighestRank:          row.HighestRank,
		Score:                round1(float64(row.ScoreRaw) / 10.0),
		ViewersNumberWatched: row.ViewersNumberWatched,
		Director:             row.DirectorName,
		DirectorHref:         "/cards?dn=" + url.QueryEscape(row.DirectorName),
		Prefix:               row.PrefixName,
		PrefixHref:           "/cards?pn=" + url.QueryEscape(row.PrefixName),
		BusUrl:               buildBusURL(r.cfg.Fetcher.BusAddress, row.MovieName),
		SearchUrl:            buildSukebeiSearchURL(row.MovieName),
		JavUrl:               buildJavURL(r.cfg.Fetcher.JavAddress, row.MovieJavID),
		ReleasingDate:        formatDate(row.ReleasingDate),
		Label:                row.LabelName,
		LabelHref:            "/cards?ln=" + url.QueryEscape(row.LabelName),
		Maker:                row.MakerName,
		MakerHref:            "/cards?mn=" + url.QueryEscape(row.MakerName),
		NeedDownload:         row.NeedDownload,
	}
	if row.FirstRankDayNumber > 0 && row.DaysInRank > 0 {
		card.FirstRankingDay = fmt.Sprintf("%s(%d)", consts.GetDateStringByRankDayNumber(row.FirstRankDayNumber), row.DaysInRank)
	}
	if row.WMediaID > 0 {
		card.OwnedWMedia = deriveOwnedWMedia(row.WMediaHasSub, row.WMediaIsRemoved)
		card.VideoUrlWMedia = filepath.Join(row.WMediaFullDir, row.WMediaFileName)
		card.FilmBirthDateWMedia = formatDate(row.WMediaBirthTime)
	}
	return card
}

func deriveOwnedWMedia(hasSub, isRemoved int64) int64 {
	if isRemoved == 2 {
		return consts.OwnedRemoved
	}
	if hasSub == 2 {
		return consts.OwnedHasSubNotRemoved
	}
	return consts.OwnedNoSubNotRemoved
}

func cardGroupedOrderClause(orderBy, order string) string {
	dir := "DESC"
	if strings.EqualFold(order, "asc") {
		dir = "ASC"
	}
	nameTieBreaker := "MAX(m.name) DESC"
	switch orderBy {
	case consts.OrderByCastAgeAsc:
		if dir == "DESC" {
			return "MAX(m.cast_average_age) DESC, " + nameTieBreaker
		}
		return "COALESCE(MIN(NULLIF(m.cast_average_age, 0)), 999999) ASC, " + nameTieBreaker
	case consts.OrderByCastAgeDesc:
		if dir == "ASC" {
			return "COALESCE(MIN(NULLIF(m.cast_average_age, 0)), 999999) ASC, " + nameTieBreaker
		}
		return "MAX(m.cast_average_age) DESC, " + nameTieBreaker
	case consts.OrderByViewerWatched:
		return "MAX(m.viewers_number_watched) " + dir + ", " + nameTieBreaker
	case consts.OrderByRankDate:
		return "MAX(mi.first_rank_day_number) " + dir + ", " + nameTieBreaker
	case consts.OrderByHighestRank:
		if dir == "ASC" {
			return "COALESCE(MIN(NULLIF(mi.highest_rank, 0)), 999999) ASC, " + nameTieBreaker
		}
		return "MAX(COALESCE(mi.highest_rank, 0)) DESC, " + nameTieBreaker
	case consts.OrderByDaysInRank:
		return "MAX(mi.days_in_rank) " + dir + ", " + nameTieBreaker
	case consts.OrderByBirthTime, consts.OrderByMediaBirthTime:
		return "MAX(wm.birth_time) " + dir + ", " + nameTieBreaker
	case consts.OrderByScTimes:
		return "MAX(gs.sc_times) " + dir + ", " + nameTieBreaker
	case consts.OrderByComeTimes:
		return "MAX(gs.come_times) " + dir + ", " + nameTieBreaker
	case consts.OrderByLastScTime:
		return "MAX(gs.last_sc_time) " + dir + ", " + nameTieBreaker
	case consts.OrderByDetailUpdateTime:
		return "MAX(m.detail_update_time) " + dir + ", " + nameTieBreaker
	default:
		return "MAX(m.releasing_date) " + dir + ", " + nameTieBreaker
	}
}

func parseDate(raw string) (int64, bool) {
	parsed, err := time.ParseInLocation(time.DateOnly, raw, time.Local)
	if err != nil {
		return 0, false
	}
	return parsed.Unix(), true
}

func formatDate(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(time.DateOnly)
}

func formatDateTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func startOfDay(now time.Time) int64 {
	value := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return value.Unix()
}

func normalizePage(page, fallback int64) int64 {
	if page <= 0 {
		return fallback
	}
	return page
}

func normalizePageSize(pageSize, fallback int64) int64 {
	if pageSize <= 0 {
		return fallback
	}
	return pageSize
}

func parseList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '|' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, item := range parts {
		value := strings.TrimSpace(item)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func parseInt64List(raw string) []int64 {
	parts := parseList(raw)
	out := make([]int64, 0, len(parts))
	for _, item := range parts {
		value, err := strconv.ParseInt(item, 10, 64)
		if err == nil {
			out = append(out, value)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func buildLocalJacketPath(rootDir, localPath string) string {
	if strings.TrimSpace(rootDir) == "" || strings.TrimSpace(localPath) == "" {
		return ""
	}
	return filepath.Join(rootDir, localPath)
}

func buildBusURL(address, movieName string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		address = "www.javbus.com"
	}
	return "https://" + strings.TrimRight(address, "/") + "/" + strings.TrimLeft(movieName, "/")
}

func buildJavURL(address, movieJavID string) string {
	address = strings.TrimSpace(address)
	if address == "" {
		address = "www.c97k.com"
	}
	return "https://" + strings.TrimRight(address, "/") + "/cn/?v=" + url.QueryEscape(movieJavID)
}

func buildSukebeiSearchURL(movieName string) string {
	parts := strings.Split(movieName, "-")
	if len(parts) != 2 {
		return ""
	}
	return "https://sukebei.nyaa.si/?f=0&c=0_0&q=" + url.QueryEscape(parts[0]+" "+parts[1])
}

func buildSukebeiViewURL(viewID int64) string {
	if viewID <= 0 {
		return ""
	}
	return fmt.Sprintf("https://sukebei.nyaa.si/view/%d", viewID)
}

func buildCastHref(personID int64, castName string) string {
	if personID > 0 {
		return "/cast/" + strconv.FormatInt(personID, 10)
	}
	return "/cards?cn=" + url.QueryEscape(castName)
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func cardReleaseTS(_ string, releasingDate string) int64 {
	if strings.TrimSpace(releasingDate) == "" {
		return 0
	}
	parsed, err := time.ParseInLocation(time.DateOnly, releasingDate, time.Local)
	if err != nil {
		return 0
	}
	return parsed.Unix()
}

func formatResourceSizeBytes(sizeBytes int64) string {
	if sizeBytes <= 0 {
		return "-"
	}
	type unitDef struct {
		label string
		value float64
	}
	units := []unitDef{
		{label: "TB", value: 1024 * 1024 * 1024 * 1024},
		{label: "GB", value: 1024 * 1024 * 1024},
		{label: "MB", value: 1024 * 1024},
		{label: "KB", value: 1024},
	}
	size := float64(sizeBytes)
	for _, unit := range units {
		if size >= unit.value {
			return fmt.Sprintf("%.2f %s", size/unit.value, unit.label)
		}
	}
	return fmt.Sprintf("%d B", sizeBytes)
}

func isRandomView(view string) bool {
	return strings.EqualFold(strings.TrimSpace(view), "cardsrandom")
}

func fetchStatusText(status int64) string {
	switch status {
	case 1:
		return "待抓取"
	case 2:
		return "抓取中"
	case 3:
		return "成功"
	case 4:
		return "失败"
	default:
		return "未入队"
	}
}

func mapFetchStatus(movieJavID, movieName string, releaseDate, fetchStatus, tryCount, lastFetchTime int64, lastError string, torrentHashCount, latestPublishTime int64, sourceURL string) *types.MovieFetchSiteStatus {
	return &types.MovieFetchSiteStatus{
		MovieJavID:        movieJavID,
		MovieName:         movieName,
		ReleaseDate:       formatDate(releaseDate),
		FetchStatus:       fetchStatus,
		FetchStatusText:   fetchStatusText(fetchStatus),
		TryCount:          tryCount,
		LastFetchTime:     formatDateTime(lastFetchTime),
		LastFetchAgo:      daysAgo(lastFetchTime),
		LastError:         strings.TrimSpace(lastError),
		TorrentHashCount:  torrentHashCount,
		LatestPublishTime: formatDate(latestPublishTime),
		SourceURL:         sourceURL,
	}
}

func daysAgo(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f 天", time.Since(time.Unix(ts, 0)).Hours()/24)
}

func (r *MovieReadRepo) TJavbusMagnetFetchRow(ctx context.Context, movieJavID string) (*TJavbusMagnetFetch, error) {
	var row TJavbusMagnetFetch
	if err := r.conn.QueryRowCtx(ctx, &row, `SELECT * FROM t_javbus_magnet_fetch WHERE movie_jav_id = ? LIMIT 1`, movieJavID); err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *MovieReadRepo) TSukebeiFetchRow(ctx context.Context, movieJavID string) (*TSukebeiTorrentFetch, error) {
	var row TSukebeiTorrentFetch
	if err := r.conn.QueryRowCtx(ctx, &row, `SELECT * FROM t_sukebei_torrent_fetch WHERE movie_jav_id = ? LIMIT 1`, movieJavID); err != nil {
		if err == sqlc.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func mapKeys[V any](input map[string]V) []string {
	out := make([]string, 0, len(input))
	for key := range input {
		out = append(out, key)
	}
	return out
}

func buildPlaceholders(n int) string {
	if n <= 0 {
		return "NULL"
	}
	holders := make([]string, 0, n)
	for i := 0; i < n; i++ {
		holders = append(holders, "?")
	}
	return strings.Join(holders, ",")
}
