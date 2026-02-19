// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


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
