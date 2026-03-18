package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"rudy_gc/internal/types"
)

type castRankInfo struct {
	Rank500MovieNumber int64
	Rank20MovieNumber  int64
	Rank1MovieNumber   int64
	HighestRank        int64
	RankTimes          int64
}

func (l *CrawlLogic) rebuildCastRankStatsByIDs(ctx context.Context, castIDs map[int64]struct{}) error {
	for castID := range castIDs {
		castRow, err := l.deps.CastRepo.FindOne(ctx, castID)
		if err != nil {
			return fmt.Errorf("CastRepo.FindOne %d: %w", castID, err)
		}

		movieJavIDs, err := l.deps.MovieCastRepo.ListMovieJavIDsByCastID(ctx, castID)
		if err != nil {
			return fmt.Errorf("MovieCastRepo.ListMovieJavIDsByCastID %d: %w", castID, err)
		}

		info, err := l.buildCastRankInfo(ctx, movieJavIDs)
		if err != nil {
			return fmt.Errorf("buildCastRankInfo %d: %w", castID, err)
		}

		if castRow.Rank500MovieNumber == info.Rank500MovieNumber &&
			castRow.Rank20MovieNumber == info.Rank20MovieNumber &&
			castRow.Rank1MovieNumber == info.Rank1MovieNumber &&
			castRow.HighestRank == info.HighestRank &&
			castRow.RankTimes == info.RankTimes {
			continue
		}

		castRow.Rank500MovieNumber = info.Rank500MovieNumber
		castRow.Rank20MovieNumber = info.Rank20MovieNumber
		castRow.Rank1MovieNumber = info.Rank1MovieNumber
		castRow.HighestRank = info.HighestRank
		castRow.RankTimes = info.RankTimes

		if _, err := l.deps.CastRepo.Upsert(ctx, castRow); err != nil {
			return fmt.Errorf("CastRepo.Upsert %s: %w", castRow.Name, err)
		}
	}

	return nil
}

func (l *CrawlLogic) buildCastRankInfo(ctx context.Context, movieJavIDs []string) (castRankInfo, error) {
	var out castRankInfo

	for _, movieJavID := range uniqueMovieJavIDs(movieJavIDs) {
		minfo, err := l.deps.MinfoRepo.FindOneByJavId(ctx, movieJavID)
		if err != nil {
			if errors.Is(err, types.ErrNotFound) {
				continue
			}
			return out, fmt.Errorf("MinfoRepo.FindOneByJavId %s: %w", movieJavID, err)
		}
		if minfo == nil {
			continue
		}

		if minfo.HighestRank <= 0 || minfo.HighestRank >= 1000 {
			continue
		}

		if minfo.DaysInRank > 0 {
			out.RankTimes += minfo.DaysInRank
		}

		if out.HighestRank == 0 || minfo.HighestRank < out.HighestRank {
			out.HighestRank = minfo.HighestRank
		}
		if minfo.HighestRank <= 500 {
			out.Rank500MovieNumber++
		}
		if minfo.HighestRank <= 20 {
			out.Rank20MovieNumber++
		}
		if minfo.HighestRank == 1 {
			out.Rank1MovieNumber++
		}
	}

	return out, nil
}

func uniqueMovieJavIDs(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (l *CrawlLogic) RebuildAllCastRankStats(ctx context.Context) error {
	castIDs, err := l.deps.CastRepo.ListAllIDs(ctx)
	if err != nil {
		return fmt.Errorf("CastRepo.ListAllIDs: %w", err)
	}

	castIDSet := make(map[int64]struct{}, len(castIDs))
	for _, castID := range castIDs {
		if castID > 0 {
			castIDSet[castID] = struct{}{}
		}
	}

	if err := l.rebuildCastRankStatsByIDs(ctx, castIDSet); err != nil {
		return fmt.Errorf("rebuild cast rank stats: %w", err)
	}

	return nil
}

func (l *CrawlLogic) RebuildCastRankStatsByName(ctx context.Context, actorName string) error {
	actorName = strings.TrimSpace(actorName)
	if actorName == "" {
		return fmt.Errorf("empty actor name")
	}

	castRow, err := l.deps.CastRepo.FindOneByName(ctx, actorName)
	if err != nil {
		return fmt.Errorf("CastRepo.FindOneByName %s: %w", actorName, err)
	}
	if castRow == nil {
		return fmt.Errorf("actor not found: %s", actorName)
	}

	return l.rebuildCastRankStatsByIDs(ctx, map[int64]struct{}{castRow.Id: {}})
}
