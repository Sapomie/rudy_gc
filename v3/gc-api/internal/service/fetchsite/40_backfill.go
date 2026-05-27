package fetchsite

import (
	"context"
	"fmt"

	"rudy_gc/internal/model/modelx/moviex"
	"rudy_gc/internal/taskctx"
)

type BackfillResult struct {
	Scanned int64
	Created int64
}

func (s *Service) BackfillFetchTasks(ctx context.Context, pageSize int64) (*BackfillResult, error) {
	if pageSize <= 0 {
		pageSize = 200
	}

	taskctx.ReportLog(ctx, taskctx.Log{
		Level:   "info",
		Message: fmt.Sprintf("开始回填外站抓取任务: page_size=%d", pageSize),
		Line:    fmt.Sprintf("开始回填外站抓取任务: page_size=%d", pageSize),
	})

	result := &BackfillResult{}
	offset := int64(0)
	for {
		rows, _, err := s.deps.MovieModel.ListPageWithTotal(ctx, offset, pageSize, "rd")
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			taskctx.ReportLog(ctx, taskctx.Log{
				Level:   "info",
				Message: fmt.Sprintf("外站抓取任务回填完成: scanned=%d | created=%d", result.Scanned, result.Created),
				Line:    fmt.Sprintf("外站抓取任务回填完成: scanned=%d | created=%d", result.Scanned, result.Created),
			})
			return result, nil
		}

		for _, row := range rows {
			if row == nil {
				continue
			}
			result.Scanned++
			createdNow := int64(0)

			if _, err := s.deps.JavbusMagnetFetchModel.FindOneByMovieJavId(ctx, row.JavId); err == moviex.ErrNotFound {
				createdNow++
			} else if err != nil {
				return nil, err
			}
			if _, err := s.deps.SukebeiTorrentFetchModel.FindOneByMovieJavId(ctx, row.JavId); err == moviex.ErrNotFound {
				createdNow++
			} else if err != nil {
				return nil, err
			}

			if err := s.EnsureFetchTasksForMovie(ctx, row.JavId, row.Name, row.ReleasingDate); err != nil {
				return nil, err
			}
			result.Created += createdNow
		}

		offset += int64(len(rows))
		if int64(len(rows)) < pageSize {
			taskctx.ReportLog(ctx, taskctx.Log{
				Level:   "info",
				Message: fmt.Sprintf("外站抓取任务回填完成: scanned=%d | created=%d", result.Scanned, result.Created),
				Line:    fmt.Sprintf("外站抓取任务回填完成: scanned=%d | created=%d", result.Scanned, result.Created),
			})
			return result, nil
		}
	}
}
