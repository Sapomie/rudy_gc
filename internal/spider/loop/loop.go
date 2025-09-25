package loop

import (
	"context"
	"fmt"
	"rudy_gc/internal/spider/types"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoopServer struct {
	ctx  context.Context
	deps *svc.Deps

	InvCh    <-chan *types.Notification
	DetailCh chan<- *types.Notification

	refInvSemaphore chan struct{} // 并发限制
	goingOnInv      int32         // 运行中互斥
}

func NewLoopServer(ctx context.Context, deps *svc.Deps, invCh <-chan *types.Notification, detailCh chan<- *types.Notification, invConcurrency int) *LoopServer {
	if invConcurrency <= 0 {
		invConcurrency = 1
	}
	return &LoopServer{
		ctx:             ctx,
		deps:            deps,
		InvCh:           invCh,
		DetailCh:        detailCh,
		refInvSemaphore: make(chan struct{}, invConcurrency),
	}
}

func (m *LoopServer) Start() {
	go m.CrawlJavInvLoop()
	// 将来还会陆续加：go m.CrawlDetailLoop() ...
}

func (m *LoopServer) logInfo(msg string, kv ...any) {
	logx.WithContext(m.ctx).Infow(msg, toFields(kv...)...)
}

func (m *LoopServer) logWarn(msg string, kv ...any) {
	fs := append([]logx.LogField{logx.Field("level", "warn")}, toFields(kv...)...)
	logx.WithContext(m.ctx).Infow(msg, fs...)
}

func (m *LoopServer) logErr(err error, kv ...any) {
	fs := append([]logx.LogField{logx.Field("error", err.Error())}, toFields(kv...)...)
	logx.WithContext(m.ctx).Errorw("error", fs...)
}

// 把 "k1,v1,k2,v2,..." 变成 []logx.LogField
func toFields(kv ...any) []logx.LogField {
	if len(kv) == 0 {
		return nil
	}
	fs := make([]logx.LogField, 0, len(kv)/2+1)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			key = fmt.Sprintf("f%d", i/2)
		}
		fs = append(fs, logx.Field(key, kv[i+1]))
	}
	if len(kv)%2 == 1 { // 奇数个参数，最后那个没有 key
		fs = append(fs, logx.Field("arg", kv[len(kv)-1]))
	}
	return fs
}
