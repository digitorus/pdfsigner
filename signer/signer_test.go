package signer

import (
	"log"
	"testing"
	"time"

	"bitbucket.org/digitorus/pdfsign/sign"
	"bitbucket.org/digitorus/pdfsigner/license"
)

func TestSigner(t *testing.T) {
	err := license.Load()
	if err != nil {
		log.Fatal(err)
	}
	// create signer
	signData := SignData{
		Signature: sign.SignDataSignature{
			Info: sign.SignDataSignatureInfo{
				Name:        "Tim",
				Location:    "Spain",
				Reason:      "Test",
				ContactInfo: "None",
				Date:        time.Now().Local(),
			},
			CertType: 2,
			Approval: false,
		},
	}
	signData.SetPEM("../testfiles/test.crt", "../testfiles//test.pem", "")

	for i := 0; i <= 20; i++ {
		err = SignFile("../testfiles/testfile20.pdf", "../testfiles/testfile20_signed.pdf", signData)
		if err != nil {
			t.Fatal(err)
		}
	}
}
