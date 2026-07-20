package dailybest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	dailybesttypes "rudy_gc/remote/api/internal/types/dailybest"
	contracts "rudy_gc/shared/dailybest"
)

const (
	rankPeriodTypeWeek int64 = 1 + iota
	rankPeriodTypeMonth
	rankPeriodTypeQuarter
	rankPeriodTypeYear
)

const (
	rankPeriodStatusReady int64 = 2
	rankPeriodWeekTopN    int64 = 200
	rankPeriodMonthTopN   int64 = 300
	rankPeriodQuarterTopN int64 = 500
	rankPeriodYearTopN    int64 = 1000
)

type rankPeriodSpec struct {
	PeriodType     int64
	PeriodKey      string
	StartDayNumber int64
	EndDayNumber   int64
	PickDays       int64
	TopN           int64
}

type rankPeriodCandidate struct {
	MovieJavID         string
	Score              float64
	DaysInRank         int64
	UsedPickDays       int64
	FirstRankDayNumber int64
	LastRankDayNumber  int64
	BestRank           int64
	WorstPickedRank    int64
}

func (s *Service) publishRanks(ctx context.Context, runID int64, draft dailybesttypes.RankDraft, finishedAt int64) error {
	existing, err := s.repository.ListAllRanks(ctx)
	if err != nil {
		return err
	}
	current := make([]contracts.DailyBestRank, 0)
	combined := make([]contracts.DailyBestRank, 0, len(existing)+len(draft.Ranks))
	for _, rank := range existing {
		if rank.DayNumber == draft.DayNumber && rank.Category == draft.Category {
			current = append(current, rank)
			continue
		}
		combined = append(combined, rank)
	}
	if len(current) > 0 && contracts.ComputeDailyBestRankChecksum(draft.BusinessDate, draft.DayNumber, draft.Category, current) == draft.Checksum {
		return s.repository.MarkFetchRunNoChange(
			ctx, runID, finishedAt, draft.PageCount, int64(len(draft.Ranks)), draft.Checksum, draft.Movies,
		)
	}
	for _, rank := range draft.Ranks {
		rank.CreatedOn = finishedAt
		rank.UpdatedOn = finishedAt
		combined = append(combined, rank)
	}
	projection := buildRankProjection(combined, finishedAt, s.loc)
	return s.repository.PublishRanks(ctx, runID, draft, projection, finishedAt)
}

func buildRankProjection(ranks []contracts.DailyBestRank, calculatedOn int64, loc *time.Location) dailybesttypes.RankProjection {
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
	return dailybesttypes.RankProjection{
		DayStats: buildRankDayStats(ordered, calculatedOn),
		Periods:  buildRankPeriods(ordered, calculatedOn, loc),
	}
}

func buildRankDayStats(rows []contracts.DailyBestRank, calculatedOn int64) []contracts.DailyBestRankDayStat {
	stats := make([]contracts.DailyBestRankDayStat, 0)
	seenMovies := make(map[string]struct{})
	prevBestRank := make(map[string]int64)
	prevDayNumber := int64(0)
	for start := 0; start < len(rows); {
		end := start + 1
		dayNumber := rows[start].DayNumber
		for end < len(rows) && rows[end].DayNumber == dayNumber {
			end++
		}
		dayRows := rows[start:end]
		currentBestRank := make(map[string]int64)
		hasher := sha256.New()
		for _, row := range dayRows {
			if best, ok := currentBestRank[row.MovieJavID]; !ok || row.RankPos < best {
				currentBestRank[row.MovieJavID] = row.RankPos
			}
			_, _ = fmt.Fprintf(hasher, "%d:%d:%d:%s\n", row.DayNumber, row.Category, row.RankPos, row.MovieJavID)
		}
		stat := contracts.DailyBestRankDayStat{
			DayNumber: dayNumber, RankDate: contracts.RankDate(dayNumber), PrevDayNumber: prevDayNumber,
			RankRowCount: int64(len(dayRows)), MovieCount: int64(len(currentBestRank)),
			SourceChecksum: hex.EncodeToString(hasher.Sum(nil)), CalculatedOn: calculatedOn,
			CreatedOn: calculatedOn, UpdatedOn: calculatedOn,
		}
		for movieJavID, currentRank := range currentBestRank {
			if prevRank, ok := prevBestRank[movieJavID]; ok {
				stat.ContinuedMovieCount++
				switch {
				case currentRank < prevRank:
					stat.RiseMovieCount++
				case currentRank > prevRank:
					stat.FallMovieCount++
				default:
					stat.UnchangedMovieCount++
				}
			} else if _, ok := seenMovies[movieJavID]; ok {
				stat.ReturningMovieCount++
			} else {
				stat.FirstSeenMovieCount++
			}
		}
		for movieJavID := range prevBestRank {
			if _, ok := currentBestRank[movieJavID]; !ok {
				stat.DroppedMovieCount++
			}
		}
		for movieJavID := range currentBestRank {
			seenMovies[movieJavID] = struct{}{}
		}
		stats = append(stats, stat)
		prevBestRank = currentBestRank
		prevDayNumber = dayNumber
		start = end
	}
	return stats
}

