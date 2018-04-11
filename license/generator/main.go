package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/gtank/cryptopasta"
	"github.com/hyperboloide/lk"
)

func main() {

	fmt.Printf("%s", cryptopasta.NewEncryptionKey())

	// create a new Private key:
	privateKey, err := lk.NewPrivateKey()
	if err != nil {
		log.Fatal(err)

	}

	// create a license document:
	doc := license.LicenseData{
		Email: "test@example.com",
		End:   time.Now().Add(time.Hour * 24 * 365), // 1 year
		Limits: []*ratelimiter.Limit{
			&ratelimiter.Limit{Unlimited: false, MaxCount: 2, Interval: time.Second},
			&ratelimiter.Limit{Unlimited: false, MaxCount: 10, Interval: time.Minute},
			&ratelimiter.Limit{Unlimited: false, MaxCount: 2000, Interval: time.Hour},
			&ratelimiter.Limit{Unlimited: false, MaxCount: 200000, Interval: 24 * time.Hour},
			&ratelimiter.Limit{Unlimited: false, MaxCount: 2000000, Interval: 720 * time.Hour},
		},
	}

	// marshall the document to json bytes:
	docBytes, err := json.Marshal(doc)
	if err != nil {
		log.Fatal(err)
	}

	// generate your license with the private key and the document:
	lic, err := lk.NewLicense(privateKey, docBytes)
	if err != nil {
		log.Fatal(err)

	}

	// encode the new license to b32, this is what you give to your customer.
	str32, err := lic.ToB32String()
	if err != nil {
		log.Fatal(err)

	}
	log.Println("LicenseData:", str32)

	// get the public key. The public key should be hardcoded in your app to check licences.
	// Do not distribute the private key!
	publicKey := privateKey.GetPublicKey()
	log.Println("Public key:", publicKey.ToB32String())

	// validate the license:
	if ok, err := lic.Verify(publicKey); err != nil {
		log.Fatal(err)
	} else if !ok {
		log.Fatal("Invalid license signature")
	}

	// unmarshal the document and check the end date:
	res := license.LicenseData{}
	if err := json.Unmarshal(lic.Data, &res); err != nil {
		log.Fatal(err)
	} else if res.End.Before(time.Now()) {
		log.Fatal("LicenseData expired on: %s", res.End.String())
	} else {
		fmt.Printf(`Licensed to %s until %s \n, with limits: %v`, res.Email, res.End.Format("2006-01-02"), res.Limits)
	}

}
