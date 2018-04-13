package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/license/ratelimiter"
	"github.com/hyperboloide/lk"
)

func main() {

	const privateKeyB64 = "KP+BAwEBC3BrQ29udGFpbmVyAf+CAAECAQNQdWIBCgABAUQB/4QAAAAK/4MFAQL/hgAAAP+Z/4IBYQQIH/7ItGy07UvY4MVC11nA21c9td4wn7N73Pz/nHF3CbkuOHMxJMxR9EUUPkQzwLEuXN8iQolO3vO9a3507Wr5cENSYQhAQgK/ZpMlo75uG0yPflKWg+KsOM39Etg/SFoBMQIkcq2v8M/xQF03dTg0aVXHB532/4gQ454IG4fcUOBohrYAA3t1o26+X1Ceh7rmavgA"

	// create a new Private key:
	//privateKey, err := lk.NewPrivateKey()
	privateKey, err := lk.PrivateKeyFromB64String(privateKeyB64)
	if err != nil {
		log.Fatal(err)
	}

	privateKeyStr, err := privateKey.ToB64String()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Private key", privateKeyStr)

	// create a license document:

	doc := license.LicenseData{
		Name:  "Name",
		Email: "test@example.com",
		End:   time.Now().Add(time.Hour * 24 * 365), // 1 year
		Limits: []*ratelimiter.Limit{
			&ratelimiter.Limit{MaxCount: 2, Interval: time.Second},
			&ratelimiter.Limit{MaxCount: 10, Interval: time.Minute},
			&ratelimiter.Limit{MaxCount: 2000, Interval: time.Hour},
			&ratelimiter.Limit{MaxCount: 200000, Interval: 24 * time.Hour},
			&ratelimiter.Limit{MaxCount: 2000000, Interval: 720 * time.Hour},
			&ratelimiter.Limit{MaxCount: 20000000, Interval: license.TotalLimitDuration}, //Total
		},
		MaxDirectoryWatchers: 2,
	}

	// marshall the document to json bytes:
	docBytes, err := json.Marshal(doc)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(docBytes))

	// generate your license with the private key and the document:
	lic, err := lk.NewLicense(privateKey, docBytes)
	if err != nil {
		log.Fatal(err)

	}

	// encode the new license to b32, this is what you give to your customer.
	str32, err := lic.ToB64String()
	if err != nil {
		log.Fatal(err)

	}
	log.Println("License Data:", str32)

	// get the public key. The public key should be hardcoded in your app to check licences.
	// Do not distribute the private key!
	publicKey := privateKey.GetPublicKey()
	log.Println("Public key:", publicKey.ToB64String())

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
