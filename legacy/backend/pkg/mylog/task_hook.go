package mylog

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"rudy_gc/internal/taskctx"
)

var taskHookOnce sync.Map

type taskHook struct{}

func EnsureTaskHook(logger *logrus.Logger) {
	if logger == nil {
		return
	}
	if _, loaded := taskHookOnce.LoadOrStore(logger, struct{}{}); loaded {
		return
	}
	logger.AddHook(taskHook{})
}

func (taskHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (taskHook) Fire(entry *logrus.Entry) error {
	if entry == nil || entry.Context == nil {
		return nil
	}

	event := taskctx.Log{
		Level:   strings.ToLower(entry.Level.String()),
		Message: entry.Message,
		Line:    formatTaskLogLine(entry),
		At:      entry.Time.Unix(),
	}
	taskctx.ReportLog(entry.Context, event)
	return nil
}

func formatTaskLogLine(entry *logrus.Entry) string {
	if entry == nil {
		return ""
	}

	fields := formatTaskFields(entry.Data)
	ts := entry.Time
	if ts.IsZero() {
		ts = time.Now()
	}
	line := fmt.Sprintf("[%s] %-7s %s", ts.Format("2006-01-02 15:04:05"), strings.ToUpper(entry.Level.String()), entry.Message)
	if fields != "" {
		line += " " + fields
	}
	return line
}

func formatTaskFields(data logrus.Fields) string {
	if len(data) == 0 {
		return ""
	}
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, data[key]))
	}
	return strings.Join(parts, " ")
}
