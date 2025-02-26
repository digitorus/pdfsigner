/*
Copyright 2017 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package ratelimiter implements rate limiting functionality.
package ratelimiter

import (
	"sync"
	"time"
)

// Limit represents the limit with current state.
type Limit struct {
	// MaxCount represents maximum number of allowed works
	MaxCount int `json:"m"`
	// IntervalStr represents time interval in string representation, for maximum allowed works to run
	IntervalStr string `json:"i"`
	// Interval represents time interval in for maximum allowed works to run
	Interval time.Duration `json:"-"`
	// LimitState represents current state
	LimitState `json:"-"`
}

// LimitState represents the state of the limit.
type LimitState struct {
	// CurCount represents counter of allowed works
	CurCount int `json:"c"`
	// LastTime represents time since the counter started to count
	LastTime time.Time `json:"l"`
}

// allow checks if the work is allowed.
func (l *Limit) allow() bool {
	// check if the limit is unlimited
	if l.IsUnlimited() {
		return true
	}

	// check if the work is allowed
	if time.Since(l.LastTime) < l.Interval {
		if l.CurCount > 0 {
			l.CurCount--

			return true
		}

		return false
	}

	// initialize if run first time or if the interval is past
	l.CurCount = l.MaxCount - 1
	l.LastTime = time.Now()

	return true
}

// IsUnlimited checks if the interval is unlimited.
func (l *Limit) IsUnlimited() bool {
	return l.MaxCount == -1
}

// Left returns how much time needed to wait until the limiter would allow to run work again.
func (l *Limit) Left() time.Duration {
	return l.Interval - time.Since(l.LastTime)
}

// RateLimiter was inspired by https://github.com/golang/go/wiki/RateLimiting.
// However, the go example is not good for setting high qps limits because
// it will cause the ticker to fire too often. Also, the ticker will continue
// to fire when the system is idle. This new Ratelimiter achieves the same thing,
// but by using just counters with no tickers or channels.
type RateLimiter struct {
	limits []*Limit
	mu     sync.Mutex
}

// NewRateLimiter creates a new RateLimiter. MaxCount is the max burst allowed
// while interval specifies the duration for a burst. The effective rate limit is
// equal to MaxCount/interval. For example, if you want to a max QPS of 5000,
// and want to limit bursts to no more than 500, you'd specify a MaxCount of 500
// and an interval of 100*time.Millilsecond.
func NewRateLimiter(limits ...*Limit) *RateLimiter {
	rl := RateLimiter{}
	// for _, l := range limits {
	//	if l.LastTime.IsZero() {
	//		l.CurCount = l.MaxCount
	//	}
	//}
	rl.limits = limits

	return &rl
}

// Allow returns true if a request is within the rate limit norms.
// Otherwise, it returns false.
func (rl *RateLimiter) Allow() (bool, *Limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, l := range rl.limits {
		if !l.allow() {
			return false, l
		}
	}

	return true, nil
}

// GetState returns current state of the limits.
func (rl *RateLimiter) GetState() []LimitState {
	var limitStates []LimitState

	for _, l := range rl.limits {
		limitStates = append(limitStates, l.LimitState)
	}

	return limitStates
}
