package backpressure

import (
	"sync"
	"time"
)

type CircuitState string

const (
	Closed    CircuitState = "closed"
	Open      CircuitState = "open"
	HalfOpen  CircuitState = "half_open"
)

type Circuit struct {
	mu          sync.Mutex
	state       CircuitState
	window      time.Duration
	threshold   float64
	openFor     time.Duration
	halfProbes  int
	failures    int
	total       int
	openedAt    time.Time
	allowedProbes int
}

func NewCircuit(window time.Duration, threshold float64, openFor time.Duration, probes int) *Circuit {
	return &Circuit{
		state:      Closed,
		window:     window,
		threshold:  threshold,
		openFor:    openFor,
		halfProbes: probes,
	}
}

func (c *Circuit) Record(ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total++
	if !ok { c.failures++ }
	c.evaluateUnlocked()
}

func (c *Circuit) evaluateUnlocked() {
	errRate := 0.0
	if c.total > 0 {
		errRate = float64(c.failures) / float64(c.total)
	}
	switch c.state {
	case Closed:
		if errRate >= c.threshold {
			c.state = Open
			c.openedAt = time.Now()
		}
	case Open:
		if time.Since(c.openedAt) >= c.openFor {
			c.state = HalfOpen
			c.allowedProbes = c.halfProbes
			c.total, c.failures = 0, 0
		}
	case HalfOpen:
		// allow only probes; after that, decide
		if c.allowedProbes <= 0 {
			if errRate >= c.threshold { // fail → stay open
				c.state = Open
				c.openedAt = time.Now()
			} else {
				c.state = Closed
				c.total, c.failures = 0, 0
			}
		}
	}
}

func (c *Circuit) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.state {
	case Closed:
		return true
	case Open:
		return false
	case HalfOpen:
		if c.allowedProbes > 0 {
			c.allowedProbes--
			return true
		}
		return false
	default:
		return false
	}
}
