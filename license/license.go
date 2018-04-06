package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bitbucket.org/digitorus/pdfsigner/db"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
)

var LD LicenseData // loaded license data

type LicenseData struct {
	Email     string    `json:"email"`
	End       time.Time `json:"end"`
	Limits    []Limit   `json:"limits"`
	CryptoKey [32]byte
}

type Limit struct {
	Unlimited bool          `json:"unlimited"`
	MaxCount  int           `json:"max_count"`
	Interval  time.Duration `json:"interval"`
}

// A previously generated license b32 encoded. In real life you should read it from a file...
const licenseB32 = "FT7YOAYBAEDUY2LDMVXHGZIB76EAAAIDAECEIYLUMEAQUAABAFJAD74EAAAQCUYB76CAAAAABL7YGBIBAL7YMAAAAD73H74IAFEHWITFNVQWS3BCHIRHIZLTORAGK6DBNVYGYZJOMNXW2IRMEJSW4ZBCHIRDEMBRHAWTCMBNGI3FIMJSHIYTSORTGMXDOMBZG43TIMJYHAVTAMR2GAYCE7IBGEBAPXB37ROJCUOYBVG4LAL3MSNKJKPGIKNT564PYK5X542NH62V7TAUEYHGLEOPZHRBAPH7M4SC55OHAEYQEXMKGG3JPO6BSHTDF3T5H6T42VUD7YAJ3TY5AP5MDE5QW4ZYWMSAPEK24HZOUXQ3LJ5YY34XYPVXBUAA===="

// the public key b32 encoded from the private key using: lkgen pub my_private_key_file`.
// It should be hardcoded somewhere in your app.
const PublicKeyBase32 = "ARIVIK3FHZ72ERWX6FQ6Z3SIGHPSMCDBRCONFKQRWSDIUMEEESQULEKQ7J7MZVFZMJDFO6B46237GOZETQ4M2NE32C3UUNOV5EUVE3OIV72F5LQRZ6DFMM6UJPELARG7RLJWKQRATUWD5YT46Q2TKQMPPGIA===="

func LoadLicense() ([]byte, error) {
	license, err := db.LoadByKey("license")
	if err != nil {
		return []byte{}, err
	}

	return license, nil
}

func ExtractLicense(licenseB32 []byte) error {
	// Unmarshal the public key.
	publicKey, err := lk.PublicKeyFromB32String(PublicKeyBase32)
	if err != nil {
		return err
	}

	// Unmarshal the customer license.
	license, err := lk.LicenseFromBytes(licenseB32)
	if err != nil {
		return err
	}

	// validate the license signature.
	if ok, err := license.Verify(publicKey); err != nil {
		return err
	} else if !ok {
		err = errors.New("Invalid license signature")
		return err
	}

	// unmarshal the document.
	if err := json.Unmarshal(license.Data, &LD); err != nil {
		return err
	}

	// Now you just have to check that the end date is after time.Now() then you can continue!
	if LD.End.Before(time.Now()) {
		return errors.New(fmt.Sprintf("License expired on: %s", LD.End.Format("2006-01-02")))
	} else {
		fmt.Printf(`Licensed to %s until %s`, LD.Email, LD.End.Format("2006-01-02"))
	}

	// set byte versions of the license
	licenseBytes, err := license.ToBytes()
	if err != nil {
		return nil
	}
	// set byte versions of the public key
	publicKeyBytes := publicKey.ToBytes()
	licenseBytes = append(licenseBytes, publicKeyBytes...)
	hash, err := cryptopasta.HashPassword(licenseBytes)
	if err != nil {
		return nil
	}
	copy(LD.CryptoKey[:], hash[:32])

	return nil
}
