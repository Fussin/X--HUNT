package backpressure

import (
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type bucket struct {
	lim *rate.Limiter
	exp time.Time
}

type TokenLimiter struct {
	mu     sync.Mutex
	perIP  map[string]*bucket
	global *rate.Limiter
	ttl    time.Duration
}

func NewTokenLimiter(globalRPS float64, ttl time.Duration) *TokenLimiter {
	return &TokenLimiter{
		perIP:  make(map[string]*bucket),
		global: rate.NewLimiter(rate.Limit(globalRPS), int(globalRPS)),
		ttl:    ttl,
	}
}

func (t *TokenLimiter) Allow(ip net.IP, perIPRPS float64) bool {
	if !t.global.Allow() { return false }

	key := ip.String()
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.perIP[key]
	if !ok || now.After(b.exp) {
		b = &bucket{
			lim: rate.NewLimiter(rate.Limit(perIPRPS), int(perIPRPS)),
			exp: now.Add(t.ttl),
		}
		t.perIP[key] = b
	}
	return b.lim.Allow()
}
