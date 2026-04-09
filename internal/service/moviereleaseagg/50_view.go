package moviereleaseagg

import (
	"context"
	"fmt"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BuildReleaseView(ctx context.Context, p AggParams) (*AggResult, error) {
	mode := NormalizeAggMode(p.Mode)
	level := detectLevel(p.Year, p.Quarter, p.Month, p.Day)
	res := &AggResult{
		AggMode:         mode,
		AggModeLabel:    aggModeLabel(mode),
		CardFilterQuery: cardFilterQuery(mode),
		Breadcrumbs:     buildBreadcrumbs(mode, p.Year, p.Quarter, p.Month, p.Day),
	}

	switch level {
	case levelRoot:
		res.Title = fmt.Sprintf("电影聚合（%s）· 上映日", res.AggModeLabel)
		res.Level = levelRoot
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, mode, levelYear, 0, 0, 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows, mode)
		topScope := buildScope(0, 0, 0, 0)
		if err := s.fillTopStats(ctx, res, mode, topScope.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelYear:
		sc := buildScope(p.Year, 0, 0, 0)
		res.Title = fmt.Sprintf("上映日（%s）· %d 年", res.AggModeLabel, p.Year)
		res.Level = levelYear
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		qRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, mode, levelQuarter, int64(p.Year), 0, 0, true)
		if err != nil {
			return nil, err
		}
		mRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, mode, levelMonth, int64(p.Year), 0, 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsQAll = buildBuckets(qRows, mode)
		res.BucketsMAll = buildBuckets(mRows, mode)
		if err := s.fillTopStats(ctx, res, mode, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelQuarter:
		sc := buildScope(p.Year, p.Quarter, 0, 0)
		res.Title = fmt.Sprintf("上映日（%s）· %d 年 Q%d", res.AggModeLabel, p.Year, p.Quarter)
		res.Level = levelQuarter
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, mode, levelMonth, int64(p.Year), int64(p.Quarter), 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows, mode)
		if err := s.fillTopStats(ctx, res, mode, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelMonth:
		sc := buildScope(p.Year, 0, p.Month, 0)
		res.Title = fmt.Sprintf("上映日（%s）· %d 年 %02d 月", res.AggModeLabel, p.Year, p.Month)
		res.Level = levelMonth
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, mode, levelDay, int64(sc.Year), 0, int64(sc.Month), true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows, mode)
		if err := s.fillTopStats(ctx, res, mode, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	default:
		sc := buildScope(p.Year, 0, p.Month, p.Day)
		res.Title = fmt.Sprintf("上映日（%s）· %04d-%02d-%02d", res.AggModeLabel, p.Year, p.Month, p.Day)
		res.Level = levelDay
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		return res, nil
	}
}

func detectLevel(year, quarter, month, day int) string {
	switch {
	case year == 0:
		return levelRoot
	case day > 0:
		return levelDay
	case quarter == 0 && month == 0:
		return levelYear
	case month == 0:
		return levelQuarter
	default:
		return levelMonth
	}
}

func buildBuckets(rows []*moviex.MovieReleaseBucketStat, mode string) []Bucket {
	out := make([]Bucket, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.CountAll <= 0 {
			continue
		}
		out = append(out, Bucket{
			Label:      bucketLabel(row),
			Href:       bucketHref(row, mode),
			CountAll:   row.CountAll,
			CountOwned: row.CountOwned,
			SizeAllGB:  bytesToGB(row.SizeBytes),
		})
	}
	return out
}

func bucketLabel(row *moviex.MovieReleaseBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelYear:
		return fmt.Sprintf("%d 年", row.Year)
	case levelQuarter:
		return fmt.Sprintf("Q%d 季", row.Quarter)
	case levelMonth:
		return fmt.Sprintf("%02d 月", row.Month)
	case levelDay:
		return fmt.Sprintf("%02d 日", row.Day)
	default:
		return row.ScopeKey
	}
}

func bucketHref(row *moviex.MovieReleaseBucketStat, mode string) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelYear:
		return buildReleaseAggHref(mode, int(row.Year), 0, 0, 0)
	case levelQuarter:
		return buildReleaseAggHref(mode, int(row.Year), int(row.Quarter), 0, 0)
	case levelMonth:
		return buildReleaseAggHref(mode, int(row.Year), 0, int(row.Month), 0)
	case levelDay:
		return buildReleaseAggHref(mode, int(row.Year), 0, int(row.Month), int(row.Day))
	default:
		return buildReleaseAggHref(mode, 0, 0, 0, 0)
	}
}

func (s *Service) fillTopStats(ctx context.Context, res *AggResult, aggMode, scopeKey string) error {
	casts, err := s.deps.MovieReleaseTopStatModel.ListByAggModeScopeAggType(ctx, aggMode, scopeKey, aggTypeCast)
	if err != nil {
		return err
	}
	res.TopCastsAll = toTopStats(casts)

	directors, err := s.deps.MovieReleaseTopStatModel.ListByAggModeScopeAggType(ctx, aggMode, scopeKey, aggTypeDirector)
	if err != nil {
		return err
	}
	res.TopDirectorsAll = toTopStats(directors)

	labels, err := s.deps.MovieReleaseTopStatModel.ListByAggModeScopeAggType(ctx, aggMode, scopeKey, aggTypeLabel)
	if err != nil {
		return err
	}
	res.TopLabelsAll = toTopStats(labels)

	prefixes, err := s.deps.MovieReleaseTopStatModel.ListByAggModeScopeAggType(ctx, aggMode, scopeKey, aggTypePrefix)
	if err != nil {
		return err
	}
	res.TopPrefixesAll = toTopStats(prefixes)
	return nil
}

func toTopStats(rows []*moviex.MovieReleaseTopStat) []TopStat {
	out := make([]TopStat, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.AggName == "" {
			continue
		}
		out = append(out, TopStat{
			PersonId:   row.AggId,
			Name:       row.AggName,
			CountAll:   row.CountAll,
			CountOwned: row.CountOwned,
		})
	}
	return out
}

func cardFilterQuery(mode string) string {
	if NormalizeAggMode(mode) == AggModeOwned {
		return "mowned=3"
	}
	return "owned=1"
}
