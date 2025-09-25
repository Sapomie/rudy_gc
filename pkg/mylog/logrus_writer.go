package mylog

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zeromicro/go-zero/core/logx"
)

// -------- Formatter: 时间戳灰色，Level 彩色，消息体固定白/灰 --------
type LevelColorFormatter struct {
	TimestampFormat string // 默认 "2006-01-02 15:04:05"
}

func (f *LevelColorFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	tsFmt := f.TimestampFormat
	if tsFmt == "" {
		tsFmt = "2006-01-02 15:04:05"
	}
	ts := entry.Time.Format(tsFmt)

	level := strings.ToUpper(entry.Level.String())

	// 只给 level 上色
	var lvlColor string
	switch entry.Level {
	case logrus.InfoLevel:
		lvlColor = "\033[36m" // 青色
	case logrus.WarnLevel:
		lvlColor = "\033[33m" // 黄色
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		lvlColor = "\033[31m" // 红色
	case logrus.DebugLevel:
		lvlColor = "\033[34m" // 亮蓝
	default:
		lvlColor = "\033[0m"
	}

	// 渲染 fields（暗灰色，避免影响正文颜色）
	var kvs string
	if len(entry.Data) > 0 {
		var b strings.Builder
		first := true
		for k, v := range entry.Data {
			if !first {
				b.WriteByte(' ')
			}
			first = false
			fmt.Fprintf(&b, "%s=%v", k, v)
		}
		kvs = " \033[90m" + b.String() + "\033[0m"
	}

	// 行首做全量复位；时间戳灰色；Level 着色；消息体强制白/灰（37）
	// WARNING 比较长，这里用 %-7s 对齐一下
	line := fmt.Sprintf(
		"\033[0m\033[39m\033[49m"+"\033[97m[%s]\033[0m %s%-7s\033[0m \033[97m%s\033[0m%s\n",
		ts, lvlColor, level, entry.Message, kvs,
	)

	return []byte(line), nil
}

// -------- logx.Writer 适配器（基于 logrus） --------

type Options struct {
	JSON            bool         // 若为 true，则使用 JSONFormatter（此时不会着色）
	TimestampFormat string       // 默认 "2006-01-02 15:04:05"
	Level           logrus.Level // 默认 InfoLevel
}

type logrusWriter struct {
	logger *logrus.Logger
}

// NewLogrusWriter: 将 logx 输出交给 logrus
func NewLogrusWriter(opts ...Options) logx.Writer {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.TimestampFormat == "" {
		o.TimestampFormat = "2006-01-02 15:04:05"
	}
	if o.Level == 0 {
		o.Level = logrus.InfoLevel
	}

	l := logrus.New()
	l.SetLevel(o.Level)

	if o.JSON {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: o.TimestampFormat,
		})
	} else {
		l.SetFormatter(&LevelColorFormatter{
			TimestampFormat: o.TimestampFormat,
		})
	}

	return &logrusWriter{logger: l}
}

// ---- 实现 logx.Writer 接口 ----

func (w *logrusWriter) Alert(v any) {
	w.logger.WithField("tag", "alert").Warn(anyToString(v))
}

func (w *logrusWriter) Close() error {
	// logrus 无需清理
	return nil
}

func (w *logrusWriter) Debug(v any, fields ...logx.LogField) {
	w.logger.WithFields(toFields(fields...)).Debug(anyToString(v))
}

func (w *logrusWriter) Error(v any, fields ...logx.LogField) {
	w.logger.WithFields(toFields(fields...)).Error(anyToString(v))
}

func (w *logrusWriter) Info(v any, fields ...logx.LogField) {
	w.logger.WithFields(toFields(fields...)).Info(anyToString(v))
}

func (w *logrusWriter) Severe(v any) {
	w.logger.WithField("severity", "severe").Error(anyToString(v))
}

func (w *logrusWriter) Slow(v any, fields ...logx.LogField) {
	fs := toFields(fields...)
	fs["tag"] = "slow"
	w.logger.WithFields(fs).Warn(anyToString(v))
}

func (w *logrusWriter) Stack(v any) {
	w.logger.WithField("stack", string(debug.Stack())).Error(anyToString(v))
}

func (w *logrusWriter) Stat(v any, fields ...logx.LogField) {
	fs := toFields(fields...)
	fs["tag"] = "stat"
	w.logger.WithFields(fs).Info(anyToString(v))
}

// ---- 辅助 ----

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case error:
		return strings.TrimSpace(t.Error())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

// 假设 go-zero 的 LogField 为 {Key string; Value any}
func toFields(fields ...logx.LogField) logrus.Fields {
	if len(fields) == 0 {
		return nil
	}
	fs := logrus.Fields{}
	for _, f := range fields {
		if f.Key != "" {
			fs[f.Key] = f.Value
		}
	}
	return fs
}
