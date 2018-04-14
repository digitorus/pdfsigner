package license

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const LicenseB64 = "LP+HAwEBB0xpY2Vuc2UB/4gAAQMBBERhdGEBCgABAVIB/4QAAQFTAf+EAAAACv+DBQEC/4YAAAD+AV//iAH/83sibiI6Ik5hbWUiLCJlIjoidGVzdEBleGFtcGxlLmNvbSIsImVuZCI6IjIwMTktMDQtMTRUMjI6Mzk6MTkuOTg0MjE4NTU1KzAyOjAwIiwibCI6W3sibSI6MiwiaSI6IjFzIn0seyJtIjoxMCwiaSI6IjEwcyJ9LHsibSI6MTAwLCJpIjoiMW0ifSx7Im0iOjIwMDAsImkiOiIxaCJ9LHsibSI6MjAwMDAwLCJpIjoiMjRoIn0seyJtIjoyMDAwMDAwLCJpIjoiNzIwaCJ9LHsibSI6MjAwMDAwMDAsImkiOiI4NjQwMDBoIn1dLCJkIjoyfQExAuyC8BfWBjtTuZ+tDcY/FwlW/rrPXXNKORz9zKxj0NWZWvLhQIGZE7DtPfw69DY60gExAivdt7v7G35eeZ2BAukus3qpxfXTAl0gaIuU5XAFXaWu+Wv7GAQ2BOw6dbqs+5lUdAA="

var licData = LicenseData{
	Name:  "Name",
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		&ratelimiter.Limit{MaxCount: 2, IntervalStr: "1s", Interval: 1 * time.Second},
		&ratelimiter.Limit{MaxCount: 10, IntervalStr: "10s", Interval: 10 * time.Second},
		&ratelimiter.Limit{MaxCount: 100, IntervalStr: "1m", Interval: 1 * time.Minute},
		&ratelimiter.Limit{MaxCount: 2000, IntervalStr: "1h", Interval: 1 * time.Hour},
		&ratelimiter.Limit{MaxCount: 200000, IntervalStr: "24h", Interval: 24 * time.Hour},
		&ratelimiter.Limit{MaxCount: 2000000, IntervalStr: "720h", Interval: 720 * time.Hour},
		&ratelimiter.Limit{MaxCount: 20000000, IntervalStr: TotalLimitDuration, Interval: 864000 * time.Hour}, //Total
	},
	MaxDirectoryWatchers: 2,
}

func TestFlow(t *testing.T) {
	// test initialize
	licenseBytes := []byte(LicenseB64)
	err := Initialize(licenseBytes)
	assert.NoError(t, err)
	assert.Equal(t, len(licData.Limits), len(LD.Limits))
	assert.Equal(t, 0, len(deep.Equal(licData.Limits, LD.Limits)))
	assert.Equal(t, licData.Limits[0].MaxCount, LD.Limits[0].MaxCount)
	assert.Equal(t, licData.Limits[0].LimitState.CurCount, LD.Limits[0].LimitState.CurCount)
	assert.Equal(t, licData.MaxDirectoryWatchers, LD.MaxDirectoryWatchers)

	// test load
	err = Load()
	assert.NoError(t, err)

	allow, _ := LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 1, LD.Limits[0].CurCount)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	time.Sleep(1 * time.Second)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 1, LD.Limits[0].CurCount)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)

	allow, limit := LD.RL.Allow()
	assert.False(t, allow)
	assert.True(t, limit.Left() > 0)

	// test save
	err = LD.SaveLimitState()
	assert.NoError(t, err)

	LD = LicenseData{}
	err = Load()
	assert.NoError(t, err)
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)
}
