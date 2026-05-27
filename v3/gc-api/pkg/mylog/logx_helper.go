package mylog

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

func Warn(ctx context.Context, msg string, kv ...any) {
	fs := []logx.LogField{logx.Field("level", "warn")}
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			fs = append(fs, logx.Field(k, kv[i+1]))
		}
	}
	logx.WithContext(ctx).Infow(msg, fs...)
}
