package wmediaagg

import (
	"context"
	"fmt"

	"rudy_gc/internal/model/modelx/moviex"
)

func (s *Service) BuildBirthView(ctx context.Context, p AggParams) (*AggResult, error) {
	if p.TopN <= 0 {
		p.TopN = 30
	}

	sc := buildScope(p.Year, p.Quarter, p.Month, p.Day)
	res := &AggResult{
		Title:       buildTitle(sc),
		Breadcrumbs: buildBreadcrumbs(sc),
		Level:       sc.Level,
		RangeStart:  sc.StartDate,
		RangeEnd:    sc.EndDate,
	}
	if sc.Level == levelRoot {
		res.RangeStart = ""
		res.RangeEnd = ""
	}

	var err error
	switch sc.Level {
	case levelRoot:
		res.BucketsY, err = s.loadBuckets(ctx, levelYear, 0, 0, 0)
	case levelYear:
		res.BucketsQ, err = s.loadBuckets(ctx, levelQuarter, int64(sc.Year), 0, 0)
		if err == nil {
			res.BucketsM, err = s.loadBuckets(ctx, levelMonth, int64(sc.Year), 0, 0)
		}
	case levelQuarter:
		res.BucketsM, err = s.loadBuckets(ctx, levelMonth, int64(sc.Year), int64(sc.Quarter), 0)
	case levelMonth:
		res.BucketsD, err = s.loadBuckets(ctx, levelDay, int64(sc.Year), 0, int64(sc.Month))
	}
	if err != nil {
		return nil, err
	}

	if sc.Level != levelDay {
		if res.TopCasts, err = s.loadTop(ctx, sc.Key, aggTypeCast, p.TopN); err != nil {
			return nil, err
		}
		if res.TopDirectors, err = s.loadTop(ctx, sc.Key, aggTypeDirector, p.TopN); err != nil {
			return nil, err
		}
		if res.TopLabels, err = s.loadTop(ctx, sc.Key, aggTypeLabel, p.TopN); err != nil {
			return nil, err
		}
		if res.TopPrefixes, err = s.loadTop(ctx, sc.Key, aggTypePrefix, p.TopN); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *Service) loadBuckets(ctx context.Context, level string, year, quarter, month int64) ([]Bucket, error) {
	rows, err := s.deps.WMediaBirthBucketStatModel.ListByLevel(ctx, level, year, quarter, month, true)
	if err != nil {
		return nil, err
	}
	out := make([]Bucket, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		label, href := bucketLabelHref(level, row)
		out = append(out, Bucket{
			Label:  label,
			Href:   href,
			Count:  row.MediaCount,
			SizeGB: bytesToGB(row.SizeBytes),
		})
	}
	return out, nil
}

func (s *Service) loadTop(ctx context.Context, scopeKey, aggType string, topN int) ([]TopStat, error) {
	rows, err := s.deps.WMediaBirthTopStatModel.ListByScopeAggType(ctx, scopeKey, aggType)
	if err != nil {
		return nil, err
	}
	if topN > 0 && len(rows) > topN {
		rows = rows[:topN]
	}
	out := make([]TopStat, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, TopStat{
			AggID:     row.AggId,
			Name:      row.AggName,
			Count:     row.MediaCount,
			SizeBytes: row.SizeBytes,
		})
	}
	return out, nil
}

func buildTitle(sc scope) string {
	switch sc.Level {
	case levelYear:
		return fmt.Sprintf("Media下载日 · %d 年", sc.Year)
	case levelQuarter:
		return fmt.Sprintf("Media下载日 · %d 年 Q%d", sc.Year, sc.Quarter)
	case levelMonth:
		return fmt.Sprintf("Media下载日 · %d 年 %02d 月", sc.Year, sc.Month)
	case levelDay:
		return fmt.Sprintf("Media下载日 · %04d-%02d-%02d", sc.Year, sc.Month, sc.Day)
	default:
		return "Media聚合 · 下载日"
	}
}

func buildBreadcrumbs(sc scope) []Breadcrumb {
	bcs := []Breadcrumb{{Title: "Media下载日", Href: rootPath}}
	if sc.Level == levelRoot {
		bcs[0].Href = ""
		return bcs
	}
	yearHref := fmt.Sprintf("%s/%d", rootPath, sc.Year)
	bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("%d 年", sc.Year), Href: yearHref})
	switch sc.Level {
	case levelYear:
		bcs[len(bcs)-1].Href = ""
	case levelQuarter:
		bcs = append(bcs, Breadcrumb{Title: fmt.Sprintf("Q%d 季", sc.Quarter)})
	case levelMonth:
		quarterHref := fmt.Sprintf("%s/%d/q/%d", rootPath, sc.Year, sc.Quarter)
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", sc.Quarter), Href: quarterHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 月", sc.Month)},
		)
	case levelDay:
		quarterHref := fmt.Sprintf("%s/%d/q/%d", rootPath, sc.Year, sc.Quarter)
		monthHref := fmt.Sprintf("%s/%d/%d", rootPath, sc.Year, sc.Month)
		bcs = append(bcs,
			Breadcrumb{Title: fmt.Sprintf("Q%d 季", sc.Quarter), Href: quarterHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 月", sc.Month), Href: monthHref},
			Breadcrumb{Title: fmt.Sprintf("%02d 日", sc.Day)},
		)
	}
	return bcs
}

func bucketLabelHref(level string, row *moviex.WMediaBirthBucketStat) (string, string) {
	switch level {
	case levelYear:
		return fmt.Sprintf("%04d", row.Year), fmt.Sprintf("%s/%d", rootPath, row.Year)
	case levelQuarter:
		return fmt.Sprintf("Q%d", row.Quarter), fmt.Sprintf("%s/%d/q/%d", rootPath, row.Year, row.Quarter)
	case levelMonth:
		return fmt.Sprintf("%02d 月", row.Month), fmt.Sprintf("%s/%d/%d", rootPath, row.Year, row.Month)
	case levelDay:
		return fmt.Sprintf("%02d 日", row.Day), fmt.Sprintf("%s/%d/%d/%d", rootPath, row.Year, row.Month, row.Day)
	default:
		return row.ScopeKey, ""
	}
}
