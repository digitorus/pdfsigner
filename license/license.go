package license

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisbrodbeck/machineid"
	log "github.com/sirupsen/logrus"

	"github.com/digitorus/pdfsigner/db"
	"github.com/digitorus/pdfsigner/license/ratelimiter"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
	"github.com/pkg/errors"
)

// TestLicense used in unit tests
const TestLicense = "LP+HAwEBB0xpY2Vuc2UB/4gAAQMBBERhdGEBCgABAVIB/4QAAQFTAf+EAAAACv+DBQEC/4YAAAD+AV3/iAH/8XsibiI6Ik5hbWUiLCJlIjoidGVzdEBleGFtcGxlLmNvbSIsImVuZCI6IjIxMjItMDYtMTFUMTM6Mzc6MDUuODg0OTIxMyswMjowMCIsImwiOlt7Im0iOjIsImkiOiIxcyJ9LHsibSI6MTAsImkiOiIxMHMifSx7Im0iOjEwMCwiaSI6IjFtIn0seyJtIjoyMDAwLCJpIjoiMWgifSx7Im0iOjIwMDAwMCwiaSI6IjI0aCJ9LHsibSI6MjAwMDAwMCwiaSI6IjcyMGgifSx7Im0iOjIwMDAwMDAwLCJpIjoiODY0MDAwaCJ9XSwiZCI6Mn0BMQIOpEnubsOkG6SGq8IjqBAtv7uFwY0aZJDLd4+JMA3DZWxQyg5OAavJ8AFQ3nPyORMBMQKsLzLxRDHhFf2wQG5gyaBpuSkIV1okdw06pg3cAAD0pcjaDQNj/+E9VQGc5I3QNckA"

// HMACKeyForLimitsEncryption
const HMACKeyForLimitsEncryption = "HMACKeyForLimitsEncryption"

// ErrOverLimit contains error for over limit
var ErrOverLimit = errors.New("limit is over")

// TotalLimitDuration is a time duration to be used for total time range in a string representation
// 864000h is equal to 100 years
var TotalLimitDuration = "864000h"

// LD stores all the license related data
var LD LicenseData

// LicenseData represents all the license related data
type LicenseData struct {
	// Name represents the name license assigned to
	Name string `json:"n"`
	// Email represents email address of the license owner
	Email string `json:"e"`
	// End represents the date when license ends
	End time.Time `json:"end"`
	// Limits represents limits assigned to the license
	Limits []*ratelimiter.Limit `json:"l"`
	// MaxDirectoryWatchers represents maximum allowed directories to watch with watch and sign service
	MaxDirectoryWatchers int `json:"d"`

	// RL contains rate limiter
	RL *ratelimiter.RateLimiter `json:"-"`
	// cryptoKey stores the key to encrypt limits before storing it into db
	cryptoKey [32]byte
	// lastState is used to check save limits if state is changed
	lastState []ratelimiter.LimitState
}

// the public key b64 encoded from the private key using: lkgen pub my_private_key_file`.
const PublicKeyBase64 = "BAgf/si0bLTtS9jgxULXWcDbVz213jCfs3vc/P+ccXcJuS44czEkzFH0RRQ+RDPAsS5c3yJCiU7e871rfnTtavlwQ1JhCEBCAr9mkyWjvm4bTI9+UpaD4qw4zf0S2D9IWg=="
const appNameMachineID = "PDFSigner_unique_key_"

// Initialize extracts license from the bytes provided to LD variable and stores it inside the db
func Initialize(licenseBytes []byte) error {
	log.Info("Initializing license...")

	// load license data
	ld, err := newExtractLicense(licenseBytes)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// save license to db
	err = db.SaveByKey("license", licenseBytes)
	if err != nil {
		return errors.Wrap(err, "")
	}

	// initialize rate limiter
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	// save limit state to the db
	err = ld.SaveLimitState()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// save machine id
	err = saveMachineID()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// assign license data to LD variable
	LD = ld

	return nil
}

// Load loads the license from the db and extracts it to LD variable
func Load() error {
	log.Info("Loading license from the DB...")

	// load license from the db
	license, err := db.LoadByKey("license")
	if err != nil {
		return errors.Wrap(err, "couldn't load license from the db")
	}

	// check machine id
	err = checkMachineID()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// load license data
	ld, err := newExtractLicense(license)
	if err != nil {
		return errors.Wrap(err, "couldn't extract license")
	}

	// load limit state from the db
	err = ld.loadLimitState()
	if err != nil {
		return errors.Wrap(err, "couldn't load license limits")
	}

	// initialize rate limiter
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	// assign license data to LD variable
	LD = ld

	return nil
}