func buildRankPeriods(rows []contracts.DailyBestRank, calculatedOn int64, loc *time.Location) []dailybesttypes.PeriodProjection {
	if len(rows) == 0 {
		return []dailybesttypes.PeriodProjection{}
	}
	earliestDay, latestDay := rows[0].DayNumber, rows[len(rows)-1].DayNumber
	specs := buildHistoricalPeriodSpecs(earliestDay, latestDay, loc)
	categories := make(map[int64]struct{})
	for _, row := range rows {
		categories[row.Category] = struct{}{}
	}
	categoryList := make([]int64, 0, len(categories))
	for category := range categories {
		categoryList = append(categoryList, category)
	}
	sort.Slice(categoryList, func(i, j int) bool { return categoryList[i] < categoryList[j] })

	previousBySeries := make(map[string]dailybesttypes.PeriodProjection)
	peakBySeries := make(map[string]map[string]int64)
	seenBySeries := make(map[string]map[string]int64)
	out := make([]dailybesttypes.PeriodProjection, 0)
	for _, spec := range specs {
		for _, category := range categoryList {
			sourceRows := ranksInPeriod(rows, spec, category)
			if len(sourceRows) == 0 {
				continue
			}
			seriesKey := fmt.Sprintf("%d:%d", spec.PeriodType, category)
			prev := previousBySeries[seriesKey]
			prevRanks := make(map[string]int64, len(prev.Items))
			for _, item := range prev.Items {
				prevRanks[item.MovieJavID] = item.RankPos
			}
			peakRanks := peakBySeries[seriesKey]
			if peakRanks == nil {
				peakRanks = make(map[string]int64)
				peakBySeries[seriesKey] = peakRanks
			}
			seenBefore := seenBySeries[seriesKey]
			if seenBefore == nil {
				seenBefore = make(map[string]int64)
				seenBySeries[seriesKey] = seenBefore
			}

			resolved, items := buildRankPeriodItems(spec, sourceRows, prevRanks, peakRanks, calculatedOn)
			key := fmt.Sprintf("%d:%s:%d", resolved.PeriodType, resolved.PeriodKey, category)
			projection := dailybesttypes.PeriodProjection{
				Key: key, PrevKey: prev.Key,
				Period: contracts.DailyBestRankPeriod{
					PeriodType: resolved.PeriodType, PeriodKey: resolved.PeriodKey, Category: category,
					StartDayNumber: resolved.StartDayNumber, EndDayNumber: resolved.EndDayNumber,
					PickDays: resolved.PickDays, TopN: resolved.TopN, Status: rankPeriodStatusReady,
					CreatedOn: calculatedOn, UpdatedOn: calculatedOn,
				},
				Items: items,
			}
			projection.Stat = buildRankPeriodStat(projection, sourceRows, prev.Items, seenBefore, calculatedOn)
			out = append(out, projection)
			for _, item := range items {
				seenBefore[item.MovieJavID] = item.RankPos
				if peakRanks[item.MovieJavID] == 0 || item.RankPos < peakRanks[item.MovieJavID] {
					peakRanks[item.MovieJavID] = item.RankPos
				}
			}
			previousBySeries[seriesKey] = projection
		}
	}
	return out
}

func ranksInPeriod(rows []contracts.DailyBestRank, spec rankPeriodSpec, category int64) []contracts.DailyBestRank {
	out := make([]contracts.DailyBestRank, 0)
	for _, row := range rows {
		if row.Category == category && row.DayNumber >= spec.StartDayNumber && row.DayNumber <= spec.EndDayNumber {
			out = append(out, row)
		}
	}
	return out
}

func buildHistoricalPeriodSpecs(earliestDay, latestDay int64, loc *time.Location) []rankPeriodSpec {
	specMap := make(map[string]rankPeriodSpec)
	for day := earliestDay; day <= latestDay; day++ {
		date, _ := time.ParseInLocation(time.DateOnly, contracts.RankDate(day), loc)
		weekStart := startOfISOWeek(date, loc)
		monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, loc)
		quarterStart := startOfQuarter(date, loc)
		yearStart := time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, loc)
		specs := []rankPeriodSpec{
			newRankPeriodSpec(rankPeriodTypeWeek, weekPeriodKey(date), weekStart, date, rankPeriodWeekTopN),
			newRankPeriodSpec(rankPeriodTypeMonth, date.Format("2006-01"), monthStart, date, rankPeriodMonthTopN),
			newRankPeriodSpec(rankPeriodTypeQuarter, quarterPeriodKey(date), quarterStart, date, rankPeriodQuarterTopN),
			newRankPeriodSpec(rankPeriodTypeYear, date.Format("2006"), yearStart, date, rankPeriodYearTopN),
		}
		for _, spec := range specs {
			specMap[fmt.Sprintf("%d:%s", spec.PeriodType, spec.PeriodKey)] = spec
		}
	}
	out := make([]rankPeriodSpec, 0, len(specMap))
	for _, spec := range specMap {
		out = append(out, spec)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PeriodType != out[j].PeriodType {
			return out[i].PeriodType < out[j].PeriodType
		}
		return out[i].StartDayNumber < out[j].StartDayNumber
	})
	return out
}

