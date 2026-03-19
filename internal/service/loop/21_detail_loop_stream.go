package loop

import (
	"sync"
	"time"
)

const detailLoopHistorySize = 500

type DetailLoopEvent struct {
	Level   string `json:"level,omitempty"`
	Message string `json:"message,omitempty"`
	Line    string `json:"line,omitempty"`
	At      int64  `json:"at"`
}

type detailLoopLogHub struct {
	subscribeBuffer int

	mu          sync.Mutex
	history     []DetailLoopEvent
	lastEvent   *DetailLoopEvent
	subscribers map[chan DetailLoopEvent]struct{}
}

func newDetailLoopLogHub(subscribeBuffer int) *detailLoopLogHub {
	if subscribeBuffer <= 0 {
		subscribeBuffer = 64
	}
	return &detailLoopLogHub{
		subscribeBuffer: subscribeBuffer,
		history:         make([]DetailLoopEvent, 0, detailLoopHistorySize),
		subscribers:     make(map[chan DetailLoopEvent]struct{}),
	}
}

func (h *detailLoopLogHub) subscribe() ([]DetailLoopEvent, <-chan DetailLoopEvent, func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	history := append([]DetailLoopEvent(nil), h.history...)
	ch := make(chan DetailLoopEvent, h.subscribeBuffer)
	h.subscribers[ch] = struct{}{}
	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.subscribers, ch)
	}
	return history, ch, cancel
}

func (h *detailLoopLogHub) publish(event DetailLoopEvent) {
	if h == nil {
		return
	}
	if event.At == 0 {
		event.At = time.Now().Unix()
	}

	h.mu.Lock()
	h.history = append(h.history, event)
	if len(h.history) > detailLoopHistorySize {
		h.history = append([]DetailLoopEvent(nil), h.history[len(h.history)-detailLoopHistorySize:]...)
	}
	cloned := event
	h.lastEvent = &cloned
	subs := make([]chan DetailLoopEvent, 0, len(h.subscribers))
	for ch := range h.subscribers {
		subs = append(subs, ch)
	}
	h.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

func (h *detailLoopLogHub) latest() (DetailLoopEvent, bool) {
	if h == nil {
		return DetailLoopEvent{}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.lastEvent == nil {
		return DetailLoopEvent{}, false
	}
	return *h.lastEvent, true
}
