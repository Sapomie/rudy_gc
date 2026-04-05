package wmediaagg

import (
	"context"
	"strconv"
	"strings"

	"rudy_gc/internal/model/modelx/moviex"
)

type BucketListParams struct {
	Level        string
	ScopeKeyLike string
	Year         *int64
	Quarter      *int64
	Month        *int64
	Sort         string
	Dir          string
	Page         int64
	PageSize     int64
}

type BucketListRow struct {
	BucketDisplay   string
	MediaCount      int64
	RemovedCount    int64
	SizeBytes       int64
	HasSubCount     int64
	LatestBirthTime int64
	ViewHref        string
}

type BucketListResult struct {
	Rows  []BucketListRow
	Total int64
}

func (s *Service) BuildBucketList(ctx context.Context, p BucketListParams) (*BucketListResult, error) {
	rows, total, err := s.deps.WMediaBirthBucketStatModel.ListPage(ctx, moviex.WMediaBirthBucketStatListFilter{
		Level:        strings.TrimSpace(p.Level),
		ScopeKeyLike: strings.TrimSpace(p.ScopeKeyLike),
		Year:         p.Year,
		Quarter:      p.Quarter,
		Month:        p.Month,
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
			BucketDisplay:   bucketDisplay(row),
			MediaCount:      row.MediaCount,
			RemovedCount:    row.RemovedCount,
			SizeBytes:       row.SizeBytes,
			HasSubCount:     row.HasSubCount,
			LatestBirthTime: row.LatestBirthTime,
			ViewHref:        bucketListHref(row),
		})
	}

	return &BucketListResult{
		Rows:  out,
		Total: total,
	}, nil
}

func bucketListHref(row *moviex.WMediaBirthBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelRoot:
		return rootPath
	case levelYear:
		return rootPath + "/" + itoa64(row.Year)
	case levelQuarter:
		return rootPath + "/" + itoa64(row.Year) + "/q/" + itoa64(row.Quarter)
	case levelMonth:
		return rootPath + "/" + itoa64(row.Year) + "/" + itoa64(row.Month)
	case levelDay:
		return rootPath + "/" + itoa64(row.Year) + "/" + itoa64(row.Month) + "/" + itoa64(row.Day)
	default:
		return ""
	}
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func bucketDisplay(row *moviex.WMediaBirthBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelRoot:
		return "全部时间"
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

func pad2(v int64) string {
	if v < 10 {
		return "0" + strconv.FormatInt(v, 10)
	}
	return strconv.FormatInt(v, 10)
}
