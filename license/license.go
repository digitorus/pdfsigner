package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/digitorus/pdfsigner/db"
	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
	errors2 "github.com/pkg/errors"
)

var ErrOverLimit = errors.New("limit is over")
var TotalLimitDuration = time.Duration(999999999)

var LD LicenseData // loaded license data

type LicenseData struct {
	Name                 string               `json:"n"`
	Email                string               `json:"e"`
	End                  time.Time            `json:"end"`
	Limits               []*ratelimiter.Limit `json:"l"`
	MaxDirectoryWatchers int                  `json:"d"`

	RL        *ratelimiter.RateLimiter `json:",omitempty"`
	cryptoKey [32]byte
	lastState []ratelimiter.LimitState
}

// the public key b64 encoded from the private key using: lkgen pub my_private_key_file`.
// It should be hardcoded somewhere in your app.
const PublicKeyBase64 = "BAgf/si0bLTtS9jgxULXWcDbVz213jCfs3vc/P+ccXcJuS44czEkzFH0RRQ+RDPAsS5c3yJCiU7e871rfnTtavlwQ1JhCEBCAr9mkyWjvm4bTI9+UpaD4qw4zf0S2D9IWg=="

func Initialize(licenseBytes []byte) error {
	// load license data
	ld, err := newExtractLicense(licenseBytes)
	if err != nil {
		return errors2.Wrap(err, "")
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

func newExtractLicense(licenseB64 []byte) (LicenseData, error) {
	ld := LicenseData{}
	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB64String(PublicKeyBase64)
	if err != nil {
		return ld, errors2.Wrap(err, "")
	}

	// Unmarshal the customer license.
	license, err := lk.LicenseFromB64String(string(licenseB64))
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
	}

	// check limits
	if len(ld.Limits) == 0 {
		return ld, errors2.Wrap(errors.New("no limits provided for license"), "")
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
	copy(ld.cryptoKey[:], hash[:32])

	return ld, nil
}

func (ld *LicenseData) isStateChanged() bool {
	if len(ld.lastState) == 0 {
		return true
	}

	for i, s := range ld.RL.GetState() {
		if s.CurCount != ld.lastState[i].CurCount || s.LastTime != ld.lastState[i].LastTime {
			return true
		}
	}

	return false
}

func (ld *LicenseData) SaveLimitState() error {
	if !ld.isStateChanged() {
		return nil
	}

	limitStates := ld.RL.GetState()
	limitStatesBytes, err := json.Marshal(limitStates)
	limitsStatesCiphered, err := cryptopasta.Encrypt(limitStatesBytes, &ld.cryptoKey)
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

	limitStatesBytes, err := cryptopasta.Decrypt(limitStatesCiphered, &ld.cryptoKey)
	if err != nil {
		return err
	}

	var limitStates []ratelimiter.LimitState
	err = json.Unmarshal(limitStatesBytes, &limitStates)
	if err != nil {
		return err
	}

	if len(limitStates) == 0 {
		return errors.New("no limits provided within license")
	}

	for i := 0; i < len(ld.Limits); i++ {
		ld.Limits[i].LimitState = limitStates[i]
	}

	return nil
}

func (ld *LicenseData) AutoSave() {
	go func(ld *LicenseData) {
		time.Sleep(1 * time.Second)
		ld.SaveLimitState()
	}(ld)
}

func (ld *LicenseData) Info() string {
	var res string
	res += fmt.Sprintf("Licensed to %s until %v\n\n", ld.Email, ld.End.Format("02 Jan 2006"))

	for _, l := range ld.Limits {
		if IsTotalLimit(l) {
			res += fmt.Sprintf("Interval: Total ")
		} else {
			res += fmt.Sprintf("Interval: %v, ", l.Interval)
		}

		if l.IsUnlimited() {
			res += "Unlimited"
		} else {
			res += fmt.Sprintf("Maximum: %v, ", l.MaxCount)
		}
		res += fmt.Sprintf("Signed: %v, Counted from: %v\n", l.CurCount, l.LastTime.Format("02 Jan 2006 15:04:05"))
	}
	return res
}

func IsTotalLimit(limit *ratelimiter.Limit) bool {
	return limit.Interval == TotalLimitDuration
}
