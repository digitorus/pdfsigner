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

type Limit struct {
	Unlimited bool          `json:"unlimited"`
	MaxCount  int           `json:"max_count"`
	Interval  time.Duration `json:"interval"`
	CurCount  int           `json:"cur_count"`
	LastTime  time.Time     `json:"last_time"`
}

func (l *Limit) allow() bool {
	if l.Unlimited {
		return true
	}

	if time.Now().Sub(l.LastTime) < l.Interval {
		if l.CurCount > 0 {
			l.CurCount--
			return true
		}
		return false
	}

	l.CurCount = l.MaxCount - 1
	l.LastTime = time.Now()
	return true
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
func NewRateLimiter(limits ...Limit) *RateLimiter {
	rl := RateLimiter{}

	for _, l := range limits {
		l.CurCount = l.MaxCount
		rl.limits = append(rl.limits, &l)
	}

	return &rl
}

// Allow returns true if a request is within the rate limit norms.
// Otherwise, it returns false.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, l := range rl.limits {
		if !l.allow() {
			return false
		}
	}

	return true
}

func (rl *RateLimiter) GetLimits() []*Limit {
	return rl.limits
}
