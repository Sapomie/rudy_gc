package html

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"rudy_gc/internal/consts"
)

type ListRecordQuery struct {
	Type  string
	Limit int
}

// 传给模板的视图模型
type recordView struct {
	Name         string
	Type         string
	DetailNumber int64
	DurationMin  int64
	Date         string // ✅ 新增：格式化后的 StartTime
}

func (h *MovieHTMLHandler) ListRecordsPage(c *gin.Context) {
	q := ListRecordQuery{
		Type:  c.Query("type"),
		Limit: parseIntDefault(c.Query("limit"), 200),
	}
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 200
	}

	const startFrom int64 = 0
	records, err := h.movieSvc.ListRecords(c, startFrom, q.Type, q.Limit)
	if err != nil {
		c.String(http.StatusInternalServerError, "query records failed: %v", err)
		return
	}

	views := make([]recordView, 0, len(records))
	for _, r := range records {
		durSec := r.EndTime - r.StartTime
		if durSec < 0 {
			durSec = 0
		}
		durMin := int64(0)
		if durSec > 0 {
			durMin = (durSec + 59) / 60 // 向上取整
		}

		dateStr := "-"
		if r.StartTime > 0 {
			dateStr = time.Unix(r.StartTime, 0).Format("2006-01-02 15:04")
		}

		views = append(views, recordView{
			Name:         r.Name,
			Type:         r.Type,
			DetailNumber: r.DetailNumber,
			DurationMin:  durMin,
			Date:         dateStr,
		})
	}

	typeOptions := []string{
		consts.RecordTypeDailyBest,
		consts.RecordTypeSeedsActive,
		consts.RecordTypeSeedName,
	}

	c.HTML(http.StatusOK, "page.list_record", gin.H{
		"Title":       "Records",
		"Records":     views,
		"Query":       q,
		"TypeOptions": typeOptions,
	})
}
func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
