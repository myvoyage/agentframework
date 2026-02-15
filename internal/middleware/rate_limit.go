package middleware

import (
	"time"
)

// Simple token bucket rate limiter (per process).
type RateLimiter struct {
	tokens chan struct{}
	quit   chan struct{}
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	rl := &RateLimiter{tokens: make(chan struct{}, burst), quit: make(chan struct{})}
	// fill initial burst
	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}
	// refill at interval 1/rps seconds
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rps))
		defer ticker.Stop()
		for {
			select {
			case <-rl.quit:
				return
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			}
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.quit)
}
