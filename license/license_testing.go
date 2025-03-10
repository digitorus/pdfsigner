package license

import (
	"testing"
	"time"

	"github.com/digitorus/pdfsigner/license/ratelimiter"
	log "github.com/sirupsen/logrus"
)

func init() {
	if !testing.Testing() {
		// we are not in a test at the moment
		return
	} else if publicKeyBase64 != "" {
		// the license has been configured already
		return
	}

	// Automatically create a test license for testing purposes
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}

	publicKeyBase64 = publicKey
	hmacKey = "PDFSIGNER-TESTING"

	// Generate a test license
	testLicense, err := GenerateLicense(privateKey, LicenseData{
		Name:  "Name",
		Email: "test@example.com",
		End:   time.Now().Add(30 * time.Minute),
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
	})
	if err != nil {
		log.Fatal(err)
	}

	err = Initialize([]byte(testLicense))
	if err != nil {
		log.Fatal(err)
	}
}
