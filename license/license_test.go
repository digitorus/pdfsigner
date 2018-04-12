package license

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const LicenseB32 = "LP+HAwEBB0xpY2Vuc2UB/4gAAQMBBERhdGEBCgABAVIB/4QAAQFTAf+EAAAACv+DBQEC/4YAAAD+AyX/iAH+Arh7ImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImVuZCI6IjIwMTktMDQtMTJUMTU6MzY6NDAuNzE3MzUxMzU3KzAyOjAwIiwicmF0ZV9saW1pdHMiOlt7InVubGltaXRlZCI6ZmFsc2UsIm1heF9jb3VudCI6MiwiaW50ZXJ2YWwiOjEwMDAwMDAwMDAsImxhc3RfdGltZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIn0seyJ1bmxpbWl0ZWQiOmZhbHNlLCJtYXhfY291bnQiOjEwLCJpbnRlcnZhbCI6NjAwMDAwMDAwMDAsImxhc3RfdGltZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIn0seyJ1bmxpbWl0ZWQiOmZhbHNlLCJtYXhfY291bnQiOjIwMDAsImludGVydmFsIjozNjAwMDAwMDAwMDAwLCJsYXN0X3RpbWUiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiJ9LHsidW5saW1pdGVkIjpmYWxzZSwibWF4X2NvdW50IjoyMDAwMDAsImludGVydmFsIjo4NjQwMDAwMDAwMDAwMCwibGFzdF90aW1lIjoiMDAwMS0wMS0wMVQwMDowMDowMFoifSx7InVubGltaXRlZCI6ZmFsc2UsIm1heF9jb3VudCI6MjAwMDAwMCwiaW50ZXJ2YWwiOjI1OTIwMDAwMDAwMDAwMDAsImxhc3RfdGltZSI6IjAwMDEtMDEtMDFUMDA6MDA6MDBaIn0seyJ1bmxpbWl0ZWQiOmZhbHNlLCJtYXhfY291bnQiOjIwMDAwMDAwLCJpbnRlcnZhbCI6OTk5OTk5OTk5LCJsYXN0X3RpbWUiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiJ9XSwibWF4X2RpcmVjdG9yeV93YXRjaGVycyI6Mn0BMQK/5I1B5IfsJcQ2LIwzicd0ARy2lcx1Sr3jgpv6Oue6sEszeXlAsHm8JX3bjm1wi8QBMQJTHxeHPq8aqPS2XkU7b+IiTo6y1G9LRGS9N4MHT7WUzeMZL19vQSKZanCpT/ogGx4A"

var licData = LicenseData{
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2, Interval: time.Second},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 10, Interval: time.Minute},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000, Interval: time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 200000, Interval: 24 * time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000000, Interval: 720 * time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 20000000, Interval: TotalLimitDuration}, //Total
	},
	MaxDirectoryWatchers: 2,
}

func TestFlow(t *testing.T) {
	// test initialize
	licenseBytes := []byte(LicenseB32)
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
