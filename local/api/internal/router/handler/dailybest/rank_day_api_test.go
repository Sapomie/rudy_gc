package dailybest

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseRankDayQueryAcceptsOpenEndedDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"GET",
		"/api/dailybest/rank-days?business_date_start=2026-07-01&p=2&ps=24",
		nil,
	)
	query, err := parseRankDayQuery(context)
	if err != nil {
		t.Fatal(err)
	}
	if query.BusinessDateStart != "2026-07-01" || query.BusinessDateEnd != "" || query.Page != 2 || query.PageSize != 24 {
		t.Fatalf("query=%+v", query)
	}
}

func TestParseRankDayQueryAcceptsDayNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"GET",
		"/api/dailybest/rank-days?day_number=1305&p=2&ps=24",
		nil,
	)
	query, err := parseRankDayQuery(context)
	if err != nil {
		t.Fatal(err)
	}
	if query.DayNumber != 1305 || query.Page != 2 || query.PageSize != 24 {
		t.Fatalf("query=%+v", query)
	}
}

func TestParseRankDayQueryRejectsInvalidDayNumber(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, value := range []string{"0", "-1", "invalid"} {
		t.Run(value, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(
				"GET",
				"/api/dailybest/rank-days?day_number="+value,
				nil,
			)
			if _, err := parseRankDayQuery(context); err == nil || err.Error() != "day_number 必须为正整数" {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestParseRankDayQueryRejectsInvertedDateRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		"GET",
		"/api/dailybest/rank-days?business_date_start=2026-07-21&business_date_end=2026-07-01",
		nil,
	)
	if _, err := parseRankDayQuery(context); err == nil {
		t.Fatal("expected inverted date range error")
	}
}
