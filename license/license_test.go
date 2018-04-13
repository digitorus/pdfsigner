package license

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const LicenseB64 = "LP+HAwEBB0xpY2Vuc2UB/4gAAQMBBERhdGEBCgABAVIB/4QAAQFTAf+EAAAACv+DBQEC/4YAAAD+AXX/iAH+AQh7Im4iOiJOYW1lIiwiZSI6InRlc3RAZXhhbXBsZS5jb20iLCJlbmQiOiIyMDE5LTA0LTEzVDE1OjI3OjM1LjU0MDMwMDk0NCswMjowMCIsImwiOlt7Im0iOjIsImkiOjEwMDAwMDAwMDB9LHsibSI6MTAsImkiOjYwMDAwMDAwMDAwfSx7Im0iOjIwMDAsImkiOjM2MDAwMDAwMDAwMDB9LHsibSI6MjAwMDAwLCJpIjo4NjQwMDAwMDAwMDAwMH0seyJtIjoyMDAwMDAwLCJpIjoyNTkyMDAwMDAwMDAwMDAwfSx7Im0iOjIwMDAwMDAwLCJpIjo5OTk5OTk5OTl9XSwiZCI6Mn0BMQL/ICI2rsJuizo5QBeebH/f+feWXtrcVV8ljkThmJqPQtY08Hft/fNSSq5ZOgG1PugBMQL/fi0AEFBNC2iuJ0HJNrxwhRVfwrHR5jKCOkRA4DMv5U3D6Kd/KSdD3N/ntmTxSEkA"

var licData = LicenseData{
	Name:  "Name",
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		&ratelimiter.Limit{MaxCount: 2, Interval: time.Second},
		&ratelimiter.Limit{MaxCount: 10, Interval: time.Minute},
		&ratelimiter.Limit{MaxCount: 2000, Interval: time.Hour},
		&ratelimiter.Limit{MaxCount: 200000, Interval: 24 * time.Hour},
		&ratelimiter.Limit{MaxCount: 2000000, Interval: 720 * time.Hour},
		&ratelimiter.Limit{MaxCount: 20000000, Interval: TotalLimitDuration}, //Total
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
