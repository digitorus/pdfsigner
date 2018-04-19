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

// ErrOverLimit contains error for over limit
var ErrOverLimit = errors.New("limit is over")

// TotalLimitDuration is a time duration to be used for total time range in a string representation
// 864000h is equal to 100 years
var TotalLimitDuration = "864000h"

// LD stores all the license related data
var LD LicenseData

// LicenseData represents all the license related data
type LicenseData struct {
	Name                 string               `json:"n"`
	Email                string               `json:"e"`
	End                  time.Time            `json:"end"`
	Limits               []*ratelimiter.Limit `json:"l"`
	MaxDirectoryWatchers int                  `json:"d"`

	RL        *ratelimiter.RateLimiter `json:"-"`
	cryptoKey [32]byte
	lastState []ratelimiter.LimitState
}

// the public key b64 encoded from the private key using: lkgen pub my_private_key_file`.
const PublicKeyBase64 = "BAgf/si0bLTtS9jgxULXWcDbVz213jCfs3vc/P+ccXcJuS44czEkzFH0RRQ+RDPAsS5c3yJCiU7e871rfnTtavlwQ1JhCEBCAr9mkyWjvm4bTI9+UpaD4qw4zf0S2D9IWg=="

// Initialize extracts license from the bytes provided to LD variable and stores it inside the db
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

	// initialize rate limiter
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	// save limit state to the db
	err = ld.SaveLimitState()
	if err != nil {
		return errors2.Wrap(err, "")
	}

	// assign license data to LD variable
	LD = ld

	return nil
}

// Load loads the license from the db and extracts it to LD variable
func Load() error {
	// load license from the db
	license, err := db.LoadByKey("license")
	if err != nil {
		return errors2.Wrap(err, "")
	}

	// load license data
	ld, err := newExtractLicense(license)
	if err != nil {
		return errors2.Wrap(err, "")
	}

	// load limit state from the db
	err = ld.loadLimitState()
	if err != nil {
		return errors2.Wrap(err, "")
	}

	// initialize rate limiter
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	// assign license data to LD variable
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

	// parse time limits
	for _, l := range ld.Limits {
		i, err := time.ParseDuration(l.IntervalStr)
		if err != nil {
			return ld, errors2.Wrap(errors.New("parse interval error"), "")
		}
		l.Interval = i
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

// SaveLimitState saves the limit state to the db
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

// isStateChanged checks if the state is changed since last saving to the db
func (ld *LicenseData) isStateChanged() bool {
	// TODO: check if that correct way to do it
	if len(ld.lastState) == 0 {
		return true
	}

	// find changed state
	for i, s := range ld.RL.GetState() {
		if s.CurCount != ld.lastState[i].CurCount || s.LastTime != ld.lastState[i].LastTime {
			return true
		}
	}

	return false
}

// loadLimitState loads state from the db
func (ld LicenseData) loadLimitState() error {
	// load limit state from the db
	limitStatesCiphered, err := db.LoadByKey("limits")
	if err != nil {
		return err
	}

	// decrypt state
	limitStatesBytes, err := cryptopasta.Decrypt(limitStatesCiphered, &ld.cryptoKey)
	if err != nil {
		return err
	}

	// unmarshal state
	var limitStates []ratelimiter.LimitState
	err = json.Unmarshal(limitStatesBytes, &limitStates)
	if err != nil {
		return err
	}

	// check if limits provided
	if len(limitStates) == 0 {
		return errors.New("no limits provided within license")
	}

	// assign state to the limits
	for i := 0; i < len(ld.Limits); i++ {
		ld.Limits[i].LimitState = limitStates[i]
	}

	return nil
}

// AutoSave saves state every second
func (ld *LicenseData) AutoSave() {
	go func(ld *LicenseData) {
		time.Sleep(1 * time.Second)
		ld.SaveLimitState()
	}(ld)
}

// Info returns formatted information about the license
func (ld *LicenseData) Info() string {
	var res string

	// get basic license information
	res += fmt.Sprintf("Licensed to %s until %v\n", ld.Email, ld.End.Format("02 Jan 2006"))

	// get limits information
	for _, l := range ld.Limits {
		// get interval
		if IsTotalLimit(l) {
			res += fmt.Sprintf("Interval: Total ")
		} else {
			res += fmt.Sprintf("Interval: %v, ", l.Interval)
		}

		// get maximum
		if l.IsUnlimited() {
			res += "Unlimited"
		} else {
			res += fmt.Sprintf("Maximum: %v", l.MaxCount)
		}

		// get current count
		if l.CurCount != 0 && !l.LastTime.IsZero() {
			res += fmt.Sprintf(", Signed: %v, Counted from: %v\n", l.CurCount, l.LastTime.Format("02 Jan 2006 15:04:05"))
		} else {
			res += "\n"
		}

	}
	return res
}

// IsTotalLimit checks if the provided limit is a total limit
func IsTotalLimit(limit *ratelimiter.Limit) bool {
	return limit.IntervalStr == TotalLimitDuration
}
