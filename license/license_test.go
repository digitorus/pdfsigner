package license

import (
	"testing"
	"time"

	"github.com/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

var licData = LicenseData{
	Name:  "Name",
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		{MaxCount: 30, IntervalStr: "1s", Interval: 1 * time.Second},
		{MaxCount: 10, IntervalStr: "10s", Interval: 10 * time.Second},
		{MaxCount: 100, IntervalStr: "1m", Interval: 1 * time.Minute},
		{MaxCount: 2000, IntervalStr: "1h", Interval: 1 * time.Hour},
		{MaxCount: 200000, IntervalStr: "24h", Interval: 24 * time.Hour},
		{MaxCount: 2000000, IntervalStr: "720h", Interval: 720 * time.Hour},
		{MaxCount: 20000000, IntervalStr: TotalLimitDuration, Interval: 864000 * time.Hour}, // Total
	},
	MaxDirectoryWatchers: 2,
}

func TestFlow(t *testing.T) {
	// Continue with the existing test...
	assert.Equal(t, len(licData.Limits), len(LD.Limits))
	assert.Empty(t, deep.Equal(licData.Limits, LD.Limits))
	assert.Equal(t, licData.Limits[0].MaxCount, LD.Limits[0].MaxCount)
	assert.Equal(t, licData.Limits[0].LimitState.CurCount, LD.Limits[0].LimitState.CurCount)
	assert.Equal(t, licData.MaxDirectoryWatchers, LD.MaxDirectoryWatchers)

	// test load
	err := Load()
	assert.NoError(t, err)

	allow, _ := LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 29, LD.Limits[0].CurCount)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 28, LD.Limits[0].CurCount)
	time.Sleep(1 * time.Second)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 29, LD.Limits[0].CurCount)

	allow, _ = LD.RL.Allow()
	assert.True(t, allow)
	assert.Equal(t, 28, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)

	for i := 0; i < LD.Limits[0].CurCount; i++ {
		_, _ = LD.RL.Allow()
	}
	allow, limit := LD.RL.Allow()
	assert.False(t, allow)
	assert.Positive(t, limit.Left())

	// test save
	err = LD.SaveLimitState()
	assert.NoError(t, err)

	LD = LicenseData{}
	err = Load()
	assert.NoError(t, err)
	assert.Equal(t, 13, LD.Limits[0].CurCount)
	assert.Equal(t, 0, LD.Limits[1].CurCount)
}
