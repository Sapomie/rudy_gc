package shthreadagg

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rudy_gc/local/api/internal/model/modelx/moviex"
)

type BucketListParams struct {
	AggScope string
	Level    string
	TimeFrom *int64
	TimeTo   *int64
	Sort     string
	Dir      string
	Page     int64
	PageSize int64
}

type BucketListRow struct {
	BucketDisplay     string
	AggScope          string
	CountAll          int64
	CountPending      int64
	CountFetched      int64
	CountFailed       int64
	CountWithMagnet   int64
	CountWithInfoHash int64
	LatestPostTime    int64
	BucketStart       int64
	BucketEnd         int64
	ViewHref          string
}

type BucketListResult struct {
	Rows  []BucketListRow
	Total int64
}

func (s *Service) BuildBucketList(ctx context.Context, p BucketListParams) (*BucketListResult, error) {
	rows, total, err := s.deps.ShThreadBucketStatModel.ListPage(ctx, moviex.ShThreadBucketStatListFilter{
		AggScope: p.AggScope, Level: p.Level,
		BucketEndFrom: p.TimeFrom, BucketStartTo: p.TimeTo,
		Sort: p.Sort, Dir: p.Dir, Page: p.Page, PageSize: p.PageSize,
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
			BucketDisplay:     shBucketDisplay(row),
			AggScope:          row.AggScope,
			CountAll:          row.CountAll,
			CountPending:      row.CountPending,
			CountFetched:      row.CountFetched,
			CountFailed:       row.CountFailed,
			CountWithMagnet:   row.CountWithMagnet,
			CountWithInfoHash: row.CountWithInfoHash,
			LatestPostTime:    row.LatestPostTime,
			BucketStart:       row.BucketStart,
			BucketEnd:         row.BucketEnd,
			ViewHref:          shBucketViewHref(row),
		})
	}
	return &BucketListResult{Rows: out, Total: total}, nil
}

func shBucketDisplay(row *moviex.ShThreadBucketStat) string {
	if row == nil {
		return ""
	}
	switch row.Level {
	case levelYear:
		return strconv.FormatInt(row.Year, 10)
	case levelQuarter:
		return strconv.FormatInt(row.Year, 10) + "-Q" + strconv.FormatInt(row.Quarter, 10)
	case levelMonth:
		return strconv.FormatInt(row.Year, 10) + "-" + pad2(row.Month)
	case levelWeek:
		return strconv.FormatInt(row.Year, 10) + "-W" + pad2(row.Week)
	case levelDay:
		return strconv.FormatInt(row.Year, 10) + "-" + pad2(row.Month) + "-" + pad2(row.Day)
	default:
		return row.ScopeKey
	}
}

func shBucketViewHref(row *moviex.ShThreadBucketStat) string {
	if row == nil || row.BucketStart <= 0 || row.BucketEnd <= 0 {
		return ""
	}
	q := url.Values{}
	q.Set("post_time_from", time.Unix(row.BucketStart, 0).In(time.Local).Format(time.DateOnly))
	q.Set("post_time_to", time.Unix(row.BucketEnd, 0).In(time.Local).Format(time.DateOnly))
	if strings.HasPrefix(row.AggScope, "forum:") {
		forumID := strings.TrimSpace(strings.TrimPrefix(row.AggScope, "forum:"))
		if forumID != "" {
			q.Set("source_forum_ids", forumID)
		}
	}
	return "/shpool?" + q.Encode()
}

func pad2(v int64) string {
	if v < 10 {
		return "0" + strconv.FormatInt(v, 10)
	}
	return strconv.FormatInt(v, 10)
}
