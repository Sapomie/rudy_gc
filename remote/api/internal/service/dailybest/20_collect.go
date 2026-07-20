package dailybest

import (
	"context"
	"errors"
	"fmt"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) collectAndPublish(ctx context.Context, runID int64, businessDate string, category int64) error {
	pages := make([]dailybesttypes.SourcePage, 0)
	seenPageChecksums := make(map[string]struct{})
	for pageNumber := int64(1); ; pageNumber++ {
		page, err := s.source.FetchPage(ctx, pageNumber)
		if errors.Is(err, dailybesttypes.ErrBlankPage) {
			if pageNumber == 1 {
				return fmt.Errorf("DailyBest 第一页为空")
			}
			break
		}
		if err != nil {
			return err
		}
		if _, exists := seenPageChecksums[page.Checksum]; exists {
			return fmt.Errorf("DailyBest 第 %d 页与已有页面内容重复，停止发布", pageNumber)
		}
		seenPageChecksums[page.Checksum] = struct{}{}
		pages = append(pages, page)
	}

	parsedRanks, err := ParsePages(pages)
	if err != nil {
		return err
	}
	dayNumber, err := contracts.RankDayNumber(businessDate)
	if err != nil {
		return fmt.Errorf("计算排行日期序号失败: %w", err)
	}
	ranks := make([]contracts.DailyBestRank, 0, len(parsedRanks))
	movies := make([]dailybesttypes.MovieDraft, 0, len(parsedRanks))
	for _, parsed := range parsedRanks {
		rank := parsed.Rank
		rank.RankKey = fmt.Sprintf("%s_%d_%03d", businessDate, category, rank.RankPos)
		rank.DayNumber = dayNumber
		rank.Category = category
		ranks = append(ranks, rank)
		movies = append(movies, parsed.Movie)
	}
	checksum := contracts.ComputeDailyBestRankChecksum(businessDate, dayNumber, category, ranks)
	return s.publishRanks(ctx, runID, dailybesttypes.RankDraft{
		BusinessDate: businessDate,
		DayNumber:    dayNumber,
		Category:     category,
		PageCount:    int64(len(pages)),
		Ranks:        ranks,
		Movies:       movies,
		Checksum:     checksum,
	}, time.Now().Unix())
}