func newExtractLicense(licenseB64 []byte) (LicenseData, error) {
	log.Info("Extracting license...")

	ld := LicenseData{}
	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB64String(PublicKeyBase64)
	if err != nil {
		return ld, errors.Wrap(err, "")
	}

	// Unmarshal the customer license.
	license, err := lk.LicenseFromB64String(string(licenseB64))
	if err != nil {
		return ld, errors.Wrap(err, "")
	}

	// validate the license signature.
	if ok, err := license.Verify(publicKey); err != nil {
		return ld, errors.Wrap(err, "")
	} else if !ok {
		err = errors.New("Invalid license signature")
		return ld, errors.Wrap(err, "")
	}

	// unmarshal the document.
	if err := json.Unmarshal(license.Data, &ld); err != nil {
		return ld, errors.Wrap(err, "")
	}

	// Now you just have to check that the end date is after time.Now() then you can continue!
	if ld.End.Before(time.Now()) {
		return ld, errors.Wrap(errors.New(fmt.Sprintf("License expired on: %s", ld.End.Format("2006-01-02"))), "")
	}

	// check limits
	if len(ld.Limits) == 0 {
		return ld, errors.Wrap(errors.New("no limits provided for license"), "")
	}

	// parse time limits
	for _, l := range ld.Limits {
		i, err := time.ParseDuration(l.IntervalStr)
		if err != nil {
			return ld, errors.Wrap(errors.New("parse interval error"), "")
		}
		l.Interval = i
	}

	// create a key to encrypt the limits before storing into db
	// set byte versions of the license
	licenseBytes, err := license.ToBytes()
	if err != nil {
		return ld, errors.Wrap(err, "")
	}
	// set byte versions of the public key
	publicKeyBytes := publicKey.ToBytes()
	licenseBytes = append(licenseBytes, publicKeyBytes...)
	hash := cryptopasta.Hash(HMACKeyForLimitsEncryption, licenseBytes)
	copy(ld.cryptoKey[:], hash[:32])

	return ld, nil
}

// SaveLimitState saves the limit state to the db
func (ld *LicenseData) SaveLimitState() error {
	if !ld.isStateChanged() {
		return nil
	}

	// get limit states
	limitStates := ld.RL.GetState()
	// marshal limit states
	limitStatesBytes, err := json.Marshal(limitStates)
	if err != nil {
		return err
	}
	// encrypt limit states
	limitsStatesCiphered, err := cryptopasta.Encrypt(limitStatesBytes, &ld.cryptoKey)
	if err != nil {
		return err
	}

	// store limit states in the db
	err = db.SaveByKey("license_limits", limitsStatesCiphered)
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
	limitStatesCiphered, err := db.LoadByKey("license_limits")
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
		_ = ld.SaveLimitState()
	}(ld)
}

// Wait validates license end data and limits, if one of the limits are reached,
// waits the time till the work is allowed again
func (ld *LicenseData) Wait() error {
	// validate license data
	if time.Now().After(ld.End) {
		return errors.Wrap(errors.New(fmt.Sprintf("license is valid until: %v, please update the license", ld.End)), "")
	}

	// check if the work is allowed by license limiters, if not wait
	for {
		allow, limit := ld.RL.Allow()
		if allow {
			break
		} else {
			// check the total limit
			if isTotalLimit(limit) {
				return errors.Wrap(errors.New("total license limits exceeded, please update the license"), "")
			}

			// log sleep time information
			log.Println(ErrOverLimit, "wait for:", limit.Left())

			// sleep
			time.Sleep(limit.Left())
		}
	}
	return nil
}

// Info returns formatted information about the license
func (ld *LicenseData) Info() string {
	var res string

	// get basic license information
	res += fmt.Sprintf("Licensed to %s until %v\n", ld.Email, ld.End.Format("02 Jan 2006"))

	// get limits information
	for _, l := range ld.Limits {
		// get interval
		if isTotalLimit(l) {
			res += "Interval: Total "
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

func saveMachineID() error {
	// load machine id
	machineID, err := machineid.ProtectedID(appNameMachineID)
	if err != nil {
		log.Fatal(err)
	}

	// save machine id
	err = db.SaveByKey("license_machineid", []byte(machineID))
	if err != nil {
		return errors.Wrap(err, "couldn't save host info")
	}

	return nil
}

// laod and check machine id
func checkMachineID() error {
	// load machine id from the db
	savedMachineID, err := db.LoadByKey("license_machineid")
	if err != nil {
		return errors.Wrap(err, "couldn't load host info from the db")
	}
	savedMachineIDStr := string(savedMachineID[:])

	// get current machine id
	machineID, err := machineid.ProtectedID(appNameMachineID)
	if err != nil {
		return errors.Wrap(err, "couldn't get host info")
	}

	// check that ids are not nil
	if savedMachineIDStr == "" || machineID == "" {
		return errors.Wrap(errors.New("machine id check failed"), "host info is not working")
	}

	// compare saved and current machine ids
	if savedMachineIDStr != machineID {
		return errors.Wrap(errors.New("the license is bound to another computer"), "saved and current host info comparison failed")
	}

	return nil
}

// isTotalLimit checks if the provided limit is a total limit
func isTotalLimit(limit *ratelimiter.Limit) bool {
	return limit.IntervalStr == TotalLimitDuration
}
