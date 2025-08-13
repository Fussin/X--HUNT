package backpressure

import (
	"net"
	"time"

	"sentinelx/core-proxy/internal/syncx"
	"sentinelx/core-proxy/internal/metrics"
)

type Policy struct {
	LowWM        float64
	HighWM       float64
	HardCap      int
	MinSleep     time.Duration
	MaxSleep     time.Duration
}

type Controller struct {
	sem     *syncx.Semaphore
	tokens  *TokenLimiter
	circuit *Circuit
	policy  Policy
}

func NewController(sem *syncx.Semaphore, tokens *TokenLimiter, circuit *Circuit, p Policy) *Controller {
	return &Controller{sem: sem, tokens: tokens, circuit: circuit, policy: p}
}

func (c *Controller) Admit(ip net.IP, perIPRPS float64) (allow bool, sleep time.Duration) {
	// Circuit breaker first
	if !c.circuit.Allow() {
		metrics.IncBackpressureEvent("circuit_open")
		return false, 0
	}
	// Token buckets
	if !c.tokens.Allow(ip, perIPRPS) {
		metrics.IncBackpressureEvent("rate_limit")
		return false, 0
	}
	// Hard cap
	depth := c.sem.Size()
	if depth >= c.policy.HardCap {
		metrics.IncBackpressureEvent("hard_cap_reject")
		return false, 0
	}
	// Adaptive pacing based on occupancy ratio
	occ := float64(depth) / float64(c.policy.HardCap)
	if occ >= c.policy.HighWM {
		// Heavy pacing
		sleep = c.policy.MaxSleep
		metrics.IncBackpressureEvent("pacing_high")
	} else if occ >= c.policy.LowWM {
		sleep = c.policy.MinSleep
		metrics.IncBackpressureEvent("pacing_low")
	}
	return true, sleep
}
