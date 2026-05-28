package agent

import (
	"sync"
	"time"
)

// ghModelsLimiter enforces the GitHub Models free-tier limit of 1 req/60s.
// It uses elapsed time — if enough time has already passed, Wait() returns immediately.
var ghModelsLimiter = newRateLimiter(62 * time.Second)

type rateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func newRateLimiter(interval time.Duration) *rateLimiter {
	// Set lastCall far in the past so the first call is always immediate
	return &rateLimiter{
		interval: interval,
		lastCall: time.Now().Add(-interval),
	}
}

// Wait blocks only for the remaining gap since the last call, then returns.
// If enough time has already passed, it returns instantly.
func (r *rateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastCall)
	if elapsed < r.interval {
		time.Sleep(r.interval - elapsed)
	}
	r.lastCall = time.Now()
}