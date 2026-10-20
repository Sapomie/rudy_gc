package moviex

import (
	"context"
	"os"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestRankStatModelsAgainstConfiguredDatabase(t *testing.T) {
	dsn := os.Getenv("RANK_STAT_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("RANK_STAT_INTEGRATION_DSN is empty")
	}

	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	dayPage, err := NewCRankDayStatModel(conn).ListPage(ctx, CRankDayStatListQuery{
		Page:     1,
		PageSize: 7,
		Sort:     "day_number",
		Order:    "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(dayPage.Items) != 7 || dayPage.Total <= int64(len(dayPage.Items)) || dayPage.Summary.DayCount != dayPage.Total {
		t.Fatalf("unexpected day page: %+v", dayPage)
	}
	if dayPage.Items[0].DayNumber <= 0 || dayPage.Items[0].RankDate == "" {
		t.Fatalf("unexpected first day row: %+v", dayPage.Items[0])
	}
	lastDayPage, err := NewCRankDayStatModel(conn).ListPage(ctx, CRankDayStatListQuery{
		Page:     dayPage.Total + 1,
		PageSize: 7,
		Sort:     "day_number",
		Order:    "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedLastDayPage := dayPage.Total / 7
	if dayPage.Total%7 != 0 {
		expectedLastDayPage++
	}
	if lastDayPage.Page != expectedLastDayPage || len(lastDayPage.Items) == 0 {
		t.Fatalf("unexpected clamped day page: page=%d items=%d", lastDayPage.Page, len(lastDayPage.Items))
	}

	periodPage, err := NewCRankPeriodStatModel(conn).ListPage(ctx, CRankPeriodStatListQuery{
		Page:     1,
		PageSize: 5,
		Sort:     "start_day_number",
		Order:    "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(periodPage.Items) != 5 || periodPage.Total <= int64(len(periodPage.Items)) || periodPage.Summary.PeriodCount != periodPage.Total {
		t.Fatalf("unexpected period page: %+v", periodPage)
	}
	if periodPage.Items[0].Id <= 0 || periodPage.Items[0].PeriodType <= 0 || periodPage.Items[0].PeriodKey == "" {
		t.Fatalf("unexpected first period row: %+v", periodPage.Items[0])
	}
	lastPeriodPage, err := NewCRankPeriodStatModel(conn).ListPage(ctx, CRankPeriodStatListQuery{
		Page:     periodPage.Total + 1,
		PageSize: 5,
		Sort:     "start_day_number",
		Order:    "desc",
	})
	if err != nil {
		t.Fatal(err)
	}
	expectedLastPeriodPage := periodPage.Total / 5
	if periodPage.Total%5 != 0 {
		expectedLastPeriodPage++
	}
	if lastPeriodPage.Page != expectedLastPeriodPage || len(lastPeriodPage.Items) == 0 {
		t.Fatalf("unexpected clamped period page: page=%d items=%d", lastPeriodPage.Page, len(lastPeriodPage.Items))
	}
}
