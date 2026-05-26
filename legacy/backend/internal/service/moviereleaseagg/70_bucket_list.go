package moviereleaseagg

import (
	"context"
	"strconv"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

type BucketListParams struct {
	AggMode      string
	Level        string
	ScopeKeyLike string
	Year         *int64
	Quarter      *int64
	Month        *int64
	Day          *int64
	Sort         string
	Dir          string
	Page         int64
	PageSize     int64
}

type BucketListRow struct {
	BucketDisplay       string
	CountAll            int64
	CountOwned          int64
	SizeBytes           int64
	LatestReleasingDate int64
	ViewHref            string
}

type BucketListResult struct {
	Rows  []BucketListRow
	Total int64
}

func (s *Service) BuildBucketList(ctx context.Context, p BucketListParams) (*BucketListResult, error) {
	mode := NormalizeAggMode(p.AggMode)
	rows, total, err := s.deps.MovieReleaseBucketStatModel.ListPage(ctx, moviex.MovieReleaseBucketStatListFilter{
		AggMode:      mode,
		Level:        strings.TrimSpace(p.Level),
		ScopeKeyLike: strings.TrimSpace(p.ScopeKeyLike),
		Year:         p.Year,
		Quarter:      p.Quarter,
		Month:        p.Month,
		Day:          p.Day,
		Sort:         strings.TrimSpace(p.Sort),
		Dir:          strings.TrimSpace(p.Dir),
		Page:         p.Page,
		PageSize:     p.PageSize,
	})
	if err != nil {
		return nil, err
	}

	out := make([]BucketListRow, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		out = append(out, BucketListRow{
			BucketDisplay:       bucketDisplay(row),
			CountAll:            row.CountAll,
			CountOwned:          row.CountOwned,
			SizeBytes:           row.SizeBytes,
			LatestReleasingDate: row.LatestReleasingDate,
			ViewHref:            bucketListHref(row, mode),
		})
	}

	return &BucketListResult{Rows: out, Total: total}, nil
}

func bucketListHref(row *moviex.MovieReleaseBucketStat, mode string) string {
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

func bucketDisplay(row *moviex.MovieReleaseBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelYear:
		return strconv.FormatInt(row.Year, 10)
	case levelQuarter:
		return strconv.FormatInt(row.Year, 10) + "-Q" + pad2(row.Quarter)
	case levelMonth:
		return strconv.FormatInt(row.Year, 10) + "-M" + pad2(row.Month)
	case levelDay:
		return strconv.FormatInt(row.Year, 10) + "-" + pad2(row.Month) + "-" + pad2(row.Day)
	default:
		return row.ScopeKey
	}
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func pad2(v int64) string {
	if v < 10 {
		return "0" + strconv.FormatInt(v, 10)
	}
	return strconv.FormatInt(v, 10)
}
