package mylog

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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

func NewLogrusLogger(level string) *logrus.Logger {
	l := logrus.New()
	var logLevel logrus.Level
	switch level {
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	}
	l.SetLevel(logLevel)

	l.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.DateTime,
	})

	l.SetFormatter(&LevelColorFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return l
}
