package license

import (
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

const LicenseB32 = "FT7YCAYBAEDUY2LDMVXHGZIB76BAAAIDAECEIYLUMEAQUAABAFJAD74EAAAQCUYB76CAAAAABL7YGBIBAL7YMAAAAD7AGLP7QIA74AWAPMRGK3LBNFWCEORCORSXG5CAMV4GC3LQNRSS4Y3PNURCYITFNZSCEORCGIYDCOJNGA2C2MJSKQYTIORRGA5DKOBOHAYTSMBRGQYTQMRLGAZDUMBQEIWCE4TBORSV63DJNVUXI4ZCHJNXWITVNZWGS3LJORSWIIR2MZQWY43FFQRG2YLYL5RW65LOOQRDUMRMEJUW45DFOJ3GC3BCHIYTAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORRGAWCE2LOORSXE5TBNQRDUNRQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5FR5SE5LONRUW22LUMVSCEOTGMFWHGZJMEJWWC6C7MNXXK3TUEI5DEMBQGAWCE2LOORSXE5TBNQRDUMZWGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQFQRGS3TUMVZHMYLMEI5DQNRUGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQGAWCE2LOORSXE5TBNQRDUMRVHEZDAMBQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5FR5SE5LONRUW22LUMVSCEOTGMFWHGZJMEJWWC6C7MNXXK3TUEI5DEMBQGAYDAMBQFQRGS3TUMVZHMYLMEI5DGMJVGM3DAMBQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5LUWCE3LBPBPWI2LSMVRXI33SPFPXOYLUMNUGK4TTEI5DE7IBGEBCX52PIXPQPP5YKP3F3PFICFH7MEZ3OZZ2US2T2N6GQNSETIH7ZBJ23YAPJP542XIEWVETGIRA566BAEYQFZJZR6HX6C7ICZ3JAYJHVXBRPUVWHJF5QS5PW5FTFPZRQLDGPBFASPCPSQ2YEKJU7CK4B6WQRIEODYAA===="

var licensePeriod = time.Hour * 24 * 365 // 1 year
var licData = LicenseData{
	Email: "test@example.com",
	Limits: []*ratelimiter.Limit{
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2, Interval: time.Second},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 10, Interval: time.Minute},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000, Interval: time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 200000, Interval: 24 * time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 2000000, Interval: 720 * time.Hour},
		&ratelimiter.Limit{Unlimited: false, MaxCount: 20000000, Interval: licensePeriod}, //Total
	},
	MaxDirectoryWatchers: 2,
}

func TestFlow(t *testing.T) {
	// test initialize
	licenseBytes := []byte(LicenseB32)
	err := Initialize(licenseBytes)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(LD.Limits))
	assert.Equal(t, 0, len(deep.Equal(licData.Limits, LD.Limits)))
	assert.Equal(t, 2, LD.Limits[0].MaxCount)
	assert.Equal(t, 0, LD.Limits[0].LimitState.CurCount)
	assert.Equal(t, 2, LD.MaxDirectoryWatchers)
	assert.Equal(t, licensePeriod, LD.Limits[5].Interval)

	// test load
	err = Load()
	assert.NoError(t, err)

	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 1, LD.Limits[0].CurCount)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	time.Sleep(1 * time.Second)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 1, LD.Limits[0].CurCount)
	assert.True(t, LD.RL.Allow())
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)
	assert.False(t, LD.RL.Allow())
	assert.True(t, LD.RL.Left() > 0)

	// test save
	err = LD.SaveLimitState()
	assert.NoError(t, err)

	LD = LicenseData{}
	err = Load()
	assert.NoError(t, err)
	assert.Equal(t, 0, LD.Limits[0].CurCount)
	assert.Equal(t, 6, LD.Limits[1].CurCount)

}