func newRankPeriodSpec(periodType int64, periodKey string, startTime, endTime time.Time, topN int64) rankPeriodSpec {
	startDay, _ := contracts.RankDayNumber(startTime.Format(time.DateOnly))
	endDay, _ := contracts.RankDayNumber(endTime.Format(time.DateOnly))
	return rankPeriodSpec{PeriodType: periodType, PeriodKey: periodKey, StartDayNumber: startDay, EndDayNumber: endDay, TopN: topN}
}

func buildRankPeriodItems(spec rankPeriodSpec, rows []contracts.DailyBestRank, prevRankMap, peakRankMap map[string]int64, now int64) (rankPeriodSpec, []contracts.DailyBestRankPeriodItem) {
	grouped := make(map[string][]contracts.DailyBestRank)
	for _, row := range rows {
		if row.MovieJavID != "" {
			grouped[row.MovieJavID] = append(grouped[row.MovieJavID], row)
		}
	}
	spec.PickDays = calcRankPeriodPickDays(grouped, spec.TopN)
	candidates := make([]rankPeriodCandidate, 0, len(grouped))
	for movieJavID, movieRows := range grouped {
		candidate, ok := buildRankPeriodCandidate(movieJavID, movieRows, spec.PickDays)
		if ok {
			candidates = append(candidates, candidate)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score < candidates[j].Score
		}
		if candidates[i].BestRank != candidates[j].BestRank {
			return candidates[i].BestRank < candidates[j].BestRank
		}
		if candidates[i].DaysInRank != candidates[j].DaysInRank {
			return candidates[i].DaysInRank > candidates[j].DaysInRank
		}
		return candidates[i].MovieJavID < candidates[j].MovieJavID
	})
	if spec.TopN > 0 && int64(len(candidates)) > spec.TopN {
		candidates = candidates[:spec.TopN]
	}
	items := make([]contracts.DailyBestRankPeriodItem, 0, len(candidates))
	for index, candidate := range candidates {
		rankPos := int64(index + 1)
		prevRank := prevRankMap[candidate.MovieJavID]
		rankChange := int64(0)
		if prevRank > 0 {
			rankChange = prevRank - rankPos
		}
		peakRank := rankPos
		if oldPeak := peakRankMap[candidate.MovieJavID]; oldPeak > 0 && oldPeak < peakRank {
			peakRank = oldPeak
		}
		items = append(items, contracts.DailyBestRankPeriodItem{
			MovieJavID: candidate.MovieJavID, RankPos: rankPos, Score: candidate.Score,
			DaysInRank: candidate.DaysInRank, UsedPickDays: candidate.UsedPickDays,
			FirstRankDayNumber: candidate.FirstRankDayNumber, LastRankDayNumber: candidate.LastRankDayNumber,
			BestRank: candidate.BestRank, WorstPickedRank: candidate.WorstPickedRank,
			PrevRank: prevRank, RankChange: rankChange, PeakRank: peakRank,
			CreatedOn: now, UpdatedOn: now,
		})
	}
	return spec, items
}

func buildRankPeriodCandidate(movieJavID string, rows []contracts.DailyBestRank, pickDays int64) (rankPeriodCandidate, bool) {
	if pickDays <= 0 || int64(len(rows)) < pickDays {
		return rankPeriodCandidate{}, false
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].RankPos != rows[j].RankPos {
			return rows[i].RankPos < rows[j].RankPos
		}
		return rows[i].DayNumber < rows[j].DayNumber
	})
	firstDay, lastDay, bestRank := rows[0].DayNumber, rows[0].DayNumber, rows[0].RankPos
	for _, row := range rows[1:] {
		if row.DayNumber < firstDay {
			firstDay = row.DayNumber
		}
		if row.DayNumber > lastDay {
			lastDay = row.DayNumber
		}
		if row.RankPos < bestRank {
			bestRank = row.RankPos
		}
	}
	picked := rows[:pickDays]
	var scoreSum float64
	worstPickedRank := picked[0].RankPos
	for _, row := range picked {
		scoreSum += float64(row.RankPos)
		if row.RankPos > worstPickedRank {
			worstPickedRank = row.RankPos
		}
	}
	return rankPeriodCandidate{
		MovieJavID: movieJavID, Score: scoreSum / float64(pickDays), DaysInRank: int64(len(rows)),
		UsedPickDays: pickDays, FirstRankDayNumber: firstDay, LastRankDayNumber: lastDay,
		BestRank: bestRank, WorstPickedRank: worstPickedRank,
	}, true
}

