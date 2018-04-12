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

package ratelimiter

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter1(t *testing.T) {
	rl := NewRateLimiter(&Limit{MaxCount: 1, Interval: 10 * time.Millisecond})
	var result bool
	result = rl.Allow()
	if !result {
		t.Error("Allow: false, want true")
	}
	result = rl.Allow()
	if result {
		t.Error("Allow: true, want false")
	}

	time.Sleep(11 * time.Millisecond)
	result = rl.Allow()
	if !result {
		t.Error("Allow: false, want true")
	}
	result = rl.Allow()
	if result {
		t.Error("Allow: true, want false")
	}
}

func TestLimiter2(t *testing.T) {
	rl := NewRateLimiter(&Limit{MaxCount: 2, Interval: 10 * time.Millisecond})
	var result bool
	for i := 0; i < 2; i++ {
		result = rl.Allow()
		if !result {
			t.Errorf("Allow(%d): false, want true", i)
		}
	}
	result = rl.Allow()
	if result {
		t.Error("Allow: true, want false")
	}

	time.Sleep(11 * time.Millisecond)
	for i := 0; i < 2; i++ {
		result = rl.Allow()
		if !result {
			t.Errorf("Allow(%d): false, want true", i)
		}
	}
	result = rl.Allow()
	if result {
		t.Error("Allow: true, want false")
	}
}

func TestLimiter3(t *testing.T) {
	rl := NewRateLimiter(
		&Limit{Unlimited: false, MaxCount: 2, Interval: time.Second},
		&Limit{Unlimited: false, MaxCount: 10, Interval: time.Minute},
		&Limit{Unlimited: false, MaxCount: 2000, Interval: time.Hour},
		&Limit{Unlimited: false, MaxCount: 200000, Interval: 24 * time.Hour},
		&Limit{Unlimited: false, MaxCount: 2000000, Interval: 720 * time.Hour},
		&Limit{Unlimited: false, MaxCount: 20000000, Interval: 9999999999},
	)

	for i := 0; i < 20; i++ {
		allowed := rl.Allow()
		if !allowed {
			left, _ := rl.Left()
			time.Sleep(left)
			log.Println("---")
			continue
		}
		log.Println("allowed")
	}
}

func TestState(t *testing.T) {
	rl := NewRateLimiter(&Limit{MaxCount: 2, Interval: 10 * time.Millisecond})
	state := rl.GetState()
	assert.Equal(t, 0, state[0].CurCount)
	rl.Allow()
	state = rl.GetState()
	assert.Equal(t, 1, state[0].CurCount)
	rl.Allow()
	state = rl.GetState()
	assert.Equal(t, 0, state[0].CurCount)
}
