package dailybest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

func (s *Service) publishRanks(ctx context.Context, runID int64, draft dailybesttypes.RankDraft, finishedAt int64) error {
	existing, err := s.repository.ListAllRanks(ctx)
	if err != nil {
		return err
	}
	current := make([]contracts.DailyBestRank, 0)
	for _, rank := range existing {
		if rank.DayNumber == draft.DayNumber && rank.Category == draft.Category {
			current = append(current, rank)
		}
	}
	if len(current) > 0 {
		currentChecksum := contracts.ComputeDailyBestRankChecksum(
			draft.BusinessDate,
			draft.DayNumber,
			draft.Category,
			current,
		)
		if currentChecksum == draft.Checksum {
			return s.repository.MarkFetchRunNoChange(
				ctx, runID, finishedAt, draft.PageCount, int64(len(draft.Ranks)), draft.Checksum, draft.Movies,
			)
		}
		return fmt.Errorf(
			"DailyBest date=%s category=%d 已发布且校验和冲突: existing=%s fetched=%s",
			draft.BusinessDate,
			draft.Category,
			currentChecksum,
			draft.Checksum,
		)
	}
	for index := range draft.Ranks {
		draft.Ranks[index].CreatedOn = finishedAt
		draft.Ranks[index].UpdatedOn = finishedAt
	}
	stat, err := buildRankDayStat(
		append(append([]contracts.DailyBestRank(nil), existing...), draft.Ranks...),
		draft.DayNumber,
		finishedAt,
		finishedAt,
		finishedAt,
	)
	if err != nil {
		return err
	}
	return s.repository.PublishRanks(ctx, runID, draft, stat, finishedAt)
}

func buildRankDayStat(
	ranks []contracts.DailyBestRank,
	dayNumber int64,
	fetchedOn int64,
	syncedOn int64,
	calculatedOn int64,
) (contracts.DailyBestRankDayStat, error) {
	ordered := append([]contracts.DailyBestRank(nil), ranks...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].DayNumber != ordered[j].DayNumber {
			return ordered[i].DayNumber < ordered[j].DayNumber
		}
		if ordered[i].Category != ordered[j].Category {
			return ordered[i].Category < ordered[j].Category
		}
		if ordered[i].RankPos != ordered[j].RankPos {
			return ordered[i].RankPos < ordered[j].RankPos
		}
		return ordered[i].MovieJavID < ordered[j].MovieJavID
	})

	currentRows := make([]contracts.DailyBestRank, 0)
	seenMovies := make(map[string]struct{})
	prevMovies := make(map[string]struct{})
	prevDayNumber := int64(0)
	for _, row := range ordered {
		switch {
		case row.DayNumber < dayNumber:
			seenMovies[row.MovieJavID] = struct{}{}
			if row.DayNumber > prevDayNumber {
				prevDayNumber = row.DayNumber
				prevMovies = map[string]struct{}{row.MovieJavID: {}}
			} else if row.DayNumber == prevDayNumber {
				prevMovies[row.MovieJavID] = struct{}{}
			}
		case row.DayNumber == dayNumber:
			currentRows = append(currentRows, row)
		}
	}
	if len(currentRows) == 0 {
		return contracts.DailyBestRankDayStat{}, fmt.Errorf("DailyBest day_number=%d 没有可计算的排名", dayNumber)
	}
	rankStoredOn := int64(0)
	for _, row := range currentRows {
		if row.CreatedOn > 0 && (rankStoredOn == 0 || row.CreatedOn < rankStoredOn) {
			rankStoredOn = row.CreatedOn
		}
	}
	if fetchedOn <= 0 {
		fetchedOn = rankStoredOn
	}
	if syncedOn <= 0 {
		syncedOn = rankStoredOn
	}

	currentMovies := make(map[string]struct{})
	hasher := sha256.New()
	for _, row := range currentRows {
		currentMovies[row.MovieJavID] = struct{}{}
		_, _ = fmt.Fprintf(hasher, "%d:%d:%d:%s\n", row.DayNumber, row.Category, row.RankPos, row.MovieJavID)
	}
	stat := contracts.DailyBestRankDayStat{
		DayNumber: dayNumber, RankDate: contracts.RankDate(dayNumber), PrevDayNumber: prevDayNumber,
		RankRowCount: int64(len(currentRows)), MovieCount: int64(len(currentMovies)),
		SourceChecksum: hex.EncodeToString(hasher.Sum(nil)), FetchedAt: fetchedOn, SyncedAt: syncedOn,
		CalculatedOn: calculatedOn,
		CreatedOn:    calculatedOn, UpdatedOn: calculatedOn,
	}
	for movieJavID := range currentMovies {
		if _, ok := prevMovies[movieJavID]; ok {
			stat.ContinuedMovieCount++
		} else if _, ok := seenMovies[movieJavID]; ok {
			stat.ReturningMovieCount++
		} else {
			stat.FirstSeenMovieCount++
		}
	}
	for movieJavID := range prevMovies {
		if _, ok := currentMovies[movieJavID]; !ok {
			stat.DroppedMovieCount++
		}
	}
	return stat, nil
}
