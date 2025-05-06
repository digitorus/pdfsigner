package license

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/digitorus/pdfsigner/db"
	"github.com/digitorus/pdfsigner/license/ratelimiter"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Environment variable names
const (
	EnvPublicKey = "PDFSIGNER_LICENSE_PUBLIC_KEY"
	EnvLicense   = "PDFSIGNER_LICENSE"
	appID        = "PDFSigner_"
)

// These variables can be set at build time using -ldflags
// Example: go build -ldflags "-X github.com/digitorus/pdfsigner/license.publicKeyBase64=PUBLICKEY"
// Example: go build -ldflags "-X github.com/digitorus/pdfsigner/license.licenseBase64=LICENSE"
// Example: go build -ldflags "-X github.com/digitorus/pdfsigner/license.hmacKey=HMACKEY"
var (
	// publicKeyBase64 is the public key used to verify licenses
	publicKeyBase64 string

	// licenseBase64 is the license that can be hardcoded into the binary
	licenseBase64 string

	// hmacKey is used for encryption of license limits
	hmacKey string
)

// 864000h is equal to 100 years.
var TotalLimitDuration = "864000h"

// ErrOverLimit contains error for over limit.
var ErrOverLimit = errors.New("exceeded license")

// ErrMissingRequiredValue indicates a required value is missing
var ErrMissingRequiredValue = errors.New("missing required value")

// LD stores all the license related data.
var LD LicenseData

// LicenseData represents all the license related data.
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

// getLicenseBytesFromEnv attempts to read license bytes from environment variable
// License should always come from environment to allow for license changes without rebuilding
func getLicenseBytesFromEnv() ([]byte, bool) {
	if license := os.Getenv(EnvLicense); license != "" {
		return []byte(license), true
	}
	return nil, false
}

// Initialize extracts license from the bytes provided to LD variable and stores it inside the db.
// If licenseBytes is empty, tries to read from environment variable
func Initialize(licenseBytes []byte) error {
	log.Debug("Initializing license...")

	// Check if license bytes should be loaded from environment
	if len(licenseBytes) == 0 {
		if envBytes, exists := getLicenseBytesFromEnv(); exists {
			licenseBytes = envBytes
			log.Debug("Using license from environment variable")
		} else if licenseBase64 != "" {
			licenseBytes = []byte(licenseBase64)
			log.Debug("Using license provided at build")
		} else {
			return errors.Wrap(ErrMissingRequiredValue,
				fmt.Sprintf("license must be provided via %s environment variable", EnvLicense))
		}
	}

	if len(licenseBytes) == 0 {
		return errors.New("no license configured")
	}

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

// Load loads the license from the db and extracts it to LD variable.
func Load() error {
	log.Debug("Loading license from the DB...")

	// load license from the db
	license, err := db.LoadByKey("license")
	if err != nil {
		return fmt.Errorf("couldn't load license from the db: %w", err)
	}

	// check machine id
	err = checkMachineID()
	if err != nil {
		return errors.Wrap(err, "")
	}

	// load license data
	ld, err := newExtractLicense(license)
	if err != nil {
		return fmt.Errorf("couldn't extract license: %w", err)
	}

	// load limit state from the db
	err = ld.loadLimitState()
	if err != nil {
		return fmt.Errorf("couldn't load license limits: %w", err)
	}

	// initialize rate limiter
	ld.RL = ratelimiter.NewRateLimiter(ld.Limits...)

	// assign license data to LD variable
	LD = ld

	return nil
}

func newExtractLicense(licenseB64 []byte) (LicenseData, error) {
	log.Debug("Extracting license...")

	ld := LicenseData{}

	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB64String(publicKeyBase64)
	if err != nil {
		return ld, fmt.Errorf("invalid public key format: %w", err)
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
		return ld, errors.Wrap(errors.New("License expired on: "+ld.End.Format("2006-01-02")), "")
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
	hash := cryptopasta.Hash(hmacKey, licenseBytes)
	copy(ld.cryptoKey[:], hash[:32])

	return ld, nil
}

// SaveLimitState saves the limit state to the db.
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

// isStateChanged checks if the state is changed since last saving to the db.
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

// loadLimitState loads state from the db.
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
	for i := range len(ld.Limits) {
		ld.Limits[i].LimitState = limitStates[i]
	}

	return nil
}

// AutoSave saves state every second.
func (ld *LicenseData) AutoSave() {
	go func(ld *LicenseData) {
		time.Sleep(1 * time.Second)

		_ = ld.SaveLimitState()
	}(ld)
}

// waits the time till the work is allowed again.
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
			log.Printf("%s (%d signatures per %s), wait for %s,", ErrOverLimit, limit.MaxCount, limit.IntervalStr, limit.Left())

			// sleep
			time.Sleep(limit.Left())
		}
	}

	return nil
}

// Info returns formatted information about the license.
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
	machineID, err := machineid.ProtectedID(appID)
	if err != nil {
		log.Fatal(err)
	}

	// save machine id
	err = db.SaveByKey("license_machineid", []byte(machineID))
	if err != nil {
		return fmt.Errorf("couldn't save host info: %w", err)
	}

	return nil
}

// check machine id.
func checkMachineID() error {
	// load machine id from the db
	savedMachineID, err := db.LoadByKey("license_machineid")
	if err != nil {
		return fmt.Errorf("couldn't load host info from the db: %w", err)
	}

	savedMachineIDStr := string(savedMachineID)

	// get current machine id
	machineID, err := machineid.ProtectedID(appID)
	if err != nil {
		return fmt.Errorf("couldn't get host info: %w", err)
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

// isTotalLimit checks if the provided limit is a total limit.
func isTotalLimit(limit *ratelimiter.Limit) bool {
	return limit.IntervalStr == TotalLimitDuration
}

// GenerateKeyPair creates a new public/private key pair for license generation
func GenerateKeyPair() (privateKeyStr, publicKeyStr string, err error) {
	privateKey, err := lk.NewPrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	privateKeyStr, err = privateKey.ToB64String()
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKey := privateKey.GetPublicKey()
	publicKeyStr = publicKey.ToB64String()

	return privateKeyStr, publicKeyStr, nil
}

// GenerateLicense creates a license using the provided private key and license data
func GenerateLicense(privateKeyStr string, licenseData interface{}) (string, error) {
	// Parse private key
	privateKey, err := lk.PrivateKeyFromB64String(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Marshal the license data to JSON
	docBytes, err := json.Marshal(licenseData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal license data: %w", err)
	}

	// Generate license
	lic, err := lk.NewLicense(privateKey, docBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate license: %w", err)
	}

	// Encode license to base64
	licStr, err := lic.ToB64String()
	if err != nil {
		return "", fmt.Errorf("failed to encode license: %w", err)
	}

	return licStr, nil
}

// GenerateRandomHMACKey creates a random HMAC key
func GenerateRandomHMACKey() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes)[:32], nil
}

// GetPublicKeyFromPrivate extracts the public key from a private key
func GetPublicKeyFromPrivate(privateKeyStr string) (string, error) {
	privateKey, err := lk.PrivateKeyFromB64String(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	publicKey := privateKey.GetPublicKey()
	return publicKey.ToB64String(), nil
}