func calcRankPeriodPickDays(grouped map[string][]contracts.DailyBestRank, topN int64) int64 {
	if len(grouped) == 0 {
		return 0
	}
	counts := make([]int64, 0, len(grouped))
	for _, rows := range grouped {
		counts = append(counts, int64(len(rows)))
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i] > counts[j] })
	if topN <= 0 || topN > int64(len(counts)) {
		return counts[len(counts)-1]
	}
	return counts[topN-1]
}

func buildRankPeriodStat(projection dailybesttypes.PeriodProjection, sourceRows []contracts.DailyBestRank, prevItems []contracts.DailyBestRankPeriodItem, seenBefore map[string]int64, now int64) contracts.DailyBestRankPeriodStat {
	stat := contracts.DailyBestRankPeriodStat{
		SourceRankRowCount: int64(len(sourceRows)), PeriodMovieCount: int64(len(projection.Items)),
		CalculatedOn: now, CreatedOn: now, UpdatedOn: now,
	}
	sourceDays := make(map[int64]struct{})
	sourceMovies := make(map[string]struct{})
	hasher := sha256.New()
	for _, row := range sourceRows {
		sourceDays[row.DayNumber] = struct{}{}
		sourceMovies[row.MovieJavID] = struct{}{}
		_, _ = fmt.Fprintf(hasher, "%d:%d:%d:%s\n", row.DayNumber, row.Category, row.RankPos, row.MovieJavID)
		if stat.SourceFetchedStartOn == 0 || (row.CreatedOn > 0 && row.CreatedOn < stat.SourceFetchedStartOn) {
			stat.SourceFetchedStartOn = row.CreatedOn
		}
		if row.CreatedOn > stat.SourceFetchedEndOn {
			stat.SourceFetchedEndOn = row.CreatedOn
		}
	}
	stat.SourceRankDayCount = int64(len(sourceDays))
	stat.SourceMovieCount = int64(len(sourceMovies))
	stat.MissingRankDayCount = projection.Period.EndDayNumber - projection.Period.StartDayNumber + 1 - stat.SourceRankDayCount
	if stat.MissingRankDayCount < 0 {
		stat.MissingRankDayCount = 0
	}
	stat.SourceChecksum = hex.EncodeToString(hasher.Sum(nil))
	prevRanks := make(map[string]int64, len(prevItems))
	for _, item := range prevItems {
		prevRanks[item.MovieJavID] = item.RankPos
	}
	current := make(map[string]struct{}, len(projection.Items))
	var scoreTotal float64
	var daysTotal int64
	for _, item := range projection.Items {
		current[item.MovieJavID] = struct{}{}
		scoreTotal += item.Score
		daysTotal += item.DaysInRank
		if item.RankPos == 1 {
			stat.RankOneMovieJavID = item.MovieJavID
		}
		if prevRank, ok := prevRanks[item.MovieJavID]; ok {
			stat.ContinuedMovieCount++
			switch {
			case item.RankPos < prevRank:
				stat.RiseMovieCount++
			case item.RankPos > prevRank:
				stat.FallMovieCount++
			default:
				stat.UnchangedMovieCount++
			}
		} else if _, ok := seenBefore[item.MovieJavID]; ok {
			stat.ReturningMovieCount++
		} else {
			stat.FirstSeenMovieCount++
		}
		if item.PeakRank > 0 && item.RankPos == item.PeakRank {
			stat.PeakMovieCount++
		}
	}
	for movieJavID := range prevRanks {
		if _, ok := current[movieJavID]; !ok {
			stat.DroppedMovieCount++
		}
	}
	if len(projection.Items) > 0 {
		stat.AverageScore = scoreTotal / float64(len(projection.Items))
		stat.AverageDaysInRank = float64(daysTotal) / float64(len(projection.Items))
	}
	return stat
}

func startOfISOWeek(t time.Time, loc *time.Location) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, loc)
}

func startOfQuarter(t time.Time, loc *time.Location) time.Time {
	month := ((int(t.Month())-1)/3)*3 + 1
	return time.Date(t.Year(), time.Month(month), 1, 0, 0, 0, 0, loc)
}

func weekPeriodKey(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%04d-W%02d", year, week)
}

func quarterPeriodKey(t time.Time) string {
	quarter := ((int(t.Month()) - 1) / 3) + 1
	return fmt.Sprintf("%04d-Q%d", t.Year(), quarter)
}
