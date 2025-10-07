package modelx

import (
	"context"

	"github.com/Masterminds/squirrel"
)

type FindMovieV1Request struct {
	MovieIds           []int64
	MovieJavIds        []string
	DirectorId         int64
	PrefixId           int64
	LabelId            int64
	MakerId            int64
	OrderBy            string
	Word               string
	Page               int64
	PageSize           int64
	CastAgeMin         float64
	CastAgeMax         float64
	FilmBirthTimeStart int64
	FilmBirthTimeEnd   int64
	ReleasingDateStart int64
	ReleasingDateEnd   int64
	StartRankingDay    int64
	ComeTimesMin       int64
	ScTimesMax         int64
	ScTimesMin         int64
	LastScTimeMin      int64
	Owned              int64
}

func (m *defaultAMovieModel) FindMoviesComplex(ctx context.Context, r *FindMovieV1Request) ([]*AMovie, int64, error) {
	builder := squirrel.Select("*").
		From(m.tableName()).
		Where("1=1")

	// Apply filters
	switch r.Owned {
	case MovieOwned, MovieOwnedAndSub, MovieIsRemoved:
		builder = builder.Where("`movie_owned` = ?", r.Owned)
	case 5:
		builder = builder.Where(squirrel.Eq{"`movie_owned`": []int64{MovieOwned, MovieOwnedAndSub}})
	case 6:
		builder = builder.Where(squirrel.Eq{"`movie_owned`": []int64{MovieOwned, MovieOwnedAndSub, MovieIsRemoved}})
	default:
	}

	if r.FilmBirthTimeStart > 0 {
		builder = builder.Where("`film_birth_time` >= ?", r.FilmBirthTimeStart)
	}
	if r.FilmBirthTimeEnd > 0 {
		builder = builder.Where("`film_birth_time` <= ?", r.FilmBirthTimeEnd)
	}
	if r.ReleasingDateStart > 0 {
		builder = builder.Where("`releasing_date` >= ?", r.ReleasingDateStart)
	}
	if r.ReleasingDateEnd > 0 {
		builder = builder.Where("`releasing_date` <= ?", r.ReleasingDateEnd)
	}
	if len(r.MovieIds) > 0 {
		builder = builder.Where(squirrel.Eq{"id": r.MovieIds})
	}
	if len(r.MovieJavIds) > 0 {
		builder = builder.Where(squirrel.Eq{"jav_id": r.MovieJavIds})
	}
	if r.Word != "" {
		likeQuery := "%" + r.Word + "%"
		builder = builder.Where("(title LIKE ? OR chinese LIKE ?)", likeQuery, likeQuery)
	}
	if r.StartRankingDay > 0 {
		builder = builder.Where("`first_rank_day_number` >= ?", r.StartRankingDay)
	}
	if r.OrderBy == "cast_age asc" || r.OrderBy == "cast_age desc" {
		builder = builder.Where("cast_age > ?", 0)
	}
	if r.CastAgeMax > 0 {
		builder = builder.Where("`cast_age` <= ? AND cast_age > ?", r.CastAgeMax, 0)
	}
	if r.CastAgeMin > 0 {
		builder = builder.Where("`cast_age` >= ? AND cast_age > ?", r.CastAgeMin, 0)
	}
	if r.DirectorId > 0 {
		builder = builder.Where("`director_id` = ?", r.DirectorId)
	}
	if r.PrefixId > 0 {
		builder = builder.Where("`prefix_id` = ?", r.PrefixId)
	}
	if r.LabelId > 0 {
		builder = builder.Where("`label_id` = ?", r.LabelId)
	}
	if r.MakerId > 0 {
		builder = builder.Where("`maker_id` = ?", r.MakerId)
	}
	if r.ScTimesMin > 0 {
		builder = builder.Where("`sc_times` >= ?", r.ScTimesMin)
	}

	builder = builder.Where("`sc_times` <= ?", r.ScTimesMax)

	if r.ComeTimesMin > 0 {
		builder = builder.Where("`come_times` >= ?", r.ComeTimesMin)
	}

	if r.LastScTimeMin > 0 {
		builder = builder.Where("`last_sc_time` >= ?", r.LastScTimeMin)
	}

	return m.executeQuery(ctx, builder, r.Page, r.PageSize, r.OrderBy)
}

func (m *defaultAMovieModel) FindMoviesNeedDownload(ctx context.Context, orderBy string, owned, page, pageSize int64) ([]*AMovie, int64, error) {
	builder := squirrel.Select("*").
		From(m.tableName()).
		Where(squirrel.Eq{"`need_download`": MovieNeedDownload})

	if owned != 0 {
		builder = builder.Where("`film_birth_time` > ?", 0)
	}
	return m.executeQuery(ctx, builder, page, pageSize, orderBy)
}

func (m *defaultAMovieModel) executeQuery(
	ctx context.Context,
	db squirrel.SelectBuilder,
	page, pageSize int64,
	orderBy string,
) ([]*AMovie, int64, error) {
	// Count query
	var total int64
	if page == 0 {
		page = 1
	}
	countQuery, countArgs, err := db.ToSql()
	if err != nil {
		return nil, 0, err
	}
	finalCountQuery := "SELECT COUNT(*) FROM (" + countQuery + ") AS count_query"
	if err := m.QueryRowNoCacheCtx(ctx, &total, finalCountQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	// Result query
	var results []*AMovie
	resultQuery, resultArgs, err := db.
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize)).
		OrderBy(orderBy).
		ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := m.QueryRowsNoCacheCtx(ctx, &results, resultQuery, resultArgs...); err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
