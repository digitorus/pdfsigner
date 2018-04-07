package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"bitbucket.org/digitorus/pdfsigner/db"
	"bitbucket.org/digitorus/pdfsigner/ratelimiter"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
	errors2 "github.com/pkg/errors"
)

var LD LicenseData // loaded license data

type LicenseData struct {
	Email     string              `json:"email"`
	End       time.Time           `json:"end"`
	Limits    []ratelimiter.Limit `json:"limits"`
	CryptoKey [32]byte
	RL        *ratelimiter.RateLimiter
}

// A previously generated license b32 encoded. In real life you should read it from a file...
const LicenseB32 = "FT7YCAYBAEDUY2LDMVXHGZIB76BAAAIDAECEIYLUMEAQUAABAFJAD74EAAAQCUYB76CAAAAABL7YGBIBAL7YMAAAAD7AF7H7QIA74AUPPMRGK3LBNFWCEORCORSXG5CAMV4GC3LQNRSS4Y3PNURCYITFNZSCEORCGIYDCOJNGA2C2MBXKQYTQORRGY5DENZOGU2TSNRYGAZDGNBLGAZDUMBQEIWCE3DJNVUXI4ZCHJNXWITVNZWGS3LJORSWIIR2MZQWY43FFQRG2YLYL5RW65LOOQRDUMRMEJUW45DFOJ3GC3BCHIYTAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORRGAWCE2LOORSXE5TBNQRDUNRQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5FR5SE5LONRUW22LUMVSCEOTGMFWHGZJMEJWWC6C7MNXXK3TUEI5DEMBQGAWCE2LOORSXE5TBNQRDUMZWGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQFQRGS3TUMVZHMYLMEI5DQNRUGAYDAMBQGAYDAMBQGAWCE3DBON2F65DJNVSSEORCGAYDAMJNGAYS2MBRKQYDAORQGA5DAMC2EJ6SY6ZCOVXGY2LNNF2GKZBCHJTGC3DTMUWCE3LBPBPWG33VNZ2CEORSGAYDAMBQGAWCE2LOORSXE5TBNQRDUMRVHEZDAMBQGAYDAMBQGAYDAMBMEJWGC43UL52GS3LFEI5CEMBQGAYS2MBRFUYDCVBQGA5DAMB2GAYFUIT5LUWCEQ3SPFYHI32LMV4SEOS3GAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDALBQFQYCYMBMGAWDAXJMEJJEYIR2NZ2WY3D5AEYQFVEO4FTR5GJ6XT6YL4EU4OOPGP73D6AAH5DKOPIFYPXLA6DNQFHULFQME5SLIEP4KRZYR6KUD2PILIATCAXSNKKQPNJ6O2UUTS7IODUZ6DSXQWZU33UDHIK7LMZ45IMOOKAFQJXJ6MF74RVHNPCZVUFRYOXFZCAAA==="

// the public key b32 encoded from the private key using: lkgen pub my_private_key_file`.
// It should be hardcoded somewhere in your app.
const PublicKeyBase32 = "ARYV7JMXD2ESN57WTBTHMFFPQDN4OX7NQAZXX6WBUNUDTPBNHQW4MP6KNY5S7MK2OVM34QU3PBIMXOEFR5GTCRMBAO3NBYHMP7NXCRMY2FRD7DOOP7P5QUHESJ3KWZMXC3QCLZI6KAMJDPYAWYVP64TEUCDQ===="

func Initialize(licenseBytes []byte) error {
	// load license data
	ld, err := newExtractLicense(licenseBytes)
	if err != nil {
		return errors2.Wrap(err, "")
	}
	if len(ld.Limits) == 0 {
		return errors2.Wrap(errors.New("no limits provided for license"), "")
	}

	// save license to db
	err = db.SaveByKey("license", licenseBytes)
	if err != nil {
		return errors2.Wrap(err, "")
	}

	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)
	err = ld.SaveLimitState()
	if err != nil {
		return errors2.Wrap(err, "")
	}
	LD = ld

	return nil
}

func Load() error {
	license, err := db.LoadByKey("license")
	if err != nil {
		return errors2.Wrap(err, "")
	}

	// load license data
	ld, err := newExtractLicense(license)
	if err != nil {
		return errors2.Wrap(err, "")
	}

	err = ld.loadLimitState()
	if err != nil {
		return errors2.Wrap(err, "")
	}
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	LD = ld

	return nil
}

func newExtractLicense(licenseB32 []byte) (LicenseData, error) {
	ld := LicenseData{}
	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB32String(PublicKeyBase32)
	if err != nil {
		return ld, errors2.Wrap(err, "")
	}

	// Unmarshal the customer license.
	license, err := lk.LicenseFromB32String(string(licenseB32))
	if err != nil {
		return ld, errors2.Wrap(err, "")
	}

	// validate the license signature.
	if ok, err := license.Verify(publicKey); err != nil {
		return ld, errors2.Wrap(err, "")
	} else if !ok {
		err = errors.New("Invalid license signature")
		return ld, errors2.Wrap(err, "")
	}

	// unmarshal the document.
	if err := json.Unmarshal(license.Data, &ld); err != nil {
		return ld, errors2.Wrap(err, "")
	}

	// Now you just have to check that the end date is after time.Now() then you can continue!
	if ld.End.Before(time.Now()) {
		return ld, errors2.Wrap(errors.New(fmt.Sprintf("License expired on: %s", ld.End.Format("2006-01-02"))), "")
	} else {
	}

	// set byte versions of the license
	licenseBytes, err := license.ToBytes()
	if err != nil {
		return ld, errors2.Wrap(err, "")
	}
	// set byte versions of the public key
	publicKeyBytes := publicKey.ToBytes()
	licenseBytes = append(licenseBytes, publicKeyBytes...)
	hash := cryptopasta.Hash("hash for license", licenseBytes)
	copy(ld.CryptoKey[:], hash[:32])

	return ld, nil
}

func (ld LicenseData) SaveLimitState() error {
	limitStates := ld.RL.GetState()
	log.Println(limitStates)
	limitStatesBytes, err := json.Marshal(limitStates)
	limitsStatesCiphered, err := cryptopasta.Encrypt(limitStatesBytes, &ld.CryptoKey)
	if err != nil {
		return err
	}

	err = db.SaveByKey("limits", limitsStatesCiphered)
	if err != nil {
		return err
	}

	return nil
}

func (ld LicenseData) loadLimitState() error {
	limitStatesCiphered, err := db.LoadByKey("limits")
	if err != nil {
		return err
	}

	limitStatesBytes, err := cryptopasta.Decrypt(limitStatesCiphered, &ld.CryptoKey)
	if err != nil {
		return err
	}

	var limitStates []ratelimiter.LimitState
	err = json.Unmarshal(limitStatesBytes, &limitStates)
	if err != nil {
		return err
	}

	for i := 0; i < len(ld.Limits); i++ {
		ld.Limits[i].LimitState = limitStates[i]
	}

	return nil
}

func (ld LicenseData) Info() {
	log.Println(ld.RL.GetState())
	log.Println(ld.Limits)
}
