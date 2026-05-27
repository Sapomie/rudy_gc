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
	Length               int64   `db:"length"`
	ReleasingDate        int64   `db:"releasing_date"`
	DetailUpdateTime     int64   `db:"detail_update_time"`
	ScoreRaw             int64   `db:"score_raw"`
	ViewersNumberWant    int64   `db:"viewers_number_want"`
	ViewersNumberOwned   int64   `db:"viewers_number_owned"`
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

type movieBaseRow struct {
	MovieJavID           string `db:"movie_jav_id"`
	MovieName            string `db:"movie_name"`
	Title                string `db:"title"`
	EncodeName           string `db:"encode_name"`
	Length               int64  `db:"length"`
	ReleasingDate        int64  `db:"releasing_date"`
	DetailUpdateTime     int64  `db:"detail_update_time"`
	ScoreRaw             int64  `db:"score_raw"`
	ViewersNumberWant    int64  `db:"viewers_number_want"`
	ViewersNumberOwned   int64  `db:"viewers_number_owned"`
	ViewersNumberWatched int64  `db:"viewers_number_watched"`
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
	return r.listMovieIDsLegacy(ctx, req)
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
	return r.listMovieIDsLegacy(ctx, req)
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

	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From("w_media"), where)
	listBuilder := applyWhere(squirrel.Select("movie_jav_id").From("w_media"), where).
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
	countBuilder := applyWhere(squirrel.Select("COUNT(*)").From(from).Join(join), where)
	listBuilder := applyWhere(squirrel.Select("cai.movie_jav_id").From(from).Join(join), where).
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
	tie := "name DESC"
	switch orderBy {
	case consts.OrderByDetailUpdateTime:
		return "detail_update_time " + dir + ", " + tie
	case consts.OrderByCastAgeAsc:
		return "cast_average_age ASC, " + tie
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
	tie := "name DESC"
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
	if orderBy == consts.OrderByReleasingDate {
		return "releasing_date " + dir + ", movie_name DESC"
	}
	return "birth_time " + dir + ", movie_name DESC"
}

