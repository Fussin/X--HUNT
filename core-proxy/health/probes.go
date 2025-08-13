package health

import (
	"sync/atomic"
	"time"
)

type AcceptHeartbeat struct {
	last int64 // unix nanos
}

func (h *AcceptHeartbeat) Tick() {
	atomic.StoreInt64(&h.last, time.Now().UnixNano())
}

func (h *AcceptHeartbeat) Age() time.Duration {
	ts := atomic.LoadInt64(&h.last)
	if ts == 0 { return time.Hour } // never ticked
	return time.Since(time.Unix(0, ts))
}

type ListenerStatus struct {
	Name                string
	Bound               bool
	State               string // running|draining|stopped
	AcceptStall         bool
	HandshakeP95Ms      int64
	HandshakeErrRate    float64
	QueueDepth          int64
	QueueHighWatermark  float64
}

type ReadinessReport struct {
	Status     string // ready|not_ready
	Listeners  []ListenerStatus
	TLSManager struct {
		RootLoaded      bool
		LeafCacheItems  int
	}
}
