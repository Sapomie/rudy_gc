package moviereleaseagg

import (
	"context"
	"fmt"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BuildReleaseView(ctx context.Context, p AggParams) (*AggResult, error) {
	level := detectLevel(p.Year, p.Quarter, p.Month, p.Day)
	res := &AggResult{
		Breadcrumbs: buildBreadcrumbs(p.Year, p.Quarter, p.Month, p.Day),
	}

	switch level {
	case levelRoot:
		res.Title = "电影聚合（全量）· 上映日"
		res.Level = levelRoot
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, levelYear, 0, 0, 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows)
		topScope := buildScope(0, 0, 0, 0)
		if err := s.fillTopStats(ctx, res, topScope.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelYear:
		sc := buildScope(p.Year, 0, 0, 0)
		res.Title = fmt.Sprintf("上映日（全量）· %d 年", p.Year)
		res.Level = levelYear
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		qRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, levelQuarter, int64(p.Year), 0, 0, true)
		if err != nil {
			return nil, err
		}
		mRows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, levelMonth, int64(p.Year), 0, 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsQAll = buildBuckets(qRows)
		res.BucketsMAll = buildBuckets(mRows)
		if err := s.fillTopStats(ctx, res, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelQuarter:
		sc := buildScope(p.Year, p.Quarter, 0, 0)
		res.Title = fmt.Sprintf("上映日（全量）· %d 年 Q%d", p.Year, p.Quarter)
		res.Level = levelQuarter
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, levelMonth, int64(p.Year), int64(p.Quarter), 0, true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows)
		if err := s.fillTopStats(ctx, res, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	case levelMonth:
		sc := buildScope(p.Year, 0, p.Month, 0)
		res.Title = fmt.Sprintf("上映日（全量）· %d 年 %02d 月", p.Year, p.Month)
		res.Level = levelMonth
		res.RangeStart = dateOnly(sc.StartUnix)
		res.RangeEnd = dateOnly(sc.EndUnix)
		rows, err := s.deps.MovieReleaseBucketStatModel.ListByLevel(ctx, levelDay, int64(sc.Year), 0, int64(sc.Month), true)
		if err != nil {
			return nil, err
		}
		res.BucketsAll = buildBuckets(rows)
		if err := s.fillTopStats(ctx, res, sc.Key); err != nil {
			return nil, err
		}
		return res, nil
	default:
		sc := buildScope(p.Year, 0, p.Month, p.Day)
		res.Title = fmt.Sprintf("上映日（全量）· %04d-%02d-%02d", p.Year, p.Month, p.Day)
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

func buildBuckets(rows []*moviex.MovieReleaseBucketStat) []Bucket {
	out := make([]Bucket, 0, len(rows))
	for _, row := range rows {
		if row == nil || row.CountAll <= 0 {
			continue
		}
		out = append(out, Bucket{
			Label:      bucketLabel(row),
			Href:       bucketHref(row),
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

func bucketHref(row *moviex.MovieReleaseBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelYear:
		return fmt.Sprintf("%s/%d", rootPath, row.Year)
	case levelQuarter:
		return fmt.Sprintf("%s/%d/q/%d", rootPath, row.Year, row.Quarter)
	case levelMonth:
		return fmt.Sprintf("%s/%d/%02d", rootPath, row.Year, row.Month)
	case levelDay:
		return fmt.Sprintf("%s/%d/%02d/%02d", rootPath, row.Year, row.Month, row.Day)
	default:
		return rootPath
	}
}

func (s *Service) fillTopStats(ctx context.Context, res *AggResult, scopeKey string) error {
	casts, err := s.deps.MovieReleaseTopStatModel.ListByScopeAggType(ctx, scopeKey, aggTypeCast)
	if err != nil {
		return err
	}
	res.TopCastsAll = toTopStats(casts)

	directors, err := s.deps.MovieReleaseTopStatModel.ListByScopeAggType(ctx, scopeKey, aggTypeDirector)
	if err != nil {
		return err
	}
	res.TopDirectorsAll = toTopStats(directors)

	labels, err := s.deps.MovieReleaseTopStatModel.ListByScopeAggType(ctx, scopeKey, aggTypeLabel)
	if err != nil {
		return err
	}
	res.TopLabelsAll = toTopStats(labels)

	prefixes, err := s.deps.MovieReleaseTopStatModel.ListByScopeAggType(ctx, scopeKey, aggTypePrefix)
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