func cardAlbumOrderClause(order string) string {
	return "releasing_date " + cardOrderDirection(order) + ", movie_name DESC, movie_jav_id DESC"
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

	cardMap := make(map[string]*types.MovieCard, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		card, err := r.loadMovieCardByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if card != nil {
			cardMap[id] = card
		}
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

func (r *MovieReadRepo) loadMovieCardByID(ctx context.Context, javID string) (*types.MovieCard, error) {
	row := &cardBaseRow{}
	var base movieBaseRow
	if err := r.conn.QueryRowCtx(ctx, &base, `
SELECT
  jav_id AS movie_jav_id,
  name AS movie_name,
  title AS title,
  encode_name AS encode_name,
  length AS length,
  releasing_date AS releasing_date,
  detail_update_time AS detail_update_time,
  score AS score_raw,
  viewers_number_want AS viewers_number_want,
  viewers_number_owned AS viewers_number_owned,
  viewers_number_watched AS viewers_number_watched
FROM a_movie
WHERE jav_id = ?
LIMIT 1`, javID); err != nil {
		return nil, err
	}
	row.MovieJavID = base.MovieJavID
	row.MovieName = base.MovieName
	row.Title = base.Title
	row.EncodeName = base.EncodeName
	row.Length = base.Length
	row.ReleasingDate = base.ReleasingDate
	row.DetailUpdateTime = base.DetailUpdateTime
	row.ScoreRaw = base.ScoreRaw
	row.ViewersNumberWant = base.ViewersNumberWant
	row.ViewersNumberOwned = base.ViewersNumberOwned
	row.ViewersNumberWatched = base.ViewersNumberWatched

	_ = r.conn.QueryRowCtx(ctx, &row.ChineseTitle, `SELECT chinese FROM bm_minfo WHERE jav_id = ? LIMIT 1`, javID)
	type minfoRow struct {
		HighestRank        int64 `db:"highest_rank"`
		DaysInRank         int64 `db:"days_in_rank"`
		FirstRankDayNumber int64 `db:"first_rank_day_number"`
	}
	var mi minfoRow
	if err := r.conn.QueryRowCtx(ctx, &mi, `SELECT highest_rank, days_in_rank, first_rank_day_number FROM bm_minfo WHERE jav_id = ? LIMIT 1`, javID); err == nil {
		row.HighestRank = mi.HighestRank
		row.DaysInRank = mi.DaysInRank
		row.FirstRankDayNumber = mi.FirstRankDayNumber
	}
	_ = r.conn.QueryRowCtx(ctx, &row.JacketImg, `SELECT jacket_img FROM bm_murl WHERE jav_id = ? LIMIT 1`, javID)
	_ = r.conn.QueryRowCtx(ctx, &row.JacketImgLocal, `SELECT jacket_img_local FROM bm_murl WHERE jav_id = ? LIMIT 1`, javID)
	_ = r.conn.QueryRowCtx(ctx, &row.PrefixName, `SELECT px.name FROM am_prefix px JOIN a_movie m ON m.prefix_id = px.id WHERE m.jav_id = ? LIMIT 1`, javID)
	_ = r.conn.QueryRowCtx(ctx, &row.MakerName, `SELECT mk.name FROM am_maker mk JOIN a_movie m ON m.maker_id = mk.id WHERE m.jav_id = ? LIMIT 1`, javID)
	_ = r.conn.QueryRowCtx(ctx, &row.LabelName, `SELECT lb.name FROM am_label lb JOIN a_movie m ON m.label_id = lb.id WHERE m.jav_id = ? LIMIT 1`, javID)
	_ = r.conn.QueryRowCtx(ctx, &row.DirectorName, `SELECT dr.name FROM am_director dr JOIN a_movie m ON m.director_id = dr.id WHERE m.jav_id = ? LIMIT 1`, javID)

	type statRow struct {
		ScTimes    int64 `db:"sc_times"`
		ComeTimes  int64 `db:"come_times"`
		LastScTime int64 `db:"last_sc_time"`
	}
	var stat statRow
	if err := r.conn.QueryRowCtx(ctx, &stat, `SELECT sc_times, come_times, last_sc_time FROM g_sc_stat WHERE movie_jav_id = ? LIMIT 1`, javID); err == nil {
		row.ScTimes = stat.ScTimes
		row.ComeTimes = stat.ComeTimes
		row.LastScTime = stat.LastScTime
	}

	var needDownload int64
	if err := r.conn.QueryRowCtx(ctx, &needDownload, `
SELECT 1
FROM c_movie_album_item cai
JOIN c_movie_album ca ON ca.id = cai.album_id
WHERE ca.name = ? AND cai.movie_jav_id = ?
LIMIT 1`, consts.MovieNeedDownloadAlbumName, javID); err == nil && needDownload == 1 {
		row.NeedDownload = consts.MovieNeedDownloadOK
	} else {
		row.NeedDownload = consts.MovieNeedDownloadNone
	}

	type mediaRow struct {
		ID        int64  `db:"id"`
		BirthTime int64  `db:"birth_time"`
		HasSub    int64  `db:"has_sub"`
		IsRemoved int64  `db:"is_removed"`
		FullDir   string `db:"full_dir"`
		FileName  string `db:"file_name"`
	}
	var media mediaRow
	if err := r.conn.QueryRowCtx(ctx, &media, `
SELECT id, birth_time, has_sub, is_removed, full_dir, file_name
FROM w_media
WHERE movie_jav_id = ? AND source_type = ?
LIMIT 1`, javID, consts.WMediaSourceNative); err == nil {
		row.WMediaID = media.ID
		row.WMediaBirthTime = media.BirthTime
		row.WMediaHasSub = media.HasSub
		row.WMediaIsRemoved = media.IsRemoved
		row.WMediaFullDir = media.FullDir
		row.WMediaFileName = media.FileName
	}
	return r.mapCardBaseRow(row), nil
}

func (r *MovieReadRepo) FindMovieCardByName(ctx context.Context, movieName string) (*types.MovieCard, error) {
	javID, err := r.FindMovieJavIDByName(ctx, movieName)
	if err != nil {
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

func (r *MovieReadRepo) FindMovieJavIDByName(ctx context.Context, movieName string) (string, error) {
	sqlText := `SELECT jav_id FROM a_movie WHERE name = ? LIMIT 1`
	var javID string
	if err := r.conn.QueryRowCtx(ctx, &javID, sqlText, movieName); err != nil {
		return "", err
	}
	return javID, nil
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

func (r *MovieReadRepo) LoadRandomRankPeriod(ctx context.Context, periodType, category int64, nextInt63 func() int64) (*CRankPeriod, error) {
	var total int64
	if err := r.conn.QueryRowCtx(ctx, &total, `SELECT COUNT(*) FROM c_rank_period WHERE period_type = ? AND category = ?`, periodType, category); err != nil {
		return nil, err
	}
	if total <= 0 {
		return nil, nil
	}
	var offset int64
	if nextInt63 != nil && total > 1 {
		offset = nextInt63() % total
		if offset < 0 {
			offset = -offset
		}
	}
	query := `SELECT * FROM c_rank_period FORCE INDEX (idx_period_range) WHERE period_type = ? AND category = ? ORDER BY start_day_number DESC LIMIT 1 OFFSET ?`
	var row CRankPeriod
	if err := r.conn.QueryRowCtx(ctx, &row, query, periodType, category, offset); err != nil {
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
		Frame:             round1(row.FrameAverage),
		SelfMake:          formatSelfMake(row.SelfMake),
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

func (r *MovieReadRepo) LoadMovieAlbums(ctx context.Context, movieJavID string) ([]*types.MovieAlbumOption, error) {
	type row struct {
		ID       int64  `db:"id"`
		Name     string `db:"name"`
		Selected int64  `db:"selected"`
	}
	var rows []*row
	query := `
SELECT
  ca.id AS id,
  ca.name AS name,
  CASE WHEN cai.id IS NULL THEN 0 ELSE 1 END AS selected
FROM c_movie_album ca
LEFT JOIN c_movie_album_item cai ON cai.album_id = ca.id AND cai.movie_jav_id = ?
ORDER BY ca.id ASC`
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, movieJavID); err != nil {
		return nil, err
	}
	out := make([]*types.MovieAlbumOption, 0, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		out = append(out, &types.MovieAlbumOption{
			ID:       item.ID,
			Name:     name,
			Selected: item.Selected == 1,
		})
	}
	return out, nil
}

func (r *MovieReadRepo) IsMovieMarkedDelete(ctx context.Context, movieJavID string) (bool, error) {
	var exists int64
	query := `
SELECT 1
FROM c_movie_album_item cai
JOIN c_movie_album ca ON ca.id = cai.album_id
WHERE ca.name = ? AND cai.movie_jav_id = ?
LIMIT 1`
	if err := r.conn.QueryRowCtx(ctx, &exists, query, consts.MovieDeleteAlbumName, movieJavID); err != nil {
		if err == sqlc.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

func (r *MovieReadRepo) ApplyFetchResourceStatus(ctx context.Context, movieJavID string, javbus []*types.MovieJavbusMagnet, sukebei []*types.MovieSukebeiTorrent, sehuatang []*types.MovieSehuatangMagnet) error {
	applyMatchedFetchSiteHashes(javbus, sukebei, sehuatang)
	if len(javbus) == 0 && len(sukebei) == 0 && len(sehuatang) == 0 {
		return nil
	}

	type row struct {
		SourceType  string `db:"source_type"`
		SourceRowID int64  `db:"source_row_id"`
		AlbumName   string `db:"album_name"`
	}
	var rows []*row
	query := `
SELECT i.source_type AS source_type, i.source_row_id AS source_row_id, a.name AS album_name
FROM tm_album_item i
JOIN t_album a ON a.id = i.album_id
WHERE i.movie_jav_id = ? AND a.name IN (?, ?)`
	if err := r.conn.QueryRowsCtx(ctx, &rows, query, movieJavID, "下载中", "待下载"); err != nil {
		return err
	}
	status := make(map[string]map[string]bool, len(rows))
	for _, item := range rows {
		if item == nil {
			continue
		}
		key := fetchResourceKey(item.SourceType, item.SourceRowID)
		if key == "" {
			continue
		}
		if status[key] == nil {
			status[key] = map[string]bool{}
		}
		status[key][strings.TrimSpace(item.AlbumName)] = true
	}
	for _, row := range javbus {
		if row == nil {
			continue
		}
		fillFetchResourceAlbumStatus(status[fetchResourceKey("javbus_magnet", row.RowID)], &row.IsInDownload, &row.IsInPending)
	}
	for _, row := range sukebei {
		if row == nil {
			continue
		}
		fillFetchResourceAlbumStatus(status[fetchResourceKey("sukebei_torrent", row.RowID)], &row.IsInDownload, &row.IsInPending)
	}
	for _, row := range sehuatang {
		if row == nil {
			continue
		}
		fillFetchResourceAlbumStatus(status[fetchResourceKey("sehuatang_magnet", row.RowID)], &row.IsInDownload, &row.IsInPending)
	}
	return nil
}

func (r *MovieReadRepo) baseCardQuery(req *types.CardsListRequest) squirrel.SelectBuilder {
	return applyWhere(squirrel.Select().From("a_movie"), r.cardAMovieWhere(req))
}

func (r *MovieReadRepo) cardFilters(req *types.CardsListRequest) []squirrel.Sqlizer {
	return nil
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
ORDER BY rg.genre_id ASC`
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
		EncodeName:           row.EncodeName,
		JacketImg:            firstNonEmpty(buildLocalJacketPath(r.cfg.Fetcher.LocalImageDir, row.JacketImgLocal), row.JacketImg),
		ComeTimes:            row.ComeTimes,
		ScTimes:              row.ScTimes,
		HighestRank:          row.HighestRank,
		LastScTime:           row.LastScTime,
		Length:               row.Length,
		Score:                round1(float64(row.ScoreRaw) / 10.0),
		ViewersNumberWant:    row.ViewersNumberWant,
		ViewersNumberOwned:   row.ViewersNumberOwned,
		ViewersNumberWatched: row.ViewersNumberWatched,
		Director:             row.DirectorName,
		DirectorHref:         "/cards?dn=" + url.QueryEscape(row.DirectorName),
		Prefix:               row.PrefixName,
		PrefixHref:           "/cards?pn=" + url.QueryEscape(row.PrefixName),
		BusUrl:               buildBusURL(r.cfg.Fetcher.BusAddress, row.MovieName),
		SearchUrl:            buildSukebeiSearchURL(row.MovieName),
		JavUrl:               buildJavURL(r.cfg.Fetcher.JavAddress, row.MovieJavID),
		ReleasingDate:        formatDate(row.ReleasingDate),
		UpdateDate:           formatDate(row.DetailUpdateTime),
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
	return cardAMovieOrderClause(orderBy, order)
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

func formatSelfMake(value int64) string {
	if value > 0 {
		return "YES"
	}
	return "NO"
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

func applyMatchedFetchSiteHashes(javbus []*types.MovieJavbusMagnet, sukebei []*types.MovieSukebeiTorrent, sehuatang []*types.MovieSehuatangMagnet) {
	javbusHashes := make(map[string]struct{}, len(javbus))
	sukebeiHashes := make(map[string]struct{}, len(sukebei))
	sehuatangHashes := make(map[string]struct{}, len(sehuatang))
	for _, row := range javbus {
		if row == nil {
			continue
		}
		if hash := normalizeInfoHash(row.InfoHash); hash != "" {
			javbusHashes[hash] = struct{}{}
		}
	}
	for _, row := range sukebei {
		if row == nil {
			continue
		}
		if hash := normalizeInfoHash(row.InfoHash); hash != "" {
			sukebeiHashes[hash] = struct{}{}
		}
	}
	for _, row := range sehuatang {
		if row == nil {
			continue
		}
		if hash := normalizeInfoHash(row.InfoHash); hash != "" {
			sehuatangHashes[hash] = struct{}{}
		}
	}
	for _, row := range javbus {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, inSukebei := sukebeiHashes[hash]
		_, inSehuatang := sehuatangHashes[hash]
		row.HasMatchedHash = inSukebei || inSehuatang
	}
	for _, row := range sukebei {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, inJavbus := javbusHashes[hash]
		_, inSehuatang := sehuatangHashes[hash]
		row.HasMatchedHash = inJavbus || inSehuatang
	}
	for _, row := range sehuatang {
		if row == nil {
			continue
		}
		hash := normalizeInfoHash(row.InfoHash)
		_, inJavbus := javbusHashes[hash]
		_, inSukebei := sukebeiHashes[hash]
		row.HasMatchedHash = inJavbus || inSukebei
	}
}

func normalizeInfoHash(hash string) string {
	return strings.ToUpper(strings.TrimSpace(hash))
}

func fetchResourceKey(sourceType string, rowID int64) string {
	sourceType = strings.TrimSpace(sourceType)
	if sourceType == "" || rowID <= 0 {
		return ""
	}
	return sourceType + ":" + strconv.FormatInt(rowID, 10)
}

func fillFetchResourceAlbumStatus(status map[string]bool, inDownload, inPending *bool) {
	if status == nil {
		return
	}
	*inDownload = status["下载中"]
	*inPending = status["待下载"]
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
