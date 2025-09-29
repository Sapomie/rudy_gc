package loop

import (
	"context"
	"fmt"
	"sync/atomic"

	"rudy_gc/internal/spider/types"
	"rudy_gc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoopServer struct {
	ctx  context.Context
	deps *svc.Deps

	// 各环节通知通道
	InvCh           <-chan *types.Notification
	DetailCh        chan *types.Notification
	TranslationCh   chan struct{}
	DownloadCoverCh chan struct{}

	// 并发/互斥控制
	refInvSemaphore    chan struct{}
	refDetailSemaphore chan struct{}
	goingOnInv         int32
	goingOnDetail      int32
}

func NewLoopServer(ctx context.Context, deps *svc.Deps, invConcurrency int) *LoopServer {
	if invConcurrency <= 0 {
		invConcurrency = 1
	}

	return &LoopServer{
		ctx:                ctx,
		deps:               deps,
		InvCh:              make(chan *types.Notification, 4), // 带缓冲，避免阻塞
		DetailCh:           make(chan *types.Notification, 4),
		TranslationCh:      make(chan struct{}, 4),
		DownloadCoverCh:    make(chan struct{}, 4),
		refInvSemaphore:    make(chan struct{}, invConcurrency),
		refDetailSemaphore: make(chan struct{}, 1),
	}
}

func (m *LoopServer) Start() {
	go m.CrawlJavInvLoop()
	go m.CrawlDetailLoop()
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

// 互斥位辅助（供各自 loop 使用）
func (m *LoopServer) startInv() bool    { return atomic.CompareAndSwapInt32(&m.goingOnInv, 0, 1) }
func (m *LoopServer) stopInv()          { atomic.StoreInt32(&m.goingOnInv, 0) }
func (m *LoopServer) startDetail() bool { return atomic.CompareAndSwapInt32(&m.goingOnDetail, 0, 1) }
func (m *LoopServer) stopDetail()       { atomic.StoreInt32(&m.goingOnDetail, 0) }

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
